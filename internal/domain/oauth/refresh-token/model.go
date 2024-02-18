package refresh_token

type RefreshToken struct {
	ID            string `json:"id"`
	AccessTokenId string `json:"accessTokenId"`
	Revoked       bool   `json:"revoked"`
	ExpiresAt     int64  `json:"expiresAt"`
}
