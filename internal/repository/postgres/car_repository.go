package postgres

import (
	"context"
	"database/sql"
	"fmt"

	"automarket/internal/domain"
)

type CarRepository struct {
	db *sql.DB
}

func NewCarRepository(db *sql.DB) *CarRepository {
	return &CarRepository{db: db}
}

func (r *CarRepository) Create(ctx context.Context, car domain.Car) error {
	_, err := r.db.ExecContext(ctx, `
		INSERT INTO cars (id, brand, model, year, description, price, seller_id, created_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)`,
		car.ID, car.Brand, car.Model, car.Year, car.Description, car.Price, car.SellerID, car.CreatedAt,
	)
	if err != nil {
		return fmt.Errorf("postgres: create car: %w", err)
	}
	return nil
}

func (r *CarRepository) List(ctx context.Context) ([]domain.Car, error) {
	rows, err := r.db.QueryContext(ctx, `
		SELECT id, brand, model, year, description, price, seller_id, created_at
		FROM cars
		ORDER BY created_at DESC`)
	if err != nil {
		return nil, fmt.Errorf("postgres: list cars: %w", err)
	}
	defer rows.Close()

	var cars []domain.Car
	for rows.Next() {
		var car domain.Car
		if err := rows.Scan(&car.ID, &car.Brand, &car.Model, &car.Year, &car.Description, &car.Price, &car.SellerID, &car.CreatedAt); err != nil {
			return nil, fmt.Errorf("postgres: scan car: %w", err)
		}
		car.CreatedAt = car.CreatedAt.UTC()
		cars = append(cars, car)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("postgres: iterate cars: %w", err)
	}
	return cars, nil
}
