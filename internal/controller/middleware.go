package controller

import (
	"log"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/myacey/avito-shop/internal/apperror"
)

func (h *Controller) AuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		if c.Request.Method == http.MethodOptions {
			c.Next()
		}

		if h.testingStatus {
			log.Print(">> TESTING, skip auth")
			c.Set("username", "testuser")
			c.Next()
			return
		}

		authHeader := c.GetHeader("Authorization")

		bearerToken := strings.Split(authHeader, " ")
		if bearerToken[0] != "Bearer" || len(bearerToken) != 2 {
			h.JSONError(c, apperror.NewUnauthorized("invalid token", nil))
			c.Abort()
			return
		}

		usrname, err := h.srv.CheckAuthToken(c, bearerToken[1])
		if err != nil {
			h.JSONError(c, err)
			c.Abort()
			return
		}

		c.Set("username", usrname)
		c.Next()
	}
}
