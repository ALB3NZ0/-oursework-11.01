package admin

import (
    "context"
    "encoding/json"
    "fmt"
    "net/http"
    "strconv"

    "github.com/gorilla/mux"
    "shoes-store-backend/db"
    "shoes-store-backend/handlers"
    "shoes-store-backend/models"
)


// ---------------------------
// Get All Reviews (Admin)
// ---------------------------

// @Summary Получить все отзывы (Admin, с пагинацией)
// @Tags Admin Reviews
// @Produce json
// @Security BearerAuth
// @Param page query int false "Номер страницы (по умолчанию 1)"
// @Param limit query int false "Количество элементов на странице (по умолчанию 20, максимум 100)"
// @Success 200 {object} handlers.PaginatedResponse
// @Failure 500 {string} string "Internal Server Error"
// @Router /admin/reviews [get]
func AdminGetReviewsHandler(w http.ResponseWriter, r *http.Request) {
    params := handlers.ParsePaginationParams(r)

    // Получаем общее количество отзывов
    var total int
    err := db.Pool.QueryRow(context.Background(),
        `SELECT COUNT(*) FROM reviews`).Scan(&total)
    if err != nil {
        http.Error(w, err.Error(), http.StatusInternalServerError)
        return
    }

    // Получаем отзывы с пагинацией
    rows, err := db.Pool.Query(context.Background(),
        `SELECT idreview, idproduct, iduser, rating, comment, reviewdate 
         FROM reviews ORDER BY reviewdate DESC LIMIT $1 OFFSET $2`,
        params.Limit, params.GetOffset())
    if err != nil {
        http.Error(w, err.Error(), http.StatusInternalServerError)
        return
    }
    defer rows.Close()

    var reviews []models.Review
    for rows.Next() {
        var rev models.Review
        if err := rows.Scan(&rev.ID, &rev.ProductID, &rev.UserID, &rev.Rating, &rev.Comment, &rev.Date); err != nil {
            http.Error(w, err.Error(), http.StatusInternalServerError)
            return
        }
        reviews = append(reviews, rev)
    }

    totalPages := handlers.CalculateTotalPages(total, params.Limit)
    response := handlers.PaginatedResponse{
        Data:       reviews,
        Page:       params.Page,
        Limit:      params.Limit,
        Total:      total,
        TotalPages: totalPages,
    }

    w.Header().Set("Content-Type", "application/json")
    json.NewEncoder(w).Encode(response)
}

// ---------------------------
// Get Review by ID (Admin)
// ---------------------------

// @Summary Получить отзыв по ID (Admin)
// @Tags Admin Reviews
// @Produce json
// @Security BearerAuth
// @Param id path int true "Review ID"
// @Success 200 {object} models.Review
// @Failure 400 {string} string "Invalid ID"
// @Failure 404 {string} string "Review not found"
// @Router /admin/reviews/{id} [get]
func AdminGetReviewByIDHandler(w http.ResponseWriter, r *http.Request) {
    vars := mux.Vars(r)
    id, err := strconv.Atoi(vars["id"])
    if err != nil {
        http.Error(w, "Invalid ID", http.StatusBadRequest)
        return
    }

    var rev models.Review
    err = db.Pool.QueryRow(context.Background(),
        `SELECT idreview, idproduct, iduser, rating, comment, reviewdate 
         FROM reviews WHERE idreview=$1`, id).
        Scan(&rev.ID, &rev.ProductID, &rev.UserID, &rev.Rating, &rev.Comment, &rev.Date)
    if err != nil {
        http.Error(w, "Review not found", http.StatusNotFound)
        return
    }

    w.Header().Set("Content-Type", "application/json")
    json.NewEncoder(w).Encode(rev)
}

// ---------------------------
// Update Review (Admin)
// ---------------------------

// @Summary Обновить отзыв (Admin)
// @Tags Admin Reviews
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path int true "Review ID"
// @Param review body models.Review true "Review data"
// @Success 200 {object} models.Review
// @Failure 400 {string} string "Invalid request"
// @Failure 404 {string} string "Review not found"
// @Router /admin/reviews/{id} [put]
func AdminUpdateReviewHandler(w http.ResponseWriter, r *http.Request) {
    vars := mux.Vars(r)
    id, err := strconv.Atoi(vars["id"])
    if err != nil {
        http.Error(w, "Invalid ID", http.StatusBadRequest)
        return
    }

    var rev models.Review
    if err := json.NewDecoder(r.Body).Decode(&rev); err != nil {
        http.Error(w, "Invalid request body", http.StatusBadRequest)
        return
    }

    // Проверяем валидность рейтинга (1-5)
    if rev.Rating < 1 || rev.Rating > 5 {
        http.Error(w, "Рейтинг должен быть от 1 до 5", http.StatusBadRequest)
        return
    }

    res, err := db.Pool.Exec(context.Background(),
        `UPDATE reviews SET rating=$1, comment=$2 WHERE idreview=$3`,
        rev.Rating, rev.Comment, id)
    if err != nil {
        http.Error(w, err.Error(), http.StatusInternalServerError)
        return
    }

    rowsAffected := res.RowsAffected()
    if rowsAffected == 0 {
        http.Error(w, "Review not found", http.StatusNotFound)
        return
    }

    // Логируем действие админа
    LogUserAction(r, "UPDATE", "review", id, fmt.Sprintf("Обновлен отзыв с ID: %d", id))

    // Получаем обновленный отзыв
    err = db.Pool.QueryRow(context.Background(),
        `SELECT idreview, idproduct, iduser, rating, comment, reviewdate 
         FROM reviews WHERE idreview=$1`, id).
        Scan(&rev.ID, &rev.ProductID, &rev.UserID, &rev.Rating, &rev.Comment, &rev.Date)
    if err != nil {
        http.Error(w, "Review not found", http.StatusNotFound)
        return
    }

    w.Header().Set("Content-Type", "application/json")
    json.NewEncoder(w).Encode(rev)
}

// ---------------------------
// Delete Review (Admin)
// ---------------------------

// @Summary Удалить отзыв (Admin)
// @Tags Admin Reviews
// @Produce json
// @Security BearerAuth
// @Param id path int true "Review ID"
// @Success 204 "No Content"
// @Failure 400 {string} string "Invalid ID"
// @Failure 404 {string} string "Review not found"
// @Router /admin/reviews/{id} [delete]
func AdminDeleteReviewHandler(w http.ResponseWriter, r *http.Request) {
    vars := mux.Vars(r)
    id, err := strconv.Atoi(vars["id"])
    if err != nil {
        http.Error(w, "Invalid ID", http.StatusBadRequest)
        return
    }

    res, err := db.Pool.Exec(context.Background(),
        `DELETE FROM reviews WHERE idreview=$1`, id)
    if err != nil {
        http.Error(w, err.Error(), http.StatusInternalServerError)
        return
    }

    rowsAffected := res.RowsAffected()
    if rowsAffected == 0 {
        http.Error(w, "Review not found", http.StatusNotFound)
        return
    }

    // Логируем действие админа
    LogUserAction(r, "DELETE", "review", id, fmt.Sprintf("Удален отзыв с ID: %d", id))

    w.WriteHeader(http.StatusNoContent)
}
