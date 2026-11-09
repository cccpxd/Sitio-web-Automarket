package postgres

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"automarket/internal/domain"
)

// ModerationRepository applies administrative transitions and stores their audit
// trail in the same transaction.
type ModerationRepository struct {
	db *sql.DB
}

func NewModerationRepository(db *sql.DB) *ModerationRepository {
	return &ModerationRepository{db: db}
}

func (r *ModerationRepository) RemoveWithAudit(ctx context.Context, entry domain.ModerationEntry) error {
	return r.transitionWithAudit(ctx, entry, "removed", "'pending', 'active', 'rejected', 'withdrawn'", domain.ModerationActionRemoved)
}

func (r *ModerationRepository) ApproveWithAudit(ctx context.Context, entry domain.ModerationEntry) error {
	return r.transitionWithAudit(ctx, entry, "active", "'pending'", domain.ModerationActionApproved)
}

func (r *ModerationRepository) RejectWithAudit(ctx context.Context, entry domain.ModerationEntry) error {
	return r.transitionWithAudit(ctx, entry, "rejected", "'pending', 'active'", domain.ModerationActionRejected)
}

func (r *ModerationRepository) transitionWithAudit(ctx context.Context, entry domain.ModerationEntry, to, from string, action domain.ModerationAction) error {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("postgres: begin moderation transition: %w", err)
	}
	defer func() { _ = tx.Rollback() }()

	result, err := tx.ExecContext(ctx,
		fmt.Sprintf(`UPDATE cars SET status = '%s' WHERE id = $1 AND status IN (%s)`, to, from),
		entry.CarID,
	)
	if err != nil {
		return fmt.Errorf("postgres: transition car: %w", err)
	}
	affected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("postgres: transition car rows affected: %w", err)
	}
	if affected == 0 {
		var status string
		err := tx.QueryRowContext(ctx, `SELECT status FROM cars WHERE id = $1`, entry.CarID).Scan(&status)
		if errors.Is(err, sql.ErrNoRows) {
			return domain.ErrNotFound
		}
		if err != nil {
			return fmt.Errorf("postgres: find transition car: %w", err)
		}
		if action == domain.ModerationActionApproved && domain.CarStatus(status) == domain.CarStatusActive {
			return nil
		}
		return domain.ErrConflict
	}

	if _, err := tx.ExecContext(ctx, `
		INSERT INTO moderation_logs (id, admin_id, car_id, reason, action, created_at)
		VALUES ($1, $2, $3, $4, $5, $6)`,
		entry.ID, entry.AdminID, entry.CarID, entry.Reason, string(action), entry.CreatedAt,
	); err != nil {
		return fmt.Errorf("postgres: insert moderation transition log: %w", err)
	}
	if err := tx.Commit(); err != nil {
		return fmt.Errorf("postgres: commit moderation transition: %w", err)
	}
	return nil
}

func (r *ModerationRepository) List(ctx context.Context) ([]domain.ModerationEntry, error) {
	rows, err := r.db.QueryContext(ctx, `
		SELECT log.id, log.admin_id, admin.name, log.car_id,
		       c.brand || ' ' || c.model, log.reason, log.action, log.created_at
		FROM moderation_logs log
		JOIN users admin ON admin.id = log.admin_id
		JOIN cars c ON c.id = log.car_id
		ORDER BY log.created_at DESC, log.id DESC`)
	if err != nil {
		return nil, fmt.Errorf("postgres: list moderation log: %w", err)
	}
	defer rows.Close()

	var entries []domain.ModerationEntry
	for rows.Next() {
		var entry domain.ModerationEntry
		if err := rows.Scan(&entry.ID, &entry.AdminID, &entry.AdminName, &entry.CarID,
			&entry.CarTitle, &entry.Reason, &entry.Action, &entry.CreatedAt); err != nil {
			return nil, fmt.Errorf("postgres: scan moderation log: %w", err)
		}
		entry.CreatedAt = entry.CreatedAt.UTC()
		entries = append(entries, entry)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("postgres: iterate moderation log: %w", err)
	}
	return entries, nil
}
