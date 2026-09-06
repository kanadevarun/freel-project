package audit

import (
	"context"
	"sync"

	"github.com/freel/backend/internal/audit/domain"
	"github.com/freel/backend/internal/audit/service"
)

var (
	defaultService service.Service
	mu             sync.RWMutex
)

// SetDefaultService sets the globally accessible audit service instance.
func SetDefaultService(s service.Service) {
	mu.Lock()
	defer mu.Unlock()
	defaultService = s
}

// GetDefaultService retrieves the globally configured audit service instance.
func GetDefaultService() service.Service {
	mu.RLock()
	defer mu.RUnlock()
	return defaultService
}

// Record persists a synchronous audit log using the global service.
func Record(ctx context.Context, params domain.CreateAuditLogParams) (*domain.AuditLog, error) {
	mu.RLock()
	s := defaultService
	mu.RUnlock()

	if s != nil {
		return s.Record(ctx, params)
	}
	return nil, nil
}

// RecordAsync persists an asynchronous audit log using the global service in a background goroutine.
func RecordAsync(ctx context.Context, params domain.CreateAuditLogParams) {
	mu.RLock()
	s := defaultService
	mu.RUnlock()

	if s != nil {
		s.RecordAsync(ctx, params)
	}
}
