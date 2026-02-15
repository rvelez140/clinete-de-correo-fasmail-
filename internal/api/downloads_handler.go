package api

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"time"

	"github.com/fasmail/panel/internal/models"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

type DownloadsHandler struct {
	pool *pgxpool.Pool
}

func NewDownloadsHandler(pool *pgxpool.Pool) *DownloadsHandler {
	return &DownloadsHandler{pool: pool}
}

func (h *DownloadsHandler) GenerateToken(c *gin.Context) {
	var req struct {
		Platform  string `json:"platform" binding:"required"`
		ServerURL string `json:"server_url" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		jsonError(c, http.StatusBadRequest, "plataforma y server_url son requeridos")
		return
	}

	if req.Platform != "android" && req.Platform != "windows" && req.Platform != "linux" {
		jsonError(c, http.StatusBadRequest, "plataforma inválida (android, windows, linux)")
		return
	}

	userID := c.MustGet("user_id").(uuid.UUID)

	rawBytes := make([]byte, 32)
	if _, err := rand.Read(rawBytes); err != nil {
		jsonError(c, http.StatusInternalServerError, "error al generar token")
		return
	}
	rawToken := hex.EncodeToString(rawBytes)

	hash := sha256.Sum256([]byte(rawToken))
	tokenHash := hex.EncodeToString(hash[:])

	repo := models.NewAppTokenRepository(h.pool)
	appToken := &models.AppToken{
		UserID:    userID,
		TokenHash: tokenHash,
		ServerURL: req.ServerURL,
		Platform:  req.Platform,
		ExpiresAt: time.Now().Add(24 * time.Hour),
	}

	if err := repo.Create(c.Request.Context(), appToken); err != nil {
		jsonError(c, http.StatusInternalServerError, "error al generar token")
		return
	}

	setupURL := fmt.Sprintf("fasmail://setup?token=%s&server=%s", rawToken, req.ServerURL)

	jsonCreated(c, gin.H{
		"token":      rawToken,
		"setup_url":  setupURL,
		"expires_at": appToken.ExpiresAt,
	})
}

func (h *DownloadsHandler) ValidateToken(c *gin.Context) {
	rawToken := c.Query("token")
	if rawToken == "" {
		jsonError(c, http.StatusBadRequest, "token requerido")
		return
	}

	hash := sha256.Sum256([]byte(rawToken))
	tokenHash := hex.EncodeToString(hash[:])

	repo := models.NewAppTokenRepository(h.pool)
	appToken, err := repo.GetByTokenHash(c.Request.Context(), tokenHash)
	if err != nil {
		jsonError(c, http.StatusNotFound, "token inválido")
		return
	}

	if appToken.IsUsed {
		jsonError(c, http.StatusGone, "token ya utilizado")
		return
	}

	if time.Now().After(appToken.ExpiresAt) {
		jsonError(c, http.StatusGone, "token expirado")
		return
	}

	if err := repo.MarkUsed(c.Request.Context(), appToken.ID); err != nil {
		jsonError(c, http.StatusInternalServerError, "error al procesar token")
		return
	}

	userRepo := models.NewUserRepository(h.pool)
	user, err := userRepo.GetByID(c.Request.Context(), appToken.UserID)
	if err != nil {
		jsonError(c, http.StatusInternalServerError, "error al obtener usuario")
		return
	}

	emailRepo := models.NewEmailAccountRepository(h.pool)
	accounts, _ := emailRepo.ListByUserID(c.Request.Context(), user.ID)

	type emailAccountInfo struct {
		EmailAddress string `json:"email_address"`
		DisplayName  string `json:"display_name"`
		IMAPHost     string `json:"imap_host"`
		IMAPPort     int    `json:"imap_port"`
		IMAPUseTLS   bool   `json:"imap_use_tls"`
		SMTPHost     string `json:"smtp_host"`
		SMTPPort     int    `json:"smtp_port"`
		SMTPUseTLS   bool   `json:"smtp_use_tls"`
		Username     string `json:"username"`
		IsDefault    bool   `json:"is_default"`
	}

	var accountList []emailAccountInfo
	for _, a := range accounts {
		accountList = append(accountList, emailAccountInfo{
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

	jsonOK(c, gin.H{
		"user_id":        user.ID,
		"email":          user.Email,
		"display_name":   user.DisplayName,
		"role":           user.Role,
		"server_url":     appToken.ServerURL,
		"company_id":     user.CompanyID,
		"email_accounts": accountList,
	})
}

func (h *DownloadsHandler) ServeApp(c *gin.Context) {
	filename := c.Param("filename")

	allowedFiles := map[string]string{
		"fasmail.apk":        "application/vnd.android.package-archive",
		"fasmail-setup.exe":  "application/octet-stream",
		"fasmail-linux":      "application/octet-stream",
		"fasmail-linux.deb":  "application/vnd.debian.binary-package",
		"fasmail-linux.AppImage": "application/octet-stream",
	}

	contentType, ok := allowedFiles[filename]
	if !ok {
		jsonError(c, http.StatusNotFound, "archivo no encontrado")
		return
	}

	filePath := filepath.Join("/data/apps", filename)
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		jsonError(c, http.StatusNotFound, "archivo no disponible aún")
		return
	}

	c.Header("Content-Type", contentType)
	c.Header("Content-Disposition", fmt.Sprintf("attachment; filename=%s", filename))
	c.File(filePath)
}
