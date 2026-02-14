package main

import (
	"context"
	"html/template"
	"io/fs"
	"log"
	"net/http"
	"os"
	"os/signal"
	"sync/atomic"
	"syscall"
	"time"

	"github.com/fasmail/panel/internal/admin"
	"github.com/fasmail/panel/internal/auth"
	"github.com/fasmail/panel/internal/config"
	"github.com/fasmail/panel/internal/database"
	dockermgr "github.com/fasmail/panel/internal/docker"
	"github.com/fasmail/panel/internal/installer"
	"github.com/fasmail/panel/internal/models"
	"github.com/fasmail/panel/web"
	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/redis/go-redis/v9"
)

var installed atomic.Bool

func main() {
	cfg, err := config.LoadConfig()
	if err != nil {
		log.Printf("Warning: could not load config: %v", err)
		cfg = config.DefaultConfig()
	}

	if cfg.JWT.Secret == "" {
		cfg.JWT.Secret = config.GenerateJWTSecret()
	}

	var pool *pgxpool.Pool
	var redisClient *redis.Client

	ctx := context.Background()

	// Try to connect to database if config exists
	if config.ConfigFileExists() {
		p, err := database.NewPostgresPool(ctx, cfg.Database)
		if err != nil {
			log.Printf("Warning: could not connect to database: %v", err)
		} else {
			pool = p
		}

		rc, err := database.NewRedisClient(cfg.Redis)
		if err != nil {
			log.Printf("Warning: could not connect to Redis: %v", err)
		} else {
			redisClient = rc
		}
	}

	// Check installation status
	if pool != nil && database.CheckTablesExist(ctx, pool) {
		systemRepo := models.NewSystemConfigRepository(pool)
		isInstalled, err := systemRepo.IsInstalled(ctx)
		if err == nil && isInstalled {
			installed.Store(true)
		}
	}

	// Docker client (best effort)
	dockerClient, err := dockermgr.NewClient()
	if err != nil {
		log.Printf("Warning: Docker not available: %v", err)
	}

	// Setup Gin
	gin.SetMode(gin.ReleaseMode)
	r := gin.Default()

	// Load templates
	tmpl := loadTemplates()
	r.SetHTMLTemplate(tmpl)

	// Serve static files from embedded FS
	staticFS, _ := fs.Sub(web.StaticFS, "static")
	r.StaticFS("/static", http.FS(staticFS))

	// Health endpoint
	r.GET("/health", func(c *gin.Context) {
		pgOK := pool != nil && database.PingPool(ctx, pool) == nil
		redisOK := redisClient != nil && database.PingRedis(ctx, redisClient) == nil

		status := "ok"
		code := http.StatusOK
		if installed.Load() && (!pgOK || !redisOK) {
			status = "degraded"
			code = http.StatusServiceUnavailable
		}

		c.JSON(code, gin.H{
			"status":    status,
			"postgres":  pgOK,
			"redis":     redisOK,
			"installed": installed.Load(),
		})
	})

	// Installer guard middleware
	r.Use(func(c *gin.Context) {
		path := c.Request.URL.Path

		// Always allow static, health
		if path == "/health" || len(path) >= 7 && path[:7] == "/static" {
			c.Next()
			return
		}

		isInst := installed.Load()

		// If installed, block installer routes
		if isInst && len(path) >= 8 && path[:8] == "/install" {
			c.Redirect(http.StatusFound, "/auth/login")
			c.Abort()
			return
		}

		// If not installed, only allow installer routes
		if !isInst && !(len(path) >= 8 && path[:8] == "/install") {
			c.Redirect(http.StatusFound, "/install/")
			c.Abort()
			return
		}

		c.Next()
	})

	// Root redirect
	r.GET("/", func(c *gin.Context) {
		if installed.Load() {
			c.Redirect(http.StatusFound, "/admin/")
		} else {
			c.Redirect(http.StatusFound, "/install/")
		}
	})

	// ---- INSTALLER ROUTES ----
	installerSvc := installer.NewService(dockerClient, cfg)
	installerHandler := installer.NewHandler(installerSvc, &pool, func() {
		installed.Store(true)
		// Reconnect redis if needed
		if redisClient == nil {
			rc, err := database.NewRedisClient(cfg.Redis)
			if err == nil {
				redisClient = rc
			}
		}
	})

	installGroup := r.Group("/install")
	{
		installGroup.GET("/", installerHandler.ShowWelcome)
		installGroup.GET("/database", installerHandler.ShowDatabaseConfig)
		installGroup.POST("/database", installerHandler.HandleDatabaseConfig)
		installGroup.GET("/docker", installerHandler.ShowDockerSetup)
		installGroup.POST("/docker", installerHandler.HandleDockerSetup)
		installGroup.POST("/test-connection", installerHandler.HandleTestConnection)
		installGroup.GET("/migrate", installerHandler.ShowMigration)
		installGroup.POST("/migrate", installerHandler.HandleMigration)
		installGroup.GET("/admin", installerHandler.ShowAdminSetup)
		installGroup.POST("/admin", installerHandler.HandleAdminSetup)
		installGroup.GET("/complete", installerHandler.ShowComplete)
		installGroup.POST("/finalize", installerHandler.HandleFinalize)
	}

	// ---- AUTH ROUTES ----
	jwtMgr := auth.NewJWTManager(cfg.JWT)

	authSvc := auth.NewService(&pool, &redisClient, jwtMgr)
	authHandler := auth.NewHandler(authSvc)

	authGroup := r.Group("/auth")
	{
		authGroup.GET("/login", authHandler.ShowLoginPage)
		authGroup.POST("/login", authHandler.HandleLogin)
	}

	authProtected := r.Group("/auth")
	authProtected.Use(auth.AuthRequired(jwtMgr, &redisClient))
	{
		authProtected.POST("/logout", authHandler.HandleLogout)
		authProtected.GET("/change-password", authHandler.ShowChangePassword)
		authProtected.POST("/change-password", authHandler.HandleChangePassword)
	}

	// ---- ADMIN ROUTES ----
	adminGroup := r.Group("/admin")
	adminGroup.Use(auth.AuthRequired(jwtMgr, &redisClient))
	adminGroup.Use(auth.ForcePasswordChange())
	{
		adminGroup.GET("/", func(c *gin.Context) {
			if pool == nil {
				c.Redirect(http.StatusFound, "/auth/login")
				return
			}
			userRepo := models.NewUserRepository(pool)
			systemRepo := models.NewSystemConfigRepository(pool)
			adminSvc := admin.NewService(userRepo, systemRepo, pool, redisClient)
			adminHandler := admin.NewHandler(adminSvc)
			adminHandler.ShowDashboard(c)
		})
		adminGroup.GET("/settings", func(c *gin.Context) {
			if pool == nil {
				c.Redirect(http.StatusFound, "/auth/login")
				return
			}
			userRepo := models.NewUserRepository(pool)
			systemRepo := models.NewSystemConfigRepository(pool)
			adminSvc := admin.NewService(userRepo, systemRepo, pool, redisClient)
			adminHandler := admin.NewHandler(adminSvc)
			adminHandler.ShowSettings(c)
		})
		adminGroup.POST("/settings", func(c *gin.Context) {
			if pool == nil {
				c.Redirect(http.StatusFound, "/auth/login")
				return
			}
			userRepo := models.NewUserRepository(pool)
			systemRepo := models.NewSystemConfigRepository(pool)
			adminSvc := admin.NewService(userRepo, systemRepo, pool, redisClient)
			adminHandler := admin.NewHandler(adminSvc)
			adminHandler.HandleSettings(c)
		})
	}

	// Start server
	port := cfg.Server.Port
	if port == "" {
		port = ":8080"
	}

	srv := &http.Server{
		Addr:         port,
		Handler:      r,
		ReadTimeout:  cfg.Server.ReadTimeout,
		WriteTimeout: cfg.Server.WriteTimeout,
	}

	go func() {
		log.Printf("FasMail Panel starting on %s", port)
		if installed.Load() {
			log.Println("Mode: Admin (system installed)")
		} else {
			log.Println("Mode: Installer (first run)")
		}
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Server error: %v", err)
		}
	}()

	// Graceful shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	log.Println("Shutting down server...")

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := srv.Shutdown(shutdownCtx); err != nil {
		log.Printf("Server forced to shutdown: %v", err)
	}

	if pool != nil {
		pool.Close()
	}
	if redisClient != nil {
		redisClient.Close()
	}
	if dockerClient != nil {
		dockerClient.Close()
	}

	log.Println("Server exited")
}

func loadTemplates() *template.Template {
	tmpl := template.New("")

	templateDirs := []string{
		"templates/installer/*.html",
		"templates/auth/*.html",
		"templates/admin/*.html",
		"templates/layouts/*.html",
	}

	for _, pattern := range templateDirs {
		matches, err := fs.Glob(web.TemplateFS, pattern)
		if err != nil {
			log.Printf("Warning: glob pattern %s: %v", pattern, err)
			continue
		}
		for _, match := range matches {
			data, err := fs.ReadFile(web.TemplateFS, match)
			if err != nil {
				log.Printf("Warning: read template %s: %v", match, err)
				continue
			}
			_, err = tmpl.Parse(string(data))
			if err != nil {
				log.Printf("Warning: parse template %s: %v", match, err)
			}
		}
	}

	return tmpl
}

