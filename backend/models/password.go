package models

import "time"

// PasswordResetRequest представляет запрос на восстановление пароля
type PasswordResetRequest struct {
	Email string `json:"email" binding:"required,email"`
}

// PasswordResetConfirm представляет подтверждение восстановления пароля
type PasswordResetConfirm struct {
	Email    string `json:"email" binding:"required,email"`
	Code     string `json:"code" binding:"required,len=6"`
	Password string `json:"password" binding:"required,min=8"`
}

// PasswordChangeRequest представляет запрос на смену пароля
type PasswordChangeRequest struct {
	OldPassword string `json:"old_password" binding:"required"`
	NewPassword string `json:"new_password" binding:"required,min=8"`
}

// PasswordChangeConfirm представляет подтверждение смены пароля
type PasswordChangeConfirm struct {
	Code string `json:"code" binding:"required,len=6"`
}

// PasswordResetCode представляет код восстановления в базе данных
type PasswordResetCode struct {
	ID        int       `json:"id"`
	Email     string    `json:"email"`
	Code      string    `json:"code"`
	ExpiresAt time.Time `json:"expires_at"`
	CreatedAt time.Time `json:"created_at"`
	Used      bool      `json:"used"`
}

// PasswordResponse представляет ответ при операциях с паролем
type PasswordResponse struct {
	Message string `json:"message"`
	Success bool   `json:"success"`
}
































