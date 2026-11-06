package service

import (
	"context"
	"errors"

	"automarket/internal/domain"
)

// ModerationService removes listings on behalf of an administrator and records
// an immutable audit entry.
type ModerationService struct {
	store ModerationStore
	ids   IDGenerator
	clock Clock
}

type ModerationServiceDeps struct {
	Store ModerationStore
	IDs   IDGenerator
	Clock Clock
}

func NewModerationService(deps ModerationServiceDeps) (*ModerationService, error) {
	if deps.Store == nil || deps.IDs == nil || deps.Clock == nil {
		return nil, errors.New("moderation service: missing dependency")
	}
	return &ModerationService{store: deps.Store, ids: deps.IDs, clock: deps.Clock}, nil
}

// Remove withdraws a listing and stores the mandatory reason atomically.
func (s *ModerationService) Remove(ctx context.Context, adminID, carID, reason string) error {
	if adminID == "" {
		return domain.ErrForbidden
	}
	normalized, err := domain.NormalizeModerationReason(reason)
	if err != nil {
		return err
	}
	entryID, err := s.ids.New()
	if err != nil {
		return err
	}
	entry := domain.ModerationEntry{
		ID:        entryID,
		AdminID:   adminID,
		CarID:     carID,
		Reason:    normalized,
		CreatedAt: s.clock.Now(),
	}
	return s.store.RemoveWithAudit(ctx, entry)
}

// List returns the historical moderation log.
func (s *ModerationService) List(ctx context.Context) ([]domain.ModerationEntry, error) {
	return s.store.List(ctx)
}
