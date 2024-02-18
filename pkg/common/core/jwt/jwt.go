package jwt

import (
	"github.com/golang-jwt/jwt/v5"
	"time"
)

func GenerateToken(
	data map[string]any,
	tokenTTL time.Duration,
	tokenSecret string,
) (string, error) {
	token := jwt.New(jwt.SigningMethodHS512)

	expAccessToken := time.Now().Add(tokenTTL).Unix()

	claims := token.Claims.(jwt.MapClaims)
	for idx, val := range data {
		claims[idx] = val
	}

	claims["exp"] = expAccessToken
	accessToken, err := token.SignedString([]byte(tokenSecret))

	if err != nil {
		return "", err
	}

	return accessToken, nil
}
