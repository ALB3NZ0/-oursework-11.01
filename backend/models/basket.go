package models

// Чистая запись в БД
type Basket struct {
	ID            int `json:"id"`
	UserID        int `json:"user_id"`
	ProductSizeID int `json:"product_size_id"`
	Quantity      int `json:"quantity"`
}

// Расширенная для ответа (с инфой о товаре)
type BasketItem struct {
	ID           int     `json:"id"`
	UserID       int     `json:"user_id"`
	ProductSizeID int    `json:"product_size_id"`
	Quantity     int     `json:"quantity"`
	ProductID    int     `json:"product_id"`
	ProductName  string  `json:"product_name"`
	Size         int     `json:"size"`
	Available    int     `json:"available"`
	Price        float64 `json:"price"`
	ImageURL     string  `json:"image_url"`
}
