package service

import (
	"context"
	"errors"

	"automarket/internal/domain"
)

// CarStore is the persistence port for vehicles.
type CarStore interface {
	Create(ctx context.Context, car domain.Car) error
	List(ctx context.Context) ([]domain.Car, error)
	Delete(ctx context.Context, id string) error
	Report(ctx context.Context, id string) error
}

// CreateCarInput is the application-level input for creating a car.
type CreateCarInput struct {
	Brand       string
	Model       string
	Year        int
	Description string
	Price       string
}

// CarService contains car business rules.
type CarService struct {
	cars  CarStore
	ids   IDGenerator
	clock Clock
}

type CarServiceDeps struct {
	Cars  CarStore
	IDs   IDGenerator
	Clock Clock
}

func NewCarService(deps CarServiceDeps) (*CarService, error) {
	if deps.Cars == nil || deps.IDs == nil || deps.Clock == nil {
		return nil, errors.New("car service: missing dependency")
	}
	return &CarService{cars: deps.Cars, ids: deps.IDs, clock: deps.Clock}, nil
}

func (s *CarService) ListCars(ctx context.Context) ([]domain.Car, error) {
	return s.cars.List(ctx)
}

func (s *CarService) CreateCar(ctx context.Context, sellerID string, input CreateCarInput) (domain.Car, error) {
	normalized, err := domain.NormalizeCarInput(domain.CarInput{
		Brand:       input.Brand,
		Model:       input.Model,
		Year:        input.Year,
		Description: input.Description,
		Price:       input.Price,
	})
	if err != nil {
		return domain.Car{}, err
	}

	id, err := s.ids.New()
	if err != nil {
		return domain.Car{}, err
	}

	car := domain.Car{
		ID:          id,
		Brand:       normalized.Brand,
		Model:       normalized.Model,
		Year:        normalized.Year,
		Description: normalized.Description,
		Price:       normalized.Price,
		SellerID:    sellerID,
		CreatedAt:   s.clock.Now(),
	}
	if err := s.cars.Create(ctx, car); err != nil {
		return domain.Car{}, err
	}
	return car, nil
}

func (s *CarService) DeleteCar(ctx context.Context, id string) error {
	return s.cars.Delete(ctx, id)
}

func (s *CarService) ReportCar(ctx context.Context, id string) error {
	return s.cars.Report(ctx, id)
}
