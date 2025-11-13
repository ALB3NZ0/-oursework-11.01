package models

import "time"

type Report struct {
	Id         int       `json:"id"`
	ReportName string    `json:"report_name"`
	ReportType string    `json:"report_type"`
	ReportData string    `json:"report_data"`
	UserID     int       `json:"user_id"`
	CreatedAt  time.Time `json:"created_at"`
}

// SalesReportData - данные отчета по продажам
type SalesReportData struct {
	TotalSales     float64         `json:"total_sales"`
	TotalOrders    int             `json:"total_orders"`
	ProductSales   []ProductSales  `json:"product_sales"`
	CategorySales  []CategorySales `json:"category_sales"`
	Period         string          `json:"period"`
}

type ProductSales struct {
	ProductID   int     `json:"product_id"`
	ProductName string  `json:"product_name"`
	TotalSold   int     `json:"total_sold"`
	TotalRevenue float64 `json:"total_revenue"`
}

type CategorySales struct {
	CategoryID   int     `json:"category_id"`
	CategoryName string  `json:"category_name"`
	TotalSold    int     `json:"total_sold"`
	TotalRevenue float64 `json:"total_revenue"`
}

// InventoryReportData - данные отчета по инвентарю
type InventoryReportData struct {
	TotalProducts    int               `json:"total_products"`
	LowStockProducts []LowStockProduct `json:"low_stock_products"`
	OutOfStock       []OutOfStock       `json:"out_of_stock"`
	CategoryStock    []CategoryStock    `json:"category_stock"`
}

type LowStockProduct struct {
	ProductID   int    `json:"product_id"`
	ProductName string `json:"product_name"`
	CurrentStock int   `json:"current_stock"`
	MinStock     int   `json:"min_stock"`
}

type OutOfStock struct {
	ProductID   int    `json:"product_id"`
	ProductName string `json:"product_name"`
	Size         string `json:"size"`
}

type CategoryStock struct {
	CategoryID   int `json:"category_id"`
	CategoryName string `json:"category_name"`
	TotalStock   int `json:"total_stock"`
}

// CustomerReportData - данные отчета по клиентам
type CustomerReportData struct {
	TotalCustomers int           `json:"total_customers"`
	TopCustomers   []TopCustomer  `json:"top_customers"`
	RecentOrders   []RecentOrder  `json:"recent_orders"`
}

type TopCustomer struct {
	UserID      int     `json:"user_id"`
	UserName    string  `json:"user_name"`
	TotalOrders int     `json:"total_orders"`
	TotalSpent  float64 `json:"total_spent"`
}

type RecentOrder struct {
	OrderID    int       `json:"order_id"`
	UserID     int       `json:"user_id"`
	UserName   string    `json:"user_name"`
	OrderDate  time.Time `json:"order_date"`
	TotalAmount float64   `json:"total_amount"`
}