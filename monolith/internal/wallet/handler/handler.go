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

// GetMyWallet godoc
//
//	@Summary		Get My Wallet
//	@Description	Get current authenticated user's wallet info (balance, currency, etc.)
//	@Tags			Wallets
//	@Accept			json
//	@Produce		json
//	@Success		200	{object}	map[string]interface{}	"Returns success: true and data: model.Wallet"
//	@Failure		404	{object}	errors.AppError
//	@Router			/wallets/me [get]
//	@Security		BearerAuth
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
