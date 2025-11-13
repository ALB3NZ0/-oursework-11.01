package models

type Category struct {
	ID           int    `json:"id" db:"idcategory"`
	CategoryName string `json:"category_name" db:"categoryname"`
}
