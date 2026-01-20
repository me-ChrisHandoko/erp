package handler

import (
	"net/http"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"

	"backend/internal/service/audit"
	pkgerrors "backend/pkg/errors"
)

// AuditHandler - HTTP handlers for audit log endpoints
type AuditHandler struct {
	auditService *audit.AuditService
}

// NewAuditHandler creates a new audit handler instance
func NewAuditHandler(auditService *audit.AuditService) *AuditHandler {
	return &AuditHandler{
		auditService: auditService,
	}
}

// GetAuditLogsByEntityID handles GET /api/v1/audit-logs/:entityType/:entityId
// Returns audit logs for a specific entity (e.g., stock_transfer, warehouse, product)
func (h *AuditHandler) GetAuditLogsByEntityID(c *gin.Context) {
	tenantID, exists := c.Get("tenant_id")
	if !exists {
		c.JSON(http.StatusBadRequest, pkgerrors.NewBadRequestError("Tenant context not found"))
		return
	}

	entityType := c.Param("entityType")
	if entityType == "" {
		c.JSON(http.StatusBadRequest, pkgerrors.NewBadRequestError("Entity type is required"))
		return
	}
	// Convert to uppercase to match database storage format (e.g., "stock_transfer" → "STOCK_TRANSFER")
	entityType = strings.ToUpper(entityType)

	entityID := c.Param("entityId")
	if entityID == "" {
		c.JSON(http.StatusBadRequest, pkgerrors.NewBadRequestError("Entity ID is required"))
		return
	}

	// Parse pagination parameters
	limit := 20
	offset := 0

	if limitStr := c.Query("limit"); limitStr != "" {
		if parsed, err := strconv.Atoi(limitStr); err == nil && parsed > 0 && parsed <= 100 {
			limit = parsed
		}
	}

	if offsetStr := c.Query("offset"); offsetStr != "" {
		if parsed, err := strconv.Atoi(offsetStr); err == nil && parsed >= 0 {
			offset = parsed
		}
	}

	// Get audit logs
	auditLogs, total, err := h.auditService.GetAuditLogsByEntityID(
		c.Request.Context(),
		tenantID.(string),
		entityType,
		entityID,
		limit,
		offset,
	)
	if err != nil {
		if appErr, ok := err.(*pkgerrors.AppError); ok {
			c.JSON(appErr.StatusCode, appErr)
			return
		}
		c.JSON(http.StatusInternalServerError, pkgerrors.NewInternalError(err))
		return
	}

	// Transform audit logs to response format
	responses := make([]AuditLogResponse, len(auditLogs))
	for i, log := range auditLogs {
		responses[i] = AuditLogResponse{
			ID:         log.ID,
			TenantID:   ptrToString(log.TenantID),
			CompanyID:  ptrToString(log.CompanyID),
			UserID:     ptrToString(log.UserID),
			RequestID:  ptrToString(log.RequestID),
			Action:     log.Action,
			EntityType: ptrToString(log.EntityType),
			EntityID:   ptrToString(log.EntityID),
			OldValues:  ptrToString(log.OldValues),
			NewValues:  ptrToString(log.NewValues),
			IPAddress:  ptrToString(log.IPAddress),
			UserAgent:  ptrToString(log.UserAgent),
			Status:     log.Status,
			Notes:      ptrToString(log.Notes),
			CreatedAt:  log.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    responses,
		"pagination": gin.H{
			"limit":  limit,
			"offset": offset,
			"total":  total,
		},
	})
}

// AuditLogResponse represents the audit log response structure
type AuditLogResponse struct {
	ID         string `json:"id"`
	TenantID   string `json:"tenantId,omitempty"`
	CompanyID  string `json:"companyId,omitempty"`
	UserID     string `json:"userId,omitempty"`
	RequestID  string `json:"requestId,omitempty"`
	Action     string `json:"action"`
	EntityType string `json:"entityType,omitempty"`
	EntityID   string `json:"entityId,omitempty"`
	OldValues  string `json:"oldValues,omitempty"`
	NewValues  string `json:"newValues,omitempty"`
	IPAddress  string `json:"ipAddress,omitempty"`
	UserAgent  string `json:"userAgent,omitempty"`
	Status     string `json:"status"`
	Notes      string `json:"notes,omitempty"`
	CreatedAt  string `json:"createdAt"`
}

// Helper function to safely dereference string pointer
func ptrToString(ptr *string) string {
	if ptr == nil {
		return ""
	}
	return *ptr
}
