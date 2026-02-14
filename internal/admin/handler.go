package admin

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

type Handler struct {
	service *Service
}

func NewHandler(service *Service) *Handler {
	return &Handler{service: service}
}

func (h *Handler) ShowDashboard(c *gin.Context) {
	stats, err := h.service.GetDashboardStats(c.Request.Context())
	if err != nil {
		c.HTML(http.StatusInternalServerError, "dashboard", gin.H{
			"Title": "Dashboard",
			"Error": "Error al cargar estadísticas",
		})
		return
	}

	email, _ := c.Get("email")
	role, _ := c.Get("role")
	branding, _ := c.Get("branding")

	c.HTML(http.StatusOK, "dashboard", gin.H{
		"Title":     "Dashboard",
		"Stats":     stats,
		"UserEmail": email,
		"UserRole":  role,
		"Branding":  branding,
	})
}

func (h *Handler) ShowSettings(c *gin.Context) {
	configs, err := h.service.GetSystemConfig(c.Request.Context())
	if err != nil {
		c.HTML(http.StatusInternalServerError, "settings", gin.H{
			"Title": "Configuración",
			"Error": "Error al cargar la configuración",
		})
		return
	}

	email, _ := c.Get("email")
	role, _ := c.Get("role")
	branding, _ := c.Get("branding")

	c.HTML(http.StatusOK, "settings", gin.H{
		"Title":     "Configuración del Sistema",
		"Configs":   configs,
		"UserEmail": email,
		"UserRole":  role,
		"Branding":  branding,
	})
}

func (h *Handler) HandleSettings(c *gin.Context) {
	key := c.PostForm("key")
	value := c.PostForm("value")

	if key == "" {
		c.Redirect(http.StatusFound, "/admin/settings")
		return
	}

	if key == "installed" || key == "installed_at" {
		c.HTML(http.StatusForbidden, "settings", gin.H{
			"Title": "Configuración del Sistema",
			"Error": "No se puede modificar esta configuración",
		})
		return
	}

	if err := h.service.UpdateSystemConfig(c.Request.Context(), key, value); err != nil {
		c.HTML(http.StatusInternalServerError, "settings", gin.H{
			"Title": "Configuración del Sistema",
			"Error": "Error al actualizar: " + err.Error(),
		})
		return
	}

	c.Redirect(http.StatusFound, "/admin/settings")
}
