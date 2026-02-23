package main

import (
	"context"
	"embed"
	"encoding/json"
	"flag"
	"fmt"
	"net/http"
	_http "open-website-defender/internal/adapter/controller/http"
	"open-website-defender/internal/adapter/controller/http/middleware"
	"open-website-defender/internal/infrastructure/cache"
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
	RootPath  = "/wall"
	AdminPath = "/admin"
	GuardPath = "/guard"
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
	// Paths: env/.env only — must match the frontend Vite build.
	// Changing these requires rebuilding the frontend.
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

	// Runtime-changeable: env/.env → config.yaml wall.* — no rebuild needed.
	backendHost := viper.GetString("BACKEND_HOST")
	if backendHost == "" {
		backendHost = viper.GetString("wall.backend-host")
	}
	if backendHost == "" {
		backendHost = rootPath // same-origin: relative path is sufficient
	}
	guardDomain := viper.GetString("GUARD_DOMAIN")
	if guardDomain == "" {
		guardDomain = viper.GetString("wall.guard-domain")
	}

	appConfig := &config.AppConfig{
		BaseURL:     backendHost,
		RootPath:    rootPath,
		AdminPath:   adminPath,
		GuardPath:   guardPath,
		GuardDomain: guardDomain,
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

	// Initialize cache (must be before DB and services)
	cache.InitStore(viper.GetInt("cache.size-mb"))

	err = database.InitDB()
	if err != nil {
		logging.Sugar.Fatalf("Error initializing database: %s", err)
		return
	}

	// Initialize event-driven cache invalidation
	cache.Init()

	// Start cross-instance cache sync (polls DB for version changes)
	viper.SetDefault("cache.sync-interval", 0)
	cache.InitSync(database.DB, viper.GetInt("cache.sync-interval"))

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
	guardFS, err := static.EmbedFolder(server, "ui/guard/dist")
	if err != nil {
		logging.Sugar.Fatalf("Failed to embed guard folder")
		return
	}

	// Prepare runtime config injection for SPA index.html files.
	// Only inject values that can change at runtime without rebuilding the frontend.
	// Paths are already baked into the frontend by Vite `base`.
	runtimeCfg := map[string]string{}

	// baseURL: explicit backend-host for cross-origin, otherwise rootPath for same-origin.
	runtimeBaseURL := appConfig.RootPath
	if v := viper.GetString("BACKEND_HOST"); v != "" {
		runtimeBaseURL = v
	} else if v := viper.GetString("wall.backend-host"); v != "" {
		runtimeBaseURL = v
	}
	runtimeCfg["baseURL"] = runtimeBaseURL

	if appConfig.GuardDomain != "" {
		runtimeCfg["guardDomain"] = appConfig.GuardDomain
	}

	configJSON, err := json.Marshal(runtimeCfg)
	if err != nil {
		logging.Sugar.Fatalf("Failed to marshal runtime config: %s", err)
		return
	}
	configScript := fmt.Sprintf(`<script>window.__APP_CONFIG__=%s</script>`, configJSON)

	injectConfig := func(htmlPath string) []byte {
		raw, err := server.ReadFile(htmlPath)
		if err != nil {
			logging.Sugar.Fatalf("Failed to read %s: %s", htmlPath, err)
		}
		return []byte(strings.Replace(string(raw), "</head>", configScript+"</head>", 1))
	}
	adminHTML := injectConfig("ui/admin/dist/index.html")
	guardHTML := injectConfig("ui/guard/dist/index.html")

	r.NoRoute(func(c *gin.Context) {
		if strings.HasPrefix(c.Request.URL.Path, appConfig.RootPath+appConfig.AdminPath) {
			c.Data(http.StatusOK, "text/html; charset=utf-8", adminHTML)
			return
		}
		if strings.HasPrefix(c.Request.URL.Path, appConfig.RootPath+appConfig.GuardPath) {
			c.Data(http.StatusOK, "text/html; charset=utf-8", guardHTML)
			return
		}
		if strings.HasPrefix(c.Request.URL.Path, appConfig.RootPath) {
			// Requests under root path that don't match any route → 404
			return
		}
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

	// Request filtering (SQLi, XSS, Path Traversal detection — enhanced rule engine)
	if viper.GetBool("request-filtering.enabled") {
		r.Use(middleware.WAF())
		logging.Sugar.Info("Request filtering enabled")
	}

	// Bot management — controlled by DB settings
	r.Use(middleware.BotManagement())

	// CAPTCHA page — serves CAPTCHA challenge when bot_captcha flag is set
	r.Use(middleware.CaptchaPage())

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

	// JS Challenge (Proof-of-Work) middleware — always registered, controlled by DB settings at runtime
	r.Use(middleware.JSChallenge())

	// Static files — after security middleware so challenges/blocks apply to all requests.
	// Paths must match the Vite `base` used during frontend build.
	// Self-compile: .env sets paths for both Vite and Go, so they match.
	// Pre-built binary: uses defaults; changing paths at runtime requires a frontend rebuild.
	r.Use(static.Serve(appConfig.RootPath+appConfig.AdminPath, adminFS))
	r.Use(static.Serve(appConfig.RootPath+appConfig.GuardPath, guardFS))

	// Register routes AFTER all middleware so they are included in the handler chain
	_http.Setup(r, appConfig)

	// Server configuration with timeouts
	// Priority: OS env (Docker) → config.yaml → default
	port := os.Getenv("PORT")
	if port == "" {
		port = viper.GetString("server.port")
	}
	if port == "" {
		port = "9999"
	}

	srv := &http.Server{
		Addr:              ":" + port,
		Handler:           r,
		ReadHeaderTimeout: 5 * time.Second,
		ReadTimeout:       10 * time.Second,
		WriteTimeout:      30 * time.Second,
		IdleTimeout:       60 * time.Second,
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

	cache.StopSync()
	pkg.CloseGeoIP()
	logging.Sugar.Info("Server exited gracefully")
}
