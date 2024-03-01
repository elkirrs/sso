package token

import (
	accessTokenDomain "app/internal/domain/oauth/access-token"
	refreshTokenDomain "app/internal/domain/oauth/refresh-token"
	"app/pkg/utils/crypt"
	"crypto/x509"
	"encoding/json"
	"encoding/pem"
	"github.com/golang-jwt/jwt/v5"
	"log/slog"
	"os"
	"time"
)

type UserClaim struct {
	jwt.RegisteredClaims
	UUID     string `json:"uuid"`
	Email    string `json:"email"`
	ClientID string `json:"client_id"`
	ExpAt    int64  `json:"exp_at"`
}

func GenerateAccessToken(
	payload *accessTokenDomain.Payload,
	tokenTTL time.Duration,
	tokenSecret string,
) (string, error) {
	expAccessToken := time.Now().Add(tokenTTL).Unix()

	token := jwt.NewWithClaims(jwt.SigningMethodHS512, &UserClaim{
		RegisteredClaims: jwt.RegisteredClaims{},
		UUID:             payload.UUID,
		Email:            payload.Email,
		ClientID:         payload.ClientID,
		ExpAt:            expAccessToken,
	})

	accessToken, err := token.SignedString([]byte(tokenSecret))

	if err != nil {
		return "", err
	}

	return accessToken, nil
}

func GenerateRefreshToken(
	payload *refreshTokenDomain.Payload,
) (string, error) {
	dataBytes, err := json.Marshal(payload)
	if err != nil {
		return "", err
	}

	publicKeyFile, err := os.ReadFile("./storage/secret/oauth-public.key")
	if err != nil {
		return "", err
	}
	p, _ := pem.Decode(publicKeyFile)
	publicKey, err := x509.ParsePKCS1PublicKey(p.Bytes)

	if err != nil {
		return "", err
	}

	key, err := crypt.EncryptWithPublicKey(dataBytes, publicKey)
	if err != nil {
		return "", err
	}

	return key, nil
}

func ParseRefreshToken(token string) (refreshTokenDomain.Payload, error) {

	privateKeyFile, err := os.ReadFile("./storage/secret/oauth-private.key")
	var oldDateRefreshToken refreshTokenDomain.Payload

	if err != nil {
		return oldDateRefreshToken, err
	}
	p, _ := pem.Decode(privateKeyFile)

	privateKey, err := x509.ParsePKCS1PrivateKey(p.Bytes)
	if err != nil {
		return oldDateRefreshToken, err
	}

	dataToken, err := crypt.DecryptWithPrivateKey(token, privateKey)
	if err != nil {
		return oldDateRefreshToken, err
	}

	err = json.Unmarshal(dataToken, &oldDateRefreshToken)

	if err != nil {
		return oldDateRefreshToken, err
	}

	slog.Info("dataToken", dataToken)
	return oldDateRefreshToken, nil
}
