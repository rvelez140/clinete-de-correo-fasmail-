package admin

import (
	"context"

	"github.com/fasmail/panel/internal/database"
	"github.com/fasmail/panel/internal/models"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/redis/go-redis/v9"
)

type DashboardStats struct {
	TotalUsers     int
	ActiveUsers    int
	TotalCompanies int
	SystemVersion  string
	InstalledAt    string
	PostgresOK     bool
	RedisOK        bool
}

type Service struct {
	userRepo   *models.UserRepository
	systemRepo *models.SystemConfigRepository
	pool       *pgxpool.Pool
	redis      *redis.Client
}

func NewService(userRepo *models.UserRepository, systemRepo *models.SystemConfigRepository, pool *pgxpool.Pool, redisClient *redis.Client) *Service {
	return &Service{
		userRepo:   userRepo,
		systemRepo: systemRepo,
		pool:       pool,
		redis:      redisClient,
	}
}

func (s *Service) GetDashboardStats(ctx context.Context) (*DashboardStats, error) {
	stats := &DashboardStats{}

	count, err := s.userRepo.Count(ctx)
	if err == nil {
		stats.TotalUsers = count
	}

	stats.ActiveUsers = count

	version, err := s.systemRepo.Get(ctx, "app_version")
	if err == nil {
		stats.SystemVersion = version
	}

	installedAt, err := s.systemRepo.Get(ctx, "installed_at")
	if err == nil {
		stats.InstalledAt = installedAt
	}

	stats.PostgresOK = database.PingPool(ctx, s.pool) == nil
	stats.RedisOK = database.PingRedis(ctx, s.redis) == nil

	companyRepo := models.NewCompanyRepository(s.pool)
	companyCount, err := companyRepo.Count(ctx)
	if err == nil {
		stats.TotalCompanies = companyCount
	}

	return stats, nil
}

func (s *Service) GetSystemConfig(ctx context.Context) (map[string]string, error) {
	return s.systemRepo.GetAll(ctx)
}

func (s *Service) UpdateSystemConfig(ctx context.Context, key, value string) error {
	return s.systemRepo.Set(ctx, key, value)
}

func (s *Service) ListUsers(ctx context.Context, offset, limit int) ([]models.User, int, error) {
	return s.userRepo.List(ctx, offset, limit)
}
