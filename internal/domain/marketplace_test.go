package domain

import (
	"errors"
	"strings"
	"testing"
	"time"
)

func TestPriceAndSearchNormalization(t *testing.T) {
	if got := FormatCOP(4_500_000_000); got != "$ 45.000.000" {
		t.Fatalf("FormatCOP() = %q", got)
	}
	if got, err := ParsePriceCOP("45.000.000"); err != nil || got != 4_500_000_000 {
		t.Fatalf("ParsePriceCOP() = %d, %v", got, err)
	}
	if _, err := ParsePriceCOP("45.5"); !errors.Is(err, ErrPriceInvalid) {
		t.Fatalf("decimal price error = %v", err)
	}
	filters, err := NormalizeSearchFilters(SearchQuery{
		Q: "  Mazda   2013 ", City: "  Bogotá  ", Year: "2013", MinPrice: "20.000.000", MaxPrice: "80.000.000", Page: "2",
	})
	if err != nil || len(filters.Tokens) != 2 || filters.City != "Bogotá" || filters.Page != 2 {
		t.Fatalf("normalized filters = %+v, %v", filters, err)
	}
	if _, err := NormalizeSearchFilters(SearchQuery{MinPrice: "90.000.000", MaxPrice: "80.000.000"}); !errors.Is(err, ErrFilterInvalid) {
		t.Fatalf("inverted range error = %v", err)
	}
}

func TestCarValidationAndEditWindow(t *testing.T) {
	now := time.Date(2026, time.November, 6, 12, 0, 0, 0, time.UTC)
	details, err := NormalizeCarInput(CarInput{
		Brand: " Mazda ", Model: " CX-5 ", City: " Bogotá ", Year: 2020, Description: "SUV", Price: "45.000.000",
	}, now)
	if err != nil || details.Brand != "Mazda" || details.City != "Bogotá" || details.PriceCents != 4_500_000_000 {
		t.Fatalf("valid car = %+v, %v", details, err)
	}
	if _, err := NormalizeCarInput(CarInput{Brand: "Mazda", Model: "3", City: "Bogotá", Year: 1899, Price: "1"}, now); !errors.Is(err, ErrYearInvalid) {
		t.Fatalf("invalid year error = %v", err)
	}
	if _, err := NormalizeCarInput(CarInput{Brand: "Mazda", Model: "3", City: "  ", Year: 2020, Price: "1"}, now); !errors.Is(err, ErrCityRequired) {
		t.Fatalf("empty city error = %v", err)
	}
	if _, err := NormalizeCarInput(CarInput{Brand: "Mazda", Model: "3", City: strings.Repeat("x", MaxCityLength+1), Year: 2020, Price: "1"}, now); !errors.Is(err, ErrCityTooLong) {
		t.Fatalf("long city error = %v", err)
	}
	car := Car{Status: CarStatusActive, CreatedAt: now}
	if !car.CanEdit(now.Add(EditWindow)) || car.CanEdit(now.Add(EditWindow+time.Nanosecond)) {
		t.Fatal("edit window boundary is incorrect")
	}
}

func TestPhotoAndMessageValidation(t *testing.T) {
	if err := ValidatePhotoUpload("image/jpeg", MaxPhotoBytes, 50, 50); err != nil {
		t.Fatalf("valid photo rejected: %v", err)
	}
	if err := ValidatePhotoUpload("image/jpeg", MaxPhotoBytes+1, 50, 50); !errors.Is(err, ErrPhotoTooLarge) {
		t.Fatalf("large photo error = %v", err)
	}
	if err := ValidatePhotoCount(MaxCarPhotos + 1); !errors.Is(err, ErrPhotosTooMany) {
		t.Fatalf("photo count error = %v", err)
	}
	if _, err := NormalizeMessageBody(strings.Repeat("x", MaxMessageLength+1)); !errors.Is(err, ErrMessageTooLong) {
		t.Fatalf("long message error = %v", err)
	}
	if _, err := NormalizeModerationReason(" "); !errors.Is(err, ErrReasonRequired) {
		t.Fatalf("empty reason error = %v", err)
	}
	car := Car{Brand: "Mazda", Model: "CX-5", Year: 2021, Description: strings.Repeat("x", InitialMessageLimit+1)}
	if got := InitialInterestMessage(car); got != "Estoy interesado en tu carro: Mazda CX-5 2021" {
		t.Fatalf("initial message = %q", got)
	}
}
