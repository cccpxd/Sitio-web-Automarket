package domain

import (
	"strconv"
	"strings"
)

const (
	DefaultPageSize    = 12
	MaxSearchTokens    = 5
	MaxSearchQueryLen  = 120
)

// SearchFilters is the normalized input for the catalog query. Tokens are
// derived from the free-text query and are always sent to PostgreSQL as bound
// parameters.
type SearchFilters struct {
	Query         string
	Tokens        []string
	Year          int
	MinPriceCents int64
	MaxPriceCents int64
	Page          int
	PageSize      int
}

// HasFilters reports whether any user-supplied criterion is active.
func (f SearchFilters) HasFilters() bool {
	return f.Query != "" || f.Year != 0 || f.MinPriceCents != 0 || f.MaxPriceCents != 0
}

// SearchQuery is the raw, untrusted query-string input.
type SearchQuery struct {
	Q        string
	Year     string
	MinPrice string
	MaxPrice string
	Page     string
}

// NormalizeSearchFilters trims, validates and normalizes the raw search input.
func NormalizeSearchFilters(raw SearchQuery) (SearchFilters, error) {
	query := strings.Join(strings.Fields(raw.Q), " ")
	if len([]rune(query)) > MaxSearchQueryLen {
		query = string([]rune(query)[:MaxSearchQueryLen])
	}

	tokens := strings.Fields(query)
	if len(tokens) > MaxSearchTokens {
		tokens = tokens[:MaxSearchTokens]
	}

	filters := SearchFilters{
		Query:    query,
		Tokens:   tokens,
		PageSize: DefaultPageSize,
		Page:     1,
	}

	if strings.TrimSpace(raw.Year) != "" {
		year, err := strconv.Atoi(strings.TrimSpace(raw.Year))
		if err != nil || year < MinCarYear || year > 3000 {
			return SearchFilters{}, ErrFilterInvalid
		}
		filters.Year = year
	}

	if strings.TrimSpace(raw.MinPrice) != "" {
		cents, err := ParsePriceCOP(raw.MinPrice)
		if err != nil {
			return SearchFilters{}, ErrFilterInvalid
		}
		filters.MinPriceCents = cents
	}
	if strings.TrimSpace(raw.MaxPrice) != "" {
		cents, err := ParsePriceCOP(raw.MaxPrice)
		if err != nil {
			return SearchFilters{}, ErrFilterInvalid
		}
		filters.MaxPriceCents = cents
	}
	if filters.MinPriceCents > 0 && filters.MaxPriceCents > 0 && filters.MinPriceCents > filters.MaxPriceCents {
		return SearchFilters{}, ErrFilterInvalid
	}

	if strings.TrimSpace(raw.Page) != "" {
		page, err := strconv.Atoi(strings.TrimSpace(raw.Page))
		if err != nil || page < 1 {
			return SearchFilters{}, ErrFilterInvalid
		}
		filters.Page = page
	}

	return filters, nil
}

// Offset is the number of rows to skip for the current page.
func (f SearchFilters) Offset() int {
	if f.Page <= 1 {
		return 0
	}
	return (f.Page - 1) * f.PageSize
}

// SearchResult is one page of active listings plus the total count.
type SearchResult struct {
	Cars       []Car
	Total      int
	Page       int
	PageSize   int
	TotalPages int
}

// NewSearchResult builds a result and derives the page count.
func NewSearchResult(cars []Car, total int, filters SearchFilters) SearchResult {
	totalPages := 0
	if filters.PageSize > 0 {
		totalPages = (total + filters.PageSize - 1) / filters.PageSize
	}
	return SearchResult{
		Cars:       cars,
		Total:      total,
		Page:       filters.Page,
		PageSize:   filters.PageSize,
		TotalPages: totalPages,
	}
}
