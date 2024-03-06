package login

import (
	"app/internal/config"
	accessTokenDomain "app/internal/domain/oauth/access-token"
	"app/internal/domain/oauth/clients"
	refreshTokenDomain "app/internal/domain/oauth/refresh-token"
	"app/internal/domain/user"
	resp "app/pkg/common/core/api/response"
	"app/pkg/common/core/identity"
	"app/pkg/common/core/token"
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
	Login(req *user.User) (user.User, error)
}
type AccessToken interface {
	CreateToken(aT *accessTokenDomain.AccessToken) (string, error)
}

type RefreshToken interface {
	CreateRefreshToken(rT *refreshTokenDomain.RefreshToken) (string, error)
}

type Client interface {
	GetClient(ID string) (clients.Client, error)
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
	log *slog.Logger,
	auth Auth,
	accessToken AccessToken,
	refreshToken RefreshToken,
	client Client,
	cfg config.Token,
) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		const op = "http-server.handlers.login.New"

		var dR = map[string]string{}
		log := log.With(
			slog.String("op", op),
			slog.String("request_id", middleware.GetReqID(r.Context())),
		)

		var req Request

		err := render.DecodeJSON(r.Body, &req)
		if errors.Is(err, io.EOF) {
			log.Error("request body is empty")
			dR["message"] = "empty request"
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

		var usr = &user.User{
			Email: req.Login,
			Name:  req.Login,
		}

		userStorage, err := auth.Login(usr)

		if err != nil {
			log.Info("failed login user")
			dR["message"] = "failed login user"
			resp.Error(w, r, dR)
			return
		}

		err = crypt.VerifyPassword(userStorage.Password, req.Password)
		if err != nil {
			log.Info("incorrect login or password")
			dR["login"] = "incorrect login or password"
			resp.Error(w, r, dR)
			return
		}

		clientStorage, err := client.GetClient(req.ClientId)
		if err != nil {
			log.Info("client storage")
			dR["message"] = "client storage"
			resp.Error(w, r, dR)
			return
		}

		var accessTokenPayload = &accessTokenDomain.Payload{
			UUID:     userStorage.UUID,
			Email:    userStorage.Email,
			ClientID: clientStorage.ID,
			Scopes:   "[*]",
		}

		accessTokenString, err := token.GenerateAccessToken(accessTokenPayload, cfg.TTL, clientStorage.Secret)

		if err != nil {
			log.Info("failed generate access token")
			log.Info("error", err)
			dR["message"] = "failed create token"
			resp.Error(w, r, dR)
			return
		}

		dateTime := time.Now().Unix()
		dateTimeExp := time.Now().Add(cfg.TTL).Unix()

		var aToken = &accessTokenDomain.AccessToken{
			ID:        crypt.GetMD5Hash(identity.NewGenerator().GenerateUUIDv4String()),
			UserId:    userStorage.ID,
			ClientId:  clientStorage.ID,
			Revoked:   false,
			CreatedAt: dateTime,
			UpdatedAt: dateTime,
			ExpiresAt: dateTimeExp,
		}

		accessTokenId, err := accessToken.CreateToken(aToken)

		if err != nil {
			log.Info("failed create access token")
			dR["message"] = "failed create token"
			resp.Error(w, r, dR)
			return
		}

		log.Info("client access token id ", accessTokenId)

		dateTimeExpRefresh := time.Now().Add(cfg.Refresh).Unix()
		var rToken = &refreshTokenDomain.RefreshToken{
			ID:            crypt.GetMD5Hash(time.Now().String()),
			AccessTokenId: accessTokenId,
			Revoked:       false,
			ExpiresAt:     dateTimeExpRefresh,
		}

		_, err = refreshToken.CreateRefreshToken(rToken)
		if err != nil {
			log.Info("failed create refresh token")
			dR["message"] = "failed create token"
			resp.Error(w, r, dR)
			return
		}

		var refreshTokenPayload = &refreshTokenDomain.Payload{
			UUID:           userStorage.UUID,
			Email:          userStorage.Email,
			TokenAccessId:  accessTokenId,
			TokenRefreshId: rToken.ID,
			ClientId:       clientStorage.ID,
			UserId:         userStorage.ID,
			ExpiresAt:      dateTimeExpRefresh,
			Scopes:         "[*]",
		}

		refreshTokenString, err := token.GenerateRefreshToken(refreshTokenPayload)

		if err != nil {
			log.Info("failed generate refresh token")
			log.Info("error", err)
			dR["message"] = "failed create token"
			resp.Error(w, r, dR)
			return
		}

		expAccessToken := time.Now().Add(cfg.TTL).Unix()

		var dRS = &Response{
			AccessToken:  accessTokenString,
			RefreshToken: refreshTokenString,
			ExpiredAt:    expAccessToken,
		}

		resp.Ok(w, r, dRS)
		return
	}
}
