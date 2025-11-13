package handlers

import (
	"context"
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/gorilla/mux"
	"shoes-store-backend/db"
	"shoes-store-backend/models"
)

// ------------------
// GET ProductSizes by Product ID
// ------------------
// @Summary Получить размеры для продукта
// @Tags ProductSizes
// @Produce json
// @Param product_id path int true "Product ID"
// @Success 200 {array} models.ProductSize
// @Failure 400 {string} string "Invalid Product ID"
// @Failure 404 {string} string "Sizes not found"
// @Router /products/{product_id}/sizes [get]
func GetSizesByProductHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	productID, err := strconv.Atoi(vars["product_id"])
	if err != nil {
		http.Error(w, "Invalid Product ID", http.StatusBadRequest)
		return
	}

	rows, err := db.Pool.Query(context.Background(),
		"SELECT idproductsize, idproduct, size, quantity FROM productsizes WHERE idproduct=$1 ORDER BY size",
		productID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	var sizes []models.ProductSize
	for rows.Next() {
		var ps models.ProductSize
		if err := rows.Scan(&ps.ID, &ps.IdProduct, &ps.Size, &ps.Quantity); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		sizes = append(sizes, ps)
	}

	if len(sizes) == 0 {
		http.Error(w, "Sizes not found", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(sizes)
}

// ------------------
// UPDATE ProductSize (Quantity)
// ------------------
// @Summary Обновить количество размера
// @Tags ProductSizes
// @Accept json
// @Produce json
// @Param id path int true "ProductSize ID"
// @Param product_size body models.ProductSize true "Product Size data"
// @Success 200 {object} models.ProductSize
// @Failure 400 {string} string "Invalid request"
// @Failure 500 {string} string "Internal Server Error"
// @Router /productsizes/{id} [put]
func UpdateProductSizeHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id, err := strconv.Atoi(vars["id"])
	if err != nil {
		http.Error(w, "Invalid ID", http.StatusBadRequest)
		return
	}

	var ps models.ProductSize
	if err := json.NewDecoder(r.Body).Decode(&ps); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	_, err = db.Pool.Exec(context.Background(),
		"UPDATE productsizes SET quantity=$1 WHERE idproductsize=$2",
		ps.Quantity, id)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	ps.ID = id
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(ps)
}
