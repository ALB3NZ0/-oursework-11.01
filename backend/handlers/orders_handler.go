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
// GET all orders
// ------------------
// @Summary Получить все заказы (с пагинацией)
// @Tags Orders
// @Produce json
// @Param page query int false "Номер страницы (по умолчанию 1)"
// @Param limit query int false "Количество элементов на странице (по умолчанию 20, максимум 100)"
// @Success 200 {object} PaginatedResponse
// @Failure 500 {string} string "Internal Server Error"
// @Router /orders [get]
func GetOrdersHandler(w http.ResponseWriter, r *http.Request) {
	params := ParsePaginationParams(r)

	// Получаем общее количество заказов
	var total int
	err := db.Pool.QueryRow(context.Background(),
		"SELECT COUNT(*) FROM orders").Scan(&total)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Получаем заказы с пагинацией
	rows, err := db.Pool.Query(context.Background(),
		"SELECT idorder, iduser, orderdate FROM orders ORDER BY orderdate DESC LIMIT $1 OFFSET $2",
		params.Limit, params.GetOffset())
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	var orders []models.Order
	for rows.Next() {
		var o models.Order
		if err := rows.Scan(&o.ID, &o.UserID, &o.OrderDate); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		orders = append(orders, o)
	}

	totalPages := CalculateTotalPages(total, params.Limit)
	response := PaginatedResponse{
		Data:       orders,
		Page:       params.Page,
		Limit:      params.Limit,
		Total:      total,
		TotalPages: totalPages,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// ------------------
// GET orders by user ID
// ------------------
// @Summary Получить заказы пользователя (с пагинацией)
// @Tags Orders
// @Produce json
// @Security BearerAuth
// @Param user_id path int true "User ID"
// @Param page query int false "Номер страницы (по умолчанию 1)"
// @Param limit query int false "Количество элементов на странице (по умолчанию 20, максимум 100)"
// @Success 200 {object} PaginatedResponse
// @Failure 400 {string} string "Invalid User ID"
// @Failure 500 {string} string "Internal Server Error"
// @Router /orders/user/{user_id} [get]
func GetOrdersByUserHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	userID, err := strconv.Atoi(vars["user_id"])
	if err != nil {
		http.Error(w, "Invalid User ID", http.StatusBadRequest)
		return
	}

	params := ParsePaginationParams(r)

	// Получаем общее количество заказов пользователя
	var total int
	err = db.Pool.QueryRow(context.Background(),
		"SELECT COUNT(*) FROM orders WHERE iduser=$1", userID).Scan(&total)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Получаем заказы пользователя с пагинацией
	rows, err := db.Pool.Query(context.Background(),
		"SELECT idorder, iduser, orderdate FROM orders WHERE iduser=$1 ORDER BY orderdate DESC LIMIT $2 OFFSET $3",
		userID, params.Limit, params.GetOffset())
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	var orders []models.Order
	for rows.Next() {
		var o models.Order
		if err := rows.Scan(&o.ID, &o.UserID, &o.OrderDate); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		orders = append(orders, o)
	}

	totalPages := CalculateTotalPages(total, params.Limit)
	response := PaginatedResponse{
		Data:       orders,
		Page:       params.Page,
		Limit:      params.Limit,
		Total:      total,
		TotalPages: totalPages,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// ------------------
// GET order by ID
// ------------------
// @Summary Получить заказ по ID
// @Tags Orders
// @Produce json
// @Param id path int true "Order ID"
// @Success 200 {object} models.Order
// @Failure 400 {string} string "Invalid ID"
// @Failure 404 {string} string "Order not found"
// @Router /orders/{id} [get]
func GetOrderByIDHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id, err := strconv.Atoi(vars["id"])
	if err != nil {
		http.Error(w, "Invalid ID", http.StatusBadRequest)
		return
	}

	var order models.Order
	err = db.Pool.QueryRow(context.Background(),
		"SELECT idorder, iduser, orderdate FROM orders WHERE idorder=$1", id).
		Scan(&order.ID, &order.UserID, &order.OrderDate)
	if err != nil {
		http.Error(w, "Order not found", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(order)
}

// ------------------
// POST create order
// ------------------
// @Summary Создать новый заказ
// @Tags Orders
// @Accept json
// @Produce json
// @Param order body models.Order true "Order data"
// @Success 201 {object} models.Order
// @Failure 400 {string} string "Invalid request body"
// @Failure 500 {string} string "Internal Server Error"
// @Router /orders [post]
func CreateOrderHandler(w http.ResponseWriter, r *http.Request) {
	var order models.Order
	if err := json.NewDecoder(r.Body).Decode(&order); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if order.OrderDate.IsZero() {
		order.OrderDate = time.Now()
	}

	err := db.Pool.QueryRow(context.Background(),
		"INSERT INTO orders (iduser, orderdate) VALUES ($1, $2) RETURNING idorder",
		order.UserID, order.OrderDate).Scan(&order.ID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Логируем создание заказа
	LogUserAction(r, "CREATE", "order", order.ID, fmt.Sprintf("Создан заказ для пользователя ID: %d", order.UserID))
	
	// Email будет отправлен после добавления товаров в заказ (в CreateOrderProductHandler)

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(order)
}
