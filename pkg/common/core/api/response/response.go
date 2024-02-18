package response

import (
	"fmt"
	"github.com/go-chi/render"
	"github.com/go-playground/validator/v10"
	"net/http"
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
	for _, err := range errs {

		switch err.ActualTag() {
		case "required":
			errMessage[err.Field()] = fmt.Sprintf("field is a required")
		case "email":
			errMessage[err.Field()] = fmt.Sprintf("field is not a valid string")
		case "min":
			errMessage[err.Field()] = fmt.Sprintf("the field must contain at least %s characters", err.Param())
		case "eqfield":
			errMessage[err.Field()] = fmt.Sprintf("the field must contain at least %s characters", err.Param())
		default:
			errMessage[err.Field()] = fmt.Sprintf("field must be the same as field")
		}
	}

	return errMessage
}
