package handler

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"swiftwallet/internal/model"
	"swiftwallet/internal/repository"
	"swiftwallet/internal/service"
)

type Handler struct{ svc service.Service }

func New(s service.Service) *Handler { return &Handler{svc: s} }

type opRequest struct {
	WalletID      uuid.UUID           `json:"walletId"      binding:"required,uuid"`
	OperationType model.OperationType `json:"operationType" binding:"required,oneof=DEPOSIT WITHDRAW"`
	Amount        int64               `json:"amount"        binding:"required,gt=0"`
}

type opResponse struct {
	WalletID uuid.UUID `json:"walletId"`
	Balance  int64     `json:"balance"`
}

// Operate POST /api/v1/wallet
func (h *Handler) Operate(c *gin.Context) {
	var req opRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	bal, err := h.svc.Operate(c.Request.Context(), req.WalletID, req.OperationType, req.Amount)
	if err != nil {
		switch {
		case errors.Is(err, repository.ErrNotFound):
			c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		case errors.Is(err, repository.ErrUnknownOperation),
			errors.Is(err, repository.ErrInsufficientFunds):
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		default:
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		}
		return
	}
	c.JSON(http.StatusOK, opResponse{WalletID: req.WalletID, Balance: bal})
}

// Balance GET /api/v1/wallets/:id
func (h *Handler) Balance(c *gin.Context) {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid uuid"})
		return
	}

	bal, err := h.svc.Balance(c.Request.Context(), id)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, opResponse{WalletID: id, Balance: bal})
}
