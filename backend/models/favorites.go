package models

// Чистая запись
type Favorite struct {
	ID            int `json:"id"`
	UserID        int `json:"user_id"`
	ProductSizeID int `json:"product_size_id"`
}

// Расширенная запись для ответа (с инфой о товаре)
type FavoriteItem struct {
	ID            int    `json:"id"`
	UserID        int    `json:"user_id"`
	ProductSizeID int    `json:"product_size_id"`
	ProductID     int    `json:"product_id"`
	ProductName   string `json:"product_name"`
	Size          int    `json:"size"`
	Price         string `json:"price"`
	ImageURL      string `json:"image_url"`
}
