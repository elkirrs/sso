package access_token

type AccessToken struct {
	ID        string `json:"id"`
	UserId    int64  `json:"userId"`
	ClientId  string `json:"clientId"`
	Name      string `json:"name"`
	Scopes    string `json:"scopes"`
	Revoked   bool   `json:"revoked"`
	UpdatedAt int64  `json:"updatedAt"`
	CreatedAt int64  `json:"createdAt"`
	ExpiresAt int64  `json:"expiresAt"`
}
