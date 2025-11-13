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
// CREATE Brand
// ------------------
// @Summary Создать бренд
// @Tags Brands
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param brand body models.Brand true "Brand data"
// @Success 201 {object} models.Brand
// @Failure 400 {string} string "Invalid request body"
// @Failure 500 {string} string "Internal Server Error"
// @Router /brands [post]
func CreateBrandHandler(w http.ResponseWriter, r *http.Request) {
	var brand models.Brand
	if err := json.NewDecoder(r.Body).Decode(&brand); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	err := db.Pool.QueryRow(context.Background(),
		"INSERT INTO brands (brandname) VALUES ($1) RETURNING idbrand",
		brand.BrandName,
	).Scan(&brand.ID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(brand)
}

// ------------------
// GET All Brands
// ------------------
// @Summary Получить все бренды
// @Tags Brands
// @Produce json
// @Success 200 {array} models.Brand
// @Failure 500 {string} string "Internal Server Error"
// @Router /brands [get]
func GetBrandsHandler(w http.ResponseWriter, r *http.Request) {
	rows, err := db.Pool.Query(context.Background(),
		"SELECT idbrand, brandname FROM brands")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	var brands []models.Brand
	for rows.Next() {
		var b models.Brand
		if err := rows.Scan(&b.ID, &b.BrandName); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		brands = append(brands, b)
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(brands)
}

// ------------------
// GET Brand by ID
// ------------------
// @Summary Получить бренд по ID
// @Tags Brands
// @Produce json
// @Param id path int true "Brand ID"
// @Success 200 {object} models.Brand
// @Failure 400 {string} string "Invalid ID"
// @Failure 404 {string} string "Brand not found"
// @Router /brands/{id} [get]
func GetBrandByIDHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id, err := strconv.Atoi(vars["id"])
	if err != nil {
		http.Error(w, "Invalid ID", http.StatusBadRequest)
		return
	}

	var brand models.Brand
	err = db.Pool.QueryRow(context.Background(),
		"SELECT idbrand, brandname FROM brands WHERE idbrand=$1", id).
		Scan(&brand.ID, &brand.BrandName)
	if err != nil {
		http.Error(w, "Brand not found", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(brand)
}

// ------------------
// UPDATE Brand
// ------------------
// @Summary Обновить бренд
// @Tags Brands
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path int true "Brand ID"
// @Param brand body models.Brand true "Brand data"
// @Success 200 {object} models.Brand
// @Failure 400 {string} string "Invalid request"
// @Failure 500 {string} string "Internal Server Error"
// @Router /brands/{id} [put]
func UpdateBrandHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id, err := strconv.Atoi(vars["id"])
	if err != nil {
		http.Error(w, "Invalid ID", http.StatusBadRequest)
		return
	}

	var brand models.Brand
	if err := json.NewDecoder(r.Body).Decode(&brand); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	_, err = db.Pool.Exec(context.Background(),
		"UPDATE brands SET brandname=$1 WHERE idbrand=$2",
		brand.BrandName, id)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	brand.ID = id
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(brand)
}

// ------------------
// DELETE Brand
// ------------------
// @Summary Удалить бренд
// @Tags Brands
// @Security BearerAuth
// @Param id path int true "Brand ID"
// @Success 204 {string} string "No Content"
// @Failure 400 {string} string "Invalid ID"
// @Failure 404 {string} string "Brand not found"
// @Router /brands/{id} [delete]
func DeleteBrandHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id, err := strconv.Atoi(vars["id"])
	if err != nil {
		http.Error(w, "Invalid ID", http.StatusBadRequest)
		return
	}

	tag, err := db.Pool.Exec(context.Background(),
		"DELETE FROM brands WHERE idbrand=$1", id)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	if tag.RowsAffected() == 0 {
		http.Error(w, "Brand not found", http.StatusNotFound)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
