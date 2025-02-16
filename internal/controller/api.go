package controller

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/myacey/avito-shop/internal/apperror"
)

type authReq struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

// Authorize checks providen userame and password.
//
// If user exists -> give access token.
//
// If user dont exists -> create new one and give access token.
func (h *Controller) Authorize(c *gin.Context) {
	var req authReq
	if err := c.ShouldBindJSON(&req); err != nil {
		h.JSONError(c, err)
		return
	}

	token, err := h.srv.AuthorizeUser(c, req.Username, req.Password)
	if err != nil {
		h.JSONError(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{"token": token})
}

// GetFullInfo checks providen token with middleware and
// then gets dbUser, dbTransactionHistory and dbInventory.
func (h *Controller) GetFullUserInfo(c *gin.Context) {
	username, ok := c.Get("username")
	if !ok {
		h.JSONError(c, apperror.NewInternal("no username in token", nil))
		return
	}

	usr, err := h.srv.GetFullUserInfo(c, username.(string))
	if err != nil {
		h.JSONError(c, err)
		return
	}

	c.JSON(http.StatusOK, usr)
}

type sendCoinReq struct {
	ToUser string `json:"toUser"`
	Amount int32  `json:"amount"`
}

// SendCoins checks providen token with middleware and
// than transfers money from one user to another.
func (h *Controller) SendCoins(c *gin.Context) {
	username, ok := c.Get("username")
	if !ok {
		h.JSONError(c, apperror.NewInternal("no username in token", nil))
		return
	}

	var req sendCoinReq
	if err := c.ShouldBindJSON(&req); err != nil {
		h.JSONError(c, err)
		return
	}

	err := h.srv.SendCoin(c, username.(string), req.ToUser, req.Amount)
	if err != nil {
		h.JSONError(c, err)
		return
	}

	c.JSON(http.StatusOK, nil)
}

func (h *Controller) BuyItem(c *gin.Context) {
	username, ok := c.Get("username")
	if !ok {
		h.JSONError(c, apperror.NewInternal("no username in token", nil))
		return
	}

	item := c.Param("item")
	if item == "" {
		h.JSONError(c, apperror.NewBadReq("invalid item", nil))
		return
	}

	err := h.srv.BuyItem(c, username.(string), item)
	if err != nil {
		h.JSONError(c, err)
		return
	}

	c.JSON(http.StatusOK, nil)
}
