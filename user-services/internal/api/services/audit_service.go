package services

import (
	"context"
	"user-services/internal/api/dto"
	"user-services/internal/api/repositories"

	"github.com/google/uuid"
)

type AuditService interface {
	LogAction(ctx context.Context, userID, actorID *uuid.UUID, action, ipAddr string, metadata map[string]any) error
	GetUserAuditLogs(ctx context.Context, userID uuid.UUID, page, pageSize int) (*dto.PaginatedResponse, error)
	GetActionAuditLogs(ctx context.Context, action string, page, pageSize int) (*dto.PaginatedResponse, error)
}

type auditService struct {
	auditLogRepo repositories.AuditLogRepository
}

func NewAuditService(auditLogRepo repositories.AuditLogRepository) AuditService {
	return &auditService{
		auditLogRepo: auditLogRepo,
	}
}

func (s *auditService) LogAction(ctx context.Context, userID, actorID *uuid.UUID, action, ipAddr string, metadata map[string]any) error {
	// TODO: implement
	return nil
}

func (s *auditService) GetUserAuditLogs(ctx context.Context, userID uuid.UUID, page, pageSize int) (*dto.PaginatedResponse, error) {
	// TODO: implement
	return nil, nil
}

func (s *auditService) GetActionAuditLogs(ctx context.Context, action string, page, pageSize int) (*dto.PaginatedResponse, error) {
	// TODO: implement
	return nil, nil
}
