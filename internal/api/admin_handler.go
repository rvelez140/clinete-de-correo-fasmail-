package api

import (
	"net/http"
	"strconv"

	"github.com/fasmail/panel/internal/admin"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type AdminHandler struct {
	adminSvc   *admin.Service
	companySvc *admin.CompanyService
}

func NewAdminHandler(adminSvc *admin.Service, companySvc *admin.CompanyService) *AdminHandler {
	return &AdminHandler{
		adminSvc:   adminSvc,
		companySvc: companySvc,
	}
}

func (h *AdminHandler) Dashboard(c *gin.Context) {
	stats, err := h.adminSvc.GetDashboardStats(c.Request.Context())
	if err != nil {
		jsonError(c, http.StatusInternalServerError, "error al cargar estadísticas")
		return
	}
	jsonOK(c, stats)
}

func (h *AdminHandler) GetSettings(c *gin.Context) {
	configs, err := h.adminSvc.GetSystemConfig(c.Request.Context())
	if err != nil {
		jsonError(c, http.StatusInternalServerError, "error al cargar configuración")
		return
	}
	jsonOK(c, configs)
}

func (h *AdminHandler) UpdateSettings(c *gin.Context) {
	var req struct {
		Key   string `json:"key" binding:"required"`
		Value string `json:"value"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		jsonError(c, http.StatusBadRequest, "datos inválidos")
		return
	}

	if req.Key == "installed" || req.Key == "installed_at" {
		jsonError(c, http.StatusForbidden, "no se puede modificar esta configuración")
		return
	}

	if err := h.adminSvc.UpdateSystemConfig(c.Request.Context(), req.Key, req.Value); err != nil {
		jsonError(c, http.StatusInternalServerError, "error al actualizar: "+err.Error())
		return
	}

	jsonOK(c, gin.H{"message": "configuración actualizada"})
}

func (h *AdminHandler) ListUsers(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))

	if page < 1 {
		page = 1
	}
	if limit < 1 || limit > 100 {
		limit = 20
	}

	offset := (page - 1) * limit
	users, total, err := h.adminSvc.ListUsers(c.Request.Context(), offset, limit)
	if err != nil {
		jsonError(c, http.StatusInternalServerError, "error al listar usuarios")
		return
	}

	type userResponse struct {
		ID          uuid.UUID  `json:"id"`
		Email       string     `json:"email"`
		DisplayName string     `json:"display_name"`
		Role        string     `json:"role"`
		IsActive    bool       `json:"is_active"`
		CompanyID   *uuid.UUID `json:"company_id"`
	}

	var response []userResponse
	for _, u := range users {
		response = append(response, userResponse{
			ID:          u.ID,
			Email:       u.Email,
			DisplayName: u.DisplayName,
			Role:        u.Role,
			IsActive:    u.IsActive,
			CompanyID:   u.CompanyID,
		})
	}

	jsonPaginated(c, response, total, page, limit)
}

func (h *AdminHandler) ListCompanies(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))

	if page < 1 {
		page = 1
	}
	if limit < 1 || limit > 100 {
		limit = 20
	}

	offset := (page - 1) * limit
	companies, total, err := h.companySvc.ListCompanies(c.Request.Context(), offset, limit)
	if err != nil {
		jsonError(c, http.StatusInternalServerError, "error al listar empresas")
		return
	}

	jsonPaginated(c, companies, total, page, limit)
}

func (h *AdminHandler) CreateCompany(c *gin.Context) {
	var req struct {
		Name         string `json:"name" binding:"required"`
		Slug         string `json:"slug" binding:"required"`
		PrimaryColor string `json:"primary_color" binding:"required"`
		SuccessColor string `json:"success_color" binding:"required"`
		DangerColor  string `json:"danger_color" binding:"required"`
		WarningColor string `json:"warning_color" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		jsonError(c, http.StatusBadRequest, "datos inválidos")
		return
	}

	company, err := h.companySvc.CreateCompany(c.Request.Context(),
		req.Name, req.Slug, req.PrimaryColor, req.SuccessColor, req.DangerColor, req.WarningColor)
	if err != nil {
		jsonError(c, http.StatusBadRequest, err.Error())
		return
	}

	jsonCreated(c, company)
}

func (h *AdminHandler) GetCompany(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		jsonError(c, http.StatusBadRequest, "ID inválido")
		return
	}

	company, err := h.companySvc.GetCompany(c.Request.Context(), id)
	if err != nil {
		jsonError(c, http.StatusNotFound, "empresa no encontrada")
		return
	}

	jsonOK(c, company)
}

func (h *AdminHandler) UpdateCompany(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		jsonError(c, http.StatusBadRequest, "ID inválido")
		return
	}

	company, err := h.companySvc.GetCompany(c.Request.Context(), id)
	if err != nil {
		jsonError(c, http.StatusNotFound, "empresa no encontrada")
		return
	}

	var req struct {
		Name         string `json:"name"`
		Slug         string `json:"slug"`
		PrimaryColor string `json:"primary_color"`
		SuccessColor string `json:"success_color"`
		DangerColor  string `json:"danger_color"`
		WarningColor string `json:"warning_color"`
		IsActive     *bool  `json:"is_active"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		jsonError(c, http.StatusBadRequest, "datos inválidos")
		return
	}

	if req.Name != "" {
		company.Name = req.Name
	}
	if req.Slug != "" {
		company.Slug = req.Slug
	}
	if req.PrimaryColor != "" {
		company.PrimaryColor = req.PrimaryColor
	}
	if req.SuccessColor != "" {
		company.SuccessColor = req.SuccessColor
	}
	if req.DangerColor != "" {
		company.DangerColor = req.DangerColor
	}
	if req.WarningColor != "" {
		company.WarningColor = req.WarningColor
	}
	if req.IsActive != nil {
		company.IsActive = *req.IsActive
	}

	if err := h.companySvc.UpdateCompany(c.Request.Context(), company); err != nil {
		jsonError(c, http.StatusBadRequest, err.Error())
		return
	}

	jsonOK(c, company)
}

func (h *AdminHandler) DeleteCompany(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		jsonError(c, http.StatusBadRequest, "ID inválido")
		return
	}

	if err := h.companySvc.DeleteCompany(c.Request.Context(), id); err != nil {
		jsonError(c, http.StatusBadRequest, err.Error())
		return
	}

	jsonOK(c, gin.H{"message": "empresa eliminada"})
}

