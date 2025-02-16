package controller

import (
	"errors"
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/gin-gonic/gin"
	"github.com/myacey/avito-shop/internal/apperror"
	"github.com/myacey/avito-shop/internal/service"
)

type Controller struct {
	srv service.Interface

	testingStatus bool
}

func NewController(srv service.Interface) *Controller {
	testingStatus := os.Getenv("STATUS") == "testing"
	return &Controller{srv, testingStatus}
}

func (h *Controller) JSONError(c *gin.Context, err error) {
	var appErr *apperror.AppError
	if errors.As(err, &appErr) {
		if appErr.HTTPCode == http.StatusInternalServerError {
			log.Printf("Internal error! user message: %s", appErr.Message)
		}
		log.Printf("error: %s", fmt.Sprint(appErr.Err))
		c.JSON(appErr.HTTPCode, gin.H{"errors": appErr.Message})
		return
	}

	log.Printf("Unexpected error: %v", err.Error())
	c.JSON(http.StatusInternalServerError, gin.H{"errors": "internal server error"})
}
