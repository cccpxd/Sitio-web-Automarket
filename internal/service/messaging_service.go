package service

import (
	"context"
	"errors"

	"automarket/internal/domain"
)

// MessagingService implements conversations between an interested user and the
// seller of a listing, including participant-only access.
type MessagingService struct {
	store ConversationStore
	cars  CarListStore
	ids   IDGenerator
	clock Clock
}

type MessagingServiceDeps struct {
	Store ConversationStore
	Cars  CarListStore
	IDs   IDGenerator
	Clock Clock
}

func NewMessagingService(deps MessagingServiceDeps) (*MessagingService, error) {
	if deps.Store == nil || deps.Cars == nil || deps.IDs == nil || deps.Clock == nil {
		return nil, errors.New("messaging service: missing dependency")
	}
	return &MessagingService{store: deps.Store, cars: deps.Cars, ids: deps.IDs, clock: deps.Clock}, nil
}

// Start opens (or reuses) the conversation for a listing and appends the first
// message. The seller cannot contact their own listing.
func (s *MessagingService) Start(ctx context.Context, carID, buyerID, body string) (domain.Conversation, error) {
	if buyerID == "" {
		return domain.Conversation{}, domain.ErrForbidden
	}
	car, err := s.cars.FindActive(ctx, carID)
	if err != nil {
		return domain.Conversation{}, err
	}
	if car.SellerID == buyerID {
		return domain.Conversation{}, domain.ErrSelfMessage
	}
	normalized, err := domain.NormalizeMessageBody(body)
	if err != nil {
		return domain.Conversation{}, err
	}

	now := s.clock.Now()
	conversationID, err := s.ids.New()
	if err != nil {
		return domain.Conversation{}, err
	}
	messageID, err := s.ids.New()
	if err != nil {
		return domain.Conversation{}, err
	}

	conversation := domain.Conversation{
		ID:        conversationID,
		CarID:     car.ID,
		Car:       car.Summary(),
		BuyerID:   buyerID,
		SellerID:  car.SellerID,
		CreatedAt: now,
	}
	message := domain.Message{
		ID:             messageID,
		ConversationID: conversationID,
		SenderID:       buyerID,
		Body:           normalized,
		CreatedAt:      now,
	}
	return s.store.Start(ctx, conversation, message)
}

// Get returns a conversation only to its participants or, read-only, to an
// administrator.
func (s *MessagingService) Get(ctx context.Context, id string, viewer domain.User) (domain.Conversation, error) {
	conversation, err := s.store.Find(ctx, id)
	if err != nil {
		return domain.Conversation{}, err
	}
	if viewer.Role != domain.RoleAdmin && !conversation.Participant(viewer.ID) {
		return domain.Conversation{}, domain.ErrForbidden
	}
	if conversation.Participant(viewer.ID) {
		if err := s.store.MarkRead(ctx, conversation.ID, viewer.ID, s.clock.Now()); err != nil {
			return domain.Conversation{}, err
		}
	}
	return conversation, nil
}

// List returns the conversations visible in the user's inbox.
func (s *MessagingService) List(ctx context.Context, userID string) ([]domain.Conversation, error) {
	if userID == "" {
		return nil, domain.ErrForbidden
	}
	return s.store.ListForUser(ctx, userID)
}

// Reply appends a message to a conversation. Administrators cannot reply.
func (s *MessagingService) Reply(ctx context.Context, id string, viewer domain.User, body string) (domain.Message, error) {
	conversation, err := s.store.Find(ctx, id)
	if err != nil {
		return domain.Message{}, err
	}
	if viewer.Role == domain.RoleAdmin {
		return domain.Message{}, domain.ErrConversationRead
	}
	if !conversation.Participant(viewer.ID) {
		return domain.Message{}, domain.ErrForbidden
	}
	normalized, err := domain.NormalizeMessageBody(body)
	if err != nil {
		return domain.Message{}, err
	}
	messageID, err := s.ids.New()
	if err != nil {
		return domain.Message{}, err
	}
	message := domain.Message{
		ID:             messageID,
		ConversationID: conversation.ID,
		SenderID:       viewer.ID,
		Body:           normalized,
		CreatedAt:      s.clock.Now(),
	}
	return s.store.AddMessage(ctx, message)
}

// InitialMessage exposes the prefilled contact text for the HTTP layer.
func (s *MessagingService) InitialMessage(car domain.Car) string {
	message := domain.InitialInterestMessage(car)
	if len([]rune(message)) > domain.MaxMessageLength {
		runes := []rune(message)
		message = string(runes[:domain.MaxMessageLength])
	}
	return message
}
