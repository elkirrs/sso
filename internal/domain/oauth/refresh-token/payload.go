package refresh_token

type Payload struct {
	UUID           string `json:"uuid"`
	Email          string `json:"email"`
	TokenAccessId  string `json:"token_access_id"`
	TokenRefreshId string `json:"token_refresh_id"`
	ClientId       string `json:"client_id"`
	UserId         int64  `json:"user_id"`
	ExpiresAt      int64  `json:"exp_at"`
	Scopes         any    `json:"scopes"`
}
