package admin

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


// ---------------------------
// Create Category (Admin)
// ---------------------------

// @Summary Создать категорию (Admin)
// @Tags Admin Categories
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param category body models.Category true "Данные категории"
// @Success 201 {object} models.Category
// @Failure 400 {string} string "Invalid request body"
// @Failure 500 {string} string "Internal Server Error"
// @Router /admin/categories [post]
func AdminCreateCategoryHandler(w http.ResponseWriter, r *http.Request) {
    var category models.Category
    if err := json.NewDecoder(r.Body).Decode(&category); err != nil {
        http.Error(w, "Invalid request body", http.StatusBadRequest)
        return
    }

    err := db.Pool.QueryRow(context.Background(),
        `INSERT INTO categories (categoryname) VALUES ($1) RETURNING idcategory`,
        category.CategoryName).Scan(&category.ID)
    if err != nil {
        http.Error(w, err.Error(), http.StatusInternalServerError)
        return
    }

    // Логируем действие админа
    LogUserAction(r, "CREATE", "category", category.ID, fmt.Sprintf("Создана категория: %s", category.CategoryName))

    w.Header().Set("Content-Type", "application/json")
    w.WriteHeader(http.StatusCreated)
    json.NewEncoder(w).Encode(category)
}

// ---------------------------
// Get All Categories (Admin)
// ---------------------------

// @Summary Получить все категории (Admin)
// @Tags Admin Categories
// @Produce json
// @Security BearerAuth
// @Success 200 {array} models.Category
// @Failure 500 {string} string "Internal Server Error"
// @Router /admin/categories [get]
func AdminGetCategoriesHandler(w http.ResponseWriter, r *http.Request) {
    rows, err := db.Pool.Query(context.Background(),
        `SELECT idcategory, categoryname FROM categories ORDER BY idcategory`)
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

// ---------------------------
// Get Category by ID (Admin)
// ---------------------------

// @Summary Получить категорию по ID (Admin)
// @Tags Admin Categories
// @Produce json
// @Security BearerAuth
// @Param id path int true "Category ID"
// @Success 200 {object} models.Category
// @Failure 400 {string} string "Invalid ID"
// @Failure 404 {string} string "Category not found"
// @Router /admin/categories/{id} [get]
func AdminGetCategoryByIDHandler(w http.ResponseWriter, r *http.Request) {
    vars := mux.Vars(r)
    id, err := strconv.Atoi(vars["id"])
    if err != nil {
        http.Error(w, "Invalid ID", http.StatusBadRequest)
        return
    }

    var category models.Category
    err = db.Pool.QueryRow(context.Background(),
        `SELECT idcategory, categoryname FROM categories WHERE idcategory=$1`, id).
        Scan(&category.ID, &category.CategoryName)
    if err != nil {
        http.Error(w, "Category not found", http.StatusNotFound)
        return
    }

    w.Header().Set("Content-Type", "application/json")
    json.NewEncoder(w).Encode(category)
}

// ---------------------------
// Update Category (Admin)
// ---------------------------

// @Summary Обновить категорию (Admin)
// @Tags Admin Categories
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path int true "Category ID"
// @Param category body models.Category true "Обновлённые данные категории"
// @Success 200 {object} models.Category
// @Failure 400 {string} string "Invalid request"
// @Failure 404 {string} string "Category not found"
// @Failure 500 {string} string "Internal Server Error"
// @Router /admin/categories/{id} [put]
func AdminUpdateCategoryHandler(w http.ResponseWriter, r *http.Request) {
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
        `UPDATE categories SET categoryname=$1 WHERE idcategory=$2`,
        category.CategoryName, id)
    if err != nil {
        http.Error(w, err.Error(), http.StatusInternalServerError)
        return
    }

    // Логируем действие админа
    LogUserAction(r, "UPDATE", "category", id, fmt.Sprintf("Обновлена категория: %s", category.CategoryName))

    category.ID = id
    w.Header().Set("Content-Type", "application/json")
    json.NewEncoder(w).Encode(category)
}

// ---------------------------
// Delete Category (Admin)
// ---------------------------

// @Summary Удалить категорию (Admin)
// @Tags Admin Categories
// @Produce json
// @Security BearerAuth
// @Param id path int true "Category ID"
// @Success 204 "No Content"
// @Failure 400 {string} string "Invalid ID"
// @Failure 404 {string} string "Category not found"
// @Router /admin/categories/{id} [delete]
func AdminDeleteCategoryHandler(w http.ResponseWriter, r *http.Request) {
    vars := mux.Vars(r)
    id, err := strconv.Atoi(vars["id"])
    if err != nil {
        http.Error(w, "Invalid ID", http.StatusBadRequest)
        return
    }

    res, err := db.Pool.Exec(context.Background(),
        `DELETE FROM categories WHERE idcategory=$1`, id)
    if err != nil {
        http.Error(w, err.Error(), http.StatusInternalServerError)
        return
    }

    rowsAffected := res.RowsAffected()
    if rowsAffected == 0 {
        http.Error(w, "Category not found", http.StatusNotFound)
        return
    }

    // Логируем действие админа
    LogUserAction(r, "DELETE", "category", id, fmt.Sprintf("Удалена категория с ID: %d", id))

    w.WriteHeader(http.StatusNoContent)
}
