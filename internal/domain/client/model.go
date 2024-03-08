package client

type Client struct {
	ID                   string `json:"id"`
	UserId               *int64 `json:"userId"`
	Name                 string `json:"name"`
	Secret               string `json:"secret"`
	Provider             string `json:"provider"`
	Redirect             string `json:"redirect"`
	PersonalAccessClient bool   `json:"personalAccessClient"`
	PasswordClient       bool   `json:"passwordClient"`
	Revoked              bool   `json:"revoked"`
	CreatedAt            int64  `json:"createdAt"`
	UpdatedAt            int64  `json:"updatedAt"`
}
