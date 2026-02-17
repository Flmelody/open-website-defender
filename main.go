package main

import (
	"context"
	"embed"
	"flag"
	"net/http"
	_http "open-website-defender/internal/adapter/controller/http"
	"open-website-defender/internal/adapter/controller/http/middleware"
	"open-website-defender/internal/infrastructure/config"
	"open-website-defender/internal/infrastructure/database"
	"open-website-defender/internal/infrastructure/logging"
	"open-website-defender/internal/pkg"
	"os"
	"os/signal"
	"path/filepath"
	"strings"
	"syscall"
	"time"

	"github.com/gin-contrib/static"
	"github.com/gin-gonic/gin"
	"github.com/spf13/viper"
)

var (
	BackendHost = "http://localhost:9999/wall"
	RootPath    = "/wall"
	AdminPath   = "/admin"
	GuardPath   = "/guard"
)

func loadConfig() error {
	viper.SetConfigFile(".env")
	if err := viper.ReadInConfig(); err != nil {
		// .env file is optional
	}

	configPath := flag.String("config", "./config/config.yaml", "config file path")
	flag.Parse()

	configDir := filepath.Dir(*configPath)
	configFile := filepath.Base(*configPath)
	configExt := filepath.Ext(configFile)
	configName := configFile[:len(configFile)-len(configExt)]

	vp := viper.New()
	vp.SetConfigName(configName)
	vp.SetConfigType(configExt[1:])
	vp.AddConfigPath(configDir)

	if err := vp.ReadInConfig(); err != nil {
		return err
	}

	return viper.MergeConfigMap(vp.AllSettings())
}

func getAppConfig() *config.AppConfig {
	backendHost := viper.GetString("BACKEND_HOST")
	if backendHost == "" {
		backendHost = BackendHost
	}

	rootPath := viper.GetString("ROOT_PATH")
	if rootPath == "" {
		rootPath = RootPath
	}

	adminPath := viper.GetString("ADMIN_PATH")
	if adminPath == "" {
		adminPath = AdminPath
	}

	guardPath := viper.GetString("GUARD_PATH")
	if guardPath == "" {
		guardPath = GuardPath
	}

	appConfig := &config.AppConfig{
		BaseURL:   backendHost,
		RootPath:  rootPath,
		AdminPath: adminPath,
		GuardPath: guardPath,
	}

	validatePath := func(path, name string) {
		if !strings.HasPrefix(path, "/") || strings.HasSuffix(path, "/") {
			logging.Sugar.Fatalf("Incorrect %s: %s, "+
				"path should start with slash and not end with slash", name, path)
		}
	}

	validatePath(appConfig.RootPath, "root-path")
	validatePath(appConfig.AdminPath, "admin-path")
	validatePath(appConfig.GuardPath, "guard-path")

	return appConfig
}

//go:embed ui/admin/dist ui/guard/dist
var server embed.FS

func main() {
	var err error

	err = logging.InitLoggerWithEnv("dev")
	if err != nil {
		panic("Failed to initialize logger: " + err.Error())
	}

	err = loadConfig()
	if err != nil {
		logging.Sugar.Fatalf("Error reading config file: %s", err)
		return
	}

	// Initialize JWT secret from configuration
	pkg.InitJWTSecret(
		viper.GetString("security.jwt-secret"),
		viper.GetInt("security.token-expiration-hours"),
	)

	// Initialize RSA key for OIDC token signing
	if viper.GetBool("oauth.enabled") {
		pkg.InitRSAKey(viper.GetString("oauth.rsa-private-key-path"))
	}

	err = database.InitDB()
	if err != nil {
		logging.Sugar.Fatalf("Error initializing database: %s", err)
		return
	}

	appConfig := getAppConfig()

	r := gin.Default()
	r.RedirectTrailingSlash = true
	trustedProxies := viper.GetStringSlice("trustedProxies")
	err = r.SetTrustedProxies(trustedProxies)
	logging.Sugar.Info("TrustedProxies:", trustedProxies)
	if err != nil {
		logging.Sugar.Warn("Failed to set trusted proxies:", err)
		return
	}

	adminFS, err := static.EmbedFolder(server, "ui/admin/dist")
	if err != nil {
		logging.Sugar.Fatalf("Failed to embed admin folder")
		return
	}
	r.Use(static.Serve(appConfig.RootPath+appConfig.AdminPath, adminFS))
	guardFS, err := static.EmbedFolder(server, "ui/guard/dist")
	if err != nil {
		logging.Sugar.Fatalf("Failed to embed guard folder")
		return
	}
	r.Use(static.Serve(appConfig.RootPath+appConfig.GuardPath, guardFS))
	r.NoRoute(func(c *gin.Context) {
		if strings.HasPrefix(c.Request.URL.Path, appConfig.RootPath+appConfig.AdminPath) {
			c.FileFromFS("ui/admin/dist/index.html", http.FS(server))
			return
		}
		if strings.HasPrefix(c.Request.URL.Path, appConfig.RootPath+appConfig.GuardPath) {
			c.FileFromFS("ui/guard/dist/index.html", http.FS(server))
			return
		}
		c.Next()
	})

	// Security headers
	r.Use(middleware.SecurityHeaders())

	// CORS middleware (replaces inline handler)
	r.Use(middleware.CORS())

	// Request body size limit
	maxBodySizeMB := viper.GetInt64("server.max-body-size-mb")
	if maxBodySizeMB <= 0 {
		maxBodySizeMB = 10 // default 10MB
	}
	r.Use(middleware.BodyLimit(maxBodySizeMB * 1024 * 1024))

	// Access logging (must be before WAF/rate limiter to capture all actions)
	r.Use(middleware.AccessLog())

	// Geo-IP blocking
	geoDBPath := viper.GetString("geo-blocking.database-path")
	if viper.GetBool("geo-blocking.enabled") && geoDBPath != "" {
		if err := pkg.InitGeoIP(geoDBPath); err == nil {
			r.Use(middleware.GeoBlock())
			logging.Sugar.Info("Geo-IP blocking enabled")
		}
	}

	// Request filtering (SQLi, XSS, Path Traversal detection)
	if viper.GetBool("request-filtering.enabled") {
		r.Use(middleware.WAF())
		logging.Sugar.Info("Request filtering enabled")
	}

	// Request logging
	r.Use(middleware.Logger())

	// Global rate limiter
	if viper.GetBool("rate-limit.enabled") {
		globalRPM := viper.GetInt("rate-limit.requests-per-minute")
		if globalRPM <= 0 {
			globalRPM = 100
		}
		r.Use(middleware.RateLimiter("global", globalRPM))
		logging.Sugar.Infof("Global rate limiter enabled: %d requests/minute per IP", globalRPM)
	}

	// Register routes AFTER all middleware so they are included in the handler chain
	_http.Setup(r, appConfig)

	// Server configuration with timeouts
	port := os.Getenv("PORT")
	if port == "" {
		port = "9999"
	}

	srv := &http.Server{
		Addr:         ":" + port,
		Handler:      r,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 30 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	// Start server in goroutine
	go func() {
		logging.Sugar.Infof("Starting server on :%s", port)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logging.Sugar.Fatalf("Failed to start server: %s", err)
		}
	}()

	// Graceful shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	logging.Sugar.Info("Shutting down server...")

	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()
	if err := srv.Shutdown(ctx); err != nil {
		logging.Sugar.Fatalf("Server forced to shutdown: %s", err)
	}

	pkg.CloseGeoIP()
	logging.Sugar.Info("Server exited gracefully")
}
