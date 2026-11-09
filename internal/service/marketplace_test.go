package service

import (
	"context"
	"errors"
	"testing"
	"time"

	"automarket/internal/domain"
)

type fixedClock struct{ now time.Time }

func (c fixedClock) Now() time.Time { return c.now }

type sequenceIDs struct{ next int }

func (g *sequenceIDs) New() (string, error) {
	g.next++
	return "id-" + string(rune('0'+g.next)), nil
}

type fakeCarStore struct {
	cars      map[string]domain.Car
	created   int
	withdrawn bool
}

func (s *fakeCarStore) Search(_ context.Context, filters domain.SearchFilters) ([]domain.Car, int, error) {
	var result []domain.Car
	for _, car := range s.cars {
		if car.Status == domain.CarStatusActive && (filters.Year == 0 || car.Year == filters.Year) {
			result = append(result, car)
		}
	}
	return result, len(result), nil
}
func (s *fakeCarStore) FindActive(_ context.Context, id string) (domain.Car, error) {
	car, ok := s.cars[id]
	if !ok || car.Status != domain.CarStatusActive {
		return domain.Car{}, domain.ErrNotFound
	}
	return car, nil
}
func (s *fakeCarStore) FindByID(_ context.Context, id string) (domain.Car, error) {
	car, ok := s.cars[id]
	if !ok {
		return domain.Car{}, domain.ErrNotFound
	}
	return car, nil
}
func (s *fakeCarStore) ListBySeller(_ context.Context, sellerID string) ([]domain.Car, error) {
	var result []domain.Car
	for _, car := range s.cars {
		if car.SellerID == sellerID {
			result = append(result, car)
		}
	}
	return result, nil
}
func (s *fakeCarStore) ListAll(_ context.Context) ([]domain.Car, error) {
	var result []domain.Car
	for _, car := range s.cars {
		result = append(result, car)
	}
	return result, nil
}
func (s *fakeCarStore) CreateWithPhotos(_ context.Context, car domain.Car, photos []domain.CarPhoto) error {
	if _, exists := s.cars[car.ID]; exists {
		return domain.ErrDuplicate
	}
	car.Photos = photos
	s.cars[car.ID] = car
	s.created++
	return nil
}
func (s *fakeCarStore) UpdateWithPhotos(_ context.Context, car domain.Car, photos []domain.CarPhoto, replace bool, cutoff time.Time) error {
	existing, ok := s.cars[car.ID]
	if !ok || existing.CreatedAt.Before(cutoff) {
		return domain.ErrEditWindowClosed
	}
	if replace {
		car.Photos = photos
	}
	s.cars[car.ID] = car
	return nil
}
func (s *fakeCarStore) Withdraw(_ context.Context, id, sellerID string, _ time.Time) error {
	car, ok := s.cars[id]
	if !ok {
		return domain.ErrNotFound
	}
	if car.SellerID != sellerID {
		return domain.ErrForbidden
	}
	car.Status = domain.CarStatusWithdrawn
	s.cars[id] = car
	s.withdrawn = true
	return nil
}
func (s *fakeCarStore) FindByClientToken(_ context.Context, sellerID, token string) (domain.Car, error) {
	for _, car := range s.cars {
		if car.SellerID == sellerID && car.ClientToken == token {
			return car, nil
		}
	}
	return domain.Car{}, domain.ErrNotFound
}

type fakePhotoProcessor struct{}

func (fakePhotoProcessor) Process(raw RawPhoto) (domain.CarPhoto, error) {
	if len(raw.Data) == 0 {
		return domain.CarPhoto{}, domain.ErrPhotoDecodeFailed
	}
	return domain.CarPhoto{ContentType: "image/jpeg", ByteSize: int64(len(raw.Data)), Data: raw.Data}, nil
}

func TestCatalogAndPublishingServices(t *testing.T) {
	now := time.Date(2026, time.November, 6, 12, 0, 0, 0, time.UTC)
	store := &fakeCarStore{cars: map[string]domain.Car{
		"active":  {ID: "active", Status: domain.CarStatusActive, Year: 2013},
		"removed": {ID: "removed", Status: domain.CarStatusRemoved, Year: 2013},
	}}
	catalog, err := NewCatalogService(store)
	if err != nil {
		t.Fatal(err)
	}
	result, err := catalog.Search(context.Background(), domain.SearchQuery{Year: "2013"})
	if err != nil || len(result.Cars) != 1 || result.Cars[0].ID != "active" {
		t.Fatalf("catalog result = %+v, %v", result, err)
	}

	ids := &sequenceIDs{}
	publishing, err := NewPublishingService(PublishingServiceDeps{
		Cars: store, Reader: store, Photos: fakePhotoProcessor{}, IDs: ids, Clock: fixedClock{now},
	})
	if err != nil {
		t.Fatal(err)
	}
	car, err := publishing.Create(context.Background(), "seller-1", CreateCarInput{
		Brand: "Mazda", Model: "3", City: "Bogotá", Year: 2020, Price: "45.000.000", ClientToken: "token-1",
		Photos: []RawPhoto{{Name: "car.jpg", Data: []byte("photo")}},
	})
	if err != nil || car.Status != domain.CarStatusPending || store.created != 1 {
		t.Fatalf("created car = %+v, %v", car, err)
	}
	if err := publishing.Withdraw(context.Background(), "seller-1", car.ID); err != nil || !store.withdrawn {
		t.Fatalf("withdraw = %v", err)
	}
	if _, err := publishing.Update(context.Background(), "seller-1", car.ID, UpdateCarInput{
		Brand: "Mazda", Model: "3", Year: 2020, Price: "1", Description: "x",
	}); !errors.Is(err, domain.ErrConflict) {
		t.Fatalf("update withdrawn error = %v", err)
	}
}
