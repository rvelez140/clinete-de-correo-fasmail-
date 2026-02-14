package auth

import (
	"net/http"
	"strings"

	"github.com/fasmail/panel/internal/database"
	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
)

func AuthRequired(jwtMgr *JWTManager, redisClient **redis.Client) gin.HandlerFunc {
	return func(c *gin.Context) {
		tokenStr := extractToken(c)
		if tokenStr == "" {
			if isHTMLRequest(c) {
				c.Redirect(http.StatusFound, "/auth/login")
				c.Abort()
				return
			}
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "token requerido"})
			return
		}

		claims, err := jwtMgr.ValidateAccessToken(tokenStr)
		if err != nil {
			if isHTMLRequest(c) {
				c.Redirect(http.StatusFound, "/auth/login")
				c.Abort()
				return
			}
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "token inválido"})
			return
		}

		// Check session in Redis if available
		if redisClient != nil && *redisClient != nil {
			_, err = database.GetSession(c.Request.Context(), *redisClient, claims.SessionID.String())
			if err != nil {
				if isHTMLRequest(c) {
					c.Redirect(http.StatusFound, "/auth/login")
					c.Abort()
					return
				}
				c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "sesión expirada"})
				return
			}
		}

		c.Set("claims", claims)
		c.Set("user_id", claims.UserID)
		c.Set("email", claims.Email)
		c.Set("role", claims.Role)
		c.Set("must_change_password", claims.MustChangePassword)
		c.Set("session_id", claims.SessionID)

		c.Next()
	}
}

func AdminRequired() gin.HandlerFunc {
	return func(c *gin.Context) {
		role, exists := c.Get("role")
		if !exists || role != "admin" {
			if isHTMLRequest(c) {
				c.HTML(http.StatusForbidden, "error", gin.H{
					"Title":   "Acceso Denegado",
					"Message": "No tienes permisos para acceder a esta sección.",
				})
				c.Abort()
				return
			}
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"error": "permisos insuficientes"})
			return
		}
		c.Next()
	}
}

func ForcePasswordChange() gin.HandlerFunc {
	return func(c *gin.Context) {
		must, exists := c.Get("must_change_password")
		if !exists {
			c.Next()
			return
		}

		if mustChange, ok := must.(bool); ok && mustChange {
			path := c.Request.URL.Path
			if path == "/auth/change-password" || path == "/auth/logout" ||
				strings.HasPrefix(path, "/static/") {
				c.Next()
				return
			}

			if isHTMLRequest(c) {
				c.Redirect(http.StatusFound, "/auth/change-password")
				c.Abort()
				return
			}
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{
				"error":  "debe cambiar su contraseña",
				"action": "change_password",
			})
			return
		}

		c.Next()
	}
}

func extractToken(c *gin.Context) string {
	authHeader := c.GetHeader("Authorization")
	if authHeader != "" {
		parts := strings.SplitN(authHeader, " ", 2)
		if len(parts) == 2 && strings.ToLower(parts[0]) == "bearer" {
			return parts[1]
		}
	}

	if token, err := c.Cookie("access_token"); err == nil && token != "" {
		return token
	}

	return ""
}

func isHTMLRequest(c *gin.Context) bool {
	accept := c.GetHeader("Accept")
	return strings.Contains(accept, "text/html") || !strings.Contains(accept, "application/json")
}
