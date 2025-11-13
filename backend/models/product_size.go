package models

type ProductSize struct {
	ID        int `json:"id" db:"idproductsize"`
	IdProduct int `json:"id_product" db:"idproduct"`
	Size      int `json:"size" db:"size"`
	Quantity  int `json:"quantity" db:"quantity"`
}
