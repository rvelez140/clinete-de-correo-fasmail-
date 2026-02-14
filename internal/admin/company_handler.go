package admin

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"

	"github.com/fasmail/panel/internal/models"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type CompanyHandler struct {
	service *CompanyService
}

func NewCompanyHandler(service *CompanyService) *CompanyHandler {
	return &CompanyHandler{service: service}
}

func (h *CompanyHandler) ShowCompanies(c *gin.Context) {
	companies, total, err := h.service.ListCompanies(c.Request.Context(), 0, 100)
	if err != nil {
		c.HTML(http.StatusInternalServerError, "companies", gin.H{
			"Title": "Empresas",
			"Error": "Error al cargar empresas",
		})
		return
	}

	email, _ := c.Get("email")
	role, _ := c.Get("role")
	branding, _ := c.Get("branding")

	c.HTML(http.StatusOK, "companies", gin.H{
		"Title":     "Empresas",
		"Companies": companies,
		"Total":     total,
		"UserEmail": email,
		"UserRole":  role,
		"Branding":  branding,
	})
}

func (h *CompanyHandler) ShowCreateCompany(c *gin.Context) {
	email, _ := c.Get("email")
	role, _ := c.Get("role")
	branding, _ := c.Get("branding")

	c.HTML(http.StatusOK, "company_form", gin.H{
		"Title":    "Nueva Empresa",
		"IsNew":    true,
		"Company": models.Company{
			PrimaryColor: "#2563eb",
			SuccessColor: "#16a34a",
			DangerColor:  "#dc2626",
			WarningColor: "#d97706",
			IsActive:     true,
		},
		"UserEmail": email,
		"UserRole":  role,
		"Branding":  branding,
	})
}

func (h *CompanyHandler) HandleCreateCompany(c *gin.Context) {
	name := c.PostForm("name")
	slug := c.PostForm("slug")
	primaryColor := c.PostForm("primary_color")
	successColor := c.PostForm("success_color")
	dangerColor := c.PostForm("danger_color")
	warningColor := c.PostForm("warning_color")

	company, err := h.service.CreateCompany(c.Request.Context(), name, slug, primaryColor, successColor, dangerColor, warningColor)
	if err != nil {
		email, _ := c.Get("email")
		role, _ := c.Get("role")
		branding, _ := c.Get("branding")

		c.HTML(http.StatusBadRequest, "company_form", gin.H{
			"Title": "Nueva Empresa",
			"IsNew": true,
			"Company": models.Company{
				Name:         name,
				Slug:         slug,
				PrimaryColor: primaryColor,
				SuccessColor: successColor,
				DangerColor:  dangerColor,
				WarningColor: warningColor,
			},
			"Error":     err.Error(),
			"UserEmail": email,
			"UserRole":  role,
			"Branding":  branding,
		})
		return
	}

	c.Redirect(http.StatusFound, fmt.Sprintf("/admin/companies/%s/edit", company.ID))
}

func (h *CompanyHandler) ShowEditCompany(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.Redirect(http.StatusFound, "/admin/companies")
		return
	}

	company, err := h.service.GetCompany(c.Request.Context(), id)
	if err != nil {
		c.Redirect(http.StatusFound, "/admin/companies")
		return
	}

	email, _ := c.Get("email")
	role, _ := c.Get("role")
	branding, _ := c.Get("branding")

	c.HTML(http.StatusOK, "company_form", gin.H{
		"Title":     "Editar Empresa",
		"IsNew":     false,
		"Company":   company,
		"UserEmail": email,
		"UserRole":  role,
		"Branding":  branding,
	})
}

func (h *CompanyHandler) HandleEditCompany(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.Redirect(http.StatusFound, "/admin/companies")
		return
	}

	company, err := h.service.GetCompany(c.Request.Context(), id)
	if err != nil {
		c.Redirect(http.StatusFound, "/admin/companies")
		return
	}

	company.Name = c.PostForm("name")
	company.Slug = c.PostForm("slug")
	company.PrimaryColor = c.PostForm("primary_color")
	company.SuccessColor = c.PostForm("success_color")
	company.DangerColor = c.PostForm("danger_color")
	company.WarningColor = c.PostForm("warning_color")
	company.IsActive = c.PostForm("is_active") == "on"

	if err := h.service.UpdateCompany(c.Request.Context(), company); err != nil {
		email, _ := c.Get("email")
		role, _ := c.Get("role")
		branding, _ := c.Get("branding")

		c.HTML(http.StatusBadRequest, "company_form", gin.H{
			"Title":     "Editar Empresa",
			"IsNew":     false,
			"Company":   company,
			"Error":     err.Error(),
			"UserEmail": email,
			"UserRole":  role,
			"Branding":  branding,
		})
		return
	}

	c.Redirect(http.StatusFound, fmt.Sprintf("/admin/companies/%s/edit", id))
}

func (h *CompanyHandler) HandleUploadLogo(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.Redirect(http.StatusFound, "/admin/companies")
		return
	}

	file, header, err := c.Request.FormFile("logo")
	if err != nil {
		c.Redirect(http.StatusFound, fmt.Sprintf("/admin/companies/%s/edit", id))
		return
	}
	defer file.Close()

	// Validate file size (max 2MB)
	if header.Size > 2*1024*1024 {
		email, _ := c.Get("email")
		role, _ := c.Get("role")
		branding, _ := c.Get("branding")
		company, _ := h.service.GetCompany(c.Request.Context(), id)

		c.HTML(http.StatusBadRequest, "company_form", gin.H{
			"Title":     "Editar Empresa",
			"IsNew":     false,
			"Company":   company,
			"Error":     "El logo no debe superar 2MB",
			"UserEmail": email,
			"UserRole":  role,
			"Branding":  branding,
		})
		return
	}

	// Validate MIME type
	contentType := header.Header.Get("Content-Type")
	allowedTypes := map[string]string{
		"image/png":     ".png",
		"image/jpeg":    ".jpg",
		"image/svg+xml": ".svg",
	}

	ext, ok := allowedTypes[contentType]
	if !ok {
		email, _ := c.Get("email")
		role, _ := c.Get("role")
		branding, _ := c.Get("branding")
		company, _ := h.service.GetCompany(c.Request.Context(), id)

		c.HTML(http.StatusBadRequest, "company_form", gin.H{
			"Title":     "Editar Empresa",
			"IsNew":     false,
			"Company":   company,
			"Error":     "Formato no soportado. Use PNG, JPG o SVG",
			"UserEmail": email,
			"UserRole":  role,
			"Branding":  branding,
		})
		return
	}

	// Ensure logos directory exists
	logosDir := "/data/logos"
	if err := os.MkdirAll(logosDir, 0755); err != nil {
		c.Redirect(http.StatusFound, fmt.Sprintf("/admin/companies/%s/edit", id))
		return
	}

	// Save file
	filename := id.String() + ext
	destPath := filepath.Join(logosDir, filename)

	dst, err := os.Create(destPath)
	if err != nil {
		c.Redirect(http.StatusFound, fmt.Sprintf("/admin/companies/%s/edit", id))
		return
	}
	defer dst.Close()

	if _, err := io.Copy(dst, file); err != nil {
		c.Redirect(http.StatusFound, fmt.Sprintf("/admin/companies/%s/edit", id))
		return
	}

	// Update company logo path
	logoPath := "logos/" + filename
	if err := h.service.UpdateCompanyLogo(c.Request.Context(), id, logoPath); err != nil {
		c.Redirect(http.StatusFound, fmt.Sprintf("/admin/companies/%s/edit", id))
		return
	}

	c.Redirect(http.StatusFound, fmt.Sprintf("/admin/companies/%s/edit", id))
}

func (h *CompanyHandler) HandleDeleteCompany(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.Redirect(http.StatusFound, "/admin/companies")
		return
	}

	if err := h.service.DeleteCompany(c.Request.Context(), id); err != nil {
		companies, total, _ := h.service.ListCompanies(c.Request.Context(), 0, 100)
		email, _ := c.Get("email")
		role, _ := c.Get("role")
		branding, _ := c.Get("branding")

		c.HTML(http.StatusBadRequest, "companies", gin.H{
			"Title":     "Empresas",
			"Companies": companies,
			"Total":     total,
			"Error":     err.Error(),
			"UserEmail": email,
			"UserRole":  role,
			"Branding":  branding,
		})
		return
	}

	c.Redirect(http.StatusFound, "/admin/companies")
}
