package views

import (
	"fmt"
	"net/url"
	"strconv"
	"time"

	"automarket/internal/domain"
)

// BasePage groups the data shared by every HTML page.
type BasePage struct {
	Title        string
	CSRF         string
	User         *UserView
	ActivePage   string
	Flash        string
	ErrorMessage string
}

// NewBasePage builds the shared page data.
func NewBasePage(csrf string, user *domain.User, title, active string) BasePage {
	page := BasePage{Title: title, CSRF: csrf, ActivePage: active}
	if user != nil {
		view := NewUserView(*user)
		page.User = &view
	}
	return page
}

// AuthPage is the view model shared by login and registration screens.
type AuthPage struct {
	BasePage
	SuccessMessage string
	Next           string
}

// HomePage is retained as an alias for the catalog landing page.
type HomePage struct {
	BasePage
	Result   domain.SearchResult
	Filters  domain.SearchFilters
	Cards    []CarCard
	MinPrice string
	MaxPrice string
	PrevURL  string
	NextURL  string
	Query    string
	City     string
}

// CatalogPage is the public catalog and search result page.
type CatalogPage = HomePage

// CarCard is the compact representation used by catalog grids.
type CarCard struct {
	ID              string
	Title           string
	Year            int
	City            string
	Description     string
	Price           string
	CoverID         string
	PhotoCount      int
	Status          domain.CarStatus
	StatusLabel     string
	RejectionReason string
	SellerName      string
	CreatedAt       time.Time
	CanEdit         bool
	CanWithdraw     bool
	CanDelete       bool
	WithdrawURL     string
	DeleteURL       string
	PublicURL       string
}

// PhotoView is a public photo reference.
type PhotoView struct {
	ID    string
	Alt   string
	Index int
}

// CarDetailPage shows a single active listing.
type CarDetailPage struct {
	BasePage
	Car            CarView
	Photos         []PhotoView
	ContactMessage string
	IsOwner        bool
	CanContact     bool
	InitialMessage string
}

// ContactPage is the review screen for the initial interest message.
type ContactPage struct {
	BasePage
	Car          CarView
	Message      string
	Conversation string
}

// CarView is the detailed public representation of a listing.
type CarView struct {
	ID              string
	Brand           string
	Model           string
	Title           string
	Year            int
	City            string
	Description     string
	Price           string
	PriceCents      int64
	SellerName      string
	Status          domain.CarStatus
	StatusLabel     string
	RejectionReason string
	CreatedAt       time.Time
	EditedAt        *time.Time
	Photos          []PhotoView
}

// SalesPage is the seller's "Ventas" section.
type SalesPage struct {
	BasePage
	Result   domain.SearchResult
	Filters  domain.SearchFilters
	Cards    []CarCard
	Mine     []MyListing
	Query    string
	City     string
	Year     string
	MinPrice string
	MaxPrice string
	PrevURL  string
	NextURL  string
}

// MyListing is one of the seller's own listings, including withdrawn ones.
type MyListing struct {
	ID              string
	Title           string
	Year            int
	City            string
	Price           string
	Status          domain.CarStatus
	StatusLabel     string
	RejectionReason string
	CanWithdraw     bool
	CanDelete       bool
	CanEdit         bool
	PublicURL       string
	EditURL         string
	WithdrawURL     string
	DeleteURL       string
}

// CarForm preserves user input across validation errors.
type CarForm struct {
	Brand       string
	Model       string
	City        string
	Year        string
	Price       string
	Description string
	ClientToken string
}

// PublishPage renders the create or edit listing form.
type PublishPage struct {
	BasePage
	Form      CarForm
	Errors    map[string]string
	Photos    []PhotoView
	IsEdit    bool
	CarID     string
	MaxPhotos int
	MaxBytes  int64
}

// MessagesPage lists the user's conversations.
type MessagesPage struct {
	BasePage
	Conversations []ConversationView
}

// ConversationPage renders a single thread.
type ConversationPage struct {
	BasePage
	Conversation ConversationView
	ReadOnly     bool
	Error        string
}

// ConversationView is the safe representation of a thread.
type ConversationView struct {
	ID           string
	CarID        string
	CarTitle     string
	CarYear      int
	CarAvailable bool
	BuyerName    string
	SellerName   string
	UnreadCount  int
	LastMessage  string
	CreatedAt    time.Time
	Messages     []MessageView
	OtherName    string
}

// MessageView is one message in a thread.
type MessageView struct {
	ID         string
	SenderID   string
	SenderName string
	Body       string
	Class      string
	CreatedAt  time.Time
	ReadAt     *time.Time
	Mine       bool
}

// ModerationPage is the administrator panel and audit log.
type ModerationPage struct {
	BasePage
	Cars    []CarCard
	Pending []CarCard
	Entries []domain.ModerationEntry
}

// ModerationConfirmPage is the generic administrator action form.
type ModerationConfirmPage struct {
	BasePage
	Car         CarView
	Reason      string
	ActionTitle string
	ActionURL   string
	ReasonLabel string
	SubmitLabel string
	CancelURL   string
}

// ErrorPage is a friendly error page.
type ErrorPage struct {
	BasePage
	Status  int
	Message string
}

// DashboardPage is the authenticated landing page.
type DashboardPage struct {
	BasePage
	RoleText string
}

// RolePage is kept for compatibility with older handlers.
type RolePage struct {
	BasePage
	Sellers []UserView
	Cars    []CarCard
}

// UserView is the only user representation allowed in responses. It
// deliberately excludes PasswordHash.
type UserView struct {
	ID        string      `json:"id"`
	Name      string      `json:"name"`
	Email     string      `json:"email"`
	Role      domain.Role `json:"role"`
	CreatedAt time.Time   `json:"createdAt"`
}

func NewUserView(user domain.User) UserView {
	return UserView{
		ID:        user.ID,
		Name:      user.Name,
		Email:     user.Email,
		Role:      user.Role,
		CreatedAt: user.CreatedAt,
	}
}

func NewDashboardPage(csrf string, user *domain.User) DashboardPage {
	return DashboardPage{
		BasePage: NewBasePage(csrf, user, "Dashboard", "dashboard"),
		RoleText: RoleLabel(user.Role),
	}
}

func NewRolePage(csrf string, user *domain.User, sellers []domain.User, cars []domain.Car) RolePage {
	page := RolePage{BasePage: NewBasePage(csrf, user, "Panel de administración", "admin")}
	for _, seller := range sellers {
		page.Sellers = append(page.Sellers, NewUserView(seller))
	}
	for _, car := range cars {
		page.Cars = append(page.Cars, NewCarCard(car, time.Time{}))
	}
	return page
}

func NewCarCard(car domain.Car, now time.Time) CarCard {
	card := CarCard{
		ID:              car.ID,
		Title:           car.Title(),
		Year:            car.Year,
		City:            displayCity(car.City),
		Description:     car.Description,
		Price:           domain.FormatCOP(car.PriceCents),
		Status:          car.Status,
		StatusLabel:     StatusLabel(car.Status),
		RejectionReason: car.RejectionReason,
		CanWithdraw:     car.Status == domain.CarStatusPending || car.Status == domain.CarStatusActive || car.Status == domain.CarStatusRejected,
		CanDelete:       car.Status == domain.CarStatusWithdrawn || car.Status == domain.CarStatusRemoved,
		DeleteURL:       "/admin/publicaciones/" + car.ID + "/eliminar",
		SellerName:      car.SellerName,
		CreatedAt:       car.CreatedAt,
		PhotoCount:      len(car.Photos),
	}
	if len(car.Photos) > 0 {
		card.CoverID = car.Photos[0].ID
	}
	if !now.IsZero() {
		card.CanEdit = car.CanEdit(now)
	}
	return card
}

func NewMyListing(car domain.Car, now time.Time) MyListing {
	listing := MyListing{
		ID:              car.ID,
		Title:           car.Title(),
		Year:            car.Year,
		City:            displayCity(car.City),
		Price:           domain.FormatCOP(car.PriceCents),
		Status:          car.Status,
		StatusLabel:     StatusLabel(car.Status),
		RejectionReason: car.RejectionReason,
		CanWithdraw:     car.Status == domain.CarStatusPending || car.Status == domain.CarStatusActive || car.Status == domain.CarStatusRejected,
		CanDelete:       car.Status == domain.CarStatusWithdrawn || car.Status == domain.CarStatusRemoved,
		CanEdit:         car.CanEdit(now),
		PublicURL:       "/cars/" + car.ID,
		EditURL:         "/seller/publicaciones/" + car.ID + "/editar",
		WithdrawURL:     "/seller/publicaciones/" + car.ID + "/retirar",
		DeleteURL:       "/seller/publicaciones/" + car.ID + "/eliminar",
	}
	return listing
}

func NewCarView(car domain.Car) CarView {
	view := CarView{
		ID:              car.ID,
		Brand:           car.Brand,
		Model:           car.Model,
		City:            displayCity(car.City),
		Title:           car.Title(),
		Year:            car.Year,
		Description:     car.Description,
		Price:           domain.FormatCOP(car.PriceCents),
		PriceCents:      car.PriceCents,
		SellerName:      car.SellerName,
		Status:          car.Status,
		StatusLabel:     StatusLabel(car.Status),
		CreatedAt:       car.CreatedAt,
		EditedAt:        car.EditedAt,
		RejectionReason: car.RejectionReason,
	}
	for index, photo := range car.Photos {
		view.Photos = append(view.Photos, PhotoView{
			ID:    photo.ID,
			Alt:   fmt.Sprintf("%s %s, fotografía %d", car.Brand, car.Model, index+1),
			Index: index,
		})
	}
	return view
}

func NewConversationView(conversation domain.Conversation, viewerID string) ConversationView {
	view := ConversationView{
		ID:           conversation.ID,
		CarID:        conversation.CarID,
		CarTitle:     conversation.Car.Title,
		CarYear:      conversation.Car.Year,
		CarAvailable: conversation.Car.Status.IsActive(),
		BuyerName:    conversation.BuyerName,
		SellerName:   conversation.SellerName,
		UnreadCount:  conversation.UnreadCount,
		LastMessage:  conversation.LastMessage,
		CreatedAt:    conversation.CreatedAt,
	}
	if viewerID == conversation.BuyerID {
		view.OtherName = conversation.SellerName
	} else {
		view.OtherName = conversation.BuyerName
	}
	for _, message := range conversation.Messages {
		view.Messages = append(view.Messages, MessageView{
			ID:         message.ID,
			SenderID:   message.SenderID,
			SenderName: message.SenderName,
			Body:       message.Body,
			Class:      messageClass(message.SenderID == viewerID),
			CreatedAt:  message.CreatedAt,
			ReadAt:     message.ReadAt,
			Mine:       message.SenderID == viewerID,
		})
	}
	return view
}

func messageClass(mine bool) string {
	if mine {
		return "message message-mine"
	}
	return "message"
}

// NewCatalogPage builds the catalog page including pagination links.
func NewCatalogPage(csrf string, user *domain.User, result domain.SearchResult, filters domain.SearchFilters) CatalogPage {
	page := CatalogPage{
		BasePage: NewBasePage(csrf, user, "Catálogo", "catalog"),
		Result:   result,
		Filters:  filters,
		Query:    filters.Query,
		City:     filters.City,
	}
	if filters.MinPriceCents > 0 {
		page.MinPrice = domain.FormatCOPAmount(filters.MinPriceCents)
	}
	if filters.MaxPriceCents > 0 {
		page.MaxPrice = domain.FormatCOPAmount(filters.MaxPriceCents)
	}
	for _, car := range result.Cars {
		page.Cards = append(page.Cards, NewCarCard(car, time.Time{}))
	}
	page.PrevURL = paginationURL(filters, result.Page-1)
	page.NextURL = paginationURL(filters, result.Page+1)
	return page
}

func paginationURL(filters domain.SearchFilters, page int) string {
	values := url.Values{}
	if filters.Query != "" {
		values.Set("q", filters.Query)
	}
	if filters.Year > 0 {
		values.Set("year", strconv.Itoa(filters.Year))
	}
	if filters.City != "" {
		values.Set("city", filters.City)
	}
	if filters.MinPriceCents > 0 {
		values.Set("min", domain.FormatCOPAmount(filters.MinPriceCents))
	}

	if filters.MaxPriceCents > 0 {
		values.Set("max", domain.FormatCOPAmount(filters.MaxPriceCents))
	}
	if page > 1 {
		values.Set("page", strconv.Itoa(page))
	}
	if len(values) == 0 {
		return "/cars"
	}
	return "/cars?" + values.Encode()
}

func displayCity(city string) string {
	if city == "" {
		return "Ciudad no especificada"
	}
	return city
}

// StatusLabel returns the Spanish label for a listing status.
func StatusLabel(status domain.CarStatus) string {
	switch status {
	case domain.CarStatusActive:
		return "Activa"
	case domain.CarStatusPending:
		return "Pendiente de aprobación"
	case domain.CarStatusRejected:
		return "Rechazada"
	case domain.CarStatusWithdrawn:
		return "Retirada"
	case domain.CarStatusRemoved:
		return "Retirada por administración"
	default:
		return string(status)
	}
}

func RoleLabel(role domain.Role) string {
	switch role {
	case domain.RoleAdmin:
		return "Administrador"
	case domain.RoleSeller:
		return "Vendedor"
	case domain.RoleVisitor:
		return "Visitante"
	default:
		return "Desconocido"
	}
}
