package client

import (
	"app/internal/domain/client"
	"app/internal/storage"
	resp "app/pkg/common/core/api/response"
	"app/pkg/common/core/identity"
	"app/pkg/common/logging"
	"app/pkg/utils/crypt"
	"context"
	"errors"
	"github.com/go-chi/chi"
	"github.com/go-chi/chi/middleware"
	"github.com/go-chi/render"
	"github.com/go-playground/validator/v10"
	"io"
	"net/http"
	"time"
)

type Client interface {
	GetClientByName(name string) (client.Client, error)
	CreateClient(client *client.Client) error
}

type Storage struct {
	ctx    context.Context
	client Client
}

type Request struct {
	ClientName string `json:"client" validate:"required,ascii"`
}

func New(ctx context.Context, client Client) *Storage {
	return &Storage{
		ctx:    ctx,
		client: client,
	}
}

type Response struct {
	ID       string `json:"id"`
	Name     string `json:"name"`
	Redirect string `json:"redirect"`
}

type CreateRequest struct {
	Name     string `json:"name" validate:"required,ascii"`
	Redirect string `json:"redirect" validate:"required,ascii"`
}

func (s *Storage) GetClient() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		const op = "http-server.handlers.client.GetClient"
		logging.L(s.ctx).Info("op", op)

		var dR = map[string]string{}

		logging.L(s.ctx).With(
			logging.StringAttr("op", op),
			logging.StringAttr("request_id", middleware.GetReqID(r.Context())),
		)

		var req = Request{
			ClientName: chi.URLParam(r, "client"),
		}

		logging.L(s.ctx).Info("client name", req.ClientName)

		if err := validator.New().Struct(req); err != nil {
			validateErr := err.(validator.ValidationErrors)
			logging.L(s.ctx).Error("invalid request", err)
			dR := resp.ValidationError(validateErr)
			resp.Error(w, r, dR)
			return
		}

		clientStorage, err := s.client.GetClientByName(req.ClientName)

		if err != nil {
			logging.L(s.ctx).Error("client not found")
			dR["message"] = "failed login user"
			resp.Error(w, r, dR)
			return
		}

		var dRS = &Response{
			ID:       clientStorage.ID,
			Name:     clientStorage.Name,
			Redirect: clientStorage.Redirect,
		}
		resp.Ok(w, r, dRS)
		return
	}
}

func (s *Storage) CreateClient() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		const op = "http-server.handlers.client.CreateClient"
		logging.L(s.ctx).Info("op", op)

		var dR = map[string]string{}

		var req CreateRequest

		err := render.DecodeJSON(r.Body, &req)
		if errors.Is(err, io.EOF) {
			logging.L(s.ctx).Error("request body is empty")
			dR["message"] = "empty request"
			resp.Error(w, r, dR)
			return
		}

		if err := validator.New().Struct(req); err != nil {
			validateErr := err.(validator.ValidationErrors)
			logging.L(s.ctx).Error("invalid request", err)
			dR := resp.ValidationError(validateErr)
			resp.Error(w, r, dR)
			return
		}

		var oauthClient = &client.Client{
			ID:             identity.NewGenerator().GenerateUUIDv4String(),
			Name:           req.Name,
			Secret:         crypt.GetSecret(),
			Redirect:       req.Redirect,
			Provider:       "users",
			PasswordClient: true,
			CreatedAt:      time.Now().Unix(),
			UpdatedAt:      time.Now().Unix(),
		}

		err = s.client.CreateClient(oauthClient)

		if err != nil {
			if storage.ErrorCode(err) == storage.ErrCodeExists {
				logging.L(s.ctx).Info("client already exists")
				dR["message"] = "client already exists"
				resp.Error(w, r, dR)
				return
			}
			logging.L(s.ctx).Error("failed create client")
			dR["message"] = "failed create client"
			resp.Error(w, r, dR)
			return
		}

		var dRS = &Response{
			ID:       oauthClient.ID,
			Name:     oauthClient.Name,
			Redirect: oauthClient.Redirect,
		}
		resp.Ok(w, r, dRS)
		return
	}
}
