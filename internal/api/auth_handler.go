package api

import (
	"net/http"

	"github.com/fasmail/panel/internal/auth"
	"github.com/fasmail/panel/internal/models"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

type AuthHandler struct {
	authSvc *auth.Service
	jwtMgr  *auth.JWTManager
	pool    **pgxpool.Pool
}

func NewAuthHandler(authSvc *auth.Service, jwtMgr *auth.JWTManager, pool **pgxpool.Pool) *AuthHandler {
	return &AuthHandler{
		authSvc: authSvc,
		jwtMgr:  jwtMgr,
		pool:    pool,
	}
}

func (h *AuthHandler) Login(c *gin.Context) {
	var req struct {
		Email    string `json:"email" binding:"required"`
		Password string `json:"password" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		jsonError(c, http.StatusBadRequest, "email y contraseña son requeridos")
		return
	}

	user, err := h.authSvc.Authenticate(c.Request.Context(), req.Email, req.Password)
	if err != nil {
		jsonError(c, http.StatusUnauthorized, "credenciales inválidas")
		return
	}

	tokens, err := h.authSvc.CreateSession(c.Request.Context(), user, c.Request.UserAgent(), c.ClientIP())
	if err != nil {
		jsonError(c, http.StatusInternalServerError, "error al crear sesión")
		return
	}

	jsonOK(c, gin.H{
		"access_token":  tokens.AccessToken,
		"refresh_token": tokens.RefreshToken,
		"user": gin.H{
			"id":                   user.ID,
			"email":                user.Email,
			"display_name":         user.DisplayName,
			"role":                 user.Role,
			"must_change_password": user.MustChangePassword,
			"company_id":           user.CompanyID,
		},
	})
}

func (h *AuthHandler) Refresh(c *gin.Context) {
	var req struct {
		RefreshToken string `json:"refresh_token" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		jsonError(c, http.StatusBadRequest, "refresh_token es requerido")
		return
	}

	claims, err := h.jwtMgr.ValidateRefreshToken(req.RefreshToken)
	if err != nil {
		jsonError(c, http.StatusUnauthorized, "refresh token inválido")
		return
	}

	if h.pool == nil || *h.pool == nil {
		jsonError(c, http.StatusInternalServerError, "base de datos no disponible")
		return
	}

	userRepo := models.NewUserRepository(*h.pool)
	user, err := userRepo.GetByID(c.Request.Context(), claims.UserID)
	if err != nil {
		jsonError(c, http.StatusUnauthorized, "usuario no encontrado")
		return
	}

	if !user.IsActive {
		jsonError(c, http.StatusUnauthorized, "cuenta desactivada")
		return
	}

	tokens, err := h.authSvc.CreateSession(c.Request.Context(), user, c.Request.UserAgent(), c.ClientIP())
	if err != nil {
		jsonError(c, http.StatusInternalServerError, "error al crear sesión")
		return
	}

	jsonOK(c, gin.H{
		"access_token":  tokens.AccessToken,
		"refresh_token": tokens.RefreshToken,
	})
}

func (h *AuthHandler) Logout(c *gin.Context) {
	sessionIDVal, exists := c.Get("session_id")
	if !exists {
		jsonError(c, http.StatusBadRequest, "sesión no encontrada")
		return
	}

	sessionID := sessionIDVal.(uuid.UUID)
	userIDVal, _ := c.Get("user_id")
	userID := userIDVal.(uuid.UUID)

	h.authSvc.Logout(c.Request.Context(), sessionID, userID)
	jsonOK(c, gin.H{"message": "sesión cerrada"})
}

func (h *AuthHandler) Profile(c *gin.Context) {
	userIDVal, _ := c.Get("user_id")
	userID := userIDVal.(uuid.UUID)

	if h.pool == nil || *h.pool == nil {
		jsonError(c, http.StatusInternalServerError, "base de datos no disponible")
		return
	}

	userRepo := models.NewUserRepository(*h.pool)
	user, err := userRepo.GetByID(c.Request.Context(), userID)
	if err != nil {
		jsonError(c, http.StatusNotFound, "usuario no encontrado")
		return
	}

	jsonOK(c, gin.H{
		"id":                   user.ID,
		"email":                user.Email,
		"display_name":         user.DisplayName,
		"role":                 user.Role,
		"must_change_password": user.MustChangePassword,
		"is_active":            user.IsActive,
		"company_id":           user.CompanyID,
		"last_login_at":        user.LastLoginAt,
		"created_at":           user.CreatedAt,
	})
}
