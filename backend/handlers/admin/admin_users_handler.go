package admin

import (
    "context"
    "encoding/json"
    "fmt"
    "net/http"
    "strconv"

    "golang.org/x/crypto/bcrypt"
    "github.com/gorilla/mux"
    "shoes-store-backend/db"
    "shoes-store-backend/handlers"
    "shoes-store-backend/models"
)


// ---------------------------
// Create User (Admin)
// ---------------------------

// @Summary Создать пользователя (Admin)
// @Tags Admin Users
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param user body models.User true "Новый пользователь"
// @Success 201 {object} models.User
// @Failure 400 {string} string "Invalid request"
// @Failure 500 {string} string "Internal Server Error"
// @Router /admin/users [post]
func AdminCreateUserHandler(w http.ResponseWriter, r *http.Request) {
    var user models.User
    if err := json.NewDecoder(r.Body).Decode(&user); err != nil {
        http.Error(w, "Invalid request body", http.StatusBadRequest)
        return
    }

    if user.Email == "" || user.PasswordHash == "" {
        http.Error(w, "Email and password are required", http.StatusBadRequest)
        return
    }

    // Хэшируем пароль
    hashedPassword, err := bcrypt.GenerateFromPassword([]byte(user.PasswordHash), bcrypt.DefaultCost)
    if err != nil {
        http.Error(w, "Failed to hash password", http.StatusInternalServerError)
        return
    }

    // Добавляем пользователя в базу
    err = db.Pool.QueryRow(context.Background(),
        `INSERT INTO users (fullname, email, passwordhash, roleid)
         VALUES ($1, $2, $3, $4)
         RETURNING iduser`,
        user.FullName, user.Email, string(hashedPassword), user.RoleID,
    ).Scan(&user.ID)

    if err != nil {
        http.Error(w, err.Error(), http.StatusInternalServerError)
        return
    }

    // Логируем действие админа
    LogUserAction(r, "CREATE", "user", user.ID, fmt.Sprintf("Создан пользователь: %s (%s)", user.FullName, user.Email))

    user.PasswordHash = "" // Не возвращаем пароль в ответе
    w.Header().Set("Content-Type", "application/json")
    w.WriteHeader(http.StatusCreated)
    json.NewEncoder(w).Encode(user)
}

// ---------------------------
// Get All Users (Admin)
// ---------------------------

// @Summary Получить всех пользователей (Admin, с пагинацией)
// @Tags Admin Users
// @Produce json
// @Security BearerAuth
// @Param page query int false "Номер страницы (по умолчанию 1)"
// @Param limit query int false "Количество элементов на странице (по умолчанию 20, максимум 100)"
// @Success 200 {object} handlers.PaginatedResponse
// @Failure 500 {string} string "Internal Server Error"
// @Router /admin/users [get]
func AdminGetUsersHandler(w http.ResponseWriter, r *http.Request) {
    params := handlers.ParsePaginationParams(r)

    // Получаем общее количество пользователей
    var total int
    err := db.Pool.QueryRow(context.Background(),
        `SELECT COUNT(*) FROM users`).Scan(&total)
    if err != nil {
        http.Error(w, err.Error(), http.StatusInternalServerError)
        return
    }

    // Получаем пользователей с пагинацией
    rows, err := db.Pool.Query(context.Background(),
        `SELECT iduser, fullname, email, roleid FROM users ORDER BY iduser LIMIT $1 OFFSET $2`,
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

    totalPages := handlers.CalculateTotalPages(total, params.Limit)
    response := handlers.PaginatedResponse{
        Data:       users,
        Page:       params.Page,
        Limit:      params.Limit,
        Total:      total,
        TotalPages: totalPages,
    }

    w.Header().Set("Content-Type", "application/json")
    json.NewEncoder(w).Encode(response)
}

// ---------------------------
// Get User by ID (Admin)
// ---------------------------

// @Summary Получить пользователя по ID (Admin)
// @Tags Admin Users
// @Produce json
// @Security BearerAuth
// @Param id path int true "User ID"
// @Success 200 {object} models.User
// @Failure 400 {string} string "Invalid ID"
// @Failure 404 {string} string "User not found"
// @Router /admin/users/{id} [get]
func AdminGetUserByIDHandler(w http.ResponseWriter, r *http.Request) {
    vars := mux.Vars(r)
    id, err := strconv.Atoi(vars["id"])
    if err != nil {
        http.Error(w, "Invalid ID", http.StatusBadRequest)
        return
    }

    var user models.User
    err = db.Pool.QueryRow(context.Background(),
        `SELECT iduser, fullname, email, roleid FROM users WHERE iduser=$1`, id).
        Scan(&user.ID, &user.FullName, &user.Email, &user.RoleID)
    if err != nil {
        http.Error(w, "User not found", http.StatusNotFound)
        return
    }

    w.Header().Set("Content-Type", "application/json")
    json.NewEncoder(w).Encode(user)
}

// ---------------------------
// Update User (Admin)
// ---------------------------

// @Summary Обновить пользователя (Admin)
// @Tags Admin Users
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path int true "User ID"
// @Param user body models.User true "Обновлённые данные пользователя"
// @Success 200 {object} models.User
// @Failure 400 {string} string "Invalid request"
// @Failure 404 {string} string "User not found"
// @Failure 500 {string} string "Internal Server Error"
// @Router /admin/users/{id} [put]
func AdminUpdateUserHandler(w http.ResponseWriter, r *http.Request) {
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

    _, err = db.Pool.Exec(context.Background(),
        `UPDATE users SET fullname=$1, email=$2, roleid=$3 WHERE iduser=$4`,
        user.FullName, user.Email, user.RoleID, id)
    if err != nil {
        http.Error(w, err.Error(), http.StatusInternalServerError)
        return
    }

    // Логируем действие админа
    LogUserAction(r, "UPDATE", "user", id, fmt.Sprintf("Обновлен пользователь: %s (%s)", user.FullName, user.Email))

    user.ID = id
    w.Header().Set("Content-Type", "application/json")
    json.NewEncoder(w).Encode(user)
}

// ---------------------------
// Delete User (Admin)
// ---------------------------

// @Summary Удалить пользователя (Admin)
// @Tags Admin Users
// @Produce json
// @Security BearerAuth
// @Param id path int true "User ID"
// @Success 204 "No Content"
// @Failure 400 {string} string "Invalid ID"
// @Failure 404 {string} string "User not found"
// @Router /admin/users/{id} [delete]
func AdminDeleteUserHandler(w http.ResponseWriter, r *http.Request) {
    vars := mux.Vars(r)
    id, err := strconv.Atoi(vars["id"])
    if err != nil {
        http.Error(w, "Invalid ID", http.StatusBadRequest)
        return
    }

    res, err := db.Pool.Exec(context.Background(),
        `DELETE FROM users WHERE iduser=$1`, id)
    if err != nil {
        http.Error(w, err.Error(), http.StatusInternalServerError)
        return
    }

    if res.RowsAffected() == 0 {
        http.Error(w, "User not found", http.StatusNotFound)
        return
    }

    // Логируем действие админа
    LogUserAction(r, "DELETE", "user", id, fmt.Sprintf("Удален пользователь с ID: %d", id))

    w.WriteHeader(http.StatusNoContent)
}
