package handlers

import (
	"errors"
	"net/http"

	"automarket/internal/domain"
)

type errorPresentation struct {
	Status   int
	Message  string
	Internal bool
}

func presentRegistrationError(err error) errorPresentation {
	switch {
	case errors.Is(err, domain.ErrNameRequired):
		return errorPresentation{Status: http.StatusBadRequest, Message: "El nombre es obligatorio."}
	case errors.Is(err, domain.ErrNameTooLong):
		return errorPresentation{Status: http.StatusBadRequest, Message: "El nombre es demasiado largo."}
	case errors.Is(err, domain.ErrEmailRequired):
		return errorPresentation{Status: http.StatusBadRequest, Message: "El email es obligatorio."}
	case errors.Is(err, domain.ErrEmailInvalid):
		return errorPresentation{Status: http.StatusBadRequest, Message: "El email no es válido."}
	case errors.Is(err, domain.ErrEmailTaken):
		return errorPresentation{Status: http.StatusConflict, Message: "Ya existe un usuario con ese email."}
	case errors.Is(err, domain.ErrPasswordRequired):
		return errorPresentation{Status: http.StatusBadRequest, Message: "La contraseña es obligatoria."}
	case errors.Is(err, domain.ErrPasswordTooShort):
		return errorPresentation{Status: http.StatusBadRequest, Message: "La contraseña debe tener al menos 6 caracteres."}
	case errors.Is(err, domain.ErrPasswordTooLong):
		return errorPresentation{Status: http.StatusBadRequest, Message: "La contraseña es demasiado larga."}
	default:
		return errorPresentation{Status: http.StatusInternalServerError, Message: "No se pudo completar el registro.", Internal: true}
	}
}

func presentLoginError(err error) errorPresentation {
	if errors.Is(err, domain.ErrInvalidCredentials) {
		return errorPresentation{Status: http.StatusUnauthorized, Message: "Credenciales inválidas."}
	}
	return errorPresentation{Status: http.StatusInternalServerError, Message: "No se pudo iniciar sesión.", Internal: true}
}
