package auth

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type Handler struct {
	service *Service
}

func NewHandler(service *Service) *Handler {
	return &Handler{service: service}
}

func (h *Handler) ShowLoginPage(c *gin.Context) {
	c.HTML(http.StatusOK, "login", gin.H{
		"Title": "Iniciar Sesión",
	})
}

func (h *Handler) HandleLogin(c *gin.Context) {
	email := c.PostForm("email")
	password := c.PostForm("password")

	if email == "" || password == "" {
		c.HTML(http.StatusBadRequest, "login", gin.H{
			"Title": "Iniciar Sesión",
			"Error": "Email y contraseña son requeridos",
			"Email": email,
		})
		return
	}

	user, err := h.service.Authenticate(c.Request.Context(), email, password)
	if err != nil {
		c.HTML(http.StatusUnauthorized, "login", gin.H{
			"Title": "Iniciar Sesión",
			"Error": "Credenciales inválidas",
			"Email": email,
		})
		return
	}

	userAgent := c.Request.UserAgent()
	ip := c.ClientIP()

	tokens, err := h.service.CreateSession(c.Request.Context(), user, userAgent, ip)
	if err != nil {
		c.HTML(http.StatusInternalServerError, "login", gin.H{
			"Title": "Iniciar Sesión",
			"Error": "Error al crear la sesión. Intente de nuevo.",
			"Email": email,
		})
		return
	}

	c.SetCookie("access_token", tokens.AccessToken, int(h.service.jwtMgr.GetAccessTokenTTL().Seconds()), "/", "", false, true)
	c.SetCookie("refresh_token", tokens.RefreshToken, int(h.service.jwtMgr.GetRefreshTokenTTL().Seconds()), "/", "", false, true)

	if user.MustChangePassword {
		c.Redirect(http.StatusFound, "/auth/change-password")
		return
	}

	c.Redirect(http.StatusFound, "/admin/")
}

func (h *Handler) HandleLogout(c *gin.Context) {
	sessionIDVal, exists := c.Get("session_id")
	if exists {
		sessionID := sessionIDVal.(uuid.UUID)
		userIDVal, _ := c.Get("user_id")
		userID := userIDVal.(uuid.UUID)
		h.service.Logout(c.Request.Context(), sessionID, userID)
	}

	c.SetCookie("access_token", "", -1, "/", "", false, true)
	c.SetCookie("refresh_token", "", -1, "/", "", false, true)

	c.Redirect(http.StatusFound, "/auth/login")
}

func (h *Handler) ShowChangePassword(c *gin.Context) {
	c.HTML(http.StatusOK, "force_password_change", gin.H{
		"Title": "Cambiar Contraseña",
	})
}

func (h *Handler) HandleChangePassword(c *gin.Context) {
	currentPassword := c.PostForm("current_password")
	newPassword := c.PostForm("new_password")
	confirmPassword := c.PostForm("confirm_password")

	if currentPassword == "" || newPassword == "" || confirmPassword == "" {
		c.HTML(http.StatusBadRequest, "force_password_change", gin.H{
			"Title": "Cambiar Contraseña",
			"Error": "Todos los campos son requeridos",
		})
		return
	}

	if newPassword != confirmPassword {
		c.HTML(http.StatusBadRequest, "force_password_change", gin.H{
			"Title": "Cambiar Contraseña",
			"Error": "Las contraseñas no coinciden",
		})
		return
	}

	userIDVal, _ := c.Get("user_id")
	userID := userIDVal.(uuid.UUID)

	if err := h.service.ChangePassword(c.Request.Context(), userID, currentPassword, newPassword); err != nil {
		c.HTML(http.StatusBadRequest, "force_password_change", gin.H{
			"Title": "Cambiar Contraseña",
			"Error": err.Error(),
		})
		return
	}

	// Re-authenticate with new password to get fresh tokens
	user, err := h.service.Authenticate(c.Request.Context(), c.GetString("email"), newPassword)
	if err != nil {
		c.Redirect(http.StatusFound, "/auth/login")
		return
	}

	tokens, err := h.service.CreateSession(c.Request.Context(), user, c.Request.UserAgent(), c.ClientIP())
	if err != nil {
		c.Redirect(http.StatusFound, "/auth/login")
		return
	}

	c.SetCookie("access_token", tokens.AccessToken, int(h.service.jwtMgr.GetAccessTokenTTL().Seconds()), "/", "", false, true)
	c.SetCookie("refresh_token", tokens.RefreshToken, int(h.service.jwtMgr.GetRefreshTokenTTL().Seconds()), "/", "", false, true)

	c.Redirect(http.StatusFound, "/admin/")
}
