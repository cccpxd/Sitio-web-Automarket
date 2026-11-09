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
	cars  CarWriteStore
	ids   IDGenerator
	clock Clock
}

type ModerationServiceDeps struct {
	Store ModerationStore
	Cars  CarWriteStore
	IDs   IDGenerator
	Clock Clock
}

func NewModerationService(deps ModerationServiceDeps) (*ModerationService, error) {
	if deps.Store == nil || deps.Cars == nil || deps.IDs == nil || deps.Clock == nil {
		return nil, errors.New("moderation service: missing dependency")
	}
	return &ModerationService{store: deps.Store, cars: deps.Cars, ids: deps.IDs, clock: deps.Clock}, nil
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
		Action:    domain.ModerationActionRemoved,
		CreatedAt: s.clock.Now(),
	}
	return s.store.RemoveWithAudit(ctx, entry)
}

func (s *ModerationService) Approve(ctx context.Context, adminID, carID string) error {
	if adminID == "" {
		return domain.ErrForbidden
	}
	entryID, err := s.ids.New()
	if err != nil {
		return err
	}
	return s.store.ApproveWithAudit(ctx, domain.ModerationEntry{
		ID: entryID, AdminID: adminID, CarID: carID,
		Action: domain.ModerationActionApproved, CreatedAt: s.clock.Now(),
	})
}

func (s *ModerationService) Reject(ctx context.Context, adminID, carID, reason string) error {
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
	return s.store.RejectWithAudit(ctx, domain.ModerationEntry{
		ID: entryID, AdminID: adminID, CarID: carID, Reason: normalized,
		Action: domain.ModerationActionRejected, CreatedAt: s.clock.Now(),
	})
}

// DeletePermanently removes a withdrawn or removed listing and all of its
// associated records on behalf of an administrator.
func (s *ModerationService) DeletePermanently(ctx context.Context, adminID, carID string) error {
	if adminID == "" {
		return domain.ErrForbidden
	}
	return s.cars.DeletePermanently(ctx, carID, "")
}

// List returns the historical moderation log.
func (s *ModerationService) List(ctx context.Context) ([]domain.ModerationEntry, error) {
	return s.store.List(ctx)
}
