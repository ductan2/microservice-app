package controllers

import (
	"user-services/internal/api/services"

	"github.com/gin-gonic/gin"
)

type AuditController struct {
	auditService services.AuditService
}

func NewAuditController(auditService services.AuditService) *AuditController {
	return &AuditController{
		auditService: auditService,
	}
}

// GetUserAuditLogs godoc
// @Summary Get audit logs for a user (admin only)
// @Tags audit
// @Produce json
// @Param id path string true "User ID"
// @Param page query int false "Page number"
// @Param page_size query int false "Page size"
// @Success 200 {object} dto.PaginatedResponse
// @Router /audit/users/{id} [get]
func (c *AuditController) GetUserAuditLogs(ctx *gin.Context) {
	// TODO: implement
}

// GetActionAuditLogs godoc
// @Summary Get audit logs by action (admin only)
// @Tags audit
// @Produce json
// @Param action query string true "Action name"
// @Param page query int false "Page number"
// @Param page_size query int false "Page size"
// @Success 200 {object} dto.PaginatedResponse
// @Router /audit/actions [get]
func (c *AuditController) GetActionAuditLogs(ctx *gin.Context) {
	// TODO: implement
}
