package postgres

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"automarket/internal/domain"
	"github.com/jackc/pgx/v5/pgconn"
)

type UserRepository struct {
	db *sql.DB
}

func NewUserRepository(db *sql.DB) *UserRepository {
	return &UserRepository{db: db}
}

func (r *UserRepository) Create(ctx context.Context, user domain.User) error {
	_, err := r.db.ExecContext(ctx, `
		INSERT INTO users (id, name, email, password_hash, role, created_at)
		VALUES ($1, $2, $3, $4, $5, $6)`,
		user.ID, user.Name, user.Email, user.PasswordHash, string(user.Role), user.CreatedAt,
	)
	if isUniqueViolation(err) {
		return domain.ErrEmailTaken
	}
	if err != nil {
		return fmt.Errorf("postgres: create user: %w", err)
	}
	return nil
}

func (r *UserRepository) FindByEmail(ctx context.Context, email string) (domain.User, error) {
	row := r.db.QueryRowContext(ctx, `
		SELECT id, name, email, password_hash, role, created_at
		FROM users
		WHERE email = $1`, email)

	user, err := scanUser(row)
	if errors.Is(err, sql.ErrNoRows) {
		return domain.User{}, domain.ErrNotFound
	}
	if err != nil {
		return domain.User{}, fmt.Errorf("postgres: find user by email: %w", err)
	}
	return user, nil
}

type rowScanner interface {
	Scan(dest ...any) error
}

func scanUser(row rowScanner) (domain.User, error) {
	var user domain.User
	var role string

	if err := row.Scan(&user.ID, &user.Name, &user.Email, &user.PasswordHash, &role, &user.CreatedAt); err != nil {
		return domain.User{}, err
	}

	user.Role = domain.Role(role)
	if !user.Role.IsValid() {
		return domain.User{}, fmt.Errorf("postgres: invalid role %q", role)
	}
	user.CreatedAt = user.CreatedAt.UTC()
	return user, nil
}

func isUniqueViolation(err error) bool {
	var pgErr *pgconn.PgError
	return errors.As(err, &pgErr) &&
		pgErr.Code == "23505" &&
		pgErr.ConstraintName == "users_email_key"
}
