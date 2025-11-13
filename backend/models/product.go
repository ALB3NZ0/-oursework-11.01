package models

// Product модель для таблицы Products
type Product struct {
	ID        int     `json:"id"`
	Name      string  `json:"name"`
	ImageUrl  string  `json:"image_url,omitempty"`
	Price     float64 `json:"price"`
	BrandID   int     `json:"brand_id"`
	CategoryID int    `json:"category_id"`
}
