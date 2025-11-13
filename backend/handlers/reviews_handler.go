package handlers

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/gorilla/mux"
	"shoes-store-backend/db"
	"shoes-store-backend/models"
)

// ------------------
// GET reviews by product
// ------------------
// @Summary Получить отзывы о продукте (с пагинацией)
// @Tags Reviews
// @Produce json
// @Param id path int true "Product ID"
// @Param page query int false "Номер страницы (по умолчанию 1)"
// @Param limit query int false "Количество элементов на странице (по умолчанию 20, максимум 100)"
// @Success 200 {object} PaginatedResponse
// @Failure 400 {string} string "Invalid Product ID"
// @Router /reviews/product/{id} [get]
func GetReviewsByProductHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	productID, err := strconv.Atoi(vars["id"])
	if err != nil {
		http.Error(w, "Invalid Product ID", http.StatusBadRequest)
		return
	}

	params := ParsePaginationParams(r)

	// Получаем общее количество отзывов о продукте
	var total int
	err = db.Pool.QueryRow(context.Background(),
		"SELECT COUNT(*) FROM reviews WHERE idproduct=$1", productID).Scan(&total)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Получаем отзывы с пагинацией
	rows, err := db.Pool.Query(context.Background(),
		"SELECT idreview, idproduct, iduser, rating, comment, reviewdate FROM reviews WHERE idproduct=$1 ORDER BY reviewdate DESC LIMIT $2 OFFSET $3",
		productID, params.Limit, params.GetOffset())
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	var reviews []models.Review
	for rows.Next() {
		var r models.Review
		if err := rows.Scan(&r.ID, &r.ProductID, &r.UserID, &r.Rating, &r.Comment, &r.Date); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		reviews = append(reviews, r)
	}

	totalPages := CalculateTotalPages(total, params.Limit)
	response := PaginatedResponse{
		Data:       reviews,
		Page:       params.Page,
		Limit:      params.Limit,
		Total:      total,
		TotalPages: totalPages,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// ------------------
// GET reviews by user
// ------------------
// @Summary Получить отзывы пользователя (с пагинацией)
// @Tags Reviews
// @Produce json
// @Security BearerAuth
// @Param id path int true "User ID"
// @Param page query int false "Номер страницы (по умолчанию 1)"
// @Param limit query int false "Количество элементов на странице (по умолчанию 20, максимум 100)"
// @Success 200 {object} PaginatedResponse
// @Failure 400 {string} string "Invalid User ID"
// @Router /reviews/user/{id} [get]
func GetReviewsByUserHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	userID, err := strconv.Atoi(vars["id"])
	if err != nil {
		http.Error(w, "Invalid User ID", http.StatusBadRequest)
		return
	}

	params := ParsePaginationParams(r)

	// Получаем общее количество отзывов пользователя
	var total int
	err = db.Pool.QueryRow(context.Background(),
		"SELECT COUNT(*) FROM reviews WHERE iduser=$1", userID).Scan(&total)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Получаем отзывы пользователя с пагинацией
	rows, err := db.Pool.Query(context.Background(),
		"SELECT idreview, idproduct, iduser, rating, comment, reviewdate FROM reviews WHERE iduser=$1 ORDER BY reviewdate DESC LIMIT $2 OFFSET $3",
		userID, params.Limit, params.GetOffset())
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	var reviews []models.Review
	for rows.Next() {
		var r models.Review
		if err := rows.Scan(&r.ID, &r.ProductID, &r.UserID, &r.Rating, &r.Comment, &r.Date); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		reviews = append(reviews, r)
	}

	totalPages := CalculateTotalPages(total, params.Limit)
	response := PaginatedResponse{
		Data:       reviews,
		Page:       params.Page,
		Limit:      params.Limit,
		Total:      total,
		TotalPages: totalPages,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// ------------------
// POST review
// ------------------
// @Summary Добавить отзыв
// @Tags Reviews
// @Accept json
// @Produce json
// @Param review body models.Review true "Review"
// @Success 201 {object} models.Review
// @Failure 400 {string} string "Invalid request body"
// @Router /reviews [post]
func CreateReviewHandler(w http.ResponseWriter, r *http.Request) {
	// Получаем пользователя из контекста
	userID := r.Context().Value("userID")
	if userID == nil {
		http.Error(w, "Пользователь не авторизован", http.StatusUnauthorized)
		return
	}

	var review models.Review
	if err := json.NewDecoder(r.Body).Decode(&review); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Устанавливаем userID из контекста
	review.UserID = userID.(int)

	// Проверяем валидность рейтинга (1-5)
	if review.Rating < 1 || review.Rating > 5 {
		http.Error(w, "Рейтинг должен быть от 1 до 5", http.StatusBadRequest)
		return
	}

	// Проверяем, что пользователь купил этот товар
	var hasPurchase bool
	err := db.Pool.QueryRow(context.Background(),
		`SELECT EXISTS(
			SELECT 1 FROM orderproducts op
			JOIN productsizes ps ON op.idproductsize = ps.idproductsize
			JOIN orders o ON op.idorder = o.idorder
			WHERE ps.idproduct = $1 AND o.iduser = $2
		)`, review.ProductID, review.UserID).Scan(&hasPurchase)
	
	if err != nil {
		http.Error(w, "Ошибка проверки покупки товара", http.StatusInternalServerError)
		return
	}

	if !hasPurchase {
		http.Error(w, "Вы можете оставить отзыв только на товары, которые вы купили", http.StatusForbidden)
		return
	}

	// Проверяем, не оставил ли пользователь уже отзыв на этот товар
	var existingReviewID int
	err = db.Pool.QueryRow(context.Background(),
		"SELECT idreview FROM reviews WHERE idproduct=$1 AND iduser=$2",
		review.ProductID, review.UserID).Scan(&existingReviewID)
	
	if err == nil {
		// Отзыв уже существует
		http.Error(w, "Вы уже оставили отзыв на этот товар. Вы можете обновить его.", http.StatusConflict)
		return
	}

	// Создаем отзыв
	err = db.Pool.QueryRow(context.Background(),
		"INSERT INTO reviews (idproduct, iduser, rating, comment, reviewdate) VALUES ($1, $2, $3, $4, $5) RETURNING idreview",
		review.ProductID, review.UserID, review.Rating, review.Comment, time.Now()).
		Scan(&review.ID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Логируем создание отзыва
	LogUserAction(r, "CREATE", "review", review.ID, fmt.Sprintf("Создан отзыв для товара ID: %d пользователем ID: %d", review.ProductID, review.UserID))

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(review)
}

// ------------------
// PUT review
// ------------------
// @Summary Обновить отзыв
// @Tags Reviews
// @Accept json
// @Produce json
// @Param id path int true "Review ID"
// @Param review body models.Review true "Review"
// @Success 200 {object} models.Review
// @Failure 400 {string} string "Invalid request"
// @Router /reviews/{id} [put]
func UpdateReviewHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id, err := strconv.Atoi(vars["id"])
	if err != nil {
		http.Error(w, "Invalid ID", http.StatusBadRequest)
		return
	}

	var review models.Review
	if err := json.NewDecoder(r.Body).Decode(&review); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	_, err = db.Pool.Exec(context.Background(),
		"UPDATE reviews SET rating=$1, comment=$2 WHERE idreview=$3",
		review.Rating, review.Comment, id)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Логируем обновление отзыва
	LogUserAction(r, "UPDATE", "review", id, fmt.Sprintf("Обновлен отзыв ID: %d", id))

	review.ID = id
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(review)
}

// ------------------
// DELETE review
// ------------------
// @Summary Удалить отзыв
// @Tags Reviews
// @Param id path int true "Review ID"
// @Success 204 {string} string "No Content"
// @Failure 400 {string} string "Invalid ID"
// @Router /reviews/{id} [delete]
func DeleteReviewHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id, err := strconv.Atoi(vars["id"])
	if err != nil {
		http.Error(w, "Invalid ID", http.StatusBadRequest)
		return
	}

	_, err = db.Pool.Exec(context.Background(),
		"DELETE FROM reviews WHERE idreview=$1", id)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Логируем удаление отзыва
	LogUserAction(r, "DELETE", "review", id, fmt.Sprintf("Удален отзыв ID: %d", id))

	w.WriteHeader(http.StatusNoContent)
}
