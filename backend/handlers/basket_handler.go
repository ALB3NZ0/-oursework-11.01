package handlers

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"

	"github.com/gorilla/mux"
	"shoes-store-backend/db"
	"shoes-store-backend/models"
)

// ------------------
// GET Basket by User ID
// ------------------
// @Summary Получить корзину пользователя
// @Tags Basket
// @Produce json
// @Security BearerAuth
// @Param user_id path int true "User ID"
// @Success 200 {array} models.Basket
// @Failure 400 {string} string "Invalid User ID"
// @Router /basket/{user_id} [get]
func GetBasketHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	userID, err := strconv.Atoi(vars["user_id"])
	if err != nil {
		http.Error(w, "Invalid User ID", http.StatusBadRequest)
		return
	}

	rows, err := db.Pool.Query(context.Background(),
		`SELECT b.idbasket, b.iduser, b.idproductsize, b.quantity,
		        p.idproduct, p.name, ps.size, ps.quantity as available, p.price, p.imageurl
		   FROM basket b
		   JOIN productsizes ps ON b.idproductsize = ps.idproductsize
		   JOIN products p ON ps.idproduct = p.idproduct
		  WHERE b.iduser=$1`, userID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	var basket []models.BasketItem
	for rows.Next() {
		var item models.BasketItem
		if err := rows.Scan(&item.ID, &item.UserID, &item.ProductSizeID,
			&item.Quantity, &item.ProductID, &item.ProductName, &item.Size, &item.Available, &item.Price, &item.ImageURL); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		basket = append(basket, item)
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(basket)
}

// ------------------
// ADD to Basket
// ------------------
// @Summary Добавить товар в корзину
// @Tags Basket
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param basket body models.Basket true "Basket item"
// @Success 201 {object} models.Basket
// @Failure 400 {string} string "Invalid request"
// @Router /basket [post]
func AddToBasketHandler(w http.ResponseWriter, r *http.Request) {
	var basket models.Basket
	if err := json.NewDecoder(r.Body).Decode(&basket); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	err := db.Pool.QueryRow(context.Background(),
		"INSERT INTO basket (iduser, idproductsize, quantity) VALUES ($1, $2, $3) RETURNING idbasket",
		basket.UserID, basket.ProductSizeID, basket.Quantity).
		Scan(&basket.ID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Логируем добавление в корзину
	LogUserAction(r, "ADD_TO_BASKET", "basket", basket.ID, fmt.Sprintf("Добавлен товар в корзину пользователя ID: %d", basket.UserID))

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(basket)
}

// ------------------
// UPDATE Basket Item
// ------------------
// @Summary Обновить количество в корзине
// @Tags Basket
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path int true "Basket ID"
// @Param basket body models.Basket true "Basket item"
// @Success 200 {object} models.Basket
// @Failure 400 {string} string "Invalid request"
// @Router /basket/{id} [put]
func UpdateBasketHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id, err := strconv.Atoi(vars["id"])
	if err != nil {
		http.Error(w, "Invalid ID", http.StatusBadRequest)
		return
	}

	var basket models.Basket
	if err := json.NewDecoder(r.Body).Decode(&basket); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	_, err = db.Pool.Exec(context.Background(),
		"UPDATE basket SET quantity=$1 WHERE idbasket=$2", basket.Quantity, id)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Логируем обновление корзины
	LogUserAction(r, "UPDATE_BASKET", "basket", id, fmt.Sprintf("Обновлено количество в корзине ID: %d", id))

	basket.ID = id
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(basket)
}

// ------------------
// DELETE from Basket
// ------------------
// @Summary Удалить товар из корзины
// @Tags Basket
// @Security BearerAuth
// @Param id path int true "Basket ID"
// @Success 204 {string} string "No Content"
// @Failure 400 {string} string "Invalid ID"
// @Router /basket/{id} [delete]
func DeleteBasketHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id, err := strconv.Atoi(vars["id"])
	if err != nil {
		http.Error(w, "Invalid ID", http.StatusBadRequest)
		return
	}

	_, err = db.Pool.Exec(context.Background(),
		"DELETE FROM basket WHERE idbasket=$1", id)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Логируем удаление из корзины
	LogUserAction(r, "REMOVE_FROM_BASKET", "basket", id, fmt.Sprintf("Удален товар из корзины ID: %d", id))

	w.WriteHeader(http.StatusNoContent)
}
