package admin

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"net/http"
	"os"
	"time"

	"github.com/fasmail/panel/internal/models"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

type DownloadsHandler struct {
	pool      *pgxpool.Pool
	serverURL string
}

func NewDownloadsHandler(pool *pgxpool.Pool, serverURL string) *DownloadsHandler {
	return &DownloadsHandler{pool: pool, serverURL: serverURL}
}

func (h *DownloadsHandler) ShowDownloads(c *gin.Context) {
	email, _ := c.Get("email")
	role, _ := c.Get("role")
	branding, _ := c.Get("branding")

	_, androidErr := os.Stat("/data/apps/fasmail.apk")
	_, windowsErr := os.Stat("/data/apps/fasmail-setup.exe")
	_, linuxErr := os.Stat("/data/apps/fasmail-linux")

	c.HTML(http.StatusOK, "downloads", gin.H{
		"Title":            "Descargas",
		"UserEmail":        email,
		"UserRole":         role,
		"Branding":         branding,
		"AndroidAvailable": androidErr == nil,
		"WindowsAvailable": windowsErr == nil,
		"LinuxAvailable":   linuxErr == nil,
	})
}

func (h *DownloadsHandler) HandleGenerateToken(c *gin.Context) {
	platform := c.PostForm("platform")
	if platform == "" {
		platform = "android"
	}

	userID := c.MustGet("user_id").(uuid.UUID)

	rawBytes := make([]byte, 32)
	if _, err := rand.Read(rawBytes); err != nil {
		c.Redirect(http.StatusFound, "/admin/downloads")
		return
	}
	rawToken := hex.EncodeToString(rawBytes)

	hash := sha256.Sum256([]byte(rawToken))
	tokenHash := hex.EncodeToString(hash[:])

	repo := models.NewAppTokenRepository(h.pool)
	appToken := &models.AppToken{
		UserID:    userID,
		TokenHash: tokenHash,
		ServerURL: h.serverURL,
		Platform:  platform,
		ExpiresAt: time.Now().Add(24 * time.Hour),
	}

	if err := repo.Create(c.Request.Context(), appToken); err != nil {
		c.Redirect(http.StatusFound, "/admin/downloads")
		return
	}

	setupURL := fmt.Sprintf("fasmail://setup?token=%s&server=%s", rawToken, h.serverURL)

	email, _ := c.Get("email")
	role, _ := c.Get("role")
	branding, _ := c.Get("branding")

	_, androidErr := os.Stat("/data/apps/fasmail.apk")
	_, windowsErr := os.Stat("/data/apps/fasmail-setup.exe")
	_, linuxErr := os.Stat("/data/apps/fasmail-linux")

	c.HTML(http.StatusOK, "downloads", gin.H{
		"Title":            "Descargas",
		"UserEmail":        email,
		"UserRole":         role,
		"Branding":         branding,
		"AndroidAvailable": androidErr == nil,
		"WindowsAvailable": windowsErr == nil,
		"LinuxAvailable":   linuxErr == nil,
		"SetupURL":         setupURL,
		"SetupToken":       rawToken,
		"Platform":         platform,
	})
}
