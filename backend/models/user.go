package models

type User struct {
	ID           int    `json:"id,omitempty"`
	FullName     string `json:"fullname"`
	Email        string `json:"email"`
	PasswordHash string `json:"password_hash,omitempty"`
	RoleID       int    `json:"role_id"`
}

type LoginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}
