package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/gootibi/golang-wallet-microservice/monolith/internal/wallet/service"
)

type WalletHandler struct {
	svc service.WalletService
}

func NewWalletHandler(svc service.WalletService) *WalletHandler {
	return &WalletHandler{
		svc: svc,
	}
}

func (h *WalletHandler) GetMyWallet(c *gin.Context) {
	// userID from JWT context
	userID, _ := c.Get("user_id")

	wallet, err := h.svc.GetWalletByUserID(c.Request.Context(), userID.(string))
	if err != nil {
		c.Error(err)
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"seccess": true,
		"data":    wallet,
	})
}
