package service

import (
	"context"
	"errors"
	"strings"

	"automarket/internal/domain"
)

// CreateCarInput is the application-level input for creating a listing.
type CreateCarInput struct {
	Brand       string
	Model       string
	City        string
	Year        int
	Description string
	Price       string
	ClientToken string
	Photos      []RawPhoto
}

// UpdateCarInput is the application-level input for editing a listing.
type UpdateCarInput struct {
	Brand       string
	Model       string
	City        string
	Year        int
	Description string
	Price       string
	Photos      []RawPhoto
}

// PublishingService enforces listing validation, the ten-minute edit window,
// seller ownership, withdrawal and idempotent creation.
type PublishingService struct {
	cars   CarWriteStore
	reader CarListStore
	photos PhotoProcessor
	ids    IDGenerator
	clock  Clock
}

type PublishingServiceDeps struct {
	Cars   CarWriteStore
	Reader CarListStore
	Photos PhotoProcessor
	IDs    IDGenerator
	Clock  Clock
}

func NewPublishingService(deps PublishingServiceDeps) (*PublishingService, error) {
	if deps.Cars == nil || deps.Reader == nil || deps.Photos == nil || deps.IDs == nil || deps.Clock == nil {
		return nil, errors.New("publishing service: missing dependency")
	}
	return &PublishingService{
		cars:   deps.Cars,
		reader: deps.Reader,
		photos: deps.Photos,
		ids:    deps.IDs,
		clock:  deps.Clock,
	}, nil
}

// Create publishes a listing immediately, with its photos stored atomically.
func (s *PublishingService) Create(ctx context.Context, sellerID string, input CreateCarInput) (domain.Car, error) {
	if sellerID == "" {
		return domain.Car{}, domain.ErrForbidden
	}
	now := s.clock.Now()
	details, err := domain.NormalizeCarInput(domain.CarInput{
		Brand:       input.Brand,
		Model:       input.Model,
		City:        input.City,
		Year:        input.Year,
		Description: input.Description,
		Price:       input.Price,
	}, now)
	if err != nil {
		return domain.Car{}, err
	}
	if err := domain.ValidatePhotoCount(len(input.Photos)); err != nil {
		return domain.Car{}, err
	}

	carID, err := s.ids.New()
	if err != nil {
		return domain.Car{}, err
	}
	car := domain.Car{
		ID:          carID,
		Brand:       details.Brand,
		Model:       details.Model,
		City:        details.City,
		Year:        details.Year,
		Description: details.Description,
		PriceCents:  details.PriceCents,
		SellerID:    sellerID,
		Status:      domain.CarStatusActive,
		CreatedAt:   now,
	}

	photos, err := s.processPhotos(input.Photos)
	if err != nil {
		return domain.Car{}, err
	}
	for index := range photos {
		photos[index].CarID = car.ID
		photos[index].Position = index
	}
	car.Photos = photos

	clientToken := strings.TrimSpace(input.ClientToken)
	if clientToken == "" {
		clientToken, err = s.ids.New()
		if err != nil {
			return domain.Car{}, err
		}
	}
	car.ClientToken = clientToken

	if err := s.cars.CreateWithPhotos(ctx, car, photos); err != nil {
		if errors.Is(err, domain.ErrDuplicate) {
			existing, findErr := s.cars.FindByClientToken(ctx, sellerID, clientToken)
			if findErr != nil {
				return domain.Car{}, findErr
			}
			return existing, nil
		}
		return domain.Car{}, err
	}
	return car, nil
}

// Update edits a listing owned by the seller, only inside the ten-minute
// window. Photos are replaced only when the seller submits new ones.
func (s *PublishingService) Update(ctx context.Context, sellerID, carID string, input UpdateCarInput) (domain.Car, error) {
	if sellerID == "" {
		return domain.Car{}, domain.ErrForbidden
	}
	existing, err := s.reader.FindByID(ctx, carID)
	if err != nil {
		return domain.Car{}, err
	}
	if existing.SellerID != sellerID {
		return domain.Car{}, domain.ErrForbidden
	}
	if existing.Status != domain.CarStatusActive {
		return domain.Car{}, domain.ErrConflict
	}
	now := s.clock.Now()
	if now.After(existing.CreatedAt.Add(domain.EditWindow)) {
		return domain.Car{}, domain.ErrEditWindowClosed
	}

	details, err := domain.NormalizeCarInput(domain.CarInput{
		Brand:       input.Brand,
		Model:       input.Model,
		City:        input.City,
		Year:        input.Year,
		Description: input.Description,
		Price:       input.Price,
	}, now)
	if err != nil {
		return domain.Car{}, err
	}

	replacePhotos := len(input.Photos) > 0
	var photos []domain.CarPhoto
	if replacePhotos {
		if err := domain.ValidatePhotoCount(len(input.Photos)); err != nil {
			return domain.Car{}, err
		}
		photos, err = s.processPhotos(input.Photos)
		if err != nil {
			return domain.Car{}, err
		}
		for index := range photos {
			photos[index].CarID = existing.ID
			photos[index].Position = index
		}
	}

	editedAt := now
	updated := existing
	updated.Brand = details.Brand
	updated.Model = details.Model
	updated.City = details.City
	updated.Year = details.Year
	updated.Description = details.Description
	updated.PriceCents = details.PriceCents
	updated.EditedAt = &editedAt
	if replacePhotos {
		updated.Photos = photos
	}

	cutoff := now.Add(-domain.EditWindow)
	if err := s.cars.UpdateWithPhotos(ctx, updated, photos, replacePhotos, cutoff); err != nil {
		return domain.Car{}, err
	}
	return updated, nil
}

// Withdraw lets the owner retire a listing at any time.
func (s *PublishingService) Withdraw(ctx context.Context, sellerID, carID string) error {
	if sellerID == "" {
		return domain.ErrForbidden
	}
	return s.cars.Withdraw(ctx, carID, sellerID, s.clock.Now())
}

// ListMine returns the seller's own listings, including withdrawn ones.
func (s *PublishingService) ListMine(ctx context.Context, sellerID string) ([]domain.Car, error) {
	return s.reader.ListBySeller(ctx, sellerID)
}

// Get returns a listing owned by the seller in any state.
func (s *PublishingService) Get(ctx context.Context, sellerID, carID string) (domain.Car, error) {
	car, err := s.reader.FindByID(ctx, carID)
	if err != nil {
		return domain.Car{}, err
	}
	if car.SellerID != sellerID {
		return domain.Car{}, domain.ErrForbidden
	}
	return car, nil
}

func (s *PublishingService) processPhotos(raw []RawPhoto) ([]domain.CarPhoto, error) {
	photos := make([]domain.CarPhoto, 0, len(raw))
	now := s.clock.Now()
	for _, upload := range raw {
		processed, err := s.photos.Process(upload)
		if err != nil {
			return nil, err
		}
		id, err := s.ids.New()
		if err != nil {
			return nil, err
		}
		processed.ID = id
		processed.CreatedAt = now
		photos = append(photos, processed)
	}
	return photos, nil
}
