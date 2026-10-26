package domain

import "time"

// Car represents a vehicle listed for sale.
type Car struct {
	ID          string
	Brand       string
	Model       string
	Year        int
	Description string
	Price       string
	SellerID    string
	Reported    bool
	CreatedAt   time.Time
}

// CarInput is the raw data used to create a car.
type CarInput struct {
	Brand       string
	Model       string
	Year        int
	Description string
	Price       string
}

func NormalizeCarInput(input CarInput) (CarInput, error) {
	input.Brand = normalizeName(input.Brand)
	if input.Brand == "" {
		return CarInput{}, ErrNameRequired
	}
	input.Model = normalizeName(input.Model)
	if input.Model == "" {
		return CarInput{}, ErrNameRequired
	}
	if input.Year < 1900 || input.Year > 2100 {
		return CarInput{}, ErrNameRequired
	}
	input.Description = normalizeName(input.Description)
	input.Price = normalizeName(input.Price)
	if input.Price == "" {
		return CarInput{}, ErrNameRequired
	}
	return input, nil
}
