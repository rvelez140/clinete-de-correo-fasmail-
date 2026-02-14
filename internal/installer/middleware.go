package installer

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
)

func InstallerGuard(isInstalled func() bool) gin.HandlerFunc {
	return func(c *gin.Context) {
		path := c.Request.URL.Path

		if strings.HasPrefix(path, "/static/") || path == "/health" {
			c.Next()
			return
		}

		installed := isInstalled()

		if installed && strings.HasPrefix(path, "/install") {
			c.Redirect(http.StatusFound, "/auth/login")
			c.Abort()
			return
		}

		if !installed && !strings.HasPrefix(path, "/install") {
			c.Redirect(http.StatusFound, "/install/")
			c.Abort()
			return
		}

		c.Next()
	}
}
