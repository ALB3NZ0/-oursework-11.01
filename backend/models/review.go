package models

import "time"

type Review struct {
	ID        int       `json:"id"`
	ProductID int       `json:"product_id"`
	UserID    int       `json:"user_id"`
	Rating    int       `json:"rating"`
	Comment   string    `json:"comment"`
	Date      time.Time `json:"date"`
}
