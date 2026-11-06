package domain

import (
	"fmt"
	"strconv"
	"strings"
	"time"
	"unicode/utf8"
)

// CarStatus is the lifecycle state of a vehicle listing. Listings are never
// hard-deleted; withdrawing or removing changes the status instead so that
// conversations and moderation evidence survive.
type CarStatus string

const (
	CarStatusActive    CarStatus = "active"
	CarStatusWithdrawn CarStatus = "withdrawn"
	CarStatusRemoved   CarStatus = "removed"
)

const (
	MinCarYear          = 1900
	MaxBrandLength      = 60
	MaxModelLength      = 80
	MaxCarDescription   = 2000
	MaxCarPhotos        = 6
	EditWindow          = 10 * time.Minute
	maxPriceWholeCOP    = int64(1_000_000_000_000)
)

// Car is a vehicle listing. PriceCents is the authoritative amount, expressed
// in cents of Colombian pesos (COP). The UI always renders whole COP.
type Car struct {
	ID          string
	Brand       string
	Model       string
	Year        int
	Description string
	PriceCents  int64
	SellerID    string
	SellerName  string
	Status      CarStatus
	ClientToken string
	CreatedAt   time.Time
	EditedAt    *time.Time
	Photos      []CarPhoto
}

// Title returns the short human-readable name of the listing.
func (c Car) Title() string {
	return strings.TrimSpace(c.Brand + " " + c.Model)
}

// CanEdit reports whether the seller may still edit the listing.
func (c Car) CanEdit(now time.Time) bool {
	return c.Status == CarStatusActive && !now.After(c.CreatedAt.Add(EditWindow))
}

// CarInput is the raw data used to create or update a car.
type CarInput struct {
	Brand       string
	Model       string
	Year        int
	Description string
	Price       string
}

// CarDetails contains validated car fields ready to be persisted.
type CarDetails struct {
	Brand       string
	Model       string
	Year        int
	Description string
	PriceCents  int64
}

func (s CarStatus) IsValid() bool {
	switch s {
	case CarStatusActive, CarStatusWithdrawn, CarStatusRemoved:
		return true
	default:
		return false
	}
}

func (s CarStatus) IsActive() bool { return s == CarStatusActive }

// NormalizeCarInput trims and validates user-provided listing data. now is
// used to bound the model year; future model years are accepted up to the next
// calendar year.
func NormalizeCarInput(input CarInput, now time.Time) (CarDetails, error) {
	details := CarDetails{
		Brand:       normalizeName(input.Brand),
		Model:       normalizeName(input.Model),
		Year:        input.Year,
		Description: strings.TrimSpace(input.Description),
	}

	if details.Brand == "" {
		return CarDetails{}, ErrBrandRequired
	}
	if utf8.RuneCountInString(details.Brand) > MaxBrandLength {
		return CarDetails{}, ErrBrandTooLong
	}
	if details.Model == "" {
		return CarDetails{}, ErrModelRequired
	}
	if utf8.RuneCountInString(details.Model) > MaxModelLength {
		return CarDetails{}, ErrModelTooLong
	}
	if details.Year < MinCarYear || details.Year > now.Year()+1 {
		return CarDetails{}, ErrYearInvalid
	}
	if utf8.RuneCountInString(details.Description) > MaxCarDescription {
		return CarDetails{}, ErrDescriptionLong
	}

	cents, err := ParsePriceCOP(input.Price)
	if err != nil {
		return CarDetails{}, err
	}
	details.PriceCents = cents

	return details, nil
}

// ParsePriceCOP accepts whole Colombian pesos, with optional dot thousands
// separators, and returns the amount in cents. Decimals are rejected because
// vehicle prices are published as whole COP.
func ParsePriceCOP(raw string) (int64, error) {
	value := strings.TrimSpace(raw)
	if value == "" {
		return 0, ErrPriceRequired
	}
	value = strings.ReplaceAll(value, " ", "")
	if strings.ContainsAny(value, ",-") {
		return 0, ErrPriceInvalid
	}
	if !validThousandsGrouping(value) {
		return 0, ErrPriceInvalid
	}
	digits := strings.ReplaceAll(value, ".", "")
	if digits == "" {
		return 0, ErrPriceInvalid
	}
	whole, err := strconv.ParseInt(digits, 10, 64)
	if err != nil || whole <= 0 || whole > maxPriceWholeCOP {
		return 0, ErrPriceInvalid
	}
	if whole > maxPriceWholeCOP/100 {
		return 0, ErrPriceInvalid
	}
	return whole * 100, nil
}

// FormatCOP renders cents as Colombian pesos with dot thousands separators and
// no decimals, prefixed with the currency symbol: "$ 45.000.000".
func FormatCOP(cents int64) string {
	return "$ " + FormatCOPAmount(cents)
}

// FormatCOPAmount renders cents as whole Colombian pesos with dot separators,
// without the currency symbol. It is used to prefill editable price fields.
func FormatCOPAmount(cents int64) string {
	if cents < 0 {
		cents = 0
	}
	whole := cents / 100
	digits := strconv.FormatInt(whole, 10)
	if len(digits) <= 3 {
		return digits
	}

	var builder strings.Builder
	first := len(digits) % 3
	if first == 0 {
		first = 3
	}
	builder.WriteString(digits[:first])
	for index := first; index < len(digits); index += 3 {
		builder.WriteByte('.')
		builder.WriteString(digits[index : index+3])
	}
	return builder.String()
}

func validThousandsGrouping(value string) bool {
	if value == "" {
		return false
	}
	parts := strings.Split(value, ".")
	if len(parts) == 1 {
		return allDigits(parts[0]) && len(parts[0]) > 0
	}
	if len(parts[0]) < 1 || len(parts[0]) > 3 || !allDigits(parts[0]) {
		return false
	}
	for _, part := range parts[1:] {
		if len(part) != 3 || !allDigits(part) {
			return false
		}
	}
	return true
}

func allDigits(value string) bool {
	if value == "" {
		return false
	}
	for _, char := range value {
		if char < '0' || char > '9' {
			return false
		}
	}
	return true
}

// CarSummary is a compact public representation used by messaging screens.
type CarSummary struct {
	ID     string
	Title  string
	Year   int
	Status CarStatus
}

func (c Car) Summary() CarSummary {
	return CarSummary{ID: c.ID, Title: c.Title(), Year: c.Year, Status: c.Status}
}

func (c CarSummary) DisplayTitle() string {
	if c.Year == 0 {
		return c.Title
	}
	return fmt.Sprintf("%s (%d)", c.Title, c.Year)
}
