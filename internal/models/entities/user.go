package entities

import "time"

type User struct {
	UserId                 string    `json:"user_id"`
	Email                  string    `json:"email"`
	PasswordHash           string    `json:"password_hash"`
	Role                   string    `json:"role"`
	RefreshToken           string    `json:"refresh_token"`
	RefreshTokenExpiryTime time.Time `json:"refresh_token_expiry_time"`
}
