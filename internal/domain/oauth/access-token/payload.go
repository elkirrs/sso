package access_token

type Payload struct {
	UUID     string `json:"uuid"`
	Email    string `json:"email"`
	ClientID string `json:"client_id"`
	Scopes   any    `json:"scopes"`
}
