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

// Transfer godoc
//
//	@Summary		Transfer accounts
//	@Description	Transfer money to another user's wallet using email
//	@Tags			Transactions
//	@Accept			json
//	@Produce		json
//	@Param			request	body		model.TransferRequest	true	"transfer payload"
//	@Success		200		{object}	map[string]interface{}	"Returns success: true, message: Transaction successful, and data: model.Transaction"
//	@Failure		400		{object}	errors.AppError
//	@Failure		401		{object}	errors.AppError
//	@Failure		404		{object}	errors.AppError
//	@Failure		409		{object}	errors.AppError
//	@Failure		500		{object}	errors.AppError
//	@Router			/transactions/transfer [get]
//	@Security		BearerAuth
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

// GetHistory godoc
//
//	@Summary		Get Transaction History
//	@Description	Get paginated list of transactions involving the authenticated user
//	@Tags			Transactions
//	@Accept			json
//	@Produce		json
//	@Param			page	query		int							false	"page number (default: 1)"
//	@Param			limit	query		int							false	"page limit (default: 10)"
//	@Param			sort	query		string						false	"sort column (default: created_at)"
//	@Param			order	query		string						false	"sort order (default: desc)"
//	@Param			status	query		string						false	"filter by status (success/failed)"
//	@Success		200		{object}	model.PaginationResponse	"Returns success: true, data: []model.Transaction, and data: model.PaginationMeta"
//	@Failure		400		{object}	errors.AppError
//	@Failure		401		{object}	errors.AppError
//	@Failure		404		{object}	errors.AppError
//	@Failure		500		{object}	errors.AppError
//	@Router			/transactions/history [get]
//	@Security		BearerAuth
func (h *TransactionHandler) GetHistory(c *gin.Context) {
	userID, exist := c.Get("user_id")
	if !exist {
		c.Error(customError.NewAppError(http.StatusUnauthorized, "UNAUTHORIZED", "User context not found."))
		return
	}

	var params model.PaginationParams
	if err := c.ShouldBindQuery(&params); err != nil {
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
