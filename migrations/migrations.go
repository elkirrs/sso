package migrations

import "embed"

//go:embed *.sql
var Content embed.FS

var (
	TableUsers             = "users"
	TableOauthAccessToken  = "oauth_access_tokens"
	TableOauthClient       = "oauth_clients"
	TableOauthRefreshToken = "oauth_refresh_tokens"
)
