package models

type OrderProduct struct {
	Id           int `json:"id"`
	OrderID      int `json:"order_id"`
	ProductSizeID int `json:"product_size_id"`
	Quantity     int `json:"quantity"`
}
