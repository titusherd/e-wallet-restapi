package handler

import (
	"main/dto"
	"main/usecase"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
)

type Handler struct {
	service usecase.TransactionService
}

func NewTransactionHandler(service usecase.TransactionService) *Handler {
	return &Handler{service: service}
}

func (h *Handler) ListTransactions(c *gin.Context) {
	var req dto.TransactionListRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Validate date format if provided
	if req.StartDate != "" {
		if _, err := time.Parse("2006-01-02", req.StartDate); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid start date format. Use YYYY-MM-DD"})
			return
		}
	}
	if req.EndDate != "" {
		if _, err := time.Parse("2006-01-02", req.EndDate); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid end date format. Use YYYY-MM-DD"})
			return
		}
	}

	// Set default values
	if req.Limit <= 0 {
		req.Limit = 10
	}
	if req.Page <= 0 {
		req.Page = 1
	}

	// Get user ID from context (set by auth middleware)
	userID, _ := c.Get("userID")

	response, err := h.service.ListTransactions(c.Request.Context(), userID.(int), req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch transactions"})
		return
	}

	c.JSON(http.StatusOK, response)
}
