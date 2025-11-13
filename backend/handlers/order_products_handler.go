package handlers

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/smtp"
	"strconv"
	"time"

	"github.com/gorilla/mux"
	"shoes-store-backend/db"
	"shoes-store-backend/models"
)

// ------------------
// Create OrderProduct
// ------------------
// @Summary Добавить товар в заказ
// @Tags OrderProducts
// @Accept json
// @Produce json
// @Param order_product body models.OrderProduct true "Данные для добавления"
// @Success 201 {object} models.OrderProduct
// @Failure 400 {string} string "Invalid request body"
// @Failure 500 {string} string "Internal Server Error"
// @Router /order-products [post]
func CreateOrderProductHandler(w http.ResponseWriter, r *http.Request) {
	var op models.OrderProduct
	if err := json.NewDecoder(r.Body).Decode(&op); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	err := db.Pool.QueryRow(context.Background(),
		"INSERT INTO orderproducts (idorder, idproductsize, quantity) VALUES ($1, $2, $3) RETURNING idorderproduct",
		op.OrderID, op.ProductSizeID, op.Quantity,
	).Scan(&op.Id)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Проверяем, был ли уже отправлен email для этого заказа
	var emailSent bool
	err = db.Pool.QueryRow(context.Background(),
		`SELECT EXISTS(
			SELECT 1 FROM logs 
			WHERE entity='order' AND entityid=$1 AND action='EMAIL_SENT'
		)`, op.OrderID).Scan(&emailSent)
	if err != nil {
		// Если ошибка проверки, продолжаем выполнение
		emailSent = false
	}

	// Если email еще не отправлен, отправляем email асинхронно с небольшой задержкой
	// Это дает время для добавления всех товаров в заказ
	if !emailSent {
		// Получаем информацию о заказе и пользователе
		var userID int
		var userEmail string
		err = db.Pool.QueryRow(context.Background(),
			`SELECT o.iduser, u.email 
			 FROM orders o 
			 JOIN users u ON o.iduser = u.iduser 
			 WHERE o.idorder=$1`, op.OrderID).Scan(&userID, &userEmail)
		
		if err == nil && userEmail != "" {
			// Отправляем email асинхронно с задержкой в 2 секунды
			// Это дает время для добавления всех товаров в заказ
			go func(orderID int, userID int, userEmail string) {
				// Ждем 2 секунды, чтобы все товары успели добавиться
				time.Sleep(2 * time.Second)
				
				// Проверяем еще раз, не был ли отправлен email
				var emailSentCheck bool
				db.Pool.QueryRow(context.Background(),
					`SELECT EXISTS(
						SELECT 1 FROM logs 
						WHERE entity='order' AND entityid=$1 AND action='EMAIL_SENT'
					)`, orderID).Scan(&emailSentCheck)
				
				if !emailSentCheck {
					// Проверяем, что в заказе есть товары перед отправкой email
					var itemsCount int
					err := db.Pool.QueryRow(context.Background(),
						"SELECT COUNT(*) FROM orderproducts WHERE idorder=$1", orderID).Scan(&itemsCount)
					
					if err == nil && itemsCount > 0 {
						// Отправляем email о заказе с товарами
						sendOrderEmailWithProducts(userEmail, orderID, userID)
						
						// Логируем отправку email
						db.Pool.Exec(context.Background(),
							`INSERT INTO logs (iduser, action, entity, entityid, details, createdat)
							 VALUES ($1, 'EMAIL_SENT', 'order', $2, $3, NOW())`,
							userID, orderID, fmt.Sprintf("Email отправлен для заказа #%d", orderID))
					}
				}
			}(op.OrderID, userID, userEmail)
		}
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(op)
}

// sendOrderEmailWithProducts отправляет email о заказе с подробной информацией о товарах
func sendOrderEmailWithProducts(toEmail string, orderID int, userID int) {
	fromEmail := "shoesstore0507@gmail.com"
	fromPassword := "bavu udva gljd gfka"
	smtpHost := "smtp.gmail.com"
	smtpPort := "587"

	// Получаем информацию о заказе
	var orderDate time.Time
	err := db.Pool.QueryRow(context.Background(),
		"SELECT orderdate FROM orders WHERE idorder=$1", orderID).Scan(&orderDate)
	if err != nil {
		fmt.Printf("❌ Ошибка получения даты заказа: %v\n", err)
		return
	}

	// Получаем детали заказа (товары)
	rows, err := db.Pool.Query(context.Background(),
		`SELECT op.idorderproduct, op.quantity,
		         p.idproduct, p.name, p.price, p.imageurl,
		         ps.size
		  FROM orderproducts op
		  JOIN productsizes ps ON op.idproductsize = ps.idproductsize
		  JOIN products p ON ps.idproduct = p.idproduct
		  WHERE op.idorder=$1
		  ORDER BY p.name`, orderID)
	if err != nil {
		fmt.Printf("❌ Ошибка получения деталей заказа: %v\n", err)
		return
	}
	defer rows.Close()

	type OrderItem struct {
		ProductName string
		Size        string
		Price       float64
		Quantity    int
		Total       float64
	}

	var items []OrderItem
	var totalAmount float64

	for rows.Next() {
		var item OrderItem
		var productID int
		var imageURL string
		var orderProductID int

		if err := rows.Scan(&orderProductID, &item.Quantity,
			&productID, &item.ProductName, &item.Price, &imageURL, &item.Size); err != nil {
			fmt.Printf("❌ Ошибка сканирования деталей заказа: %v\n", err)
			continue
		}

		item.Total = item.Price * float64(item.Quantity)
		totalAmount += item.Total
		items = append(items, item)
	}

	// Формируем тело письма
	subject := fmt.Sprintf("✅ Заказ #%d принят - Shoes Store", orderID)
	
	body := fmt.Sprintf(`
Здравствуйте!

Ваш заказ #%d был успешно создан.

Дата заказа: %s

Детали заказа:
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
`, orderID, orderDate.Format("02.01.2006 15:04"))

	// Добавляем информацию о каждом товаре
	for i, item := range items {
		body += fmt.Sprintf(`
%d. %s
   Размер: %s
   Цена за единицу: %.2f ₽
   Количество: %d шт
   Сумма: %.2f ₽
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
`, i+1, item.ProductName, item.Size, item.Price, item.Quantity, item.Total)
	}

	// Добавляем итоговую сумму
	if len(items) > 0 {
		body += fmt.Sprintf(`
ИТОГО: %.2f ₽

Мы обработаем ваш заказ в ближайшее время и свяжемся с вами для уточнения деталей.

Спасибо за покупку в Shoes Store!

С уважением,
Команда Shoes Store
`, totalAmount)
	} else {
		body += `
В заказе пока нет товаров. Детали будут отправлены после добавления товаров в заказ.

Спасибо за покупку в Shoes Store!

С уважением,
Команда Shoes Store
`
	}

	message := fmt.Sprintf("From: %s\r\nTo: %s\r\nSubject: %s\r\nContent-Type: text/plain; charset=UTF-8\r\n\r\n%s",
		fromEmail, toEmail, subject, body)

	auth := smtp.PlainAuth("", fromEmail, fromPassword, smtpHost)

	err = smtp.SendMail(smtpHost+":"+smtpPort, auth, fromEmail, []string{toEmail}, []byte(message))
	if err != nil {
		fmt.Printf("❌ Ошибка отправки email о заказе: %v\n", err)
		return
	}

	fmt.Printf("✅ Email о заказе #%d отправлен на %s\n", orderID, toEmail)
}

// ------------------
// Get all OrderProducts
// ------------------
// @Summary Получить все товары всех заказов
// @Tags OrderProducts
// @Produce json
// @Success 200 {array} models.OrderProduct
// @Failure 500 {string} string "Internal Server Error"
// @Router /order-products [get]
func GetOrderProductsHandler(w http.ResponseWriter, r *http.Request) {
	rows, err := db.Pool.Query(context.Background(), "SELECT idorderproduct, idorder, idproductsize, quantity FROM orderproducts")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	var ops []models.OrderProduct
	for rows.Next() {
		var op models.OrderProduct
		if err := rows.Scan(&op.Id, &op.OrderID, &op.ProductSizeID, &op.Quantity); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		ops = append(ops, op)
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(ops)
}

// ------------------
// Get OrderProducts by Order ID
// ------------------
// @Summary Получить товары по заказу
// @Tags OrderProducts
// @Produce json
// @Param order_id path int true "Order ID"
// @Success 200 {array} models.OrderProduct
// @Failure 400 {string} string "Invalid ID"
// @Failure 404 {string} string "Not found"
// @Router /order-products/order/{order_id} [get]
func GetOrderProductsByOrderIDHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	orderID, err := strconv.Atoi(vars["order_id"])
	if err != nil {
		http.Error(w, "Invalid ID", http.StatusBadRequest)
		return
	}

	rows, err := db.Pool.Query(context.Background(),
		`SELECT op.idorderproduct, op.idorder, op.idproductsize, op.quantity,
		         p.idproduct, p.name, p.price, p.imageurl, ps.size
		  FROM orderproducts op
		  JOIN productsizes ps ON op.idproductsize = ps.idproductsize
		  JOIN products p ON ps.idproduct = p.idproduct
		  WHERE op.idorder=$1`, orderID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	type OrderProductDetail struct {
		Id           int     `json:"id"`
		OrderID      int     `json:"order_id"`
		ProductSizeID int    `json:"product_size_id"`
		ProductID    int     `json:"product_id"`
		ProductName  string  `json:"product_name"`
		Size         string  `json:"size"`
		Quantity     int     `json:"quantity"`
		Price        float64 `json:"price"`
		ImageURL     string  `json:"image_url"`
	}

	var ops []OrderProductDetail
	for rows.Next() {
		var op OrderProductDetail
		if err := rows.Scan(&op.Id, &op.OrderID, &op.ProductSizeID, &op.Quantity,
			&op.ProductID, &op.ProductName, &op.Price, &op.ImageURL, &op.Size); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		ops = append(ops, op)
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(ops)
}

// ------------------
// Update OrderProduct quantity
// ------------------
// @Summary Обновить количество товара в заказе
// @Tags OrderProducts
// @Accept json
// @Produce json
// @Param id path int true "OrderProduct ID"
// @Param quantity body int true "Новое количество"
// @Success 200 {object} models.OrderProduct
// @Failure 400 {string} string "Invalid request"
// @Failure 500 {string} string "Internal Server Error"
// @Router /order-products/{id} [put]
func UpdateOrderProductHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id, err := strconv.Atoi(vars["id"])
	if err != nil {
		http.Error(w, "Invalid ID", http.StatusBadRequest)
		return
	}

	var body struct {
		Quantity int `json:"quantity"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	_, err = db.Pool.Exec(context.Background(),
		"UPDATE orderproducts SET quantity=$1 WHERE idorderproduct=$2",
		body.Quantity, id)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	var op models.OrderProduct
	err = db.Pool.QueryRow(context.Background(),
		"SELECT idorderproduct, idorder, idproductsize, quantity FROM orderproducts WHERE idorderproduct=$1", id).
		Scan(&op.Id, &op.OrderID, &op.ProductSizeID, &op.Quantity)
	if err != nil {
		http.Error(w, "Not found", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(op)
}
