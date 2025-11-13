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
// CREATE Category
// ------------------
// @Summary Создать категорию
// @Tags Categories
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param category body models.Category true "Category data"
// @Success 201 {object} models.Category
// @Failure 400 {string} string "Invalid request body"
// @Failure 500 {string} string "Internal Server Error"
// @Router /categories [post]
func CreateCategoryHandler(w http.ResponseWriter, r *http.Request) {
	var category models.Category
	if err := json.NewDecoder(r.Body).Decode(&category); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	err := db.Pool.QueryRow(context.Background(),
		"INSERT INTO categories (categoryname) VALUES ($1) RETURNING idcategory",
		category.CategoryName,
	).Scan(&category.ID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(category)
}

// ------------------
// GET All Categories
// ------------------
// @Summary Получить все категории
// @Tags Categories
// @Produce json
// @Success 200 {array} models.Category
// @Failure 500 {string} string "Internal Server Error"
// @Router /categories [get]
func GetCategoriesHandler(w http.ResponseWriter, r *http.Request) {
	rows, err := db.Pool.Query(context.Background(),
		"SELECT idcategory, categoryname FROM categories")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	var categories []models.Category
	for rows.Next() {
		var c models.Category
		if err := rows.Scan(&c.ID, &c.CategoryName); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		categories = append(categories, c)
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(categories)
}

// ------------------
// GET Category by ID
// ------------------
// @Summary Получить категорию по ID
// @Tags Categories
// @Produce json
// @Param id path int true "Category ID"
// @Success 200 {object} models.Category
// @Failure 400 {string} string "Invalid ID"
// @Failure 404 {string} string "Category not found"
// @Router /categories/{id} [get]
func GetCategoryByIDHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id, err := strconv.Atoi(vars["id"])
	if err != nil {
		http.Error(w, "Invalid ID", http.StatusBadRequest)
		return
	}

	var category models.Category
	err = db.Pool.QueryRow(context.Background(),
		"SELECT idcategory, categoryname FROM categories WHERE idcategory=$1", id).
		Scan(&category.ID, &category.CategoryName)
	if err != nil {
		http.Error(w, "Category not found", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(category)
}

// ------------------
// UPDATE Category
// ------------------
// @Summary Обновить категорию
// @Tags Categories
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path int true "Category ID"
// @Param category body models.Category true "Category data"
// @Success 200 {object} models.Category
// @Failure 400 {string} string "Invalid request"
// @Failure 500 {string} string "Internal Server Error"
// @Router /categories/{id} [put]
func UpdateCategoryHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id, err := strconv.Atoi(vars["id"])
	if err != nil {
		http.Error(w, "Invalid ID", http.StatusBadRequest)
		return
	}

	var category models.Category
	if err := json.NewDecoder(r.Body).Decode(&category); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	_, err = db.Pool.Exec(context.Background(),
		"UPDATE categories SET categoryname=$1 WHERE idcategory=$2",
		category.CategoryName, id)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	category.ID = id
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(category)
}

// ------------------
// DELETE Category
// ------------------
// @Summary Удалить категорию
// @Tags Categories
// @Security BearerAuth
// @Param id path int true "Category ID"
// @Success 204 {string} string "No Content"
// @Failure 400 {string} string "Invalid ID"
// @Failure 404 {string} string "Category not found"
// @Router /categories/{id} [delete]
func DeleteCategoryHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id, err := strconv.Atoi(vars["id"])
	if err != nil {
		http.Error(w, "Invalid ID", http.StatusBadRequest)
		return
	}

	tag, err := db.Pool.Exec(context.Background(),
		"DELETE FROM categories WHERE idcategory=$1", id)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	if tag.RowsAffected() == 0 {
		http.Error(w, "Category not found", http.StatusNotFound)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
