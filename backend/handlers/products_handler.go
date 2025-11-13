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
// CREATE Product
// ------------------
// @Summary Создать продукт
// @Tags Products
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param product body models.Product true "Данные продукта"
// @Success 201 {object} models.Product
// @Failure 400 {string} string "Invalid request body"
// @Failure 500 {string} string "Internal Server Error"
// @Router /products [post]
func CreateProductHandler(w http.ResponseWriter, r *http.Request) {
	var product models.Product
	if err := json.NewDecoder(r.Body).Decode(&product); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	err := db.Pool.QueryRow(context.Background(),
		"INSERT INTO products (name, imageurl, price, idbrand, idcategory) VALUES ($1, $2, $3, $4, $5) RETURNING idproduct",
		product.Name, product.ImageUrl, product.Price, product.BrandID, product.CategoryID,
	).Scan(&product.ID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(product)
}

// ------------------
// GET All Products
// ------------------
// @Summary Получить все продукты (с пагинацией)
// @Tags Products
// @Produce json
// @Param page query int false "Номер страницы (по умолчанию 1)"
// @Param limit query int false "Количество элементов на странице (по умолчанию 20, максимум 100)"
// @Success 200 {object} PaginatedResponse
// @Failure 500 {string} string "Internal Server Error"
// @Router /products [get]
func GetProductsHandler(w http.ResponseWriter, r *http.Request) {
	params := ParsePaginationParams(r)

	// Получаем общее количество продуктов
	var total int
	err := db.Pool.QueryRow(context.Background(),
		"SELECT COUNT(*) FROM products").Scan(&total)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Получаем продукты с пагинацией
	rows, err := db.Pool.Query(context.Background(),
		"SELECT idproduct, name, imageurl, price, idbrand, idcategory FROM products ORDER BY idproduct LIMIT $1 OFFSET $2",
		params.Limit, params.GetOffset())
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	var products []models.Product
	for rows.Next() {
		var p models.Product
		if err := rows.Scan(&p.ID, &p.Name, &p.ImageUrl, &p.Price, &p.BrandID, &p.CategoryID); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		products = append(products, p)
	}

	totalPages := CalculateTotalPages(total, params.Limit)
	response := PaginatedResponse{
		Data:       products,
		Page:       params.Page,
		Limit:      params.Limit,
		Total:      total,
		TotalPages: totalPages,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// ------------------
// GET Product by ID
// ------------------
// @Summary Получить продукт по ID
// @Tags Products
// @Produce json
// @Param id path int true "Product ID"
// @Success 200 {object} models.Product
// @Failure 400 {string} string "Invalid ID"
// @Failure 404 {string} string "Product not found"
// @Router /products/{id} [get]
func GetProductByIDHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id, err := strconv.Atoi(vars["id"])
	if err != nil {
		http.Error(w, "Invalid ID", http.StatusBadRequest)
		return
	}

	var p models.Product
	err = db.Pool.QueryRow(context.Background(),
		"SELECT idproduct, name, imageurl, price, idbrand, idcategory FROM products WHERE idproduct=$1", id).
		Scan(&p.ID, &p.Name, &p.ImageUrl, &p.Price, &p.BrandID, &p.CategoryID)
	if err != nil {
		http.Error(w, "Product not found", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(p)
}

// ------------------
// UPDATE Product
// ------------------
// @Summary Обновить продукт
// @Tags Products
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path int true "Product ID"
// @Param product body models.Product true "Данные продукта"
// @Success 200 {object} models.Product
// @Failure 400 {string} string "Invalid request"
// @Failure 500 {string} string "Internal Server Error"
// @Router /products/{id} [put]
func UpdateProductHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id, err := strconv.Atoi(vars["id"])
	if err != nil {
		http.Error(w, "Invalid ID", http.StatusBadRequest)
		return
	}

	var p models.Product
	if err := json.NewDecoder(r.Body).Decode(&p); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	_, err = db.Pool.Exec(context.Background(),
		"UPDATE products SET name=$1, imageurl=$2, price=$3, idbrand=$4, idcategory=$5 WHERE idproduct=$6",
		p.Name, p.ImageUrl, p.Price, p.BrandID, p.CategoryID, id)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	p.ID = id
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(p)
}

// ------------------
// DELETE Product
// ------------------
// @Summary Удалить продукт
// @Tags Products
// @Security BearerAuth
// @Param id path int true "Product ID"
// @Success 204 {string} string "No Content"
// @Failure 400 {string} string "Invalid ID"
// @Failure 404 {string} string "Product not found"
// @Router /products/{id} [delete]
func DeleteProductHandler(w http.ResponseWriter, r *http.Request) {
    vars := mux.Vars(r)
    id, err := strconv.Atoi(vars["id"])
    if err != nil {
        http.Error(w, "Invalid ID", http.StatusBadRequest)
        return
    }

    // сначала удаляем размеры
    _, err = db.Pool.Exec(context.Background(),
        "DELETE FROM productsizes WHERE idproduct=$1", id)
    if err != nil {
        http.Error(w, err.Error(), http.StatusInternalServerError)
        return
    }

    // теперь сам продукт
    tag, err := db.Pool.Exec(context.Background(),
        "DELETE FROM products WHERE idproduct=$1", id)
    if err != nil {
        http.Error(w, err.Error(), http.StatusInternalServerError)
        return
    }
    if tag.RowsAffected() == 0 {
        http.Error(w, "Product not found", http.StatusNotFound)
        return
    }

    w.WriteHeader(http.StatusNoContent)
}
