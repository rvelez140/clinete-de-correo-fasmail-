package auth

import (
	"fmt"
	"time"

	"github.com/fasmail/panel/internal/config"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

type Claims struct {
	jwt.RegisteredClaims
	UserID             uuid.UUID  `json:"uid"`
	Email              string     `json:"email"`
	Role               string     `json:"role"`
	MustChangePassword bool       `json:"mcp"`
	SessionID          uuid.UUID  `json:"sid"`
	CompanyID          *uuid.UUID `json:"cid,omitempty"`
}

type TokenPair struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
}

type JWTManager struct {
	secret          []byte
	accessTokenTTL  time.Duration
	refreshTokenTTL time.Duration
}

func NewJWTManager(cfg config.JWTConfig) *JWTManager {
	return &JWTManager{
		secret:          []byte(cfg.Secret),
		accessTokenTTL:  cfg.AccessTokenTTL,
		refreshTokenTTL: cfg.RefreshTokenTTL,
	}
}

func (m *JWTManager) GenerateTokenPair(userID uuid.UUID, email, role string, mustChangePassword bool, sessionID uuid.UUID, companyID *uuid.UUID) (*TokenPair, error) {
	now := time.Now()

	accessClaims := Claims{
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(now.Add(m.accessTokenTTL)),
			IssuedAt:  jwt.NewNumericDate(now),
			NotBefore: jwt.NewNumericDate(now),
			Issuer:    "fasmail-panel",
			ID:        uuid.New().String(),
		},
		UserID:             userID,
		Email:              email,
		Role:               role,
		MustChangePassword: mustChangePassword,
		SessionID:          sessionID,
		CompanyID:          companyID,
	}

	accessToken := jwt.NewWithClaims(jwt.SigningMethodHS256, accessClaims)
	accessStr, err := accessToken.SignedString(m.secret)
	if err != nil {
		return nil, fmt.Errorf("sign access token: %w", err)
	}

	refreshClaims := Claims{
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(now.Add(m.refreshTokenTTL)),
			IssuedAt:  jwt.NewNumericDate(now),
			NotBefore: jwt.NewNumericDate(now),
			Issuer:    "fasmail-panel",
			ID:        uuid.New().String(),
		},
		UserID:    userID,
		Email:     email,
		Role:      role,
		SessionID: sessionID,
		CompanyID: companyID,
	}

	refreshToken := jwt.NewWithClaims(jwt.SigningMethodHS256, refreshClaims)
	refreshStr, err := refreshToken.SignedString(m.secret)
	if err != nil {
		return nil, fmt.Errorf("sign refresh token: %w", err)
	}

	return &TokenPair{
		AccessToken:  accessStr,
		RefreshToken: refreshStr,
	}, nil
}

func (m *JWTManager) ValidateAccessToken(tokenStr string) (*Claims, error) {
	return m.validateToken(tokenStr)
}

func (m *JWTManager) ValidateRefreshToken(tokenStr string) (*Claims, error) {
	return m.validateToken(tokenStr)
}

func (m *JWTManager) validateToken(tokenStr string) (*Claims, error) {
	token, err := jwt.ParseWithClaims(tokenStr, &Claims{}, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return m.secret, nil
	})
	if err != nil {
		return nil, fmt.Errorf("parse token: %w", err)
	}

	claims, ok := token.Claims.(*Claims)
	if !ok || !token.Valid {
		return nil, fmt.Errorf("invalid token claims")
	}

	return claims, nil
}

func (m *JWTManager) GetRefreshTokenTTL() time.Duration {
	return m.refreshTokenTTL
}

func (m *JWTManager) GetAccessTokenTTL() time.Duration {
	return m.accessTokenTTL
}
