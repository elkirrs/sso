package client

import (
	"app/internal/domain/client"
	resp "app/pkg/common/core/api/response"
	"app/pkg/common/logging"
	"context"
	"github.com/go-chi/chi"
	"github.com/go-chi/chi/middleware"
	"github.com/go-playground/validator/v10"
	"net/http"
)

type Client interface {
	GetClientByName(name string) (client.Client, error)
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
	Name     string `json:"name,"`
	Redirect string `json:"redirect"`
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
