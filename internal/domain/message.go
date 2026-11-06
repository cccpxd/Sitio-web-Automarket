package domain

import (
	"strings"
	"time"
	"unicode/utf8"
)

const (
	MaxMessageLength    = 1000
	InitialMessageLimit = 280
)

const initialMessagePrefix = "Estoy interesado en tu carro: "

// Conversation groups the messages exchanged about one listing between one
// interested buyer and the seller.
type Conversation struct {
	ID          string
	CarID       string
	Car         CarSummary
	BuyerID     string
	BuyerName   string
	SellerID    string
	SellerName  string
	CreatedAt   time.Time
	Messages    []Message
	LastMessage string
	UnreadCount int
}

// Participant reports whether the user takes part in the conversation.
func (c Conversation) Participant(userID string) bool {
	return userID != "" && (userID == c.BuyerID || userID == c.SellerID)
}

// Message is one entry in a conversation.
type Message struct {
	ID             string
	ConversationID string
	SenderID       string
	SenderName     string
	Body           string
	CreatedAt      time.Time
	ReadAt         *time.Time
}

// ModerationEntry is the immutable audit record of an administrative removal.
type ModerationEntry struct {
	ID        string
	AdminID   string
	AdminName string
	CarID     string
	CarTitle  string
	Reason    string
	CreatedAt time.Time
}

// NormalizeMessageBody trims and validates a message.
func NormalizeMessageBody(body string) (string, error) {
	body = strings.TrimSpace(body)
	if body == "" {
		return "", ErrMessageEmpty
	}
	if utf8.RuneCountInString(body) > MaxMessageLength {
		return "", ErrMessageTooLong
	}
	return body, nil
}

// NormalizeModerationReason trims and validates the mandatory reason.
func NormalizeModerationReason(reason string) (string, error) {
	reason = strings.TrimSpace(reason)
	if reason == "" {
		return "", ErrReasonRequired
	}
	if utf8.RuneCountInString(reason) > 500 {
		return "", ErrReasonTooLong
	}
	return reason, nil
}

// InitialInterestMessage builds the prefilled contact text. Long descriptions
// are replaced by a short brand/model/year reference.
func InitialInterestMessage(car Car) string {
	description := strings.TrimSpace(car.Description)
	if description != "" && utf8.RuneCountInString(description) <= InitialMessageLimit {
		return initialMessagePrefix + description
	}
	reference := car.Title()
	if car.Year > 0 {
		reference = strings.TrimSpace(reference + " " + itoa(car.Year))
	}
	if reference == "" {
		return initialMessagePrefix + "vehículo publicado"
	}
	return initialMessagePrefix + reference
}

func itoa(value int) string {
	if value == 0 {
		return "0"
	}
	negative := value < 0
	if negative {
		value = -value
	}
	var buffer [20]byte
	index := len(buffer)
	for value > 0 {
		index--
		buffer[index] = byte('0' + value%10)
		value /= 10
	}
	if negative {
		index--
		buffer[index] = '-'
	}
	return string(buffer[index:])
}
