package main

import (
	"embed"
	"flag"
	"net/http"
	_http "open-website-defender/internal/adapter/controller/http"
	"open-website-defender/internal/adapter/controller/http/middleware"
	"open-website-defender/internal/infrastructure/config"
	"open-website-defender/internal/infrastructure/database"
	"open-website-defender/internal/infrastructure/logging"
	"os"
	"path/filepath"
	"strings"

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

	err = database.InitDB()
	if err != nil {
		logging.Sugar.Fatalf("Error initializing database: %s", err)
		return
	}

	appConfig := getAppConfig()

	r := _http.Setup(appConfig)

	r.RedirectTrailingSlash = true
	trustedProxies := viper.GetStringSlice("trustedProxies")
	err = r.SetTrustedProxies(trustedProxies)
	logging.Sugar.Info("TrustedProxies:", trustedProxies)
	if err != nil {
		logging.Sugar.Warn("Failed to set trusted proxies:", err)
		return
	}

	//r.Use(func(c *gin.Context) {
	//	path := c.Request.URL.Path
	//	adminRoot := appConfig.RootPath + appConfig.AdminPath
	//	guardRoot := appConfig.RootPath + appConfig.GuardPath
	//
	//	if path == adminRoot || path == guardRoot {
	//		c.Redirect(http.StatusMovedPermanently, path+"/")
	//		c.Abort()
	//		return
	//	}
	//	c.Next()
	//})

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

	r.Use(middleware.Logger())
	r.Use(middleware.Recovery())
	r.Use(func(c *gin.Context) {
		c.Writer.Header().Set("Access-Control-Allow-Origin", c.GetHeader("Origin"))
		c.Writer.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		c.Writer.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization, X-Requested-With, Cookie")
		c.Writer.Header().Set("Access-Control-Allow-Credentials", "true")
		c.Writer.Header().Set("Access-Control-Max-Age", "86400") // 24 hours

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}
		c.Next()
	})
	port := os.Getenv("PORT")
	if port == "" {
		port = "9999"
	}
	err = r.Run(":" + port)
	if err != nil {
		return
	}
}
