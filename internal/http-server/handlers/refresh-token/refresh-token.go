package refresh_token

import (
	"app/internal/config"
	"app/internal/domain/client"
	accessTokenDomain "app/internal/domain/oauth/access-token"
	refreshTokenDomain "app/internal/domain/oauth/refresh-token"
	resp "app/pkg/common/core/api/response"
	"app/pkg/common/core/identity"
	"app/pkg/common/core/token"
	"app/pkg/common/logging"
	"app/pkg/utils/crypt"
	"context"
	"errors"
	"github.com/go-chi/chi/middleware"
	"github.com/go-chi/render"
	"github.com/go-playground/validator/v10"
	"io"
	"net/http"
	"sync"
	"time"
)

type AccessToken interface {
	CreateToken(aT *accessTokenDomain.AccessToken) (string, error)
	ExistsToken(aT *accessTokenDomain.AccessToken) (bool, error)
	UpdateToken(aT *accessTokenDomain.AccessToken) (bool, error)
}

type RefreshToken interface {
	CreateRefreshToken(rT *refreshTokenDomain.RefreshToken) (string, error)
	GetToken(rT *refreshTokenDomain.RefreshToken) (refreshTokenDomain.RefreshToken, error)
	GetLastReceivedToken(rT *refreshTokenDomain.RefreshToken) (refreshTokenDomain.RefreshToken, error)
	UpdateToken(rT *refreshTokenDomain.RefreshToken) (bool, error)
}

type Client interface {
	GetClient(ID string) (client.Client, error)
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
	ctx context.Context,
	accessToken AccessToken,
	refreshToken RefreshToken,
	client Client,
	cfg config.Token,
) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {

		const op = "http-server.handlers.refresh-token.New"
		logging.L(ctx).With(
			logging.StringAttr("op", op),
			logging.StringAttr("request_id", middleware.GetReqID(r.Context())),
		)

		var dR = map[string]string{}
		var req Request

		err := render.DecodeJSON(r.Body, &req)
		if errors.Is(err, io.EOF) {
			logging.L(ctx).Error("request body is empty")
			dR["message"] = "empty request"
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

		oldPayloadRefreshToken, err := token.ParseRefreshToken(req.RefreshToken)
		if err != nil {
			logging.L(ctx).Error("refresh token invalid")
			dR["message"] = "refresh token invalid"
			resp.Error(w, r, dR)
			return
		}

		if oldPayloadRefreshToken.ExpiresAt < time.Now().Unix() {
			logging.L(ctx).Error("refresh token expired")
			dR["message"] = "refresh token expired"
			resp.Error(w, r, dR)
			return
		}

		var rT = &refreshTokenDomain.RefreshToken{
			AccessTokenId: oldPayloadRefreshToken.TokenAccessId,
			ID:            oldPayloadRefreshToken.TokenRefreshId,
		}

		results := make(chan refreshTokenDomain.RefreshToken, 2)
		var wg sync.WaitGroup
		wg.Add(1)
		go getToken(refreshToken, rT, results, &wg)
		wg.Add(1)
		go getLastReceivedToken(refreshToken, rT, results, &wg)

		go func() {
			wg.Wait()
			close(results)
		}()

		for result := range results {
			if result.ID == "" {
				logging.L(ctx).Error("failed to retrieve refresh token")
				dR["message"] = "refresh token invalid"
				resp.Error(w, r, dR)
				return
			}

			if result.Revoked {
				logging.L(ctx).Error("refresh token is revoked")
				dR["message"] = "refresh token invalid"
				resp.Error(w, r, dR)
				return
			}

			if result.ID != rT.ID {
				logging.L(ctx).Error("refresh token mismatch with the last received token")
				dR["message"] = "refresh token invalid"
				resp.Error(w, r, dR)
				return
			}
		}

		var aT = &accessTokenDomain.AccessToken{
			ClientId: oldPayloadRefreshToken.ClientId,
			UserId:   oldPayloadRefreshToken.UserId,
			ID:       oldPayloadRefreshToken.TokenAccessId,
		}

		existsAT, err := accessToken.ExistsToken(aT)
		if err != nil || !existsAT {
			logging.L(ctx).Error("not fount access token in database")
			dR["message"] = "refresh token invalid"
			resp.Error(w, r, dR)
			return
		}

		updateAT, err := accessToken.UpdateToken(aT)
		if err != nil || !updateAT {
			logging.L(ctx).Error("access token was not revoked")
			dR["message"] = "refresh token invalid"
			resp.Error(w, r, dR)
			return
		}

		updateRT, err := refreshToken.UpdateToken(rT)
		if err != nil || !updateRT {
			logging.L(ctx).Error("refresh token was not revoked")
			dR["message"] = "refresh token invalid"
			resp.Error(w, r, dR)
			return
		}

		clientStorage, err := client.GetClient(oldPayloadRefreshToken.ClientId)
		if err != nil {
			logging.L(ctx).Info("client storage")
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
			logging.L(ctx).Info("failed create access token")
			dR["message"] = "failed create token"
			resp.Error(w, r, dR)
			return
		}

		logging.L(ctx).Info("generate access token")

		dateTimeExpRefresh := time.Now().Add(cfg.Refresh).Unix()
		var rToken = &refreshTokenDomain.RefreshToken{
			ID:            crypt.GetMD5Hash(time.Now().String()),
			AccessTokenId: accessTokenId,
			Revoked:       false,
			ExpiresAt:     dateTimeExpRefresh,
		}

		refreshTokenId, err := refreshToken.CreateRefreshToken(rToken)
		if err != nil {
			logging.L(ctx).Info("failed create refresh token")
			dR["message"] = "failed create token"
			resp.Error(w, r, dR)
			return
		}

		logging.L(ctx).Info("generate refresh token")

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
			logging.L(ctx).Error("failed generate refresh token", err)
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

func getToken(
	refreshToken RefreshToken,
	rT *refreshTokenDomain.RefreshToken,
	results chan<- refreshTokenDomain.RefreshToken,
	wg *sync.WaitGroup,
) {
	defer wg.Done()

	rTQ, err := refreshToken.GetToken(rT)
	if err != nil {
		results <- refreshTokenDomain.RefreshToken{}
		return
	}

	results <- rTQ
}

func getLastReceivedToken(
	refreshToken RefreshToken,
	rT *refreshTokenDomain.RefreshToken,
	results chan<- refreshTokenDomain.RefreshToken,
	wg *sync.WaitGroup,
) {
	defer wg.Done()

	rTQ, err := refreshToken.GetLastReceivedToken(rT)
	if err != nil {
		results <- refreshTokenDomain.RefreshToken{}
		return
	}

	results <- rTQ
}
