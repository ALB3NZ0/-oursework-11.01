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
// GET Favorites by User ID
// ------------------
// @Summary Получить избранные товары пользователя
// @Tags Favorites
// @Produce json
// @Security BearerAuth
// @Param user_id path int true "User ID"
// @Success 200 {array} models.FavoriteItem
// @Failure 400 {string} string "Invalid User ID"
// @Router /favorites/{user_id} [get]
func GetFavoritesHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	userID, err := strconv.Atoi(vars["user_id"])
	if err != nil {
		http.Error(w, "Invalid User ID", http.StatusBadRequest)
		return
	}

	rows, err := db.Pool.Query(context.Background(),
		`SELECT f.idfavorites, f.iduser, f.idproductsize,
		        p.idproduct, p.name, ps.size, p.price, p.imageurl
		   FROM favorites f
		   JOIN productsizes ps ON f.idproductsize = ps.idproductsize
		   JOIN products p ON ps.idproduct = p.idproduct
		  WHERE f.iduser=$1`, userID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	var favorites []models.FavoriteItem
	for rows.Next() {
		var item models.FavoriteItem
		if err := rows.Scan(&item.ID, &item.UserID, &item.ProductSizeID,
			&item.ProductID, &item.ProductName, &item.Size, &item.Price, &item.ImageURL); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		favorites = append(favorites, item)
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(favorites)
}

// ------------------
// ADD to Favorites
// ------------------
// @Summary Добавить товар в избранное
// @Tags Favorites
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param favorite body models.Favorite true "Favorite item"
// @Success 201 {object} models.Favorite
// @Failure 400 {string} string "Invalid request"
// @Router /favorites [post]
func AddToFavoritesHandler(w http.ResponseWriter, r *http.Request) {
	var fav models.Favorite
	if err := json.NewDecoder(r.Body).Decode(&fav); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	err := db.Pool.QueryRow(context.Background(),
		"INSERT INTO favorites (iduser, idproductsize) VALUES ($1, $2) RETURNING idfavorites",
		fav.UserID, fav.ProductSizeID).
		Scan(&fav.ID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Логируем добавление в избранное
	LogUserAction(r, "ADD_TO_FAVORITES", "favorite", fav.ID, fmt.Sprintf("Добавлен товар в избранное пользователя ID: %d", fav.UserID))

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(fav)
}

// ------------------
// DELETE from Favorites
// ------------------
// @Summary Удалить товар из избранного
// @Tags Favorites
// @Security BearerAuth
// @Param id path int true "Favorite ID"
// @Success 204 {string} string "No Content"
// @Failure 400 {string} string "Invalid ID"
// @Router /favorites/{id} [delete]
func DeleteFavoriteHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id, err := strconv.Atoi(vars["id"])
	if err != nil {
		http.Error(w, "Invalid ID", http.StatusBadRequest)
		return
	}

	_, err = db.Pool.Exec(context.Background(),
		"DELETE FROM favorites WHERE idfavorites=$1", id)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Логируем удаление из избранного
	LogUserAction(r, "REMOVE_FROM_FAVORITES", "favorite", id, fmt.Sprintf("Удален товар из избранного ID: %d", id))

	w.WriteHeader(http.StatusNoContent)
}
