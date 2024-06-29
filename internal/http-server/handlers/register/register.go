package register

import (
	"app/internal/domain/user"
	"app/internal/storage"
	resp "app/pkg/common/core/api/response"
	"app/pkg/common/core/identity"
	"app/pkg/common/logging"
	"app/pkg/utils/crypt"
	"context"
	"errors"
	"github.com/go-chi/chi/middleware"
	"github.com/go-chi/render"
	"github.com/go-playground/validator/v10"
	"io"
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
	ConfirmPassword string `json:"confirmPassword" validate:"required,eqfield=Password"`
}

type Response struct {
	Message string `json:"message,omitempty"`
}

func New(
	ctx context.Context,
	auth Auth,
) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		const op = "http-server.handlers.register.New"

		logging.L(ctx).With(
			logging.StringAttr("op", op),
			logging.StringAttr("request_id", middleware.GetReqID(r.Context())),
		)
		var req Request

		err := render.DecodeJSON(r.Body, &req)
		if errors.Is(err, io.EOF) {
			logging.L(ctx).Error("request body is empty")
			var dR = &Response{Message: "empty request"}
			resp.Error(w, r, dR)
			return
		}

		if err != nil {
			logging.L(ctx).Error("failed to decode request body", err)
			var dR = &Response{Message: "failed to decode request"}
			resp.Error(w, r, dR)
			return
		}

		if err := validator.New().Struct(req); err != nil {
			validateErr := err.(validator.ValidationErrors)
			logging.L(ctx).Error("invalid request", err)
			dR := resp.ValidationError(validateErr)
			resp.Error(w, r, dR)

			return
		}

		password, err := crypt.GeneratePasswordHash(req.Password)
		if err != nil {
			logging.L(ctx).Error("invalid generate hash password", err)
			var dR = &Response{Message: "failed create user"}
			resp.Error(w, r, dR)
			return
		}

		timeUnix := time.Now().Unix()

		var usr = &user.CreateUser{
			UUID:      identity.NewGenerator().GenerateUUIDv4String(),
			Password:  password,
			Email:     req.Email,
			Name:      req.Name,
			CreatedAt: timeUnix,
			UpdatedAt: timeUnix,
		}

		err = auth.Registration(usr)
		if err != nil {
			if storage.ErrorCode(err) == storage.ErrCodeExists {
				logging.L(ctx).Info("user already exists")
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
