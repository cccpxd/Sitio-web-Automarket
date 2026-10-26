package views

import (
	"time"

	"automarket/internal/domain"
)

// HomePage is the view model for the landing page.
type HomePage struct {
	Title    string
	Products []Product
}

type Product struct {
	Brand       string
	Model       string
	Year        int
	Description string
	Price       string
}

// AuthPage is the view model shared by login and registration screens.
type AuthPage struct {
	Title          string
	ActivePage     string
	ErrorMessage   string
	SuccessMessage string
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

type DashboardPage struct {
	Title    string
	User     UserView
	RoleText string
}

type RolePage struct {
	Title string
	User  UserView
	Cars  []Product
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

func NewDashboardPage(user domain.User) DashboardPage {
	return DashboardPage{
		Title:    "Dashboard",
		User:     NewUserView(user),
		RoleText: RoleLabel(user.Role),
	}
}

func NewRolePage(title string, user domain.User) RolePage {
	return RolePage{
		Title: title,
		User:  NewUserView(user),
	}
}

func NewRolePageWithCars(title string, user domain.User, cars []domain.Car) RolePage {
	products := make([]Product, 0, len(cars))
	for _, car := range cars {
		products = append(products, Product{
			Brand:       car.Brand,
			Model:       car.Model,
			Year:        car.Year,
			Description: car.Description,
			Price:       car.Price,
		})
	}
	return RolePage{
		Title: title,
		User:  NewUserView(user),
		Cars:  products,
	}
}

func NewHomePage(cars []domain.Car) HomePage {
	products := make([]Product, 0, len(cars))
	for _, car := range cars {
		products = append(products, Product{
			Brand:       car.Brand,
			Model:       car.Model,
			Year:        car.Year,
			Description: car.Description,
			Price:       car.Price,
		})
	}
	return HomePage{Title: "Automarket", Products: products}
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
