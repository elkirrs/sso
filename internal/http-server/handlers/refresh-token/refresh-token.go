package refresh_token

import (
	"app/internal/config"
	accessTokenDomain "app/internal/domain/oauth/access-token"
	"app/internal/domain/oauth/clients"
	refreshTokenDomain "app/internal/domain/oauth/refresh-token"
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

type AccessToken interface {
	CreateToken(aT *accessTokenDomain.AccessToken) (string, error)
	ExistsToken(aT *accessTokenDomain.AccessToken) (bool, error)
	UpdateToken(aT *accessTokenDomain.AccessToken) (bool, error)
}

type RefreshToken interface {
	CreateRefreshToken(rT *refreshTokenDomain.RefreshToken) (string, error)
	ExistsToken(rT *refreshTokenDomain.RefreshToken) (bool, error)
	UpdateToken(rT *refreshTokenDomain.RefreshToken) (bool, error)
}

type Client interface {
	GetClient(ID string) (clients.Client, error)
}

type Request struct {
	RefreshToken string `json:"refresh_token" validate:"required,ascii"`
	ClientID     string `json:"client_id" validate:"required,ascii"`
}

type Response struct {
	AccessToken  string `json:"access_token,omitempty"`
	RefreshToken string `json:"refresh_token,omitempty"`
	ExpiredAt    int64  `json:"expired_at,omitempty"`
	Message      string `json:"message,omitempty"`
}

func New(
	log *slog.Logger,
	accessToken AccessToken,
	refreshToken RefreshToken,
	client Client,
	cfg config.Token,
) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {

		const op = "http-server.handlers.refresh-token.New"
		log := log.With(
			slog.String("op", op),
			slog.String("request_id", middleware.GetReqID(r.Context())),
		)

		var dR = map[string]string{}
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

		oldPayloadRefreshToken, err := token.ParseRefreshToken(req.RefreshToken)
		if err != nil {
			log.Error("refresh token invalid")
			dR["message"] = "refresh token invalid"
			resp.Error(w, r, dR)
			return
		}

		if oldPayloadRefreshToken.ExpiresAt < time.Now().Unix() {
			log.Error("refresh token expired")
			dR["message"] = "refresh token expired"
			resp.Error(w, r, dR)
			return
		}

		var aT = &accessTokenDomain.AccessToken{
			ClientId: oldPayloadRefreshToken.ClientId,
			UserId:   oldPayloadRefreshToken.UserId,
			ID:       oldPayloadRefreshToken.TokenAccessId,
		}

		existsAT, err := accessToken.ExistsToken(aT)
		if err != nil || !existsAT {
			log.Error("not fount access token in database")
			dR["message"] = "refresh token invalid"
			resp.Error(w, r, dR)
			return
		}

		var rT = &refreshTokenDomain.RefreshToken{
			AccessTokenId: oldPayloadRefreshToken.TokenAccessId,
			ID:            oldPayloadRefreshToken.TokenRefreshId,
		}

		existsRT, err := refreshToken.ExistsToken(rT)
		if err != nil || !existsRT {
			log.Error("not fount refresh token in database")
			dR["message"] = "refresh token invalid"
			resp.Error(w, r, dR)
			return
		}

		updateAT, err := accessToken.UpdateToken(aT)
		if err != nil || !updateAT {
			log.Error("access token was not revoked")
			dR["message"] = "refresh token invalid"
			resp.Error(w, r, dR)
			return
		}

		updateRT, err := refreshToken.UpdateToken(rT)
		if err != nil || !updateRT {
			log.Error("refresh token was not revoked")
			dR["message"] = "refresh token invalid"
			resp.Error(w, r, dR)
			return
		}

		clientStorage, err := client.GetClient(oldPayloadRefreshToken.ClientId)
		if err != nil {
			log.Info("client storage")
			dR["message"] = "client storage"
			resp.Error(w, r, dR)
			return
		}

		var accessTokenPayload = &accessTokenDomain.Payload{
			UUID:     oldPayloadRefreshToken.UUID,
			Email:    oldPayloadRefreshToken.Email,
			ClientID: clientStorage.ID,
			Scopes:   "[*]",
		}

		accessTokenString, err := token.GenerateAccessToken(accessTokenPayload, cfg.TTL, clientStorage.Secret)

		dateTime := time.Now().Unix()
		dateTimeExp := time.Now().Add(cfg.TTL).Unix()

		var aToken = &accessTokenDomain.AccessToken{
			ID:        crypt.GetMD5Hash(identity.NewGenerator().GenerateUUIDv4String()),
			UserId:    oldPayloadRefreshToken.UserId,
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

		refreshTokenId, err := refreshToken.CreateRefreshToken(rToken)
		if err != nil {
			log.Info("failed create refresh token")
			dR["message"] = "failed create token"
			resp.Error(w, r, dR)
			return
		}

		log.Info("client refresh token id ", refreshTokenId)

		var refreshTokenPayload = &refreshTokenDomain.Payload{
			UUID:           oldPayloadRefreshToken.UUID,
			Email:          oldPayloadRefreshToken.Email,
			TokenAccessId:  accessTokenId,
			TokenRefreshId: refreshTokenId,
			ClientId:       clientStorage.ID,
			UserId:         oldPayloadRefreshToken.UserId,
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
