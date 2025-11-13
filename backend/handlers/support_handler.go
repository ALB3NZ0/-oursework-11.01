package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/smtp"
	"regexp"
	"strings"
	"time"
	"crypto/tls"
)

// ---------------------------
// Support Request Structure
// ---------------------------
type SupportRequest struct {
	Name    string `json:"name" validate:"required,min=2,max=100"`
	Email   string `json:"email" validate:"required,email"`
	Message string `json:"message" validate:"required,min=15,max=2000"`
}

// ---------------------------
// Send Support Message
// ---------------------------

// @Summary Отправить сообщение в поддержку
// @Tags Support
// @Accept json
// @Produce json
// @Param support body SupportRequest true "Данные сообщения поддержки"
// @Success 200 {object} map[string]string "Success message"
// @Failure 400 {string} string "Validation error"
// @Failure 500 {string} string "Internal Server Error"
// @Router /support [post]
func SendSupportMessageHandler(w http.ResponseWriter, r *http.Request) {
	var supportReq SupportRequest
	if err := json.NewDecoder(r.Body).Decode(&supportReq); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Валидация данных
	if err := validateSupportMessage(supportReq); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// Отправляем email администратору
	adminEmail := "shoesstore0507@gmail.com"
	subject := "📩 Новое сообщение в поддержку"
	body := fmt.Sprintf("👤 Имя: %s\n📧 Email: %s\n⏰ Время: %s\n\n💬 Сообщение:\n%s", 
		supportReq.Name, supportReq.Email, time.Now().Format("2006-01-02 15:04:05"), supportReq.Message)

	emailSent := sendSupportEmail(adminEmail, subject, body)

	// Логируем действие
	LogUserAction(r, "SEND_SUPPORT", "support", 0, 
		fmt.Sprintf("Отправлено сообщение в поддержку от %s (%s)", supportReq.Name, supportReq.Email))

	// Отправляем ответ
	response := map[string]string{
		"message": "Ваше сообщение отправлено! Мы ответим вам на email.",
		"status":  "success",
	}

	if !emailSent {
		response["message"] = "Ошибка при отправке сообщения. Попробуйте позже."
		response["status"] = "error"
	}

	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	json.NewEncoder(w).Encode(response)
}

// ---------------------------
// Helper Functions
// ---------------------------

func validateSupportMessage(req SupportRequest) error {
	// Проверка имени
	if strings.TrimSpace(req.Name) == "" {
		return fmt.Errorf("Имя обязательно для заполнения")
	}
	if len(req.Name) < 2 || len(req.Name) > 100 {
		return fmt.Errorf("Имя должно содержать от 2 до 100 символов")
	}

	// Проверка email
	if strings.TrimSpace(req.Email) == "" {
		return fmt.Errorf("Email обязателен для заполнения")
	}
	emailRegex := regexp.MustCompile(`^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`)
	if !emailRegex.MatchString(req.Email) {
		return fmt.Errorf("Некорректный формат email")
	}

	// Проверка сообщения
	trimmedMessage := strings.TrimSpace(req.Message)
	if trimmedMessage == "" {
		return fmt.Errorf("Сообщение обязательно для заполнения")
	}
	if len(trimmedMessage) < 15 {
		return fmt.Errorf("Сообщение должно содержать не менее 15 символов")
	}
	if len(req.Message) > 2000 {
		return fmt.Errorf("Сообщение не должно превышать 2000 символов")
	}

	// Проверка, что сообщение (после удаления пробелов) содержит хотя бы одну букву или цифру
	// Это обеспечивает, что сообщение не состоит только из специальных символов
	hasLetterOrDigit := regexp.MustCompile(`[А-Яа-яA-Za-z0-9]`).MatchString(trimmedMessage)
	if !hasLetterOrDigit {
		return fmt.Errorf("Сообщение должно содержать хотя бы одну букву или цифру")
	}

	return nil
}

func sendSupportEmail(toEmail, subject, body string) bool {
	// Настройки SMTP для Gmail
	fromEmail := "shoesstore0507@gmail.com"
	fromPassword := "bavu udva gljd gfka"
	smtpHost := "smtp.gmail.com"
	smtpPort := "587"

	fmt.Printf("📧 Попытка отправки email...\n")
	fmt.Printf("From: %s\n", fromEmail)
	fmt.Printf("To: %s\n", toEmail)
	fmt.Printf("Subject: %s\n", subject)
	fmt.Printf("SMTP: %s:%s\n", smtpHost, smtpPort)

	// Создаем сообщение с правильной кодировкой для кириллицы
	message := fmt.Sprintf("From: %s\r\nTo: %s\r\nSubject: %s\r\nContent-Type: text/plain; charset=UTF-8\r\n\r\n%s", 
		fromEmail, toEmail, subject, body)

	// Настройки аутентификации
	auth := smtp.PlainAuth("", fromEmail, fromPassword, smtpHost)

	// Попробуем подключиться с TLS
	conn, err := tls.Dial("tcp", smtpHost+":587", &tls.Config{
		ServerName: smtpHost,
	})
	if err != nil {
		fmt.Printf("❌ Ошибка TLS подключения: %v\n", err)
		fmt.Printf("🔄 Пробуем обычное подключение...\n")
		
		// Fallback к обычному методу
		err = smtp.SendMail(smtpHost+":587", auth, fromEmail, []string{toEmail}, []byte(message))
		if err != nil {
			fmt.Printf("❌ Ошибка отправки email: %v\n", err)
			fmt.Printf("❌ Проверьте настройки SMTP и пароль приложения\n")
			fmt.Printf("❌ Возможно, порт заблокирован провайдером или файрволом\n")
			return false
		}
	} else {
		defer conn.Close()
		
		// Создаем SMTP клиент
		client, err := smtp.NewClient(conn, smtpHost)
		if err != nil {
			fmt.Printf("❌ Ошибка создания SMTP клиента: %v\n", err)
			return false
		}
		defer client.Quit()

		// Аутентификация
		if err = client.Auth(auth); err != nil {
			fmt.Printf("❌ Ошибка аутентификации: %v\n", err)
			return false
		}

		// Отправка
		if err = client.Mail(fromEmail); err != nil {
			fmt.Printf("❌ Ошибка MAIL: %v\n", err)
			return false
		}

		if err = client.Rcpt(toEmail); err != nil {
			fmt.Printf("❌ Ошибка RCPT: %v\n", err)
			return false
		}

		writer, err := client.Data()
		if err != nil {
			fmt.Printf("❌ Ошибка DATA: %v\n", err)
			return false
		}

		_, err = writer.Write([]byte(message))
		if err != nil {
			fmt.Printf("❌ Ошибка записи: %v\n", err)
			return false
		}

		err = writer.Close()
		if err != nil {
			fmt.Printf("❌ Ошибка закрытия writer: %v\n", err)
			return false
		}
	}

	fmt.Printf("✅ Email успешно отправлен на %s\n", toEmail)
	return true
}
