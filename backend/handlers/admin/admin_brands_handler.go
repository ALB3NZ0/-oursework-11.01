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
// Create Brand (Admin)
// ---------------------------

// @Summary Создать бренд (Admin)
// @Tags Admin Brands
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param brand body models.Brand true "Данные бренда"
// @Success 201 {object} models.Brand
// @Failure 400 {string} string "Invalid request body"
// @Failure 500 {string} string "Internal Server Error"
// @Router /admin/brands [post]
func AdminCreateBrandHandler(w http.ResponseWriter, r *http.Request) {
    var brand models.Brand
    if err := json.NewDecoder(r.Body).Decode(&brand); err != nil {
        http.Error(w, "Invalid request body", http.StatusBadRequest)
        return
    }

    err := db.Pool.QueryRow(context.Background(),
        `INSERT INTO brands (brandname) VALUES ($1) RETURNING idbrand`,
        brand.BrandName).Scan(&brand.ID)
    if err != nil {
        http.Error(w, err.Error(), http.StatusInternalServerError)
        return
    }

    // Логируем действие админа
    LogUserAction(r, "CREATE", "brand", brand.ID, fmt.Sprintf("Создан бренд: %s", brand.BrandName))

    w.Header().Set("Content-Type", "application/json")
    w.WriteHeader(http.StatusCreated)
    json.NewEncoder(w).Encode(brand)
}

// ---------------------------
// Get All Brands (Admin)
// ---------------------------

// @Summary Получить все бренды (Admin)
// @Tags Admin Brands
// @Produce json
// @Security BearerAuth
// @Success 200 {array} models.Brand
// @Failure 500 {string} string "Internal Server Error"
// @Router /admin/brands [get]
func AdminGetBrandsHandler(w http.ResponseWriter, r *http.Request) {
    rows, err := db.Pool.Query(context.Background(),
        `SELECT idbrand, brandname FROM brands ORDER BY idbrand`)
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

// ---------------------------
// Get Brand by ID (Admin)
// ---------------------------

// @Summary Получить бренд по ID (Admin)
// @Tags Admin Brands
// @Produce json
// @Security BearerAuth
// @Param id path int true "Brand ID"
// @Success 200 {object} models.Brand
// @Failure 400 {string} string "Invalid ID"
// @Failure 404 {string} string "Brand not found"
// @Router /admin/brands/{id} [get]
func AdminGetBrandByIDHandler(w http.ResponseWriter, r *http.Request) {
    vars := mux.Vars(r)
    id, err := strconv.Atoi(vars["id"])
    if err != nil {
        http.Error(w, "Invalid ID", http.StatusBadRequest)
        return
    }

    var brand models.Brand
    err = db.Pool.QueryRow(context.Background(),
        `SELECT idbrand, brandname FROM brands WHERE idbrand=$1`, id).
        Scan(&brand.ID, &brand.BrandName)
    if err != nil {
        http.Error(w, "Brand not found", http.StatusNotFound)
        return
    }

    w.Header().Set("Content-Type", "application/json")
    json.NewEncoder(w).Encode(brand)
}

// ---------------------------
// Update Brand (Admin)
// ---------------------------

// @Summary Обновить бренд (Admin)
// @Tags Admin Brands
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path int true "Brand ID"
// @Param brand body models.Brand true "Обновлённые данные бренда"
// @Success 200 {object} models.Brand
// @Failure 400 {string} string "Invalid request"
// @Failure 404 {string} string "Brand not found"
// @Failure 500 {string} string "Internal Server Error"
// @Router /admin/brands/{id} [put]
func AdminUpdateBrandHandler(w http.ResponseWriter, r *http.Request) {
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
        `UPDATE brands SET brandname=$1 WHERE idbrand=$2`,
        brand.BrandName, id)
    if err != nil {
        http.Error(w, err.Error(), http.StatusInternalServerError)
        return
    }

    // Логируем действие админа
    LogUserAction(r, "UPDATE", "brand", id, fmt.Sprintf("Обновлен бренд: %s", brand.BrandName))

    brand.ID = id
    w.Header().Set("Content-Type", "application/json")
    json.NewEncoder(w).Encode(brand)
}

// ---------------------------
// Delete Brand (Admin)
// ---------------------------

// @Summary Удалить бренд (Admin)
// @Tags Admin Brands
// @Produce json
// @Security BearerAuth
// @Param id path int true "Brand ID"
// @Success 204 "No Content"
// @Failure 400 {string} string "Invalid ID"
// @Failure 404 {string} string "Brand not found"
// @Router /admin/brands/{id} [delete]
func AdminDeleteBrandHandler(w http.ResponseWriter, r *http.Request) {
    vars := mux.Vars(r)
    id, err := strconv.Atoi(vars["id"])
    if err != nil {
        http.Error(w, "Invalid ID", http.StatusBadRequest)
        return
    }

    res, err := db.Pool.Exec(context.Background(),
        `DELETE FROM brands WHERE idbrand=$1`, id)
    if err != nil {
        http.Error(w, err.Error(), http.StatusInternalServerError)
        return
    }

    rowsAffected := res.RowsAffected()
    if rowsAffected == 0 {
        http.Error(w, "Brand not found", http.StatusNotFound)
        return
    }

    // Логируем действие админа
    LogUserAction(r, "DELETE", "brand", id, fmt.Sprintf("Удален бренд с ID: %d", id))

    w.WriteHeader(http.StatusNoContent)
}
