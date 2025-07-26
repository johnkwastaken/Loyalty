package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/loyalty/ledger/internal/models"
	"github.com/loyalty/ledger/internal/repository"
)

type LedgerHandler struct {
	repo repository.TigerBeetleRepoInterface
}

func NewLedgerHandler(repo repository.TigerBeetleRepoInterface) *LedgerHandler {
	return &LedgerHandler{repo: repo}
}

func (h *LedgerHandler) CreateAccount(c *gin.Context) {
	var req models.CreateAccountRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	account, err := h.repo.CreateAccount(c.Request.Context(), &req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, account)
}

func (h *LedgerHandler) CreateTransfer(c *gin.Context) {
	var req models.CreateTransferRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	response, err := h.repo.CreateTransfer(c.Request.Context(), &req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, response)
}

func (h *LedgerHandler) GetAccount(c *gin.Context) {
	accountID := c.Param("id")
	if accountID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "account ID is required"})
		return
	}

	account, err := h.repo.GetAccount(c.Request.Context(), accountID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, account)
}

func (h *LedgerHandler) GetBalance(c *gin.Context) {
	orgID := c.Query("org_id")
	customerID := c.Query("customer_id")
	
	if orgID == "" || customerID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "org_id and customer_id are required"})
		return
	}

	balances, err := h.repo.GetBalance(c.Request.Context(), orgID, customerID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"org_id":         orgID,
		"customer_id":    customerID,
		"points_balance": balances["points"],
		"stamps_balance": balances["stamps"],
	})
}

func (h *LedgerHandler) Health(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"status": "healthy",
		"service": "ledger",
	})
}