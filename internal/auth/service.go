package auth

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"time"

	"github.com/fasmail/panel/internal/database"
	"github.com/fasmail/panel/internal/models"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/redis/go-redis/v9"
)

const (
	DefaultAdminEmail    = "admin@fasmail.local"
	DefaultAdminPassword = "admin123"
	DefaultAdminName     = "Administrador"
)

type Service struct {
	pool        **pgxpool.Pool
	redisClient **redis.Client
	jwtMgr      *JWTManager
}

func NewService(pool **pgxpool.Pool, redisClient **redis.Client, jwtMgr *JWTManager) *Service {
	return &Service{
		pool:        pool,
		redisClient: redisClient,
		jwtMgr:      jwtMgr,
	}
}

func (s *Service) getUserRepo() *models.UserRepository {
	if s.pool == nil || *s.pool == nil {
		return nil
	}
	return models.NewUserRepository(*s.pool)
}

func (s *Service) getSessionRepo() *models.SessionRepository {
	if s.pool == nil || *s.pool == nil {
		return nil
	}
	return models.NewSessionRepository(*s.pool)
}

func (s *Service) getRedis() *redis.Client {
	if s.redisClient == nil {
		return nil
	}
	return *s.redisClient
}

func (s *Service) Authenticate(ctx context.Context, email, password string) (*models.User, error) {
	userRepo := s.getUserRepo()
	if userRepo == nil {
		return nil, fmt.Errorf("base de datos no disponible")
	}

	user, err := userRepo.GetByEmail(ctx, email)
	if err != nil {
		return nil, fmt.Errorf("credenciales inválidas")
	}

	if !user.IsActive {
		return nil, fmt.Errorf("cuenta desactivada")
	}

	if !CheckPassword(password, user.PasswordHash) {
		return nil, fmt.Errorf("credenciales inválidas")
	}

	if err := userRepo.UpdateLastLogin(ctx, user.ID); err != nil {
		return nil, fmt.Errorf("update last login: %w", err)
	}

	return user, nil
}

func (s *Service) CreateSession(ctx context.Context, user *models.User, userAgent, ip string) (*TokenPair, error) {
	sessionRepo := s.getSessionRepo()
	if sessionRepo == nil {
		return nil, fmt.Errorf("base de datos no disponible")
	}

	sessionID := uuid.New()

	tokens, err := s.jwtMgr.GenerateTokenPair(
		user.ID, user.Email, user.Role,
		user.MustChangePassword, sessionID,
	)
	if err != nil {
		return nil, fmt.Errorf("generate tokens: %w", err)
	}

	refreshHash := hashToken(tokens.RefreshToken)

	session := &models.Session{
		UserID:           user.ID,
		RefreshTokenHash: refreshHash,
		UserAgent:        userAgent,
		IPAddress:        ip,
		ExpiresAt:        time.Now().Add(s.jwtMgr.GetRefreshTokenTTL()),
	}

	if err := sessionRepo.Create(ctx, session); err != nil {
		return nil, fmt.Errorf("create session: %w", err)
	}

	rc := s.getRedis()
	if rc != nil {
		if err := database.StoreSession(ctx, rc, sessionID.String(), user.ID, user.Role, s.jwtMgr.GetAccessTokenTTL()); err != nil {
			return nil, fmt.Errorf("store redis session: %w", err)
		}
	}

	return tokens, nil
}

func (s *Service) Logout(ctx context.Context, sessionID, userID uuid.UUID) error {
	rc := s.getRedis()
	if rc != nil {
		if err := database.DeleteSession(ctx, rc, sessionID.String(), userID); err != nil {
			return fmt.Errorf("delete redis session: %w", err)
		}
	}

	sessionRepo := s.getSessionRepo()
	if sessionRepo != nil {
		if err := sessionRepo.DeleteByID(ctx, sessionID); err != nil {
			return fmt.Errorf("delete db session: %w", err)
		}
	}

	return nil
}

func (s *Service) ChangePassword(ctx context.Context, userID uuid.UUID, oldPassword, newPassword string) error {
	userRepo := s.getUserRepo()
	if userRepo == nil {
		return fmt.Errorf("base de datos no disponible")
	}

	user, err := userRepo.GetByID(ctx, userID)
	if err != nil {
		return fmt.Errorf("usuario no encontrado")
	}

	if !CheckPassword(oldPassword, user.PasswordHash) {
		return fmt.Errorf("contraseña actual incorrecta")
	}

	if oldPassword == newPassword {
		return fmt.Errorf("la nueva contraseña debe ser diferente a la actual")
	}

	if err := ValidatePasswordStrength(newPassword); err != nil {
		return err
	}

	hash, err := HashPassword(newPassword)
	if err != nil {
		return fmt.Errorf("hash password: %w", err)
	}

	if err := userRepo.UpdatePassword(ctx, userID, hash); err != nil {
		return fmt.Errorf("update password: %w", err)
	}

	rc := s.getRedis()
	if rc != nil {
		if err := database.DeleteAllUserSessions(ctx, rc, userID); err != nil {
			return fmt.Errorf("invalidate sessions: %w", err)
		}
	}

	sessionRepo := s.getSessionRepo()
	if sessionRepo != nil {
		if err := sessionRepo.DeleteByUserID(ctx, userID); err != nil {
			return fmt.Errorf("delete db sessions: %w", err)
		}
	}

	return nil
}

func (s *Service) CreateDefaultAdmin(ctx context.Context) error {
	userRepo := s.getUserRepo()
	if userRepo == nil {
		return fmt.Errorf("base de datos no disponible")
	}

	exists, err := userRepo.EmailExists(ctx, DefaultAdminEmail)
	if err != nil {
		return fmt.Errorf("check admin exists: %w", err)
	}
	if exists {
		return nil
	}

	hash, err := HashPassword(DefaultAdminPassword)
	if err != nil {
		return fmt.Errorf("hash default password: %w", err)
	}

	admin := &models.User{
		Email:              DefaultAdminEmail,
		PasswordHash:       hash,
		DisplayName:        DefaultAdminName,
		Role:               "admin",
		MustChangePassword: true,
		IsActive:           true,
	}

	if err := userRepo.Create(ctx, admin); err != nil {
		return fmt.Errorf("create admin: %w", err)
	}

	return nil
}

func hashToken(token string) string {
	h := sha256.Sum256([]byte(token))
	return hex.EncodeToString(h[:])
}
