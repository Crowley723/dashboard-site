package models

import "time"

type User struct {
	Sub          string    `json:"sub"`
	Iss          string    `json:"iss"`
	Username     string    `json:"name"`
	DisplayName  string    `json:"display_name"`
	Email        string    `json:"email"`
	Groups       []string  `json:"groups"`
	LastLoggedIn time.Time `json:"last_logged_in"`
	CreatedAt    time.Time `json:"created_at"`
}
