package user

type User struct {
	ID              int64   `json:"id"`
	UUID            string  `json:"uuid"`
	Name            string  `json:"name"`
	Email           string  `json:"email"`
	EmailVerifiedAt *int64  `json:"emailVerifiedAt"`
	Password        string  `json:"password"`
	RememberToken   *string `json:"rememberToken"`
	IsActive        int     `json:"isActive"`
	CreatedAt       int64   `json:"createdAt"`
	UpdatedAt       int64   `json:"updatedAt"`
}
type CreateUser struct {
	ID              int
	UUID            string
	Name            string
	Email           string
	EmailVerifiedAt *int64
	Password        string
	RememberToken   *string
	IsActive        int
	CreatedAt       int64
	UpdatedAt       int64
}

type CreateUserSignal struct {
	UUID      string `json:"uuid"`
	Service   string `json:"service"`
	CreatedAt int64  `json:"createdAt"`
}
