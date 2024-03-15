package response

import (
	"fmt"
	"github.com/go-chi/render"
	"github.com/go-playground/validator/v10"
	"net/http"
	"strings"
)

type Response struct {
	Status string `json:"status"`
	Errors any    `json:"errors,omitempty"`
	Data   any    `json:"data,omitempty"`
}

const (
	StatusOK    = "ok"
	StatusError = "error"
)

func Ok(w http.ResponseWriter, r *http.Request, data any) {
	render.JSON(w, r, Response{
		Status: StatusOK,
		Data:   data,
	})
}

func Error(w http.ResponseWriter, r *http.Request, error any) {
	render.JSON(w, r, Response{
		Status: StatusError,
		Errors: error,
	})
}

type ValidationErr struct {
	Password string `json:"password,omitempty"`
}

func ValidationError(errs validator.ValidationErrors) any {

	var errMessage = map[string]string{}
	var message string
	var field string

	for _, err := range errs {

		field = strings.ToLower(err.Field())

		switch err.ActualTag() {
		case "required":
			message = fmt.Sprintf("field is a required")
		case "email":
			message = fmt.Sprintf("field is not a valid string")
		case "min":
			message = fmt.Sprintf("the field must contain at least %s characters", err.Param())
		case "eqfield":
			message = fmt.Sprintf("the field must contain at least %s characters", err.Param())
		default:
			message = fmt.Sprintf("field must be the same as field")
		}
		errMessage[field] = message
	}

	return errMessage
}
