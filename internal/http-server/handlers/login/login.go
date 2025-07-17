package login

import (
	"app/internal/config"
	"app/internal/domain/client"
	accessTokenDomain "app/internal/domain/oauth/access-token"
	refreshTokenDomain "app/internal/domain/oauth/refresh-token"
	"app/internal/domain/user"
	resp "app/pkg/common/core/api/response"
	"app/pkg/common/core/identity"
	"app/pkg/common/core/token"
	"app/pkg/common/logging"
	"app/pkg/utils/crypt"
	"context"
	"errors"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/render"
	"github.com/go-playground/validator/v10"
	"io"
	"net/http"
	"time"
)

type Auth interface {
	Login(req *user.User) (user.User, error)
}

type AuthToken interface {
	Create(aT *accessTokenDomain.AccessToken, rT *refreshTokenDomain.RefreshToken) error
}

type Client interface {
	GetClient(ID string) (client.Client, error)
}

type Request struct {
	Login    string `json:"login" validate:"required,ascii"`
	Password string `json:"password" validate:"required,ascii"`
	ClientId string `json:"client_id" validate:"required,ascii"`
}

type Response struct {
	AccessToken  string `json:"access_token,"`
	RefreshToken string `json:"refresh_token"`
	ExpiredAt    int64  `json:"expired_at"`
}

func New(
	ctx context.Context,
	auth Auth,
	authToken AuthToken,
	client Client,
	cfg config.Token,
) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		const op = "http-server.handlers.login.New"

		logging.L(ctx).With(
			logging.StringAttr("op", op),
			logging.StringAttr("request_id", middleware.GetReqID(r.Context())),
		)

		req, err := decodeAndValidateRequest(r, ctx)
		if err != nil {
			resp.Error(w, r, map[string]string{"message": "invalid request"})
			return
		}

		var usr = &user.User{
			Email: req.Login,
			Name:  req.Login,
		}

		userStorage, err := auth.Login(usr)

		if err != nil || crypt.VerifyPassword(userStorage.Password, req.Password) != nil {
			logging.L(ctx).Error("authentication failed")
			resp.Error(w, r, map[string]string{"message": "incorrect login or password"})
			return
		}

		clientStorage, err := client.GetClient(req.ClientId)
		if err != nil {
			logging.L(ctx).Error("client storage")
			resp.Error(w, r, map[string]string{"message": "invalid client storage"})
			return
		}

		accessTokenStr, expAt, err := generateAccessToken(userStorage, clientStorage, cfg)
		if err != nil {
			logging.L(ctx).Error("failed generate access token")
			resp.Error(w, r, map[string]string{"message": "failed create token"})
			return
		}

		accessTokenID := crypt.GetMD5Hash(identity.UUIDv7())
		now := time.Now().Unix()

		var aToken = &accessTokenDomain.AccessToken{
			ID:        accessTokenID,
			UserId:    userStorage.ID,
			ClientId:  clientStorage.ID,
			Revoked:   false,
			CreatedAt: now,
			UpdatedAt: now,
			ExpiresAt: expAt,
		}

		refreshTokenID := crypt.GetMD5Hash(identity.UUIDv7())
		refreshExp := time.Now().Add(cfg.Refresh).Unix()
		rToken := &refreshTokenDomain.RefreshToken{
			ID:            refreshTokenID,
			AccessTokenId: accessTokenID,
			Revoked:       false,
			ExpiresAt:     refreshExp,
		}

		refreshTokenStr, err := generateRefreshToken(userStorage, accessTokenID, refreshTokenID, clientStorage.ID, refreshExp)
		if err != nil {
			logging.L(ctx).Error("failed generate refresh token", err)
			resp.Error(w, r, map[string]string{"message": "failed to create token"})
			return
		}

		if err := authToken.Create(aToken, rToken); err != nil {
			logging.L(ctx).Error("failed create token")
			resp.Error(w, r, map[string]string{"message": "failed to create token"})
			return
		}

		resp.Ok(w, r, &Response{
			AccessToken:  accessTokenStr,
			RefreshToken: refreshTokenStr,
			ExpiredAt:    expAt,
		})
		return
	}
}

func decodeAndValidateRequest(r *http.Request, ctx context.Context) (*Request, error) {
	var req Request

	err := render.DecodeJSON(r.Body, &req)
	if errors.Is(err, io.EOF) {
		logging.L(ctx).Error("request body is empty")
		return nil, err
	}
	if err := validator.New().Struct(req); err != nil {
		logging.L(ctx).Error("invalid request", err)
		return nil, err
	}

	return &req, nil
}

func generateAccessToken(user user.User, client client.Client, cfg config.Token) (string, int64, error) {
	payload := &accessTokenDomain.Payload{
		UUID:     user.UUID,
		Email:    user.Email,
		ClientID: client.ID,
		Scopes:   "[*]",
	}
	tokenStr, err := token.GenerateAccessToken(payload, cfg.TTL, client.Secret)
	if err != nil {
		return "", 0, err
	}
	return tokenStr, time.Now().Add(cfg.TTL).Unix(), nil
}

func generateRefreshToken(user user.User, accessTokenID, refreshTokenID, clientID string, expiresAt int64) (string, error) {
	payload := &refreshTokenDomain.Payload{
		UUID:           user.UUID,
		Email:          user.Email,
		TokenAccessId:  accessTokenID,
		TokenRefreshId: refreshTokenID,
		ClientId:       clientID,
		UserId:         user.ID,
		ExpiresAt:      expiresAt,
		Scopes:         "[*]",
	}
	return token.GenerateRefreshToken(payload)
}

func logTime(step string, start time.Time, ctx context.Context) {
	elapsed := time.Since(start)
	logging.L(ctx).Info("perf", "step", step, "took", elapsed)
}
