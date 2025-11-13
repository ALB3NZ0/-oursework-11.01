package handlers

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"

	"github.com/gorilla/mux"
	"golang.org/x/crypto/bcrypt"
	"shoes-store-backend/db"
	"shoes-store-backend/models"
	"shoes-store-backend/middlewares"
)

func RoleIDToString(roleID int) string {
	switch roleID {
	case 1:
		return "admin"
	case 2:
		return "manager"
	case 3:
		return "user"
	default:
		return "user"
	}
}

// ------------------
// Hello
// ------------------

// @Summary Приветствие
// @Tags General
// @Produce plain
// @Success 200 {string} string "Hello, world!"
// @Router / [get]
func HelloHandler(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("Hello, world!"))
}

// ------------------
// Register
// ------------------

// @Summary Регистрация пользователя
// @Description Создаёт нового пользователя с хэшированным паролем
// @Tags Auth
// @Accept json
// @Produce json
// @Param user body models.User true "Данные пользователя"
// @Success 201 {object} models.User
// @Failure 400 {string} string "Invalid request body"
// @Failure 500 {string} string "Internal Server Error"
// @Router /register [post]
func RegisterHandler(w http.ResponseWriter, r *http.Request) {
	var user models.User
	if err := json.NewDecoder(r.Body).Decode(&user); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}
	
	fmt.Printf("📝 Register: Email=%s, PasswordHash length=%d\n", user.Email, len(user.PasswordHash))

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(user.PasswordHash), bcrypt.DefaultCost)
	if err != nil {
		fmt.Printf("❌ Register: Failed to hash password - %v\n", err)
		http.Error(w, "Failed to hash password", http.StatusInternalServerError)
		return
	}
	
	fmt.Printf("📝 Register: Password hashed successfully for: %s\n", user.Email)

	// Всегда устанавливаем роль "user" (3) для новых регистраций
	user.RoleID = 3

	err = db.Pool.QueryRow(context.Background(),
		"INSERT INTO users (fullname, email, passwordhash, roleid) VALUES ($1, $2, $3, $4) RETURNING iduser",
		user.FullName, user.Email, string(hashedPassword), user.RoleID,
	).Scan(&user.ID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Логируем регистрацию пользователя
	LogUserAction(r, "REGISTER", "user", user.ID, fmt.Sprintf("Зарегистрирован пользователь: %s (%s)", user.FullName, user.Email))

	user.PasswordHash = ""
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(user)
}

// ------------------
// Login
// ------------------

// @Summary Логин пользователя
// @Description Проверяет email и пароль пользователя
// @Tags Auth
// @Accept json
// @Produce json
// @Param login body models.LoginRequest true "Данные для входа"
// @Success 200 {object} models.User
// @Failure 400 {string} string "Invalid request"
// @Failure 401 {string} string "Invalid email or password"
// @Router /login [post]
func LoginHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Printf("🔐 LOGIN REQUEST: Method=%s, Origin=%s\n", r.Method, r.Header.Get("Origin"))
	
	var req models.LoginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		fmt.Printf("❌ Login: Invalid request body - %v\n", err)
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}
	
	fmt.Printf("📧 Login attempt for email: %s\n", req.Email)

	var user models.User
	err := db.Pool.QueryRow(context.Background(),
		"SELECT iduser, fullname, email, passwordhash, roleid FROM users WHERE email=$1", req.Email).
		Scan(&user.ID, &user.FullName, &user.Email, &user.PasswordHash, &user.RoleID)
	if err != nil {
		fmt.Printf("❌ Login: User not found - email: %s, error: %v\n", req.Email, err)
		http.Error(w, "Invalid email or password", http.StatusUnauthorized)
		return
	}
	
	fmt.Printf("✅ User found: ID=%d, Name=%s, RoleID=%d\n", user.ID, user.FullName, user.RoleID)
	fmt.Printf("   Password hash length: %d\n", len(user.PasswordHash))
	fmt.Printf("   Input password length: %d\n", len(req.Password))

	err = bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(req.Password))
	if err != nil {
		fmt.Printf("❌ Login: Invalid password for user: %s\n", req.Email)
		fmt.Printf("   Error: %v\n", err)
		http.Error(w, "Invalid email or password", http.StatusUnauthorized)
		return
	}
	
	fmt.Printf("✅ Password verified for user: %s\n", req.Email)

	user.PasswordHash = ""

	// Генерация JWT
	token, err := middlewares.GenerateJWT(user.ID, RoleIDToString(user.RoleID))
	if err != nil {
		fmt.Printf("❌ Login: Failed to generate token - %v\n", err)
		http.Error(w, "Could not generate token", http.StatusInternalServerError)
		return
	}

	fmt.Printf("✅ LOGIN SUCCESS: User=%s, ID=%d, Token generated\n", user.FullName, user.ID)

	// Return both token and full user data
	response := map[string]interface{}{
		"token": token,
		"user":  user,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}


// ------------------
// CRUD Users
// ------------------

// @Summary Получить всех пользователей (с пагинацией)
// @Description Возвращает список всех пользователей с пагинацией
// @Tags Users
// @Produce json
// @Security BearerAuth
// @Param page query int false "Номер страницы (по умолчанию 1)"
// @Param limit query int false "Количество элементов на странице (по умолчанию 20, максимум 100)"
// @Success 200 {object} PaginatedResponse
// @Failure 500 {string} string "Internal Server Error"
// @Router /users [get]
func GetUsersHandler(w http.ResponseWriter, r *http.Request) {
	params := ParsePaginationParams(r)

	// Получаем общее количество пользователей
	var total int
	err := db.Pool.QueryRow(context.Background(), "SELECT COUNT(*) FROM users").Scan(&total)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Получаем пользователей с пагинацией
	rows, err := db.Pool.Query(context.Background(),
		"SELECT iduser, fullname, email, roleid FROM users ORDER BY iduser LIMIT $1 OFFSET $2",
		params.Limit, params.GetOffset())
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	var users []models.User
	for rows.Next() {
		var u models.User
		if err := rows.Scan(&u.ID, &u.FullName, &u.Email, &u.RoleID); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		users = append(users, u)
	}

	totalPages := CalculateTotalPages(total, params.Limit)
	response := PaginatedResponse{
		Data:       users,
		Page:       params.Page,
		Limit:      params.Limit,
		Total:      total,
		TotalPages: totalPages,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// @Summary Получить пользователя по ID
// @Description Возвращает одного пользователя по ID
// @Tags Users
// @Produce json
// @Security BearerAuth
// @Param id path int true "User ID"
// @Success 200 {object} models.User
// @Failure 400 {string} string "Invalid ID"
// @Failure 404 {string} string "User not found"
// @Router /users/{id} [get]
func GetUserByIDHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id, err := strconv.Atoi(vars["id"])
	if err != nil {
		http.Error(w, "Invalid ID", http.StatusBadRequest)
		return
	}

	var user models.User
	err = db.Pool.QueryRow(context.Background(),
		"SELECT iduser, fullname, email, roleid FROM users WHERE iduser=$1", id).
		Scan(&user.ID, &user.FullName, &user.Email, &user.RoleID)
	if err != nil {
		http.Error(w, "User not found", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(user)
}

// @Summary Обновить пользователя
// @Description Обновляет данные пользователя
// @Tags Users
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path int true "User ID"
// @Param user body models.User true "Данные пользователя"
// @Success 200 {object} models.User
// @Failure 400 {string} string "Invalid request"
// @Failure 500 {string} string "Internal Server Error"
// @Router /users/{id} [put]
func UpdateUserHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id, err := strconv.Atoi(vars["id"])
	if err != nil {
		http.Error(w, "Invalid ID", http.StatusBadRequest)
		return
	}

	var user models.User
	if err := json.NewDecoder(r.Body).Decode(&user); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if user.PasswordHash != "" {
		hash, err := bcrypt.GenerateFromPassword([]byte(user.PasswordHash), bcrypt.DefaultCost)
		if err != nil {
			http.Error(w, "Failed to hash password", http.StatusInternalServerError)
			return
		}
		user.PasswordHash = string(hash)
		_, err = db.Pool.Exec(context.Background(),
			"UPDATE users SET fullname=$1, email=$2, passwordhash=$3, roleid=$4 WHERE iduser=$5",
			user.FullName, user.Email, user.PasswordHash, user.RoleID, id)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	} else {
		_, err = db.Pool.Exec(context.Background(),
			"UPDATE users SET fullname=$1, email=$2, roleid=$3 WHERE iduser=$4",
			user.FullName, user.Email, user.RoleID, id)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	}

	// Логируем обновление пользователя
	LogUserAction(r, "UPDATE", "user", id, fmt.Sprintf("Обновлен пользователь: %s (%s)", user.FullName, user.Email))

	user.ID = id
	user.PasswordHash = ""
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(user)
}

// @Summary Удалить пользователя
// @Description Удаляет пользователя по ID
// @Tags Users
// @Security BearerAuth
// @Param id path int true "User ID"
// @Success 204 {string} string "No Content"
// @Failure 400 {string} string "Invalid ID"
// @Failure 404 {string} string "User not found"
// @Failure 500 {string} string "Internal Server Error"
// @Router /users/{id} [delete]
func DeleteUserHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id, err := strconv.Atoi(vars["id"])
	if err != nil {
		http.Error(w, "Invalid ID", http.StatusBadRequest)
		return
	}

	tag, err := db.Pool.Exec(context.Background(),
		"DELETE FROM users WHERE iduser=$1", id)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	if tag.RowsAffected() == 0 {
		http.Error(w, "User not found", http.StatusNotFound)
		return
	}

	// Логируем удаление пользователя
	LogUserAction(r, "DELETE", "user", id, fmt.Sprintf("Удален пользователь с ID: %d", id))

	w.WriteHeader(http.StatusNoContent)
}



// @Summary Создать пользователя
// @Description Создаёт нового пользователя (только для админов)
// @Tags Users
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param user body models.User true "Данные пользователя"
// @Success 201 {object} models.User
// @Failure 400 {string} string "Invalid request body"
// @Failure 500 {string} string "Internal Server Error"
// @Router /users [post]
func CreateUserHandler(w http.ResponseWriter, r *http.Request) {
	var user models.User
	if err := json.NewDecoder(r.Body).Decode(&user); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if user.PasswordHash != "" {
		hash, err := bcrypt.GenerateFromPassword([]byte(user.PasswordHash), bcrypt.DefaultCost)
		if err != nil {
			http.Error(w, "Failed to hash password", http.StatusInternalServerError)
			return
		}
		user.PasswordHash = string(hash)
	}

	err := db.Pool.QueryRow(context.Background(),
		"INSERT INTO users (fullname, email, passwordhash, roleid) VALUES ($1, $2, $3, $4) RETURNING iduser",
		user.FullName, user.Email, user.PasswordHash, user.RoleID,
	).Scan(&user.ID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Логируем создание пользователя
	LogUserAction(r, "CREATE", "user", user.ID, fmt.Sprintf("Создан пользователь: %s (%s)", user.FullName, user.Email))

	user.PasswordHash = ""
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(user)
}
