package postgres

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"automarket/internal/domain"
)

// PhotoRepository reads photo metadata and binary content separately so that
// catalog queries never load image blobs.
type PhotoRepository struct {
	db *sql.DB
}

func NewPhotoRepository(db *sql.DB) *PhotoRepository {
	return &PhotoRepository{db: db}
}

func (r *PhotoRepository) FindMeta(ctx context.Context, id string) (domain.CarPhoto, error) {
	row := r.db.QueryRowContext(ctx, `
		SELECT p.id, p.car_id, p.position, p.content_type, p.byte_size, p.sha256,
		       p.width, p.height, p.original_name, p.created_at,
		       c.status, c.seller_id
		FROM car_photos p
		JOIN cars c ON c.id = p.car_id
		WHERE p.id = $1`, id)

	var photo domain.CarPhoto
	var status string
	err := row.Scan(
		&photo.ID, &photo.CarID, &photo.Position, &photo.ContentType, &photo.ByteSize, &photo.SHA256,
		&photo.Width, &photo.Height, &photo.OriginalName, &photo.CreatedAt,
		&status, &photo.CarSellerID,
	)
	if errors.Is(err, sql.ErrNoRows) {
		return domain.CarPhoto{}, domain.ErrNotFound
	}
	if err != nil {
		return domain.CarPhoto{}, fmt.Errorf("postgres: find photo metadata: %w", err)
	}
	photo.CarStatus = domain.CarStatus(status)
	photo.CreatedAt = photo.CreatedAt.UTC()
	return photo, nil
}

func (r *PhotoRepository) FindData(ctx context.Context, id string) (domain.CarPhoto, error) {
	row := r.db.QueryRowContext(ctx, `
		SELECT p.id, p.car_id, p.position, p.content_type, p.byte_size, p.sha256,
		       p.width, p.height, p.original_name, p.data, p.created_at,
		       c.status, c.seller_id
		FROM car_photos p
		JOIN cars c ON c.id = p.car_id
		WHERE p.id = $1`, id)

	var photo domain.CarPhoto
	var status string
	err := row.Scan(
		&photo.ID, &photo.CarID, &photo.Position, &photo.ContentType, &photo.ByteSize, &photo.SHA256,
		&photo.Width, &photo.Height, &photo.OriginalName, &photo.Data, &photo.CreatedAt,
		&status, &photo.CarSellerID,
	)
	if errors.Is(err, sql.ErrNoRows) {
		return domain.CarPhoto{}, domain.ErrNotFound
	}
	if err != nil {
		return domain.CarPhoto{}, fmt.Errorf("postgres: find photo data: %w", err)
	}
	photo.CarStatus = domain.CarStatus(status)
	photo.CreatedAt = photo.CreatedAt.UTC()
	return photo, nil
}
