package postgres

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"

	"automarket/internal/domain"
	"github.com/jackc/pgx/v5/pgconn"
)

type CarRepository struct {
	db *sql.DB
}

func NewCarRepository(db *sql.DB) *CarRepository {
	return &CarRepository{db: db}
}

const carSelectColumns = `
	c.id, c.brand, c.model, c.year, COALESCE(c.city, ''), c.description, c.price_cents,
	c.seller_id, u.name, c.status, COALESCE(c.client_token, ''), c.created_at, c.edited_at,
	COALESCE(p.id, ''), COALESCE(p.content_type, ''), COALESCE(p.byte_size, 0),
	COALESCE(p.sha256, ''), COALESCE(p.width, 0), COALESCE(p.height, 0),
	COALESCE(p.created_at, 'epoch'::timestamptz)`

const carFromJoin = `
	FROM cars c
	JOIN users u ON u.id = c.seller_id
	LEFT JOIN LATERAL (
		SELECT id, content_type, byte_size, sha256, width, height, created_at
		FROM car_photos
		WHERE car_id = c.id
		ORDER BY position
		LIMIT 1
	) p ON TRUE`

func (r *CarRepository) CreateWithPhotos(ctx context.Context, car domain.Car, photos []domain.CarPhoto) error {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("postgres: begin create car: %w", err)
	}
	defer func() { _ = tx.Rollback() }()

	_, err = tx.ExecContext(ctx, `
		INSERT INTO cars (
			id, brand, model, year, city, description, price, seller_id, reported,
			status, price_cents, client_token, edited_at, created_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, FALSE, $9, $10, $11, $12, $13)`,
		car.ID, car.Brand, car.Model, car.Year, car.City, car.Description,
		strconv.FormatInt(car.PriceCents/100, 10),
		car.SellerID, string(car.Status), car.PriceCents, nullableString(car.ClientToken), car.EditedAt, car.CreatedAt,
	)
	if err != nil {
		if isConstraintViolation(err, "cars_seller_client_token_key") {
			return domain.ErrDuplicate
		}
		return fmt.Errorf("postgres: create car: %w", err)
	}

	if err := insertPhotos(ctx, tx, photos); err != nil {
		return err
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("postgres: commit create car: %w", err)
	}
	return nil
}

func (r *CarRepository) UpdateWithPhotos(ctx context.Context, car domain.Car, photos []domain.CarPhoto, replacePhotos bool, cutoff time.Time) error {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("postgres: begin update car: %w", err)
	}
	defer func() { _ = tx.Rollback() }()

	result, err := tx.ExecContext(ctx, `
		UPDATE cars
		SET brand = $1, model = $2, year = $3, city = $4, description = $5,
		    price_cents = $6, price = $7, edited_at = $8
		WHERE id = $9 AND seller_id = $10 AND status = 'active' AND created_at >= $11`,
		car.Brand, car.Model, car.Year, car.City, car.Description,
		car.PriceCents, strconv.FormatInt(car.PriceCents/100, 10), car.EditedAt,
		car.ID, car.SellerID, cutoff,
	)
	if err != nil {
		return fmt.Errorf("postgres: update car: %w", err)
	}
	affected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("postgres: update car rows affected: %w", err)
	}
	if affected == 0 {
		return r.classifyUpdateFailure(ctx, car.ID, car.SellerID)
	}

	if replacePhotos {
		if _, err := tx.ExecContext(ctx, `DELETE FROM car_photos WHERE car_id = $1`, car.ID); err != nil {
			return fmt.Errorf("postgres: replace car photos: %w", err)
		}
		if err := insertPhotos(ctx, tx, photos); err != nil {
			return err
		}
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("postgres: commit update car: %w", err)
	}
	return nil
}

func (r *CarRepository) classifyUpdateFailure(ctx context.Context, id, sellerID string) error {
	var owner string
	var status string
	var createdAt time.Time
	err := r.db.QueryRowContext(ctx, `SELECT seller_id, status, created_at FROM cars WHERE id = $1`, id).
		Scan(&owner, &status, &createdAt)
	if errors.Is(err, sql.ErrNoRows) {
		return domain.ErrNotFound
	}
	if err != nil {
		return fmt.Errorf("postgres: classify update failure: %w", err)
	}
	if owner != sellerID {
		return domain.ErrForbidden
	}
	if domain.CarStatus(status) != domain.CarStatusActive {
		return domain.ErrConflict
	}
	return domain.ErrEditWindowClosed
}

func (r *CarRepository) Withdraw(ctx context.Context, id, sellerID string, at time.Time) error {
	result, err := r.db.ExecContext(ctx,
		`UPDATE cars SET status = 'withdrawn' WHERE id = $1 AND seller_id = $2 AND status = 'active'`,
		id, sellerID,
	)
	if err != nil {
		return fmt.Errorf("postgres: withdraw car: %w", err)
	}
	affected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("postgres: withdraw car rows affected: %w", err)
	}
	if affected > 0 {
		return nil
	}

	var owner string
	var status string
	err = r.db.QueryRowContext(ctx, `SELECT seller_id, status FROM cars WHERE id = $1`, id).Scan(&owner, &status)
	if errors.Is(err, sql.ErrNoRows) {
		return domain.ErrNotFound
	}
	if err != nil {
		return fmt.Errorf("postgres: withdraw classify: %w", err)
	}
	if owner != sellerID {
		return domain.ErrForbidden
	}
	if domain.CarStatus(status) == domain.CarStatusWithdrawn {
		return nil
	}
	return domain.ErrConflict
}

func (r *CarRepository) FindActive(ctx context.Context, id string) (domain.Car, error) {
	car, err := r.findDetail(ctx, `WHERE c.id = $1 AND c.status = 'active'`, id)
	if errors.Is(err, sql.ErrNoRows) {
		return domain.Car{}, domain.ErrNotFound
	}
	return car, err
}

func (r *CarRepository) FindByID(ctx context.Context, id string) (domain.Car, error) {
	car, err := r.findDetail(ctx, `WHERE c.id = $1`, id)
	if errors.Is(err, sql.ErrNoRows) {
		return domain.Car{}, domain.ErrNotFound
	}
	return car, err
}

func (r *CarRepository) findDetail(ctx context.Context, where string, args ...any) (domain.Car, error) {
	query := `SELECT ` + carSelectColumns + carFromJoin + ` ` + where
	row := r.db.QueryRowContext(ctx, query, args...)
	car, err := scanCar(row)
	if err != nil {
		return domain.Car{}, err
	}
	photos, err := r.listPhotoMetadata(ctx, car.ID)
	if err != nil {
		return domain.Car{}, err
	}
	car.Photos = photos
	return car, nil
}

func (r *CarRepository) FindByClientToken(ctx context.Context, sellerID, token string) (domain.Car, error) {
	car, err := r.findDetail(ctx, `WHERE c.seller_id = $1 AND c.client_token = $2`, sellerID, token)
	if errors.Is(err, sql.ErrNoRows) {
		return domain.Car{}, domain.ErrNotFound
	}
	return car, err
}

func (r *CarRepository) Search(ctx context.Context, filters domain.SearchFilters) ([]domain.Car, int, error) {
	conditions := []string{
		"c.status = 'active'",
		"EXISTS (SELECT 1 FROM car_photos cover WHERE cover.car_id = c.id)",
	}
	args := make([]any, 0, len(filters.Tokens)+4)
	for _, token := range filters.Tokens {
		args = append(args, "%"+escapeLike(token)+"%")
		placeholder := fmt.Sprintf("$%d", len(args))
		conditions = append(conditions, fmt.Sprintf(
			"(c.brand ILIKE %[1]s OR c.model ILIKE %[1]s OR c.description ILIKE %[1]s OR c.year::text LIKE %[1]s)",
			placeholder,
		))
	}
	if filters.Year > 0 {
		args = append(args, filters.Year)
		conditions = append(conditions, fmt.Sprintf("c.year = $%d", len(args)))
	}
	if filters.City != "" {
		args = append(args, filters.City)
		conditions = append(conditions, fmt.Sprintf("LOWER(COALESCE(c.city, '')) = LOWER($%d)", len(args)))
	}
	if filters.MinPriceCents > 0 {
		args = append(args, filters.MinPriceCents)
		conditions = append(conditions, fmt.Sprintf("c.price_cents >= $%d", len(args)))
	}
	if filters.MaxPriceCents > 0 {
		args = append(args, filters.MaxPriceCents)
		conditions = append(conditions, fmt.Sprintf("c.price_cents <= $%d", len(args)))
	}
	where := "WHERE " + strings.Join(conditions, " AND ")

	var total int
	if err := r.db.QueryRowContext(ctx, `SELECT COUNT(*) FROM cars c `+where, args...).Scan(&total); err != nil {
		return nil, 0, fmt.Errorf("postgres: count cars: %w", err)
	}

	pageArgs := append([]any{}, args...)
	pageArgs = append(pageArgs, filters.PageSize, filters.Offset())
	query := `SELECT ` + carSelectColumns + carFromJoin + ` ` + where +
		fmt.Sprintf(` ORDER BY c.created_at DESC, c.id DESC LIMIT $%d OFFSET $%d`, len(args)+1, len(args)+2)

	cars, err := r.queryCars(ctx, query, pageArgs...)
	if err != nil {
		return nil, 0, err
	}
	return cars, total, nil
}

func (r *CarRepository) ListBySeller(ctx context.Context, sellerID string) ([]domain.Car, error) {
	query := `SELECT ` + carSelectColumns + carFromJoin +
		` WHERE c.seller_id = $1 ORDER BY c.created_at DESC`
	return r.queryCars(ctx, query, sellerID)
}

func (r *CarRepository) ListAll(ctx context.Context) ([]domain.Car, error) {
	query := `SELECT ` + carSelectColumns + carFromJoin + ` ORDER BY c.created_at DESC`
	return r.queryCars(ctx, query)
}

func (r *CarRepository) queryCars(ctx context.Context, query string, args ...any) ([]domain.Car, error) {
	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("postgres: list cars: %w", err)
	}
	defer rows.Close()

	var cars []domain.Car
	for rows.Next() {
		car, err := scanCar(rows)
		if err != nil {
			return nil, fmt.Errorf("postgres: scan car: %w", err)
		}
		cars = append(cars, car)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("postgres: iterate cars: %w", err)
	}
	return cars, nil
}

type carRowScanner interface {
	Scan(dest ...any) error
}

func scanCar(scanner carRowScanner) (domain.Car, error) {
	var car domain.Car
	var status string
	var editedAt sql.NullTime
	var photoID, photoType, photoSHA string
	var photoSize int64
	var photoWidth, photoHeight int
	var photoCreated time.Time

	err := scanner.Scan(
		&car.ID, &car.Brand, &car.Model, &car.Year, &car.City, &car.Description, &car.PriceCents,
		&car.SellerID, &car.SellerName, &status, &car.ClientToken, &car.CreatedAt, &editedAt,
		&photoID, &photoType, &photoSize, &photoSHA, &photoWidth, &photoHeight, &photoCreated,
	)
	if err != nil {
		return domain.Car{}, err
	}
	car.Status = domain.CarStatus(status)
	car.CreatedAt = car.CreatedAt.UTC()
	if editedAt.Valid {
		edited := editedAt.Time.UTC()
		car.EditedAt = &edited
	}
	if photoID != "" {
		car.Photos = []domain.CarPhoto{{
			ID:          photoID,
			CarID:       car.ID,
			Position:    0,
			ContentType: photoType,
			ByteSize:    photoSize,
			SHA256:      photoSHA,
			Width:       photoWidth,
			Height:      photoHeight,
			CreatedAt:   photoCreated.UTC(),
		}}
	}
	return car, nil
}

func (r *CarRepository) listPhotoMetadata(ctx context.Context, carID string) ([]domain.CarPhoto, error) {
	rows, err := r.db.QueryContext(ctx, `
		SELECT id, content_type, byte_size, sha256, width, height, created_at
		FROM car_photos WHERE car_id = $1 ORDER BY position`, carID)
	if err != nil {
		return nil, fmt.Errorf("postgres: list car photos: %w", err)
	}
	defer rows.Close()

	var photos []domain.CarPhoto
	for rows.Next() {
		photo := domain.CarPhoto{CarID: carID}
		if err := rows.Scan(&photo.ID, &photo.ContentType, &photo.ByteSize, &photo.SHA256,
			&photo.Width, &photo.Height, &photo.CreatedAt); err != nil {
			return nil, fmt.Errorf("postgres: scan car photo: %w", err)
		}
		photo.CreatedAt = photo.CreatedAt.UTC()
		photos = append(photos, photo)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("postgres: iterate car photos: %w", err)
	}
	return photos, nil
}

func insertPhotos(ctx context.Context, tx *sql.Tx, photos []domain.CarPhoto) error {
	for _, photo := range photos {
		_, err := tx.ExecContext(ctx, `
			INSERT INTO car_photos (
				id, car_id, position, content_type, byte_size, sha256,
				width, height, original_name, data, created_at
			) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)`,
			photo.ID, photo.CarID, photo.Position, photo.ContentType, photo.ByteSize,
			photo.SHA256, photo.Width, photo.Height, photo.OriginalName, photo.Data, photo.CreatedAt,
		)
		if err != nil {
			return fmt.Errorf("postgres: insert car photo: %w", err)
		}
	}
	return nil
}

func nullableString(value string) any {
	if strings.TrimSpace(value) == "" {
		return nil
	}
	return value
}

func escapeLike(value string) string {
	replacer := strings.NewReplacer(`\`, `\\`, `%`, `\%`, `_`, `\_`)
	return replacer.Replace(value)
}

func isConstraintViolation(err error, constraint string) bool {
	var pgErr *pgconn.PgError
	return errors.As(err, &pgErr) && pgErr.Code == "23505" && pgErr.ConstraintName == constraint
}
