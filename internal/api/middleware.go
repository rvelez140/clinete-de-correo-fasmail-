package api

import (
	"github.com/fasmail/panel/internal/auth"
	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
)

func APIAuthRequired(jwtMgr *auth.JWTManager, redisClient **redis.Client) gin.HandlerFunc {
	authMiddleware := auth.AuthRequired(jwtMgr, redisClient)
	return func(c *gin.Context) {
		c.Request.Header.Set("Accept", "application/json")
		authMiddleware(c)
	}
}

func APISuperAdminRequired() gin.HandlerFunc {
	mw := auth.SuperAdminRequired()
	return func(c *gin.Context) {
		c.Request.Header.Set("Accept", "application/json")
		mw(c)
	}
}

func APIForcePasswordChange() gin.HandlerFunc {
	mw := auth.ForcePasswordChange()
	return func(c *gin.Context) {
		c.Request.Header.Set("Accept", "application/json")
		mw(c)
	}
}
