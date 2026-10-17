package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
	customError "github.com/gootibi/golang-wallet-microservice/monolith/internal/errors"
	"github.com/gootibi/golang-wallet-microservice/monolith/internal/transaction/model"
	"github.com/gootibi/golang-wallet-microservice/monolith/internal/transaction/service"
)

type TransactionHandler struct {
	svc service.TransactionService
}

func NewTransactionHandler(svc service.TransactionService) *TransactionHandler {
	return &TransactionHandler{
		svc: svc,
	}
}

func (h *TransactionHandler) Transfer(c *gin.Context) {
	// Get senderUserID from auth middleware
	senderUserID, exist := c.Get("user_id")
	if !exist {
		c.Error(customError.NewAppError(http.StatusUnauthorized, "UNAUTHORIZED", "User context not found."))
		return
	}

	var req model.TransferRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.Error(customError.NewAppError(http.StatusBadRequest, "INVALID_INPUT", err.Error()))
		return
	}

	tx, err := h.svc.Transfer(c.Request.Context(), senderUserID.(string), req)
	if err != nil {
		c.Error(err)
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "Transaction successful",
		"data":    tx,
	})
}

func (h *TransactionHandler) GetHistory(c *gin.Context) {
	userID, exist := c.Get("user_id")
	if !exist {
		c.Error(customError.NewAppError(http.StatusUnauthorized, "UNAUTHORIZED", "User context not found."))
		return
	}

	var params model.PaginationParams
	if err := c.ShouldBindJSON(&params); err != nil {
		c.Error(customError.NewAppError(http.StatusBadRequest, "INVALID_INPUT", err.Error()))
		return
	}

	txs, meta, err := h.svc.GetHistory(c.Request.Context(), userID.(string), params)
	if err != nil {
		c.Error(err)
		return
	}

	// c.JSON(http.StatusOK, model.PaginationResponse[[]model.Transaction]{
	// 	Success: true,
	// 	Data:    txs,
	// 	Meta:    *meta,
	// })

	c.JSON(http.StatusOK, model.PaginationResponse{
		Success: true,
		Data:    txs,
		Meta:    *meta,
	})
}
