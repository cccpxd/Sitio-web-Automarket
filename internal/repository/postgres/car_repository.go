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
	var sellerID any
	if car.SellerID != "" {
		sellerID = car.SellerID
	}

	_, err := r.db.ExecContext(ctx, `
		INSERT INTO cars (id, brand, model, year, description, price, seller_id, reported, created_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)`,
		car.ID, car.Brand, car.Model, car.Year, car.Description, car.Price, sellerID, car.Reported, car.CreatedAt,
	)
	if err != nil {
		return fmt.Errorf("postgres: create car: %w", err)
	}
	return nil
}

func (r *CarRepository) List(ctx context.Context) ([]domain.Car, error) {
	rows, err := r.db.QueryContext(ctx, `
		SELECT id, brand, model, year, description, price, COALESCE(seller_id, ''), reported, created_at
		FROM cars
		ORDER BY created_at DESC`)
	if err != nil {
		return nil, fmt.Errorf("postgres: list cars: %w", err)
	}
	defer rows.Close()

	var cars []domain.Car
	for rows.Next() {
		var car domain.Car
		if err := rows.Scan(&car.ID, &car.Brand, &car.Model, &car.Year, &car.Description, &car.Price, &car.SellerID, &car.Reported, &car.CreatedAt); err != nil {
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

func (r *CarRepository) Delete(ctx context.Context, id string) error {
	result, err := r.db.ExecContext(ctx, `DELETE FROM cars WHERE id = $1`, id)
	if err != nil {
		return fmt.Errorf("postgres: delete car: %w", err)
	}
	affected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("postgres: delete car rows affected: %w", err)
	}
	if affected == 0 {
		return domain.ErrNotFound
	}
	return nil
}

func (r *CarRepository) Report(ctx context.Context, id string) error {
	result, err := r.db.ExecContext(ctx, `UPDATE cars SET reported = TRUE WHERE id = $1`, id)
	if err != nil {
		return fmt.Errorf("postgres: report car: %w", err)
	}
	affected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("postgres: report car rows affected: %w", err)
	}
	if affected == 0 {
		return domain.ErrNotFound
	}
	return nil
}
