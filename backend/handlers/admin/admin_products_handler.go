package admin

import (
    "context"
    "encoding/json"
    "fmt"
    "net/http"
    "strconv"

    "github.com/gorilla/mux"
    "shoes-store-backend/db"
    "shoes-store-backend/handlers"
    "shoes-store-backend/models"
)

// ---------------------------
// Helper function to log admin actions
// ---------------------------
func LogUserAction(r *http.Request, action, entity string, entityID int, details string) {
    userID := r.Context().Value("userID")
    if userID == nil {
        return // Если нет userID, не логируем
    }
    
    _, err := db.Pool.Exec(context.Background(),
        `INSERT INTO logs (iduser, action, entity, entityid, details, createdat)
         VALUES ($1, $2, $3, $4, $5, NOW())`,
        userID, action, entity, entityID, details)
    if err != nil {
        fmt.Printf("Error logging user action: %v\n", err)
    }
}


// ---------------------------
// Create Product (Admin)
// ---------------------------

// @Summary Создать товар (Admin)
// @Tags Admin Products
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param product body models.Product true "Данные товара"
// @Success 201 {object} models.Product
// @Failure 400 {string} string "Invalid request body"
// @Failure 500 {string} string "Internal Server Error"
// @Router /admin/products [post]
func AdminCreateProductHandler(w http.ResponseWriter, r *http.Request) {
    var product models.Product
    if err := json.NewDecoder(r.Body).Decode(&product); err != nil {
        http.Error(w, "Invalid request body", http.StatusBadRequest)
        return
    }

    // Валидация обязательных полей
    if product.Name == "" {
        http.Error(w, "Name is required", http.StatusBadRequest)
        return
    }
    if product.ImageUrl == "" {
        http.Error(w, "Image URL is required", http.StatusBadRequest)
        return
    }
    if product.Price <= 0 {
        http.Error(w, "Price must be greater than 0", http.StatusBadRequest)
        return
    }
    if product.BrandID <= 0 {
        http.Error(w, "Brand ID is required", http.StatusBadRequest)
        return
    }
    if product.CategoryID <= 0 {
        http.Error(w, "Category ID is required", http.StatusBadRequest)
        return
    }

    err := db.Pool.QueryRow(context.Background(),
        `INSERT INTO products (name, imageurl, price, idbrand, idcategory)
         VALUES ($1, $2, $3, $4, $5)
         RETURNING idproduct`,
        product.Name, product.ImageUrl, product.Price, product.BrandID, product.CategoryID,
    ).Scan(&product.ID)
    if err != nil {
        http.Error(w, err.Error(), http.StatusInternalServerError)
        return
    }

    // Логируем действие админа
    LogUserAction(r, "CREATE", "product", product.ID, fmt.Sprintf("Создан товар: %s", product.Name))

    w.Header().Set("Content-Type", "application/json")
    w.WriteHeader(http.StatusCreated)
    json.NewEncoder(w).Encode(product)
}

// ---------------------------
// Get All Products (Admin)
// ---------------------------

// @Summary Получить все товары (Admin, с пагинацией)
// @Tags Admin Products
// @Produce json
// @Security BearerAuth
// @Param page query int false "Номер страницы (по умолчанию 1)"
// @Param limit query int false "Количество элементов на странице (по умолчанию 20, максимум 100)"
// @Success 200 {object} handlers.PaginatedResponse
// @Failure 500 {string} string "Internal Server Error"
// @Router /admin/products [get]
func AdminGetProductsHandler(w http.ResponseWriter, r *http.Request) {
    params := handlers.ParsePaginationParams(r)

    // Получаем общее количество продуктов
    var total int
    err := db.Pool.QueryRow(context.Background(),
        `SELECT COUNT(*) FROM products`).Scan(&total)
    if err != nil {
        http.Error(w, err.Error(), http.StatusInternalServerError)
        return
    }

    // Получаем продукты с пагинацией
    rows, err := db.Pool.Query(context.Background(),
        `SELECT idproduct, name, imageurl, price, idbrand, idcategory 
         FROM products ORDER BY idproduct LIMIT $1 OFFSET $2`,
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

    totalPages := handlers.CalculateTotalPages(total, params.Limit)
    response := handlers.PaginatedResponse{
        Data:       products,
        Page:       params.Page,
        Limit:      params.Limit,
        Total:      total,
        TotalPages: totalPages,
    }

    w.Header().Set("Content-Type", "application/json")
    json.NewEncoder(w).Encode(response)
}

// ---------------------------
// Get Product by ID (Admin)
// ---------------------------

// @Summary Получить товар по ID (Admin)
// @Tags Admin Products
// @Produce json
// @Security BearerAuth
// @Param id path int true "Product ID"
// @Success 200 {object} models.Product
// @Failure 400 {string} string "Invalid ID"
// @Failure 404 {string} string "Product not found"
// @Router /admin/products/{id} [get]
func AdminGetProductByIDHandler(w http.ResponseWriter, r *http.Request) {
    vars := mux.Vars(r)
    id, err := strconv.Atoi(vars["id"])
    if err != nil {
        http.Error(w, "Invalid ID", http.StatusBadRequest)
        return
    }

    var product models.Product
    err = db.Pool.QueryRow(context.Background(),
        `SELECT idproduct, name, imageurl, price, idbrand, idcategory
         FROM products WHERE idproduct = $1`, id,
    ).Scan(&product.ID, &product.Name, &product.ImageUrl, &product.Price, &product.BrandID, &product.CategoryID)
    if err != nil {
        http.Error(w, "Product not found", http.StatusNotFound)
        return
    }

    w.Header().Set("Content-Type", "application/json")
    json.NewEncoder(w).Encode(product)
}

// ---------------------------
// Update Product (Admin)
// ---------------------------

// @Summary Обновить товар (Admin)
// @Tags Admin Products
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path int true "Product ID"
// @Param product body models.Product true "Обновлённые данные товара"
// @Success 200 {object} models.Product
// @Failure 400 {string} string "Invalid request"
// @Failure 404 {string} string "Product not found"
// @Failure 500 {string} string "Internal Server Error"
// @Router /admin/products/{id} [put]
func AdminUpdateProductHandler(w http.ResponseWriter, r *http.Request) {
    vars := mux.Vars(r)
    id, err := strconv.Atoi(vars["id"])
    if err != nil {
        http.Error(w, "Invalid ID", http.StatusBadRequest)
        return
    }

    var product models.Product
    if err := json.NewDecoder(r.Body).Decode(&product); err != nil {
        http.Error(w, "Invalid request body", http.StatusBadRequest)
        return
    }

    // Валидация обязательных полей
    if product.Name == "" {
        http.Error(w, "Name is required", http.StatusBadRequest)
        return
    }
    if product.ImageUrl == "" {
        http.Error(w, "Image URL is required", http.StatusBadRequest)
        return
    }
    if product.Price <= 0 {
        http.Error(w, "Price must be greater than 0", http.StatusBadRequest)
        return
    }
    if product.BrandID <= 0 {
        http.Error(w, "Brand ID is required", http.StatusBadRequest)
        return
    }
    if product.CategoryID <= 0 {
        http.Error(w, "Category ID is required", http.StatusBadRequest)
        return
    }

    _, err = db.Pool.Exec(context.Background(),
        `UPDATE products 
         SET name=$1, imageurl=$2, price=$3, idbrand=$4, idcategory=$5
         WHERE idproduct=$6`,
        product.Name, product.ImageUrl, product.Price, product.BrandID, product.CategoryID, id,
    )
    if err != nil {
        http.Error(w, err.Error(), http.StatusInternalServerError)
        return
    }

    // Логируем действие админа
    LogUserAction(r, "UPDATE", "product", id, fmt.Sprintf("Обновлен товар: %s", product.Name))

    product.ID = id
    w.Header().Set("Content-Type", "application/json")
    json.NewEncoder(w).Encode(product)
}

// ---------------------------
// Delete Product (Admin)
// ---------------------------

// @Summary Удалить товар (Admin)
// @Tags Admin Products
// @Produce json
// @Security BearerAuth
// @Param id path int true "Product ID"
// @Success 204 "No Content"
// @Failure 400 {string} string "Invalid ID"
// @Failure 404 {string} string "Product not found"
// @Router /admin/products/{id} [delete]
func AdminDeleteProductHandler(w http.ResponseWriter, r *http.Request) {
    vars := mux.Vars(r)
    id, err := strconv.Atoi(vars["id"])
    if err != nil {
        http.Error(w, "Invalid ID", http.StatusBadRequest)
        return
    }

    // Проверяем, существует ли товар
    var productName string
    err = db.Pool.QueryRow(context.Background(),
        `SELECT name FROM products WHERE idproduct=$1`, id,
    ).Scan(&productName)
    if err != nil {
        http.Error(w, "Product not found", http.StatusNotFound)
        return
    }

    // Начинаем транзакцию для безопасного удаления всех связанных данных
    tx, err := db.Pool.Begin(context.Background())
    if err != nil {
        http.Error(w, err.Error(), http.StatusInternalServerError)
        return
    }
    defer tx.Rollback(context.Background())

    // 1. Удаляем записи из basket, которые ссылаются на productsizes этого товара
    _, err = tx.Exec(context.Background(),
        `DELETE FROM basket 
         WHERE idproductsize IN (
             SELECT idproductsize FROM productsizes WHERE idproduct=$1
         )`, id)
    if err != nil {
        http.Error(w, fmt.Sprintf("Error deleting basket items: %v", err), http.StatusInternalServerError)
        return
    }

    // 2. Удаляем записи из favorites, которые ссылаются на productsizes этого товара
    _, err = tx.Exec(context.Background(),
        `DELETE FROM favorites 
         WHERE idproductsize IN (
             SELECT idproductsize FROM productsizes WHERE idproduct=$1
         )`, id)
    if err != nil {
        http.Error(w, fmt.Sprintf("Error deleting favorites: %v", err), http.StatusInternalServerError)
        return
    }

    // 3. Удаляем записи из orderproducts, которые ссылаются на productsizes этого товара
    _, err = tx.Exec(context.Background(),
        `DELETE FROM orderproducts 
         WHERE idproductsize IN (
             SELECT idproductsize FROM productsizes WHERE idproduct=$1
         )`, id)
    if err != nil {
        http.Error(w, fmt.Sprintf("Error deleting order products: %v", err), http.StatusInternalServerError)
        return
    }

    // 4. Удаляем productsizes (размеры товара)
    _, err = tx.Exec(context.Background(),
        `DELETE FROM productsizes WHERE idproduct=$1`, id)
    if err != nil {
        http.Error(w, fmt.Sprintf("Error deleting product sizes: %v", err), http.StatusInternalServerError)
        return
    }

    // 5. Удаляем reviews (отзывы) этого товара
    _, err = tx.Exec(context.Background(),
        `DELETE FROM reviews WHERE idproduct=$1`, id)
    if err != nil {
        http.Error(w, fmt.Sprintf("Error deleting reviews: %v", err), http.StatusInternalServerError)
        return
    }

    // 6. Удаляем сам товар
    _, err = tx.Exec(context.Background(),
        `DELETE FROM products WHERE idproduct=$1`, id)
    if err != nil {
        http.Error(w, fmt.Sprintf("Error deleting product: %v", err), http.StatusInternalServerError)
        return
    }

    // Коммитим транзакцию
    if err = tx.Commit(context.Background()); err != nil {
        http.Error(w, fmt.Sprintf("Error committing transaction: %v", err), http.StatusInternalServerError)
        return
    }

    // Логируем действие админа
    LogUserAction(r, "DELETE", "product", id, fmt.Sprintf("Удален товар: %s (ID: %d)", productName, id))

    w.WriteHeader(http.StatusNoContent)
}
