package service

import (
	"context"
	"errors"

	"automarket/internal/domain"
)

// PhotoService enforces visibility of photos. Active listings are public;
// withdrawn and removed listings are only visible to their owner or an
// administrator.
type PhotoService struct {
	photos PhotoStore
}

func NewPhotoService(photos PhotoStore) (*PhotoService, error) {
	if photos == nil {
		return nil, errors.New("photo service: missing photo store")
	}
	return &PhotoService{photos: photos}, nil
}

// GetPublic returns the metadata and binary content of a photo when the
// viewer is allowed to see it.
func (s *PhotoService) GetPublic(ctx context.Context, id string, viewer *domain.User) (domain.CarPhoto, error) {
	meta, err := s.photos.FindMeta(ctx, id)
	if err != nil {
		return domain.CarPhoto{}, err
	}
	if !meta.CarStatus.IsActive() {
		allowed := viewer != nil && (viewer.Role == domain.RoleAdmin || viewer.ID == meta.CarSellerID)
		if !allowed {
			return domain.CarPhoto{}, domain.ErrForbidden
		}
	}
	return s.photos.FindData(ctx, id)
}
