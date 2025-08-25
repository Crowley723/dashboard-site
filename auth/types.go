package auth

type User struct {
	Sub         string   `json:"sub"`
	Iss         string   `json:"iss"`
	Username    string   `json:"username"`
	DisplayName string   `json:"display_name"`
	Email       string   `json:"email"`
	Groups      []string `json:"groups"`
}
