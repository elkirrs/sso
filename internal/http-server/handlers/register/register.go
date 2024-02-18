package register

import (
	"app/internal/domain/user"
	"app/internal/storage"
	resp "app/pkg/common/core/api/response"
	"app/pkg/common/core/identity"
	"app/pkg/utils/crypt"
	"errors"
	"github.com/go-chi/chi/middleware"
	"github.com/go-chi/render"
	"github.com/go-playground/validator/v10"
	"io"
	"log/slog"
	"net/http"
	"time"
)

type Auth interface {
	Registration(req *user.CreateUser) (err error)
}
type Request struct {
	Name            string `json:"name" validate:"required,ascii"`
	Email           string `json:"email" validate:"required,email"`
	Password        string `json:"password" validate:"required,ascii,min=9"`
	ConfirmPassword string `json:"confirm_password" validate:"required,eqfield=Password"`
}

type Response struct {
	Message string `json:"message,omitempty"`
}

func New(
	log *slog.Logger,
	auth Auth,
) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		const op = "http-server.handlers.register.New"

		log := log.With(
			slog.String("op", op),
			slog.String("request_id", middleware.GetReqID(r.Context())),
		)
		var req Request

		err := render.DecodeJSON(r.Body, &req)
		if errors.Is(err, io.EOF) {
			log.Error("request body is empty")
			var dR = &Response{Message: "empty request"}
			resp.Error(w, r, dR)
			return
		}

		if err != nil {
			log.Error("failed to decode request body", err)
			var dR = &Response{Message: "failed to decode request"}
			resp.Error(w, r, dR)
			return
		}

		if err := validator.New().Struct(req); err != nil {
			validateErr := err.(validator.ValidationErrors)
			log.Error("invalid request", err)
			dR := resp.ValidationError(validateErr)
			resp.Error(w, r, dR)

			return
		}

		password, err := crypt.GeneratePasswordHash(req.Password)
		if err != nil {
			log.Error("invalid generate hash password", err)
			var dR = &Response{Message: "failed create user"}
			resp.Error(w, r, dR)
			return
		}

		time := time.Now().Unix()

		var usr = &user.CreateUser{
			UUID:      identity.NewGenerator().GenerateUUIDv4String(),
			Password:  password,
			Email:     req.Email,
			Name:      req.Name,
			CreatedAt: time,
			UpdatedAt: time,
		}

		err = auth.Registration(usr)
		if err != nil {
			if storage.ErrorCode(err) == storage.ErrCodeExists {
				log.Info("user already exists")
				var dR = &Response{Message: "user already exists"}
				resp.Error(w, r, dR)
				return
			}
			var dR = &Response{Message: "failed create user"}
			resp.Error(w, r, dR)
			return
		}

		resp.Ok(w, r, nil)
	}
}
