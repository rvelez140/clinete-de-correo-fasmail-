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
	branding, _ := c.Get("branding")
	companySlug := c.Query("company")

	c.HTML(http.StatusOK, "login", gin.H{
		"Title":       "Iniciar Sesion",
		"Branding":    branding,
		"CompanySlug": companySlug,
	})
}

func (h *Handler) HandleLogin(c *gin.Context) {
	email := c.PostForm("email")
	password := c.PostForm("password")
	branding, _ := c.Get("branding")
	companySlug := c.Query("company")

	if email == "" || password == "" {
		c.HTML(http.StatusBadRequest, "login", gin.H{
			"Title":       "Iniciar Sesion",
			"Error":       "Email y contraseña son requeridos",
			"Email":       email,
			"Branding":    branding,
			"CompanySlug": companySlug,
		})
		return
	}

	user, err := h.service.Authenticate(c.Request.Context(), email, password)
	if err != nil {
		c.HTML(http.StatusUnauthorized, "login", gin.H{
			"Title":       "Iniciar Sesion",
			"Error":       "Credenciales invalidas",
			"Email":       email,
			"Branding":    branding,
			"CompanySlug": companySlug,
		})
		return
	}

	userAgent := c.Request.UserAgent()
	ip := c.ClientIP()

	tokens, err := h.service.CreateSession(c.Request.Context(), user, userAgent, ip)
	if err != nil {
		c.HTML(http.StatusInternalServerError, "login", gin.H{
			"Title":       "Iniciar Sesion",
			"Error":       "Error al crear la sesion. Intente de nuevo.",
			"Email":       email,
			"Branding":    branding,
			"CompanySlug": companySlug,
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
	branding, _ := c.Get("branding")

	c.HTML(http.StatusOK, "force_password_change", gin.H{
		"Title":    "Cambiar Contrasena",
		"Branding": branding,
	})
}

func (h *Handler) HandleChangePassword(c *gin.Context) {
	currentPassword := c.PostForm("current_password")
	newPassword := c.PostForm("new_password")
	confirmPassword := c.PostForm("confirm_password")
	branding, _ := c.Get("branding")

	if currentPassword == "" || newPassword == "" || confirmPassword == "" {
		c.HTML(http.StatusBadRequest, "force_password_change", gin.H{
			"Title":    "Cambiar Contrasena",
			"Error":    "Todos los campos son requeridos",
			"Branding": branding,
		})
		return
	}

	if newPassword != confirmPassword {
		c.HTML(http.StatusBadRequest, "force_password_change", gin.H{
			"Title":    "Cambiar Contrasena",
			"Error":    "Las contrasenas no coinciden",
			"Branding": branding,
		})
		return
	}

	userIDVal, _ := c.Get("user_id")
	userID := userIDVal.(uuid.UUID)

	if err := h.service.ChangePassword(c.Request.Context(), userID, currentPassword, newPassword); err != nil {
		c.HTML(http.StatusBadRequest, "force_password_change", gin.H{
			"Title":    "Cambiar Contrasena",
			"Error":    err.Error(),
			"Branding": branding,
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
