package installer

import (
	"context"
	"fmt"

	"github.com/fasmail/panel/internal/auth"
	"github.com/fasmail/panel/internal/config"
	"github.com/fasmail/panel/internal/database"
	"github.com/fasmail/panel/internal/docker"
	"github.com/fasmail/panel/internal/models"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Service struct {
	dockerClient *docker.Client
	appConfig    *config.Config
}

func NewService(dockerClient *docker.Client, appConfig *config.Config) *Service {
	return &Service{
		dockerClient: dockerClient,
		appConfig:    appConfig,
	}
}

func (s *Service) TestDatabaseConnection(ctx context.Context, cfg config.DatabaseConfig) error {
	return database.TestConnection(ctx, cfg)
}

func (s *Service) SaveDatabaseConfig(cfg config.DatabaseConfig, redisCfg config.RedisConfig) error {
	s.appConfig.Database = cfg
	s.appConfig.Redis = redisCfg

	if s.appConfig.JWT.Secret == "" {
		s.appConfig.JWT.Secret = config.GenerateJWTSecret()
	}

	return config.SaveConfig(s.appConfig, "")
}

func (s *Service) IsDockerAvailable(ctx context.Context) bool {
	if s.dockerClient == nil {
		return false
	}
	return s.dockerClient.IsAvailable(ctx)
}

func (s *Service) AutoInstallWithDocker(ctx context.Context) (*config.DatabaseConfig, *config.RedisConfig, error) {
	if s.dockerClient == nil {
		return nil, nil, fmt.Errorf("cliente Docker no disponible")
	}

	if !s.dockerClient.IsAvailable(ctx) {
		return nil, nil, fmt.Errorf("Docker daemon no está disponible")
	}

	networkName := s.appConfig.Docker.NetworkName
	if networkName == "" {
		networkName = "fasmail-network"
	}

	if err := s.dockerClient.EnsureNetwork(ctx, networkName); err != nil {
		return nil, nil, fmt.Errorf("crear red Docker: %w", err)
	}

	dbPassword := config.GeneratePassword(24)
	dbUser := "fasmail"
	dbName := "fasmail"

	pgContainerID, err := s.dockerClient.CreatePostgres(ctx, docker.PostgresOptions{
		User:        dbUser,
		Password:    dbPassword,
		DBName:      dbName,
		NetworkName: networkName,
	})
	if err != nil {
		return nil, nil, fmt.Errorf("crear contenedor PostgreSQL: %w", err)
	}

	pgIP, err := s.dockerClient.GetContainerIP(ctx, pgContainerID, networkName)
	if err != nil {
		return nil, nil, fmt.Errorf("obtener IP de PostgreSQL: %w", err)
	}

	redisContainerID, err := s.dockerClient.CreateRedis(ctx, docker.RedisOptions{
		NetworkName: networkName,
	})
	if err != nil {
		return nil, nil, fmt.Errorf("crear contenedor Redis: %w", err)
	}

	redisIP, err := s.dockerClient.GetContainerIP(ctx, redisContainerID, networkName)
	if err != nil {
		return nil, nil, fmt.Errorf("obtener IP de Redis: %w", err)
	}

	dbConfig := &config.DatabaseConfig{
		Host:     pgIP,
		Port:     5432,
		User:     dbUser,
		Password: dbPassword,
		DBName:   dbName,
		SSLMode:  "disable",
	}

	redisConfig := &config.RedisConfig{
		Addr: fmt.Sprintf("%s:6379", redisIP),
		DB:   0,
	}

	return dbConfig, redisConfig, nil
}

func (s *Service) RunMigrations(ctx context.Context, pool *pgxpool.Pool) error {
	return database.RunMigrations(ctx, pool)
}

func (s *Service) CreateAdminUser(ctx context.Context, pool *pgxpool.Pool, email, password, displayName string) error {
	userRepo := models.NewUserRepository(pool)

	exists, err := userRepo.EmailExists(ctx, email)
	if err != nil {
		return fmt.Errorf("verificar email: %w", err)
	}
	if exists {
		return fmt.Errorf("el email ya está registrado: %s", email)
	}

	hash, err := auth.HashPassword(password)
	if err != nil {
		return fmt.Errorf("hash password: %w", err)
	}

	// Assign to default company
	defaultCompanyID := uuid.MustParse("00000000-0000-0000-0000-000000000001")

	admin := &models.User{
		Email:              email,
		PasswordHash:       hash,
		DisplayName:        displayName,
		Role:               "super_admin",
		MustChangePassword: true,
		IsActive:           true,
		CompanyID:          &defaultCompanyID,
	}

	return userRepo.Create(ctx, admin)
}

func (s *Service) FinalizeInstallation(ctx context.Context, pool *pgxpool.Pool) error {
	systemRepo := models.NewSystemConfigRepository(pool)
	return systemRepo.MarkInstalled(ctx)
}

func (s *Service) GetConfig() *config.Config {
	return s.appConfig
}
