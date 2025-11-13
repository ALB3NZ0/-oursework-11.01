package handlers

import (
	"context"
	"crypto/rand"
	"encoding/json"
	"fmt"
	"math/big"
	"net/http"
	"net/smtp"
	"sync"
	"time"

	"golang.org/x/crypto/bcrypt"
	"shoes-store-backend/db"
	"shoes-store-backend/models"
)

// PasswordResetCode представляет код восстановления в памяти
type PasswordResetCode struct {
	Email       string    `json:"email"`
	Code        string    `json:"code"`
	ExpiresAt   time.Time `json:"expires_at"`
	Used        bool      `json:"used"`
	NewPassword string    `json:"new_password,omitempty"` // Для смены пароля
}

// Глобальное хранилище кодов (в реальном приложении лучше использовать Redis)
var (
	resetCodes = make(map[string]PasswordResetCode)
	codesMutex sync.RWMutex
)

// RequestPasswordResetHandler отправляет код восстановления пароля на email
// @Summary Запрос восстановления пароля
// @Description Отправляет 6-значный код восстановления пароля на указанный email
// @Tags Password
// @Accept json
// @Produce json
// @Param request body models.PasswordResetRequest true "Email для восстановления пароля"
// @Success 200 {object} models.PasswordResponse
// @Failure 400 {string} string "Ошибка валидации"
// @Failure 404 {string} string "Пользователь не найден"
// @Failure 500 {string} string "Внутренняя ошибка сервера"
// @Router /password/reset [post]
func RequestPasswordResetHandler(w http.ResponseWriter, r *http.Request) {
	var req models.PasswordResetRequest
	
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Неверный формат JSON", http.StatusBadRequest)
		return
	}

	// Проверяем, что пользователь существует
	var userExists bool
	err := db.Pool.QueryRow(context.Background(), 
		"SELECT EXISTS(SELECT 1 FROM users WHERE email = $1)", req.Email).Scan(&userExists)
	if err != nil {
		http.Error(w, "Ошибка проверки пользователя", http.StatusInternalServerError)
		return
	}

	if !userExists {
		http.Error(w, "Пользователь с таким email не найден", http.StatusNotFound)
		return
	}

	// Генерируем 6-значный код
	code := generateConfirmationCode()
	
	// Удаляем старые коды для этого email
	codesMutex.Lock()
	delete(resetCodes, req.Email)
	
	// Сохраняем новый код (действителен 10 минут)
	resetCodes[req.Email] = PasswordResetCode{
		Email:     req.Email,
		Code:      code,
		ExpiresAt: time.Now().Add(10 * time.Minute),
		Used:      false,
	}
	codesMutex.Unlock()

	// Отправляем email
	if !sendPasswordResetEmail(req.Email, code) {
		http.Error(w, "Ошибка отправки email", http.StatusInternalServerError)
		return
	}

	fmt.Printf("📧 Код восстановления отправлен на %s: %s\n", req.Email, code)

	response := models.PasswordResponse{
		Message: "Код восстановления отправлен на ваш email",
		Success: true,
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)
}

// ConfirmPasswordResetHandler подтверждает код и меняет пароль
// @Summary Подтверждение восстановления пароля
// @Description Подтверждает код восстановления и устанавливает новый пароль
// @Tags Password
// @Accept json
// @Produce json
// @Param request body models.PasswordResetConfirm true "Данные для восстановления пароля"
// @Success 200 {object} models.PasswordResponse
// @Failure 400 {string} string "Ошибка валидации"
// @Failure 404 {string} string "Код не найден или истек"
// @Failure 500 {string} string "Внутренняя ошибка сервера"
// @Router /password/reset/confirm [post]
func ConfirmPasswordResetHandler(w http.ResponseWriter, r *http.Request) {
	var req models.PasswordResetConfirm
	
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Неверный формат JSON", http.StatusBadRequest)
		return
	}

	// Проверяем код
	codesMutex.RLock()
	storedCode, exists := resetCodes[req.Email]
	codesMutex.RUnlock()
	
	if !exists || storedCode.Code != req.Code || storedCode.Used || time.Now().After(storedCode.ExpiresAt) {
		http.Error(w, "Неверный или истекший код", http.StatusNotFound)
		return
	}

	// Проверяем валидность нового пароля
	if len(req.Password) < 8 {
		http.Error(w, "Пароль должен содержать минимум 8 символов", http.StatusBadRequest)
		return
	}

	// Хэшируем пароль перед сохранением
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		http.Error(w, "Ошибка хэширования пароля", http.StatusInternalServerError)
		return
	}

	// Обновляем пароль пользователя
	_, err = db.Pool.Exec(context.Background(),
		"UPDATE users SET passwordhash = $1 WHERE email = $2",
		string(hashedPassword), req.Email)
	if err != nil {
		http.Error(w, "Ошибка обновления пароля", http.StatusInternalServerError)
		return
	}

	// Помечаем код как использованный
	codesMutex.Lock()
	if storedCode, exists := resetCodes[req.Email]; exists {
		storedCode.Used = true
		resetCodes[req.Email] = storedCode
	}
	codesMutex.Unlock()

	fmt.Printf("✅ Пароль успешно восстановлен для %s\n", req.Email)

	// Логируем восстановление пароля
	LogUserAction(r, "PASSWORD_RESET", "user", 0, fmt.Sprintf("Восстановлен пароль для пользователя: %s", req.Email))

	response := models.PasswordResponse{
		Message: "Пароль успешно восстановлен",
		Success: true,
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)
}

// ChangePasswordHandler инициирует смену пароля с подтверждением по email
// @Summary Смена пароля
// @Description Инициирует смену пароля с отправкой кода подтверждения на email
// @Tags Password
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body models.PasswordChangeRequest true "Данные для смены пароля"
// @Success 200 {object} models.PasswordResponse
// @Failure 400 {string} string "Ошибка валидации"
// @Failure 401 {string} string "Неверный текущий пароль"
// @Failure 500 {string} string "Внутренняя ошибка сервера"
// @Router /password/change [post]
func ChangePasswordHandler(w http.ResponseWriter, r *http.Request) {
	var req models.PasswordChangeRequest
	
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Неверный формат JSON", http.StatusBadRequest)
		return
	}

	// Получаем пользователя из контекста
	userID := r.Context().Value("userID")
	if userID == nil {
		http.Error(w, "Пользователь не авторизован", http.StatusUnauthorized)
		return
	}

	// Проверяем текущий пароль
	var currentPassword string
	var email string
	err := db.Pool.QueryRow(context.Background(),
		"SELECT passwordhash, email FROM users WHERE iduser = $1", userID).Scan(&currentPassword, &email)
	if err != nil {
		fmt.Printf("❌ Password Change: User not found - UserID: %v, error: %v\n", userID, err)
		http.Error(w, "Пользователь не найден", http.StatusNotFound)
		return
	}

	fmt.Printf("🔐 Password Change: UserID=%v, Email=%s\n", userID, email)
	fmt.Printf("   Stored password hash length: %d\n", len(currentPassword))
	fmt.Printf("   Old password length: %d\n", len(req.OldPassword))
	
	// Показываем первые 20 символов хеша для отладки
	hashPreview := currentPassword
	if len(currentPassword) > 20 {
		hashPreview = currentPassword[:20]
	}
	fmt.Printf("   Stored password starts with: %s\n", hashPreview)

	// Проверяем, является ли пароль bcrypt хешем
	// Bcrypt хеш всегда начинается с "$2a$", "$2b$" или "$2y$" и длиной 60 символов
	isBcryptHash := len(currentPassword) >= 10 && 
		(currentPassword[:3] == "$2a" || currentPassword[:3] == "$2b" || currentPassword[:3] == "$2y")

	fmt.Printf("   Is Bcrypt hash: %v\n", isBcryptHash)

	// Если пароль хранится как bcrypt хеш, сравниваем с bcrypt
	if isBcryptHash {
		// Сравниваем хэш пароля с использованием bcrypt
		err = bcrypt.CompareHashAndPassword([]byte(currentPassword), []byte(req.OldPassword))
		if err != nil {
			fmt.Printf("❌ Password Change: Invalid old password - %v\n", err)
			http.Error(w, "Неверный текущий пароль", http.StatusUnauthorized)
			return
		}
		fmt.Printf("✅ Password Change: Old password verified (bcrypt)\n")
	} else {
		// Если пароль в открытом виде (старая версия), сравниваем напрямую
		fmt.Printf("⚠️  Password stored as plain text, using direct comparison\n")
		if currentPassword != req.OldPassword {
			fmt.Printf("❌ Password Change: Invalid old password (plain text comparison)\n")
			http.Error(w, "Неверный текущий пароль", http.StatusUnauthorized)
			return
		}
		fmt.Printf("✅ Password Change: Old password verified (plain text)\n")
	}

	// Проверяем валидность нового пароля
	if len(req.NewPassword) < 8 {
		http.Error(w, "Пароль должен содержать минимум 8 символов", http.StatusBadRequest)
		return
	}

	// Генерируем код подтверждения
	code := generateConfirmationCode()
	
	// Удаляем старые коды для этого email
	codesMutex.Lock()
	delete(resetCodes, email)
	
	// Сохраняем новый код с новым паролем (действителен 10 минут)
	resetCodes[email] = PasswordResetCode{
		Email:       email,
		Code:        code,
		ExpiresAt:   time.Now().Add(10 * time.Minute),
		Used:        false,
		NewPassword: req.NewPassword, // Сохраняем новый пароль в открытом виде для подтверждения
	}
	codesMutex.Unlock()

	// Отправляем email
	if !sendPasswordChangeEmail(email, code) {
		http.Error(w, "Ошибка отправки email", http.StatusInternalServerError)
		return
	}

	fmt.Printf("📧 Код подтверждения смены пароля отправлен на %s: %s\n", email, code)

	response := models.PasswordResponse{
		Message: "Код подтверждения отправлен на ваш email",
		Success: true,
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)
}

// ConfirmPasswordChangeHandler подтверждает смену пароля по коду
// @Summary Подтверждение смены пароля
// @Description Подтверждает смену пароля по коду из email
// @Tags Password
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body models.PasswordChangeConfirm true "Код подтверждения"
// @Success 200 {object} models.PasswordResponse
// @Failure 400 {string} string "Ошибка валидации"
// @Failure 404 {string} string "Код не найден или истек"
// @Failure 500 {string} string "Внутренняя ошибка сервера"
// @Router /password/change/confirm [post]
func ConfirmPasswordChangeHandler(w http.ResponseWriter, r *http.Request) {
	var req models.PasswordChangeConfirm
	
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Неверный формат JSON", http.StatusBadRequest)
		return
	}

	// Получаем пользователя из контекста
	userID := r.Context().Value("userID")
	if userID == nil {
		http.Error(w, "Пользователь не авторизован", http.StatusUnauthorized)
		return
	}

	// Получаем email пользователя
	var email string
	err := db.Pool.QueryRow(context.Background(),
		"SELECT email FROM users WHERE iduser = $1", userID).Scan(&email)
	if err != nil {
		http.Error(w, "Пользователь не найден", http.StatusNotFound)
		return
	}

	// Проверяем код
	codesMutex.RLock()
	storedCode, exists := resetCodes[email]
	codesMutex.RUnlock()
	
	if !exists || storedCode.Code != req.Code || storedCode.Used || time.Now().After(storedCode.ExpiresAt) {
		http.Error(w, "Неверный или истекший код", http.StatusNotFound)
		return
	}

	// Хэшируем новый пароль перед сохранением
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(storedCode.NewPassword), bcrypt.DefaultCost)
	if err != nil {
		http.Error(w, "Ошибка хэширования пароля", http.StatusInternalServerError)
		return
	}

	// Обновляем пароль используя сохраненный новый пароль
	_, err = db.Pool.Exec(context.Background(),
		"UPDATE users SET passwordhash = $1 WHERE iduser = $2",
		string(hashedPassword), userID)
	if err != nil {
		http.Error(w, "Ошибка обновления пароля", http.StatusInternalServerError)
		return
	}

	// Помечаем код как использованный
	codesMutex.Lock()
	if storedCode, exists := resetCodes[email]; exists {
		storedCode.Used = true
		resetCodes[email] = storedCode
	}
	codesMutex.Unlock()

	fmt.Printf("✅ Пароль успешно изменен для пользователя ID: %d\n", userID)

	// Логируем смену пароля
	LogUserAction(r, "PASSWORD_CHANGE", "user", userID.(int), fmt.Sprintf("Изменен пароль для пользователя: %s", email))

	response := models.PasswordResponse{
		Message: "Пароль успешно изменен",
		Success: true,
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)
}

// generateConfirmationCode генерирует 6-значный код подтверждения
func generateConfirmationCode() string {
	// Генерируем случайное число от 100000 до 999999
	n, _ := rand.Int(rand.Reader, big.NewInt(900000))
	return fmt.Sprintf("%06d", n.Int64()+100000)
}

// isValidPassword проверяет валидность пароля
func isValidPassword(password string) bool {
	if len(password) < 8 {
		return false
	}
	
	hasDigit := false
	hasUpper := false
	
	for _, char := range password {
		if char >= '0' && char <= '9' {
			hasDigit = true
		}
		if char >= 'A' && char <= 'Z' {
			hasUpper = true
		}
	}
	
	return hasDigit && hasUpper
}

// sendPasswordResetEmail отправляет email с кодом восстановления пароля
func sendPasswordResetEmail(toEmail, code string) bool {
	fromEmail := "shoesstore0507@gmail.com"
	fromPassword := "bavu udva gljd gfka"
	smtpHost := "smtp.gmail.com"
	smtpPort := "587"

	subject := "🔐 Восстановление пароля - Shoes Store"
	body := fmt.Sprintf(`
Здравствуйте!

Вы запросили восстановление пароля для вашего аккаунта в Shoes Store.

Ваш код подтверждения: %s

Код действителен в течение 10 минут.

Если вы не запрашивали восстановление пароля, проигнорируйте это письмо.

С уважением,
Команда Shoes Store
`, code)

	return sendEmail(toEmail, subject, body, fromEmail, fromPassword, smtpHost, smtpPort)
}

// sendPasswordChangeEmail отправляет email с кодом подтверждения смены пароля
func sendPasswordChangeEmail(toEmail, code string) bool {
	fromEmail := "shoesstore0507@gmail.com"
	fromPassword := "bavu udva gljd gfka"
	smtpHost := "smtp.gmail.com"
	smtpPort := "587"

	subject := "🔑 Подтверждение смены пароля - Shoes Store"
	body := fmt.Sprintf(`
Здравствуйте!

Вы запросили смену пароля для вашего аккаунта в Shoes Store.

Ваш код подтверждения: %s

Код действителен в течение 10 минут.

Если вы не запрашивали смену пароля, проигнорируйте это письмо.

С уважением,
Команда Shoes Store
`, code)

	return sendEmail(toEmail, subject, body, fromEmail, fromPassword, smtpHost, smtpPort)
}

// sendEmail отправляет email (общая функция)
func sendEmail(toEmail, subject, body, fromEmail, fromPassword, smtpHost, smtpPort string) bool {
	fmt.Printf("📧 Отправка email на %s...\n", toEmail)

	message := fmt.Sprintf("From: %s\r\nTo: %s\r\nSubject: %s\r\n\r\n%s", 
		fromEmail, toEmail, subject, body)

	auth := smtp.PlainAuth("", fromEmail, fromPassword, smtpHost)

	err := smtp.SendMail(smtpHost+":"+smtpPort, auth, fromEmail, []string{toEmail}, []byte(message))
	if err != nil {
		fmt.Printf("❌ Ошибка отправки email: %v\n", err)
		return false
	}

	fmt.Printf("✅ Email успешно отправлен на %s\n", toEmail)
	return true
}
