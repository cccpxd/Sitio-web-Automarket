package postgres

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"automarket/internal/domain"
)

// ConversationRepository persists conversations and messages. Queries always
// scope access by participant before returning rows.
type ConversationRepository struct {
	db *sql.DB
}

func NewConversationRepository(db *sql.DB) *ConversationRepository {
	return &ConversationRepository{db: db}
}

func (r *ConversationRepository) Start(ctx context.Context, conversation domain.Conversation, first domain.Message) (domain.Conversation, error) {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return domain.Conversation{}, fmt.Errorf("postgres: begin conversation: %w", err)
	}
	defer func() { _ = tx.Rollback() }()

	_, err = tx.ExecContext(ctx, `
		INSERT INTO conversations (id, car_id, buyer_id, seller_id, created_at)
		VALUES ($1, $2, $3, $4, $5)
		ON CONFLICT (car_id, buyer_id) DO NOTHING`,
		conversation.ID, conversation.CarID, conversation.BuyerID, conversation.SellerID, conversation.CreatedAt,
	)
	if err != nil {
		return domain.Conversation{}, fmt.Errorf("postgres: create conversation: %w", err)
	}

	// ON CONFLICT DO NOTHING leaves the transaction usable, allowing the
	// existing conversation to be reused without losing the new message.
	var createdAt time.Time
	if err := tx.QueryRowContext(ctx,
		`SELECT id, created_at FROM conversations WHERE car_id = $1 AND buyer_id = $2`,
		conversation.CarID, conversation.BuyerID,
	).Scan(&conversation.ID, &createdAt); err != nil {
		return domain.Conversation{}, fmt.Errorf("postgres: find conversation after insert: %w", err)
	}
	conversation.CreatedAt = createdAt

	first.ConversationID = conversation.ID
	if _, err := tx.ExecContext(ctx, `
		INSERT INTO messages (id, conversation_id, sender_id, body, created_at)
		VALUES ($1, $2, $3, $4, $5)`,
		first.ID, first.ConversationID, first.SenderID, first.Body, first.CreatedAt,
	); err != nil {
		return domain.Conversation{}, fmt.Errorf("postgres: create first message: %w", err)
	}

	if err := tx.Commit(); err != nil {
		return domain.Conversation{}, fmt.Errorf("postgres: commit conversation: %w", err)
	}
	conversation.Messages = []domain.Message{first}
	conversation.LastMessage = first.Body
	return conversation, nil
}

func (r *ConversationRepository) Find(ctx context.Context, id string) (domain.Conversation, error) {
	row := r.db.QueryRowContext(ctx, `
		SELECT conv.id, conv.car_id, conv.buyer_id, buyer.name, conv.seller_id, seller.name,
		       conv.created_at, c.brand, c.model, c.year, c.status
		FROM conversations conv
		JOIN users buyer ON buyer.id = conv.buyer_id
		JOIN users seller ON seller.id = conv.seller_id
		JOIN cars c ON c.id = conv.car_id
		WHERE conv.id = $1`, id)

	conversation, err := scanConversation(row)
	if errors.Is(err, sql.ErrNoRows) {
		return domain.Conversation{}, domain.ErrNotFound
	}
	if err != nil {
		return domain.Conversation{}, err
	}

	messages, err := r.listMessages(ctx, conversation.ID)
	if err != nil {
		return domain.Conversation{}, err
	}
	conversation.Messages = messages
	conversation.LastMessage = lastMessageBody(messages)
	return conversation, nil
}

func (r *ConversationRepository) ListForUser(ctx context.Context, userID string) ([]domain.Conversation, error) {
	rows, err := r.db.QueryContext(ctx, `
		SELECT conv.id, conv.car_id, conv.buyer_id, buyer.name, conv.seller_id, seller.name,
		       conv.created_at, c.brand, c.model, c.year, c.status,
		       COALESCE((
		           SELECT m.body FROM messages m
		           WHERE m.conversation_id = conv.id
		           ORDER BY m.created_at DESC, m.id DESC LIMIT 1
		       ), ''),
		       (
		           SELECT COUNT(*) FROM messages m
		           WHERE m.conversation_id = conv.id AND m.sender_id <> $1 AND m.read_at IS NULL
		       )
		FROM conversations conv
		JOIN users buyer ON buyer.id = conv.buyer_id
		JOIN users seller ON seller.id = conv.seller_id
		JOIN cars c ON c.id = conv.car_id
		WHERE conv.buyer_id = $1 OR conv.seller_id = $1
		ORDER BY conv.created_at DESC`, userID)
	if err != nil {
		return nil, fmt.Errorf("postgres: list conversations: %w", err)
	}
	defer rows.Close()

	var conversations []domain.Conversation
	for rows.Next() {
		conversation, err := scanConversationWithExtras(rows)
		if err != nil {
			return nil, fmt.Errorf("postgres: scan conversation: %w", err)
		}
		conversations = append(conversations, conversation)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("postgres: iterate conversations: %w", err)
	}
	return conversations, nil
}

func (r *ConversationRepository) AddMessage(ctx context.Context, message domain.Message) (domain.Message, error) {
	_, err := r.db.ExecContext(ctx, `
		INSERT INTO messages (id, conversation_id, sender_id, body, created_at)
		VALUES ($1, $2, $3, $4, $5)`,
		message.ID, message.ConversationID, message.SenderID, message.Body, message.CreatedAt,
	)
	if err != nil {
		return domain.Message{}, fmt.Errorf("postgres: add message: %w", err)
	}
	return message, nil
}

func (r *ConversationRepository) MarkRead(ctx context.Context, conversationID, readerID string, at time.Time) error {
	_, err := r.db.ExecContext(ctx, `
		UPDATE messages SET read_at = $3
		WHERE conversation_id = $1 AND sender_id <> $2 AND read_at IS NULL`,
		conversationID, readerID, at)
	if err != nil {
		return fmt.Errorf("postgres: mark messages read: %w", err)
	}
	return nil
}

type conversationScanner interface {
	Scan(dest ...any) error
}

func scanConversation(scanner conversationScanner) (domain.Conversation, error) {
	var conversation domain.Conversation
	var status string
	var brand, model string
	err := scanner.Scan(
		&conversation.ID, &conversation.CarID, &conversation.BuyerID, &conversation.BuyerName,
		&conversation.SellerID, &conversation.SellerName, &conversation.CreatedAt,
		&brand, &model, &conversation.Car.Year, &status,
	)
	if err != nil {
		return domain.Conversation{}, err
	}
	conversation.Car = domain.CarSummary{
		ID:     conversation.CarID,
		Title:  brand + " " + model,
		Year:   conversation.Car.Year,
		Status: domain.CarStatus(status),
	}
	conversation.CreatedAt = conversation.CreatedAt.UTC()
	return conversation, nil
}

func scanConversationWithExtras(scanner conversationScanner) (domain.Conversation, error) {
	var conversation domain.Conversation
	var status string
	var brand, model string
	err := scanner.Scan(
		&conversation.ID, &conversation.CarID, &conversation.BuyerID, &conversation.BuyerName,
		&conversation.SellerID, &conversation.SellerName, &conversation.CreatedAt,
		&brand, &model, &conversation.Car.Year, &status,
		&conversation.LastMessage, &conversation.UnreadCount,
	)
	if err != nil {
		return domain.Conversation{}, err
	}
	conversation.Car = domain.CarSummary{
		ID:     conversation.CarID,
		Title:  brand + " " + model,
		Year:   conversation.Car.Year,
		Status: domain.CarStatus(status),
	}
	conversation.CreatedAt = conversation.CreatedAt.UTC()
	return conversation, nil
}

func (r *ConversationRepository) listMessages(ctx context.Context, conversationID string) ([]domain.Message, error) {
	rows, err := r.db.QueryContext(ctx, `
		SELECT m.id, m.conversation_id, m.sender_id, u.name, m.body, m.created_at, m.read_at
		FROM messages m
		JOIN users u ON u.id = m.sender_id
		WHERE m.conversation_id = $1
		ORDER BY m.created_at, m.id`, conversationID)
	if err != nil {
		return nil, fmt.Errorf("postgres: list messages: %w", err)
	}
	defer rows.Close()

	var messages []domain.Message
	for rows.Next() {
		var message domain.Message
		var readAt sql.NullTime
		if err := rows.Scan(&message.ID, &message.ConversationID, &message.SenderID,
			&message.SenderName, &message.Body, &message.CreatedAt, &readAt); err != nil {
			return nil, fmt.Errorf("postgres: scan message: %w", err)
		}
		message.CreatedAt = message.CreatedAt.UTC()
		if readAt.Valid {
			read := readAt.Time.UTC()
			message.ReadAt = &read
		}
		messages = append(messages, message)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("postgres: iterate messages: %w", err)
	}
	return messages, nil
}

func lastMessageBody(messages []domain.Message) string {
	if len(messages) == 0 {
		return ""
	}
	return messages[len(messages)-1].Body
}
