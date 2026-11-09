package postgres

import (
	"context"
	"database/sql"
	"fmt"
)

const migrationLockID int64 = 821_305_117

var migrationStatements = []string{
	`CREATE TABLE IF NOT EXISTS users (
		id TEXT PRIMARY KEY,
		name TEXT NOT NULL,
		email TEXT NOT NULL UNIQUE,
		password_hash TEXT NOT NULL,
		role TEXT NOT NULL CHECK (role IN ('seller', 'admin')),
		created_at TIMESTAMPTZ NOT NULL
	)`,
	`CREATE TABLE IF NOT EXISTS sessions (
		token_hash TEXT PRIMARY KEY,
		user_id TEXT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
		created_at TIMESTAMPTZ NOT NULL,
		expires_at TIMESTAMPTZ NOT NULL
	)`,
	`CREATE INDEX IF NOT EXISTS sessions_user_id_idx ON sessions (user_id)`,
	`CREATE INDEX IF NOT EXISTS sessions_expires_at_idx ON sessions (expires_at)`,
	`CREATE TABLE IF NOT EXISTS cars (
		id TEXT PRIMARY KEY,
		brand TEXT NOT NULL,
		model TEXT NOT NULL,
		year INTEGER NOT NULL,
		description TEXT NOT NULL DEFAULT '',
		price TEXT NOT NULL,
		seller_id TEXT REFERENCES users(id) ON DELETE SET NULL,
		reported BOOLEAN NOT NULL DEFAULT FALSE,
		created_at TIMESTAMPTZ NOT NULL
	)`,
	`ALTER TABLE cars ADD COLUMN IF NOT EXISTS reported BOOLEAN NOT NULL DEFAULT FALSE`,
	`ALTER TABLE cars ADD COLUMN IF NOT EXISTS status TEXT NOT NULL DEFAULT 'active'`,
	`ALTER TABLE cars ADD COLUMN IF NOT EXISTS price_cents BIGINT`,
	`ALTER TABLE cars ADD COLUMN IF NOT EXISTS client_token TEXT`,
	`ALTER TABLE cars ADD COLUMN IF NOT EXISTS edited_at TIMESTAMPTZ`,
	`ALTER TABLE cars ADD COLUMN IF NOT EXISTS city TEXT NOT NULL DEFAULT ''`,

	`CREATE TABLE IF NOT EXISTS car_photos (
		id TEXT PRIMARY KEY,
		car_id TEXT NOT NULL REFERENCES cars(id) ON DELETE CASCADE,
		position INTEGER NOT NULL CHECK (position >= 0),
		content_type TEXT NOT NULL,
		byte_size INTEGER NOT NULL CHECK (byte_size > 0),
		sha256 TEXT NOT NULL,
		width INTEGER NOT NULL CHECK (width > 0),
		height INTEGER NOT NULL CHECK (height > 0),
		original_name TEXT NOT NULL DEFAULT '',
		data BYTEA NOT NULL,
		created_at TIMESTAMPTZ NOT NULL,
		UNIQUE (car_id, position)
	)`,
	`CREATE INDEX IF NOT EXISTS car_photos_car_id_idx ON car_photos (car_id)`,

	`CREATE TABLE IF NOT EXISTS conversations (
		id TEXT PRIMARY KEY,
		car_id TEXT NOT NULL REFERENCES cars(id) ON DELETE RESTRICT,
		buyer_id TEXT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
		seller_id TEXT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
		created_at TIMESTAMPTZ NOT NULL,
		UNIQUE (car_id, buyer_id),
		CHECK (buyer_id <> seller_id)
	)`,
	`CREATE INDEX IF NOT EXISTS conversations_buyer_id_idx ON conversations (buyer_id)`,
	`CREATE INDEX IF NOT EXISTS conversations_seller_id_idx ON conversations (seller_id)`,

	`CREATE TABLE IF NOT EXISTS messages (
		id TEXT PRIMARY KEY,
		conversation_id TEXT NOT NULL REFERENCES conversations(id) ON DELETE CASCADE,
		sender_id TEXT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
		body TEXT NOT NULL CHECK (length(btrim(body)) > 0),
		created_at TIMESTAMPTZ NOT NULL,
		read_at TIMESTAMPTZ
	)`,
	`CREATE INDEX IF NOT EXISTS messages_conversation_id_idx ON messages (conversation_id, created_at)`,

	`CREATE TABLE IF NOT EXISTS moderation_logs (
		id TEXT PRIMARY KEY,
		admin_id TEXT NOT NULL REFERENCES users(id) ON DELETE RESTRICT,
		car_id TEXT NOT NULL REFERENCES cars(id) ON DELETE RESTRICT,
		reason TEXT NOT NULL,
		action TEXT NOT NULL DEFAULT 'removed',
		created_at TIMESTAMPTZ NOT NULL
	)`,
	`ALTER TABLE moderation_logs ADD COLUMN IF NOT EXISTS action TEXT NOT NULL DEFAULT 'removed'`,
	`ALTER TABLE moderation_logs DROP CONSTRAINT IF EXISTS moderation_logs_reason_check`,
	`DO $$
	 BEGIN
	   IF NOT EXISTS (SELECT 1 FROM pg_constraint WHERE conname = 'moderation_logs_reason_check') THEN
	     ALTER TABLE moderation_logs ADD CONSTRAINT moderation_logs_reason_check
	       CHECK (action = 'approved' OR length(btrim(reason)) > 0);
	   END IF;
	 END $$`,
	`CREATE INDEX IF NOT EXISTS moderation_logs_car_id_idx ON moderation_logs (car_id)`,
	`CREATE INDEX IF NOT EXISTS moderation_logs_created_at_idx ON moderation_logs (created_at DESC)`,

	// Remove the six vehicles seeded by older versions before requiring a seller.
	`DELETE FROM cars
	 WHERE seller_id IS NULL
	   AND (brand, model, year) IN (
		 ('Toyota', 'Corolla', 2020),
		 ('Honda', 'Civic', 2019),
		 ('Mazda', 'CX-5', 2021),
		 ('Nissan', 'Versa', 2018),
		 ('Ford', 'Ranger', 2022),
		 ('Chevrolet', 'Spark', 2017)
	   )`,

	// Backfill legacy TEXT prices, interpreted as whole Colombian pesos.
	`UPDATE cars
	 SET price_cents = (regexp_replace(price, '[^0-9]', '', 'g'))::BIGINT * 100
	 WHERE price_cents IS NULL
	   AND regexp_replace(price, '[^0-9]', '', 'g') <> ''`,

	`ALTER TABLE cars ALTER COLUMN seller_id SET NOT NULL`,
	`ALTER TABLE cars ALTER COLUMN price_cents SET NOT NULL`,

	`ALTER TABLE cars DROP CONSTRAINT IF EXISTS cars_status_check`,
	`DO $$
	 BEGIN
	   IF NOT EXISTS (SELECT 1 FROM pg_constraint WHERE conname = 'cars_status_check') THEN
	     ALTER TABLE cars ADD CONSTRAINT cars_status_check CHECK (status IN ('pending', 'active', 'rejected', 'withdrawn', 'removed'));
	   END IF;
	 END $$`,
	`DO $$
	 BEGIN
	   IF NOT EXISTS (SELECT 1 FROM pg_constraint WHERE conname = 'cars_year_check') THEN
	     ALTER TABLE cars ADD CONSTRAINT cars_year_check CHECK (year >= 1900);
	   END IF;
	 END $$`,
	`DO $$
	 BEGIN
	   IF NOT EXISTS (SELECT 1 FROM pg_constraint WHERE conname = 'cars_price_cents_check') THEN
	     ALTER TABLE cars ADD CONSTRAINT cars_price_cents_check CHECK (price_cents > 0);
	   END IF;
	 END $$`,

	`CREATE INDEX IF NOT EXISTS cars_status_created_at_idx ON cars (status, created_at DESC)`,
	`CREATE INDEX IF NOT EXISTS cars_status_year_idx ON cars (status, year)`,
	`CREATE INDEX IF NOT EXISTS cars_status_price_idx ON cars (status, price_cents)`,
	`CREATE INDEX IF NOT EXISTS cars_seller_id_idx ON cars (seller_id)`,
	`CREATE UNIQUE INDEX IF NOT EXISTS cars_seller_client_token_key
	 ON cars (seller_id, client_token) WHERE client_token IS NOT NULL`,
}

// Migrate applies the idempotent schema. The advisory lock prevents concurrent
// cold starts from running DDL at the same time.
func Migrate(ctx context.Context, db *sql.DB) error {
	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("postgres: begin migration: %w", err)
	}
	defer func() { _ = tx.Rollback() }()

	if _, err := tx.ExecContext(ctx, `SELECT pg_advisory_xact_lock($1)`, migrationLockID); err != nil {
		return fmt.Errorf("postgres: acquire migration lock: %w", err)
	}

	for _, statement := range migrationStatements {
		if _, err := tx.ExecContext(ctx, statement); err != nil {
			return fmt.Errorf("postgres: apply migration: %w", err)
		}
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("postgres: commit migration: %w", err)
	}
	return nil
}
