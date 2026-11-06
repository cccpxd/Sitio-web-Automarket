package service

import (
	"context"
	"errors"

	"automarket/internal/domain"
)

// CatalogService exposes the public catalog: active listings, detail and
// search. It never returns listings that are not active.
type CatalogService struct {
	cars CarListStore
}

func NewCatalogService(cars CarListStore) (*CatalogService, error) {
	if cars == nil {
		return nil, errors.New("catalog service: missing car store")
	}
	return &CatalogService{cars: cars}, nil
}

// Search returns one page of active listings matching the normalized filters.
func (s *CatalogService) Search(ctx context.Context, raw domain.SearchQuery) (domain.SearchResult, error) {
	filters, err := domain.NormalizeSearchFilters(raw)
	if err != nil {
		return domain.SearchResult{}, err
	}
	cars, total, err := s.cars.Search(ctx, filters)
	if err != nil {
		return domain.SearchResult{}, err
	}
	return domain.NewSearchResult(cars, total, filters), nil
}

// GetActive returns one active listing or ErrNotFound.
func (s *CatalogService) GetActive(ctx context.Context, id string) (domain.Car, error) {
	car, err := s.cars.FindActive(ctx, id)
	if err != nil {
		return domain.Car{}, err
	}
	if !car.Status.IsActive() {
		return domain.Car{}, domain.ErrNotFound
	}
	return car, nil
}

// ListAll returns every listing, including withdrawn and removed ones. It is
// reserved for the administrator panel.
func (s *CatalogService) ListAll(ctx context.Context) ([]domain.Car, error) {
	return s.cars.ListAll(ctx)
}

// Get returns one listing in any state. Callers must enforce ownership or role
// rules before exposing it.
func (s *CatalogService) Get(ctx context.Context, id string) (domain.Car, error) {
	return s.cars.FindByID(ctx, id)
}
