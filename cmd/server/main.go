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
	"github.com/fasmail/panel/internal/api"
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

	// Serve uploaded files (logos, etc.) from /data/
	r.Static("/uploads", "/data")

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

		// Always allow static, health, uploads, and API routes
		if path == "/health" ||
			len(path) >= 7 && path[:7] == "/static" ||
			len(path) >= 8 && path[:8] == "/uploads" ||
			len(path) >= 5 && path[:5] == "/api/" {
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
	authGroup.Use(auth.CompanyBrandingMiddleware(&pool))
	{
		authGroup.GET("/login", authHandler.ShowLoginPage)
		authGroup.POST("/login", authHandler.HandleLogin)
	}

	authProtected := r.Group("/auth")
	authProtected.Use(auth.AuthRequired(jwtMgr, &redisClient))
	authProtected.Use(auth.CompanyBrandingMiddleware(&pool))
	{
		authProtected.POST("/logout", authHandler.HandleLogout)
		authProtected.GET("/change-password", authHandler.ShowChangePassword)
		authProtected.POST("/change-password", authHandler.HandleChangePassword)
	}

	// ---- ADMIN ROUTES ----
	adminGroup := r.Group("/admin")
	adminGroup.Use(auth.AuthRequired(jwtMgr, &redisClient))
	adminGroup.Use(auth.ForcePasswordChange())
	adminGroup.Use(auth.CompanyBrandingMiddleware(&pool))
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

	// ---- COMPANY MANAGEMENT ROUTES (super_admin only) ----
	companiesGroup := adminGroup.Group("/companies")
	companiesGroup.Use(auth.SuperAdminRequired())
	{
		companiesGroup.GET("/", func(c *gin.Context) {
			if pool == nil {
				c.Redirect(http.StatusFound, "/auth/login")
				return
			}
			companyRepo := models.NewCompanyRepository(pool)
			userRepo := models.NewUserRepository(pool)
			companySvc := admin.NewCompanyService(companyRepo, userRepo)
			companyHandler := admin.NewCompanyHandler(companySvc)
			companyHandler.ShowCompanies(c)
		})
		companiesGroup.GET("/new", func(c *gin.Context) {
			if pool == nil {
				c.Redirect(http.StatusFound, "/auth/login")
				return
			}
			companyRepo := models.NewCompanyRepository(pool)
			userRepo := models.NewUserRepository(pool)
			companySvc := admin.NewCompanyService(companyRepo, userRepo)
			companyHandler := admin.NewCompanyHandler(companySvc)
			companyHandler.ShowCreateCompany(c)
		})
		companiesGroup.POST("/new", func(c *gin.Context) {
			if pool == nil {
				c.Redirect(http.StatusFound, "/auth/login")
				return
			}
			companyRepo := models.NewCompanyRepository(pool)
			userRepo := models.NewUserRepository(pool)
			companySvc := admin.NewCompanyService(companyRepo, userRepo)
			companyHandler := admin.NewCompanyHandler(companySvc)
			companyHandler.HandleCreateCompany(c)
		})
		companiesGroup.GET("/:id/edit", func(c *gin.Context) {
			if pool == nil {
				c.Redirect(http.StatusFound, "/auth/login")
				return
			}
			companyRepo := models.NewCompanyRepository(pool)
			userRepo := models.NewUserRepository(pool)
			companySvc := admin.NewCompanyService(companyRepo, userRepo)
			companyHandler := admin.NewCompanyHandler(companySvc)
			companyHandler.ShowEditCompany(c)
		})
		companiesGroup.POST("/:id/edit", func(c *gin.Context) {
			if pool == nil {
				c.Redirect(http.StatusFound, "/auth/login")
				return
			}
			companyRepo := models.NewCompanyRepository(pool)
			userRepo := models.NewUserRepository(pool)
			companySvc := admin.NewCompanyService(companyRepo, userRepo)
			companyHandler := admin.NewCompanyHandler(companySvc)
			companyHandler.HandleEditCompany(c)
		})
		companiesGroup.POST("/:id/logo", func(c *gin.Context) {
			if pool == nil {
				c.Redirect(http.StatusFound, "/auth/login")
				return
			}
			companyRepo := models.NewCompanyRepository(pool)
			userRepo := models.NewUserRepository(pool)
			companySvc := admin.NewCompanyService(companyRepo, userRepo)
			companyHandler := admin.NewCompanyHandler(companySvc)
			companyHandler.HandleUploadLogo(c)
		})
		companiesGroup.POST("/:id/delete", func(c *gin.Context) {
			if pool == nil {
				c.Redirect(http.StatusFound, "/auth/login")
				return
			}
			companyRepo := models.NewCompanyRepository(pool)
			userRepo := models.NewUserRepository(pool)
			companySvc := admin.NewCompanyService(companyRepo, userRepo)
			companyHandler := admin.NewCompanyHandler(companySvc)
			companyHandler.HandleDeleteCompany(c)
		})
	}

	// ---- DOWNLOADS WEB ROUTES (authenticated users) ----
	adminGroup.GET("/downloads", func(c *gin.Context) {
		if pool == nil {
			c.Redirect(http.StatusFound, "/auth/login")
			return
		}
		systemRepo := models.NewSystemConfigRepository(pool)
		serverURL, _ := systemRepo.Get(c.Request.Context(), "server_url")
		if serverURL == "" {
			serverURL = "http://localhost:8080"
		}
		downloadsHandler := admin.NewDownloadsHandler(pool, serverURL)
		downloadsHandler.ShowDownloads(c)
	})
	adminGroup.POST("/downloads/generate", func(c *gin.Context) {
		if pool == nil {
			c.Redirect(http.StatusFound, "/auth/login")
			return
		}
		systemRepo := models.NewSystemConfigRepository(pool)
		serverURL, _ := systemRepo.Get(c.Request.Context(), "server_url")
		if serverURL == "" {
			serverURL = "http://localhost:8080"
		}
		downloadsHandler := admin.NewDownloadsHandler(pool, serverURL)
		downloadsHandler.HandleGenerateToken(c)
	})

	// ---- REST API v1 ROUTES ----
	apiV1 := r.Group("/api/v1")
	{
		// Public auth endpoints
		apiAuthHandler := api.NewAuthHandler(authSvc, jwtMgr, &pool)

		apiAuthGroup := apiV1.Group("/auth")
		{
			apiAuthGroup.POST("/login", apiAuthHandler.Login)
			apiAuthGroup.POST("/refresh", apiAuthHandler.Refresh)
		}

		// Public token validation (called by app on first launch)
		apiDownloadsPublic := apiV1.Group("/downloads")
		{
			if pool != nil {
				apiDlHandler := api.NewDownloadsHandler(pool)
				apiDownloadsPublic.GET("/validate-token", apiDlHandler.ValidateToken)
				apiDownloadsPublic.GET("/apps/:filename", apiDlHandler.ServeApp)
			}
		}

		// Authenticated API endpoints
		apiProtected := apiV1.Group("")
		apiProtected.Use(api.APIAuthRequired(jwtMgr, &redisClient))
		apiProtected.Use(api.APIForcePasswordChange())
		{
			// Auth (protected)
			apiProtectedAuth := apiProtected.Group("/auth")
			{
				apiProtectedAuth.POST("/logout", apiAuthHandler.Logout)
				apiProtectedAuth.GET("/profile", apiAuthHandler.Profile)
			}

			// Dashboard
			apiProtected.GET("/admin/dashboard", func(c *gin.Context) {
				if pool == nil {
					c.JSON(http.StatusInternalServerError, gin.H{"error": "base de datos no disponible"})
					return
				}
				userRepo := models.NewUserRepository(pool)
				systemRepo := models.NewSystemConfigRepository(pool)
				adminSvc := admin.NewService(userRepo, systemRepo, pool, redisClient)
				companyRepo := models.NewCompanyRepository(pool)
				companySvc := admin.NewCompanyService(companyRepo, userRepo)
				apiAdminHandler := api.NewAdminHandler(adminSvc, companySvc)
				apiAdminHandler.Dashboard(c)
			})

			// Settings
			apiProtected.GET("/admin/settings", func(c *gin.Context) {
				if pool == nil {
					c.JSON(http.StatusInternalServerError, gin.H{"error": "base de datos no disponible"})
					return
				}
				userRepo := models.NewUserRepository(pool)
				systemRepo := models.NewSystemConfigRepository(pool)
				adminSvc := admin.NewService(userRepo, systemRepo, pool, redisClient)
				apiAdminHandler := api.NewAdminHandler(adminSvc, nil)
				apiAdminHandler.GetSettings(c)
			})
			apiProtected.PUT("/admin/settings", func(c *gin.Context) {
				if pool == nil {
					c.JSON(http.StatusInternalServerError, gin.H{"error": "base de datos no disponible"})
					return
				}
				userRepo := models.NewUserRepository(pool)
				systemRepo := models.NewSystemConfigRepository(pool)
				adminSvc := admin.NewService(userRepo, systemRepo, pool, redisClient)
				apiAdminHandler := api.NewAdminHandler(adminSvc, nil)
				apiAdminHandler.UpdateSettings(c)
			})

			// Users
			apiProtected.GET("/admin/users", func(c *gin.Context) {
				if pool == nil {
					c.JSON(http.StatusInternalServerError, gin.H{"error": "base de datos no disponible"})
					return
				}
				userRepo := models.NewUserRepository(pool)
				systemRepo := models.NewSystemConfigRepository(pool)
				adminSvc := admin.NewService(userRepo, systemRepo, pool, redisClient)
				apiAdminHandler := api.NewAdminHandler(adminSvc, nil)
				apiAdminHandler.ListUsers(c)
			})

			// Companies (super_admin only)
			apiCompanies := apiProtected.Group("/admin/companies")
			apiCompanies.Use(api.APISuperAdminRequired())
			{
				apiCompanies.GET("", func(c *gin.Context) {
					if pool == nil {
						c.JSON(http.StatusInternalServerError, gin.H{"error": "base de datos no disponible"})
						return
					}
					companyRepo := models.NewCompanyRepository(pool)
					userRepo := models.NewUserRepository(pool)
					companySvc := admin.NewCompanyService(companyRepo, userRepo)
					adminSvc := admin.NewService(userRepo, models.NewSystemConfigRepository(pool), pool, redisClient)
					apiAdminHandler := api.NewAdminHandler(adminSvc, companySvc)
					apiAdminHandler.ListCompanies(c)
				})
				apiCompanies.POST("", func(c *gin.Context) {
					if pool == nil {
						c.JSON(http.StatusInternalServerError, gin.H{"error": "base de datos no disponible"})
						return
					}
					companyRepo := models.NewCompanyRepository(pool)
					userRepo := models.NewUserRepository(pool)
					companySvc := admin.NewCompanyService(companyRepo, userRepo)
					adminSvc := admin.NewService(userRepo, models.NewSystemConfigRepository(pool), pool, redisClient)
					apiAdminHandler := api.NewAdminHandler(adminSvc, companySvc)
					apiAdminHandler.CreateCompany(c)
				})
				apiCompanies.GET("/:id", func(c *gin.Context) {
					if pool == nil {
						c.JSON(http.StatusInternalServerError, gin.H{"error": "base de datos no disponible"})
						return
					}
					companyRepo := models.NewCompanyRepository(pool)
					userRepo := models.NewUserRepository(pool)
					companySvc := admin.NewCompanyService(companyRepo, userRepo)
					adminSvc := admin.NewService(userRepo, models.NewSystemConfigRepository(pool), pool, redisClient)
					apiAdminHandler := api.NewAdminHandler(adminSvc, companySvc)
					apiAdminHandler.GetCompany(c)
				})
				apiCompanies.PUT("/:id", func(c *gin.Context) {
					if pool == nil {
						c.JSON(http.StatusInternalServerError, gin.H{"error": "base de datos no disponible"})
						return
					}
					companyRepo := models.NewCompanyRepository(pool)
					userRepo := models.NewUserRepository(pool)
					companySvc := admin.NewCompanyService(companyRepo, userRepo)
					adminSvc := admin.NewService(userRepo, models.NewSystemConfigRepository(pool), pool, redisClient)
					apiAdminHandler := api.NewAdminHandler(adminSvc, companySvc)
					apiAdminHandler.UpdateCompany(c)
				})
				apiCompanies.DELETE("/:id", func(c *gin.Context) {
					if pool == nil {
						c.JSON(http.StatusInternalServerError, gin.H{"error": "base de datos no disponible"})
						return
					}
					companyRepo := models.NewCompanyRepository(pool)
					userRepo := models.NewUserRepository(pool)
					companySvc := admin.NewCompanyService(companyRepo, userRepo)
					adminSvc := admin.NewService(userRepo, models.NewSystemConfigRepository(pool), pool, redisClient)
					apiAdminHandler := api.NewAdminHandler(adminSvc, companySvc)
					apiAdminHandler.DeleteCompany(c)
				})
			}

			// Downloads token generation
			apiProtected.POST("/downloads/generate-token", func(c *gin.Context) {
				if pool == nil {
					c.JSON(http.StatusInternalServerError, gin.H{"error": "base de datos no disponible"})
					return
				}
				apiDlHandler := api.NewDownloadsHandler(pool)
				apiDlHandler.GenerateToken(c)
			})

			// Email accounts CRUD
			apiEmail := apiProtected.Group("/email/accounts")
			{
				apiEmail.GET("", func(c *gin.Context) {
					if pool == nil {
						c.JSON(http.StatusInternalServerError, gin.H{"error": "base de datos no disponible"})
						return
					}
					emailHandler := api.NewEmailHandler(pool, cfg.JWT.Secret)
					emailHandler.ListAccounts(c)
				})
				apiEmail.POST("", func(c *gin.Context) {
					if pool == nil {
						c.JSON(http.StatusInternalServerError, gin.H{"error": "base de datos no disponible"})
						return
					}
					emailHandler := api.NewEmailHandler(pool, cfg.JWT.Secret)
					emailHandler.CreateAccount(c)
				})
				apiEmail.GET("/:id", func(c *gin.Context) {
					if pool == nil {
						c.JSON(http.StatusInternalServerError, gin.H{"error": "base de datos no disponible"})
						return
					}
					emailHandler := api.NewEmailHandler(pool, cfg.JWT.Secret)
					emailHandler.GetAccount(c)
				})
				apiEmail.PUT("/:id", func(c *gin.Context) {
					if pool == nil {
						c.JSON(http.StatusInternalServerError, gin.H{"error": "base de datos no disponible"})
						return
					}
					emailHandler := api.NewEmailHandler(pool, cfg.JWT.Secret)
					emailHandler.UpdateAccount(c)
				})
				apiEmail.DELETE("/:id", func(c *gin.Context) {
					if pool == nil {
						c.JSON(http.StatusInternalServerError, gin.H{"error": "base de datos no disponible"})
						return
					}
					emailHandler := api.NewEmailHandler(pool, cfg.JWT.Secret)
					emailHandler.DeleteAccount(c)
				})
			}
		}
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

	// Load layouts first so partials are available to all templates
	templateDirs := []string{
		"templates/layouts/*.html",
		"templates/installer/*.html",
		"templates/auth/*.html",
		"templates/admin/*.html",
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
