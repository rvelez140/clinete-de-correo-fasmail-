package installer

import (
	"context"
	"net/http"
	"strconv"

	"github.com/fasmail/panel/internal/config"
	"github.com/fasmail/panel/internal/database"
	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Handler struct {
	service    *Service
	pool       **pgxpool.Pool
	onComplete func()
}

func NewHandler(service *Service, pool **pgxpool.Pool, onComplete func()) *Handler {
	return &Handler{
		service:    service,
		pool:       pool,
		onComplete: onComplete,
	}
}

func (h *Handler) ShowWelcome(c *gin.Context) {
	dockerAvailable := h.service.IsDockerAvailable(c.Request.Context())
	c.HTML(http.StatusOK, "installer_welcome", gin.H{
		"Title":           "Bienvenido",
		"Step":            1,
		"DockerAvailable": dockerAvailable,
	})
}

func (h *Handler) ShowDatabaseConfig(c *gin.Context) {
	c.HTML(http.StatusOK, "installer_database", gin.H{
		"Title":   "Configuración de Base de Datos",
		"Step":    2,
		"Host":    "localhost",
		"Port":    "5432",
		"User":    "fasmail",
		"DBName":  "fasmail",
		"SSLMode": "disable",
	})
}

func (h *Handler) HandleDatabaseConfig(c *gin.Context) {
	host := c.PostForm("host")
	portStr := c.PostForm("port")
	user := c.PostForm("user")
	password := c.PostForm("password")
	dbname := c.PostForm("dbname")
	sslmode := c.PostForm("sslmode")
	redisAddr := c.PostForm("redis_addr")

	port, err := strconv.Atoi(portStr)
	if err != nil {
		port = 5432
	}

	if redisAddr == "" {
		redisAddr = "localhost:6379"
	}

	dbCfg := config.DatabaseConfig{
		Host:     host,
		Port:     port,
		User:     user,
		Password: password,
		DBName:   dbname,
		SSLMode:  sslmode,
	}

	if err := h.service.TestDatabaseConnection(c.Request.Context(), dbCfg); err != nil {
		c.HTML(http.StatusBadRequest, "installer_database", gin.H{
			"Title":   "Configuración de Base de Datos",
			"Step":    2,
			"Error":   "No se pudo conectar a la base de datos: " + err.Error(),
			"Host":    host,
			"Port":    portStr,
			"User":    user,
			"DBName":  dbname,
			"SSLMode": sslmode,
		})
		return
	}

	redisCfg := config.RedisConfig{
		Addr: redisAddr,
	}

	if err := h.service.SaveDatabaseConfig(dbCfg, redisCfg); err != nil {
		c.HTML(http.StatusInternalServerError, "installer_database", gin.H{
			"Title": "Configuración de Base de Datos",
			"Step":  2,
			"Error": "Error al guardar la configuración: " + err.Error(),
			"Host":  host,
			"Port":  portStr,
			"User":  user,
			"DBName": dbname,
		})
		return
	}

	pool, err := database.NewPostgresPool(c.Request.Context(), dbCfg)
	if err != nil {
		c.HTML(http.StatusInternalServerError, "installer_database", gin.H{
			"Title": "Configuración de Base de Datos",
			"Step":  2,
			"Error": "Error al crear el pool de conexiones: " + err.Error(),
		})
		return
	}

	*h.pool = pool

	c.Redirect(http.StatusFound, "/install/migrate")
}

func (h *Handler) ShowDockerSetup(c *gin.Context) {
	dockerAvailable := h.service.IsDockerAvailable(c.Request.Context())
	c.HTML(http.StatusOK, "installer_docker", gin.H{
		"Title":           "Instalación Automática con Docker",
		"Step":            2,
		"DockerAvailable": dockerAvailable,
	})
}

func (h *Handler) HandleDockerSetup(c *gin.Context) {
	ctx := c.Request.Context()

	dbCfg, redisCfg, err := h.service.AutoInstallWithDocker(ctx)
	if err != nil {
		c.HTML(http.StatusInternalServerError, "installer_docker", gin.H{
			"Title":           "Instalación Automática con Docker",
			"Step":            2,
			"Error":           "Error en la instalación: " + err.Error(),
			"DockerAvailable": true,
		})
		return
	}

	if err := h.service.SaveDatabaseConfig(*dbCfg, *redisCfg); err != nil {
		c.HTML(http.StatusInternalServerError, "installer_docker", gin.H{
			"Title": "Instalación Automática con Docker",
			"Step":  2,
			"Error": "Error al guardar la configuración: " + err.Error(),
		})
		return
	}

	pool, err := database.NewPostgresPool(ctx, *dbCfg)
	if err != nil {
		c.HTML(http.StatusInternalServerError, "installer_docker", gin.H{
			"Title": "Instalación Automática con Docker",
			"Step":  2,
			"Error": "Error al conectar a la base de datos creada: " + err.Error(),
		})
		return
	}

	*h.pool = pool

	c.Redirect(http.StatusFound, "/install/migrate")
}

func (h *Handler) HandleTestConnection(c *gin.Context) {
	var req struct {
		Host     string `json:"host"`
		Port     int    `json:"port"`
		User     string `json:"user"`
		Password string `json:"password"`
		DBName   string `json:"dbname"`
		SSLMode  string `json:"sslmode"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "error": "datos inválidos"})
		return
	}

	dbCfg := config.DatabaseConfig{
		Host:     req.Host,
		Port:     req.Port,
		User:     req.User,
		Password: req.Password,
		DBName:   req.DBName,
		SSLMode:  req.SSLMode,
	}

	if err := h.service.TestDatabaseConnection(c.Request.Context(), dbCfg); err != nil {
		c.JSON(http.StatusOK, gin.H{"success": false, "error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"success": true, "message": "Conexión exitosa"})
}

func (h *Handler) ShowMigration(c *gin.Context) {
	var migrationList []database.MigrationInfo
	if *h.pool != nil {
		list, err := database.GetMigrationStatus(c.Request.Context(), *h.pool)
		if err == nil {
			migrationList = list
		}
	}

	c.HTML(http.StatusOK, "installer_migration", gin.H{
		"Title":      "Migraciones de Base de Datos",
		"Step":       3,
		"Migrations": migrationList,
	})
}

func (h *Handler) HandleMigration(c *gin.Context) {
	if *h.pool == nil {
		c.HTML(http.StatusBadRequest, "installer_migration", gin.H{
			"Title": "Migraciones de Base de Datos",
			"Step":  3,
			"Error": "No hay conexión a la base de datos. Vuelva al paso anterior.",
		})
		return
	}

	if err := h.service.RunMigrations(c.Request.Context(), *h.pool); err != nil {
		c.HTML(http.StatusInternalServerError, "installer_migration", gin.H{
			"Title": "Migraciones de Base de Datos",
			"Step":  3,
			"Error": "Error al ejecutar las migraciones: " + err.Error(),
		})
		return
	}

	c.Redirect(http.StatusFound, "/install/admin")
}

func (h *Handler) ShowAdminSetup(c *gin.Context) {
	c.HTML(http.StatusOK, "installer_admin", gin.H{
		"Title":       "Crear Usuario Administrador",
		"Step":        4,
		"Email":       "admin@fasmail.local",
		"DisplayName": "Administrador",
	})
}

func (h *Handler) HandleAdminSetup(c *gin.Context) {
	email := c.PostForm("email")
	password := c.PostForm("password")
	confirmPassword := c.PostForm("confirm_password")
	displayName := c.PostForm("display_name")

	if email == "" || password == "" || displayName == "" {
		c.HTML(http.StatusBadRequest, "installer_admin", gin.H{
			"Title":       "Crear Usuario Administrador",
			"Step":        4,
			"Error":       "Todos los campos son requeridos",
			"Email":       email,
			"DisplayName": displayName,
		})
		return
	}

	if password != confirmPassword {
		c.HTML(http.StatusBadRequest, "installer_admin", gin.H{
			"Title":       "Crear Usuario Administrador",
			"Step":        4,
			"Error":       "Las contraseñas no coinciden",
			"Email":       email,
			"DisplayName": displayName,
		})
		return
	}

	if err := h.service.CreateAdminUser(c.Request.Context(), *h.pool, email, password, displayName); err != nil {
		c.HTML(http.StatusBadRequest, "installer_admin", gin.H{
			"Title":       "Crear Usuario Administrador",
			"Step":        4,
			"Error":       err.Error(),
			"Email":       email,
			"DisplayName": displayName,
		})
		return
	}

	c.Redirect(http.StatusFound, "/install/complete")
}

func (h *Handler) ShowComplete(c *gin.Context) {
	cfg := h.service.GetConfig()
	c.HTML(http.StatusOK, "installer_complete", gin.H{
		"Title":    "Instalación Completa",
		"Step":     5,
		"DBHost":   cfg.Database.Host,
		"DBPort":   cfg.Database.Port,
		"DBName":   cfg.Database.DBName,
		"DBUser":   cfg.Database.User,
		"RedisAddr": cfg.Redis.Addr,
	})
}

func (h *Handler) HandleFinalize(c *gin.Context) {
	if *h.pool == nil {
		c.Redirect(http.StatusFound, "/install/")
		return
	}

	if err := h.service.FinalizeInstallation(context.Background(), *h.pool); err != nil {
		c.HTML(http.StatusInternalServerError, "installer_complete", gin.H{
			"Title": "Instalación Completa",
			"Step":  5,
			"Error": "Error al finalizar la instalación: " + err.Error(),
		})
		return
	}

	if h.onComplete != nil {
		h.onComplete()
	}

	c.Redirect(http.StatusFound, "/auth/login")
}
