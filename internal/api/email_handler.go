package api

import (
	"net/http"

	"github.com/fasmail/panel/internal/auth"
	"github.com/fasmail/panel/internal/models"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

type EmailHandler struct {
	pool      *pgxpool.Pool
	jwtSecret string
}

func NewEmailHandler(pool *pgxpool.Pool, jwtSecret string) *EmailHandler {
	return &EmailHandler{pool: pool, jwtSecret: jwtSecret}
}

func (h *EmailHandler) ListAccounts(c *gin.Context) {
	userID := c.MustGet("user_id").(uuid.UUID)

	repo := models.NewEmailAccountRepository(h.pool)
	accounts, err := repo.ListByUserID(c.Request.Context(), userID)
	if err != nil {
		jsonError(c, http.StatusInternalServerError, "error al listar cuentas de correo")
		return
	}

	type accountResponse struct {
		ID           uuid.UUID `json:"id"`
		EmailAddress string    `json:"email_address"`
		DisplayName  string    `json:"display_name"`
		IMAPHost     string    `json:"imap_host"`
		IMAPPort     int       `json:"imap_port"`
		IMAPUseTLS   bool      `json:"imap_use_tls"`
		SMTPHost     string    `json:"smtp_host"`
		SMTPPort     int       `json:"smtp_port"`
		SMTPUseTLS   bool      `json:"smtp_use_tls"`
		Username     string    `json:"username"`
		IsDefault    bool      `json:"is_default"`
	}

	var response []accountResponse
	for _, a := range accounts {
		response = append(response, accountResponse{
			ID:           a.ID,
			EmailAddress: a.EmailAddress,
			DisplayName:  a.DisplayName,
			IMAPHost:     a.IMAPHost,
			IMAPPort:     a.IMAPPort,
			IMAPUseTLS:   a.IMAPUseTLS,
			SMTPHost:     a.SMTPHost,
			SMTPPort:     a.SMTPPort,
			SMTPUseTLS:   a.SMTPUseTLS,
			Username:     a.Username,
			IsDefault:    a.IsDefault,
		})
	}

	jsonOK(c, response)
}

func (h *EmailHandler) CreateAccount(c *gin.Context) {
	var req struct {
		EmailAddress string `json:"email_address" binding:"required"`
		DisplayName  string `json:"display_name"`
		IMAPHost     string `json:"imap_host" binding:"required"`
		IMAPPort     int    `json:"imap_port"`
		IMAPUseTLS   *bool  `json:"imap_use_tls"`
		SMTPHost     string `json:"smtp_host" binding:"required"`
		SMTPPort     int    `json:"smtp_port"`
		SMTPUseTLS   *bool  `json:"smtp_use_tls"`
		Username     string `json:"username" binding:"required"`
		Password     string `json:"password" binding:"required"`
		IsDefault    bool   `json:"is_default"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		jsonError(c, http.StatusBadRequest, "datos inválidos: email_address, imap_host, smtp_host, username y password son requeridos")
		return
	}

	userID := c.MustGet("user_id").(uuid.UUID)

	encrypted, err := auth.EncryptMailPassword(req.Password, h.jwtSecret)
	if err != nil {
		jsonError(c, http.StatusInternalServerError, "error de cifrado")
		return
	}

	imapPort := req.IMAPPort
	if imapPort == 0 {
		imapPort = 993
	}
	smtpPort := req.SMTPPort
	if smtpPort == 0 {
		smtpPort = 587
	}
	imapUseTLS := true
	if req.IMAPUseTLS != nil {
		imapUseTLS = *req.IMAPUseTLS
	}
	smtpUseTLS := true
	if req.SMTPUseTLS != nil {
		smtpUseTLS = *req.SMTPUseTLS
	}

	acct := &models.EmailAccount{
		UserID:            userID,
		EmailAddress:      req.EmailAddress,
		DisplayName:       req.DisplayName,
		IMAPHost:          req.IMAPHost,
		IMAPPort:          imapPort,
		IMAPUseTLS:        imapUseTLS,
		SMTPHost:          req.SMTPHost,
		SMTPPort:          smtpPort,
		SMTPUseTLS:        smtpUseTLS,
		Username:          req.Username,
		PasswordEncrypted: encrypted,
		IsDefault:         req.IsDefault,
	}

	repo := models.NewEmailAccountRepository(h.pool)
	if err := repo.Create(c.Request.Context(), acct); err != nil {
		jsonError(c, http.StatusInternalServerError, "error al crear cuenta de correo")
		return
	}

	jsonCreated(c, gin.H{
		"id":            acct.ID,
		"email_address": acct.EmailAddress,
		"display_name":  acct.DisplayName,
		"imap_host":     acct.IMAPHost,
		"imap_port":     acct.IMAPPort,
		"smtp_host":     acct.SMTPHost,
		"smtp_port":     acct.SMTPPort,
		"username":      acct.Username,
		"is_default":    acct.IsDefault,
	})
}

func (h *EmailHandler) GetAccount(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		jsonError(c, http.StatusBadRequest, "ID inválido")
		return
	}

	userID := c.MustGet("user_id").(uuid.UUID)

	repo := models.NewEmailAccountRepository(h.pool)
	acct, err := repo.GetByID(c.Request.Context(), id)
	if err != nil {
		jsonError(c, http.StatusNotFound, "cuenta de correo no encontrada")
		return
	}

	if acct.UserID != userID {
		jsonError(c, http.StatusForbidden, "acceso denegado")
		return
	}

	password, err := auth.DecryptMailPassword(acct.PasswordEncrypted, h.jwtSecret)
	if err != nil {
		jsonError(c, http.StatusInternalServerError, "error al descifrar credenciales")
		return
	}

	jsonOK(c, gin.H{
		"id":            acct.ID,
		"email_address": acct.EmailAddress,
		"display_name":  acct.DisplayName,
		"imap_host":     acct.IMAPHost,
		"imap_port":     acct.IMAPPort,
		"imap_use_tls":  acct.IMAPUseTLS,
		"smtp_host":     acct.SMTPHost,
		"smtp_port":     acct.SMTPPort,
		"smtp_use_tls":  acct.SMTPUseTLS,
		"username":      acct.Username,
		"password":      password,
		"is_default":    acct.IsDefault,
	})
}

func (h *EmailHandler) UpdateAccount(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		jsonError(c, http.StatusBadRequest, "ID inválido")
		return
	}

	userID := c.MustGet("user_id").(uuid.UUID)

	repo := models.NewEmailAccountRepository(h.pool)
	acct, err := repo.GetByID(c.Request.Context(), id)
	if err != nil {
		jsonError(c, http.StatusNotFound, "cuenta de correo no encontrada")
		return
	}

	if acct.UserID != userID {
		jsonError(c, http.StatusForbidden, "acceso denegado")
		return
	}

	var req struct {
		EmailAddress string `json:"email_address"`
		DisplayName  string `json:"display_name"`
		IMAPHost     string `json:"imap_host"`
		IMAPPort     int    `json:"imap_port"`
		IMAPUseTLS   *bool  `json:"imap_use_tls"`
		SMTPHost     string `json:"smtp_host"`
		SMTPPort     int    `json:"smtp_port"`
		SMTPUseTLS   *bool  `json:"smtp_use_tls"`
		Username     string `json:"username"`
		Password     string `json:"password"`
		IsDefault    *bool  `json:"is_default"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		jsonError(c, http.StatusBadRequest, "datos inválidos")
		return
	}

	if req.EmailAddress != "" {
		acct.EmailAddress = req.EmailAddress
	}
	if req.DisplayName != "" {
		acct.DisplayName = req.DisplayName
	}
	if req.IMAPHost != "" {
		acct.IMAPHost = req.IMAPHost
	}
	if req.IMAPPort > 0 {
		acct.IMAPPort = req.IMAPPort
	}
	if req.IMAPUseTLS != nil {
		acct.IMAPUseTLS = *req.IMAPUseTLS
	}
	if req.SMTPHost != "" {
		acct.SMTPHost = req.SMTPHost
	}
	if req.SMTPPort > 0 {
		acct.SMTPPort = req.SMTPPort
	}
	if req.SMTPUseTLS != nil {
		acct.SMTPUseTLS = *req.SMTPUseTLS
	}
	if req.Username != "" {
		acct.Username = req.Username
	}
	if req.Password != "" {
		encrypted, err := auth.EncryptMailPassword(req.Password, h.jwtSecret)
		if err != nil {
			jsonError(c, http.StatusInternalServerError, "error de cifrado")
			return
		}
		acct.PasswordEncrypted = encrypted
	}
	if req.IsDefault != nil {
		acct.IsDefault = *req.IsDefault
	}

	if err := repo.Update(c.Request.Context(), acct); err != nil {
		jsonError(c, http.StatusInternalServerError, "error al actualizar cuenta")
		return
	}

	jsonOK(c, gin.H{
		"id":            acct.ID,
		"email_address": acct.EmailAddress,
		"display_name":  acct.DisplayName,
		"imap_host":     acct.IMAPHost,
		"imap_port":     acct.IMAPPort,
		"smtp_host":     acct.SMTPHost,
		"smtp_port":     acct.SMTPPort,
		"username":      acct.Username,
		"is_default":    acct.IsDefault,
	})
}

func (h *EmailHandler) DeleteAccount(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		jsonError(c, http.StatusBadRequest, "ID inválido")
		return
	}

	userID := c.MustGet("user_id").(uuid.UUID)

	repo := models.NewEmailAccountRepository(h.pool)
	if err := repo.Delete(c.Request.Context(), id, userID); err != nil {
		jsonError(c, http.StatusNotFound, "cuenta de correo no encontrada")
		return
	}

	jsonOK(c, gin.H{"message": "cuenta de correo eliminada"})
}
