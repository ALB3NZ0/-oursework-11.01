package handlers

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/adrg/sysfont"
	"github.com/gorilla/mux"
	"github.com/signintech/gopdf"
	"github.com/xuri/excelize/v2"
	"shoes-store-backend/db"
	"shoes-store-backend/models"
)

// Helper function to create PDF with signintech/gopdf + sysfont (excellent UTF-8 support)
func createPDF() *gopdf.GoPdf {
	pdf := gopdf.GoPdf{}
	pdf.Start(gopdf.Config{PageSize: *gopdf.PageSizeA4})
	pdf.AddPage()

	// Автоматический поиск системного шрифта Arial
	fonts := sysfont.NewFinder(nil)
	font := fonts.Match("Arial")
	
	if font != nil {
		err := pdf.AddTTFFont("Arial", font.Filename)
		if err == nil {
			pdf.SetFont("Arial", "", 12)
		} else {
			// Fallback на встроенный шрифт
			pdf.SetFont("helvetica", "", 12)
		}
	} else {
		// Fallback на встроенный шрифт
		pdf.SetFont("helvetica", "", 12)
	}
	
	return &pdf
}

// Helper function to add text to PDF with proper positioning
func addTextToPDF(pdf *gopdf.GoPdf, text string, fontSize float64, x, y float64) {
	pdf.SetFont("Arial", "", fontSize)
	pdf.SetX(x)
	pdf.SetY(y)
	pdf.Cell(nil, text)
}

// Helper function to add bold text to PDF
func addBoldTextToPDF(pdf *gopdf.GoPdf, text string, fontSize float64, x, y float64) {
	pdf.SetFont("Arial", "B", fontSize)
	pdf.SetX(x)
	pdf.SetY(y)
	pdf.Cell(nil, text)
}

// Helper function to create a simple text report
func createTextReport(title string, content string) string {
	report := fmt.Sprintf("=== %s ===\n\n", title)
	report += content
	report += "\n\n--- End of Report ---"
	return report
}

// ------------------
// Create Report
// ------------------
// @Summary Создать отчёт
// @Tags Reports
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param report body models.Report true "Данные отчёта"
// @Success 201 {object} models.Report
// @Failure 400 {string} string "Invalid request body"
// @Failure 500 {string} string "Internal Server Error"
// @Router /reports [post]
func CreateReportHandler(w http.ResponseWriter, r *http.Request) {
	var report models.Report
	if err := json.NewDecoder(r.Body).Decode(&report); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	err := db.Pool.QueryRow(context.Background(),
		"INSERT INTO reports (reportname, reporttype, reportdata, iduser) VALUES ($1, $2, $3, $4) RETURNING idreport, createdat",
		report.ReportName, report.ReportType, report.ReportData, report.UserID,
	).Scan(&report.Id, &report.CreatedAt)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Логируем создание отчета
	LogUserAction(r, "CREATE", "report", report.Id, fmt.Sprintf("Создан отчет: %s (%s)", report.ReportName, report.ReportType))

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(report)
}

// ------------------
// Get all Reports
// ------------------
// @Summary Получить все отчёты
// @Tags Reports
// @Produce json
// @Security BearerAuth
// @Success 200 {array} models.Report
// @Failure 500 {string} string "Internal Server Error"
// @Router /reports [get]
func GetReportsHandler(w http.ResponseWriter, r *http.Request) {
	rows, err := db.Pool.Query(context.Background(),
		"SELECT idreport, reportname, reporttype, reportdata, iduser, createdat FROM reports")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	var reports []models.Report
	for rows.Next() {
		var r models.Report
		if err := rows.Scan(&r.Id, &r.ReportName, &r.ReportType, &r.ReportData, &r.UserID, &r.CreatedAt); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		reports = append(reports, r)
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(reports)
}

// ------------------
// Get Report by ID
// ------------------
// @Summary Получить отчёт по ID
// @Tags Reports
// @Produce json
// @Security BearerAuth
// @Param id path int true "Report ID"
// @Success 200 {object} models.Report
// @Failure 400 {string} string "Invalid ID"
// @Failure 404 {string} string "Report not found"
// @Router /reports/{id} [get]
func GetReportByIDHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id, err := strconv.Atoi(vars["id"])
	if err != nil {
		http.Error(w, "Invalid ID", http.StatusBadRequest)
		return
	}

	var report models.Report
	err = db.Pool.QueryRow(context.Background(),
		"SELECT idreport, reportname, reporttype, reportdata, iduser, createdat FROM reports WHERE idreport=$1",
		id).Scan(&report.Id, &report.ReportName, &report.ReportType, &report.ReportData, &report.UserID, &report.CreatedAt)
	if err != nil {
		http.Error(w, "Report not found", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(report)
}

// ------------------
// PDF Report Handlers
// ------------------

// @Summary Сгенерировать отчет по продажам в PDF
// @Tags Reports
// @Produce application/pdf
// @Security BearerAuth
// @Success 200 {file} file "PDF файл отчета"
// @Failure 500 {string} string "Internal Server Error"
// @Router /reports/sales/pdf [get]
func GenerateSalesPDFHandler(w http.ResponseWriter, r *http.Request) {
	// Получаем данные для отчета используя БД функцию
	var salesData models.SalesReportData
	
	// Общая статистика продаж за последние 30 дней
	err := db.Pool.QueryRow(context.Background(),
		`SELECT COALESCE(SUM(op.quantity * p.price), 0) as total_sales,
		        COUNT(DISTINCT o.idorder) as total_orders
		 FROM orders o
		 JOIN orderproducts op ON o.idorder = op.idorder
		 JOIN productsizes ps ON op.idproductsize = ps.idproductsize
		 JOIN products p ON ps.idproduct = p.idproduct
		 WHERE o.orderdate >= NOW() - INTERVAL '30 days'`).Scan(&salesData.TotalSales, &salesData.TotalOrders)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	
	// Топ продуктов за последние 30 дней
	rows, err := db.Pool.Query(context.Background(),
		`SELECT 
			p.name as product_name,
			SUM(op.quantity) as total_quantity,
			SUM(op.quantity * p.price) as total_revenue
		FROM orderproducts op
		JOIN orders o ON op.idorder = o.idorder
		JOIN productsizes ps ON op.idproductsize = ps.idproductsize
		JOIN products p ON ps.idproduct = p.idproduct
		WHERE o.orderdate >= CURRENT_DATE - INTERVAL '30 days'
		GROUP BY p.idproduct, p.name
		ORDER BY total_revenue DESC
		LIMIT 10`)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer rows.Close()
	
	for rows.Next() {
		var ps models.ProductSales
		rows.Scan(&ps.ProductName, &ps.TotalSold, &ps.TotalRevenue)
		salesData.ProductSales = append(salesData.ProductSales, ps)
	}
	
	salesData.Period = "Последние 30 дней"
	
	// Создаем PDF с поддержкой UTF-8 через phpdave11/gofpdf
	pdf := createPDF()
	
	// Заголовок
	pdf.SetFont("Arial", "B", 18)
	pdf.SetX(50)
	pdf.SetY(50)
	pdf.Cell(nil, "ОТЧЕТ ПО ПРОДАЖАМ")
	
	// Общая информация
	pdf.SetFont("Arial", "", 12)
	pdf.SetX(50)
	pdf.SetY(80)
	pdf.Cell(nil, fmt.Sprintf("Общая выручка: %.2f RUB", salesData.TotalSales))
	pdf.SetX(50)
	pdf.SetY(100)
	pdf.Cell(nil, fmt.Sprintf("Всего заказов: %d", salesData.TotalOrders))
	pdf.SetX(50)
	pdf.SetY(120)
	pdf.Cell(nil, fmt.Sprintf("Период: %s", salesData.Period))
	
	// Заголовок раздела
	pdf.SetFont("Arial", "B", 14)
	pdf.SetX(50)
	pdf.SetY(150)
	pdf.Cell(nil, "ТОП ПРОДУКТЫ")
	
	// Список продуктов
	y := 180.0
	for _, ps := range salesData.ProductSales {
		pdf.SetFont("Arial", "", 10)
		pdf.SetX(50)
		pdf.SetY(y)
		pdf.Cell(nil, fmt.Sprintf("- %s: %d шт. (%.2f RUB)", ps.ProductName, ps.TotalSold, ps.TotalRevenue))
		y += 20
	}
	
	// Сохраняем во временный файл
	tempFile := fmt.Sprintf("temp_sales_report_%d.pdf", time.Now().Unix())
	err = pdf.WritePdf(tempFile)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	
	// Читаем файл и отправляем
	fileContent, err := os.ReadFile(tempFile)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	
	// Удаляем временный файл
	os.Remove(tempFile)
	
	// Логируем генерацию отчета
	LogUserAction(r, "GENERATE_PDF", "report", 0, "Сгенерирован PDF отчет по продажам")
	
	w.Header().Set("Content-Type", "application/pdf")
	w.Header().Set("Content-Disposition", "attachment; filename=sales_report.pdf")
	w.Write(fileContent)
}

// @Summary Сгенерировать отчет по инвентарю в PDF
// @Tags Reports
// @Produce application/pdf
// @Security BearerAuth
// @Success 200 {file} file "PDF файл отчета"
// @Failure 500 {string} string "Internal Server Error"
// @Router /reports/inventory/pdf [get]
func GenerateInventoryPDFHandler(w http.ResponseWriter, r *http.Request) {
	var inventoryData models.InventoryReportData
	
	// Общее количество продуктов
	err := db.Pool.QueryRow(context.Background(),
		`SELECT COUNT(DISTINCT p.idproduct) FROM products p`).Scan(&inventoryData.TotalProducts)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	
	// Продукты с низким остатком (группируем по продукту, чтобы избежать дубликатов)
	rows, err := db.Pool.Query(context.Background(),
		`SELECT p.idproduct, p.name, MIN(ps.quantity) as min_quantity, 5 as min_stock
		 FROM products p
		 JOIN productsizes ps ON p.idproduct = ps.idproduct
		 WHERE ps.quantity <= 5
		 GROUP BY p.idproduct, p.name
		 ORDER BY min_quantity ASC`)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer rows.Close()
	
	for rows.Next() {
		var lsp models.LowStockProduct
		rows.Scan(&lsp.ProductID, &lsp.ProductName, &lsp.CurrentStock, &lsp.MinStock)
		inventoryData.LowStockProducts = append(inventoryData.LowStockProducts, lsp)
	}
	
	// Создаем PDF с поддержкой UTF-8 через phpdave11/gofpdf
	pdf := createPDF()
	
	// Заголовок
	pdf.SetFont("Arial", "B", 18)
	pdf.SetX(50)
	pdf.SetY(50)
	pdf.Cell(nil, "ОТЧЕТ ПО ИНВЕНТАРЮ")
	
	// Общая информация
	pdf.SetFont("Arial", "", 12)
	pdf.SetX(50)
	pdf.SetY(80)
	pdf.Cell(nil, fmt.Sprintf("Всего продуктов: %d", inventoryData.TotalProducts))
	pdf.SetX(50)
	pdf.SetY(100)
	pdf.Cell(nil, fmt.Sprintf("Продуктов с низким остатком: %d", len(inventoryData.LowStockProducts)))
	
	// Заголовок раздела
	pdf.SetFont("Arial", "B", 14)
	pdf.SetX(50)
	pdf.SetY(130)
	pdf.Cell(nil, "ПРОДУКТЫ С НИЗКИМ ОСТАТКОМ")
	
	// Список продуктов
	y := 160.0
	for _, lsp := range inventoryData.LowStockProducts {
		pdf.SetFont("Arial", "", 10)
		pdf.SetX(50)
		pdf.SetY(y)
		pdf.Cell(nil, fmt.Sprintf("- %s: %d шт. (мин: %d)", lsp.ProductName, lsp.CurrentStock, lsp.MinStock))
		y += 20
	}
	
	// Сохраняем во временный файл
	tempFile := fmt.Sprintf("temp_inventory_report_%d.pdf", time.Now().Unix())
	err = pdf.WritePdf(tempFile)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	
	// Читаем файл и отправляем
	fileContent, err := os.ReadFile(tempFile)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	
	// Удаляем временный файл
	os.Remove(tempFile)
	
	// Логируем генерацию отчета
	LogUserAction(r, "GENERATE_PDF", "report", 0, "Сгенерирован PDF отчет по инвентарю")
	
	w.Header().Set("Content-Type", "application/pdf")
	w.Header().Set("Content-Disposition", "attachment; filename=inventory_report.pdf")
	w.Write(fileContent)
}

// @Summary Сгенерировать отчет по клиентам в PDF
// @Tags Reports
// @Produce application/pdf
// @Security BearerAuth
// @Success 200 {file} file "PDF файл отчета"
// @Failure 500 {string} string "Internal Server Error"
// @Router /reports/customers/pdf [get]
func GenerateCustomerPDFHandler(w http.ResponseWriter, r *http.Request) {
	var customerData models.CustomerReportData
	
	// Общее количество клиентов
	err := db.Pool.QueryRow(context.Background(),
		`SELECT COUNT(*) FROM users WHERE roleid = 3`).Scan(&customerData.TotalCustomers)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	
	// Топ клиенты используя БД функцию GetUserTotalSpent
	rows, err := db.Pool.Query(context.Background(),
		`SELECT u.iduser, u.fullname, 
		        COUNT(o.idorder) as total_orders,
		        GetUserTotalSpent(u.iduser) as total_spent
		 FROM users u
		 LEFT JOIN orders o ON u.iduser = o.iduser
		 WHERE u.roleid = 3
		 GROUP BY u.iduser, u.fullname
		 ORDER BY GetUserTotalSpent(u.iduser) DESC
		 LIMIT 10`)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer rows.Close()
	
	for rows.Next() {
		var tc models.TopCustomer
		rows.Scan(&tc.UserID, &tc.UserName, &tc.TotalOrders, &tc.TotalSpent)
		customerData.TopCustomers = append(customerData.TopCustomers, tc)
	}
	
	// Создаем PDF с поддержкой UTF-8 через phpdave11/gofpdf
	pdf := createPDF()
	
	// Заголовок
	pdf.SetFont("Arial", "B", 18)
	pdf.SetX(50)
	pdf.SetY(50)
	pdf.Cell(nil, "ОТЧЕТ ПО КЛИЕНТАМ")
	
	// Общая информация
	pdf.SetFont("Arial", "", 12)
	pdf.SetX(50)
	pdf.SetY(80)
	pdf.Cell(nil, fmt.Sprintf("Всего клиентов: %d", customerData.TotalCustomers))
	
	// Заголовок раздела
	pdf.SetFont("Arial", "B", 14)
	pdf.SetX(50)
	pdf.SetY(110)
	pdf.Cell(nil, "ТОП КЛИЕНТЫ")
	
	// Список клиентов
	y := 140.0
	for _, tc := range customerData.TopCustomers {
		pdf.SetFont("Arial", "", 10)
		pdf.SetX(50)
		pdf.SetY(y)
		pdf.Cell(nil, fmt.Sprintf("- %s: %d заказов (%.2f RUB)", tc.UserName, tc.TotalOrders, tc.TotalSpent))
		y += 20
	}
	
	// Сохраняем во временный файл
	tempFile := fmt.Sprintf("temp_customer_report_%d.pdf", time.Now().Unix())
	err = pdf.WritePdf(tempFile)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	
	// Читаем файл и отправляем
	fileContent, err := os.ReadFile(tempFile)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	
	// Удаляем временный файл
	os.Remove(tempFile)
	
	// Логируем генерацию отчета
	LogUserAction(r, "GENERATE_PDF", "report", 0, "Сгенерирован PDF отчет по клиентам")
	
	w.Header().Set("Content-Type", "application/pdf")
	w.Header().Set("Content-Disposition", "attachment; filename=customer_report.pdf")
	w.Write(fileContent)
}

// @Summary Сгенерировать отчет по клиентам в текстовом формате
// @Tags Reports
// @Produce text/plain; charset=utf-8
// @Security BearerAuth
// @Success 200 {string} string "Текстовый отчет"
// @Failure 500 {string} string "Internal Server Error"
// @Router /reports/customers/text [get]
func GenerateCustomerTextHandler(w http.ResponseWriter, r *http.Request) {
	var customerData models.CustomerReportData
	
	// Общее количество клиентов
	err := db.Pool.QueryRow(context.Background(),
		`SELECT COUNT(*) FROM users WHERE roleid = 3`).Scan(&customerData.TotalCustomers)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	
	// Топ клиенты используя БД функцию GetUserTotalSpent
	rows, err := db.Pool.Query(context.Background(),
		`SELECT u.iduser, u.fullname, 
		        COUNT(o.idorder) as total_orders,
		        GetUserTotalSpent(u.iduser) as total_spent
		 FROM users u
		 LEFT JOIN orders o ON u.iduser = o.iduser
		 WHERE u.roleid = 3
		 GROUP BY u.iduser, u.fullname
		 ORDER BY GetUserTotalSpent(u.iduser) DESC
		 LIMIT 10`)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer rows.Close()
	
	for rows.Next() {
		var tc models.TopCustomer
		rows.Scan(&tc.UserID, &tc.UserName, &tc.TotalOrders, &tc.TotalSpent)
		customerData.TopCustomers = append(customerData.TopCustomers, tc)
	}
	
	// Создаем текстовый отчет
	content := fmt.Sprintf("Total Customers: %d\n\n", customerData.TotalCustomers)
	content += "TOP CUSTOMERS\n"
	
	for _, tc := range customerData.TopCustomers {
		content += fmt.Sprintf("- %s: %d orders (%.2f RUB)\n", tc.UserName, tc.TotalOrders, tc.TotalSpent)
	}
	
	report := createTextReport("CUSTOMERS REPORT", content)
	
	// Логируем генерацию отчета
	LogUserAction(r, "GENERATE_TEXT", "report", 0, "Сгенерирован текстовый отчет по клиентам")
	
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	w.Header().Set("Content-Disposition", "attachment; filename=customer_report.txt")
	w.Write([]byte(report))
}

// @Summary Сгенерировать отчет по инвентарю в текстовом формате
// @Tags Reports
// @Produce text/plain; charset=utf-8
// @Security BearerAuth
// @Success 200 {string} string "Текстовый отчет"
// @Failure 500 {string} string "Internal Server Error"
// @Router /reports/inventory/text [get]
func GenerateInventoryTextHandler(w http.ResponseWriter, r *http.Request) {
	var inventoryData models.InventoryReportData
	
	// Общее количество продуктов
	err := db.Pool.QueryRow(context.Background(),
		`SELECT COUNT(DISTINCT p.idproduct) FROM products p`).Scan(&inventoryData.TotalProducts)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	
	// Продукты с низким остатком (группируем по продукту, чтобы избежать дубликатов)
	rows, err := db.Pool.Query(context.Background(),
		`SELECT p.idproduct, p.name, MIN(ps.quantity) as min_quantity, 5 as min_stock
		 FROM products p
		 JOIN productsizes ps ON p.idproduct = ps.idproduct
		 WHERE ps.quantity <= 5
		 GROUP BY p.idproduct, p.name
		 ORDER BY min_quantity ASC`)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer rows.Close()
	
	for rows.Next() {
		var lsp models.LowStockProduct
		rows.Scan(&lsp.ProductID, &lsp.ProductName, &lsp.CurrentStock, &lsp.MinStock)
		inventoryData.LowStockProducts = append(inventoryData.LowStockProducts, lsp)
	}
	
	// Создаем текстовый отчет
	content := fmt.Sprintf("Total Products: %d\n", inventoryData.TotalProducts)
	content += fmt.Sprintf("Low Stock Products: %d\n\n", len(inventoryData.LowStockProducts))
	content += "LOW STOCK PRODUCTS\n"
	
	for _, lsp := range inventoryData.LowStockProducts {
		content += fmt.Sprintf("- %s: %d pcs (min: %d)\n", lsp.ProductName, lsp.CurrentStock, lsp.MinStock)
	}
	
	report := createTextReport("INVENTORY REPORT", content)
	
	// Логируем генерацию отчета
	LogUserAction(r, "GENERATE_TEXT", "report", 0, "Сгенерирован текстовый отчет по инвентарю")
	
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	w.Header().Set("Content-Disposition", "attachment; filename=inventory_report.txt")
	w.Write([]byte(report))
}

// @Summary Сгенерировать отчет по категориям в PDF
// @Tags Reports
// @Produce application/pdf
// @Security BearerAuth
// @Success 200 {file} file "PDF файл отчета"
// @Failure 500 {string} string "Internal Server Error"
// @Router /reports/categories/pdf [get]
func GenerateCategoriesPDFHandler(w http.ResponseWriter, r *http.Request) {
	var categoryData []models.CategorySales
	
	// Выручка по категориям за последние 30 дней
	rows, err := db.Pool.Query(context.Background(),
		`SELECT 
			c.categoryname as category_name,
			SUM(op.quantity * p.price) as total_revenue
		FROM orderproducts op
		JOIN orders o ON op.idorder = o.idorder
		JOIN productsizes ps ON op.idproductsize = ps.idproductsize
		JOIN products p ON ps.idproduct = p.idproduct
		JOIN categories c ON p.idcategory = c.idcategory
		WHERE o.orderdate >= CURRENT_DATE - INTERVAL '30 days'
		GROUP BY c.idcategory, c.categoryname
		ORDER BY total_revenue DESC`)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer rows.Close()
	
	for rows.Next() {
		var cs models.CategorySales
		rows.Scan(&cs.CategoryName, &cs.TotalRevenue)
		categoryData = append(categoryData, cs)
	}
	
	// Создаем PDF с поддержкой UTF-8 через phpdave11/gofpdf
	pdf := createPDF()
	
	// Заголовок
	pdf.SetFont("Arial", "B", 18)
	pdf.SetX(50)
	pdf.SetY(50)
	pdf.Cell(nil, "ОТЧЕТ ПО КАТЕГОРИЯМ")
	
	// Общая информация
	pdf.SetFont("Arial", "", 12)
	pdf.SetX(50)
	pdf.SetY(80)
	pdf.Cell(nil, "Период: Последние 30 дней")
	
	// Заголовок раздела
	pdf.SetFont("Arial", "B", 14)
	pdf.SetX(50)
	pdf.SetY(110)
	pdf.Cell(nil, "ВЫРУЧКА ПО КАТЕГОРИЯМ")
	
	// Список категорий
	y := 140.0
	for _, cs := range categoryData {
		pdf.SetFont("Arial", "", 10)
		pdf.SetX(50)
		pdf.SetY(y)
		pdf.Cell(nil, fmt.Sprintf("- %s: %.2f RUB (%d шт.)", cs.CategoryName, cs.TotalRevenue, cs.TotalSold))
		y += 20
	}
	
	// Сохраняем во временный файл
	tempFile := fmt.Sprintf("temp_categories_report_%d.pdf", time.Now().Unix())
	err = pdf.WritePdf(tempFile)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	
	// Читаем файл и отправляем
	fileContent, err := os.ReadFile(tempFile)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	
	// Удаляем временный файл
	os.Remove(tempFile)
	
	// Логируем генерацию отчета
	LogUserAction(r, "GENERATE_PDF", "report", 0, "Сгенерирован PDF отчет по категориям")
	
	w.Header().Set("Content-Type", "application/pdf")
	w.Header().Set("Content-Disposition", "attachment; filename=categories_report.pdf")
	w.Write(fileContent)
}

// ------------------
// Excel Report Handlers
// ------------------

// @Summary Сгенерировать отчет по продажам в Excel
// @Tags Reports
// @Produce application/vnd.openxmlformats-officedocument.spreadsheetml.sheet
// @Security BearerAuth
// @Success 200 {file} file "Excel файл отчета"
// @Failure 500 {string} string "Internal Server Error"
// @Router /reports/sales/excel [get]
func GenerateSalesExcelHandler(w http.ResponseWriter, r *http.Request) {
	var salesData models.SalesReportData
	
	// Общая статистика продаж за последние 30 дней
	err := db.Pool.QueryRow(context.Background(),
		`SELECT COALESCE(SUM(op.quantity * p.price), 0) as total_sales,
		        COUNT(DISTINCT o.idorder) as total_orders
		 FROM orders o
		 JOIN orderproducts op ON o.idorder = op.idorder
		 JOIN productsizes ps ON op.idproductsize = ps.idproductsize
		 JOIN products p ON ps.idproduct = p.idproduct
		 WHERE o.orderdate >= NOW() - INTERVAL '30 days'`).Scan(&salesData.TotalSales, &salesData.TotalOrders)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	
	// Топ продуктов за последние 30 дней
	rows, err := db.Pool.Query(context.Background(),
		`SELECT 
			p.name as product_name,
			SUM(op.quantity) as total_quantity,
			SUM(op.quantity * p.price) as total_revenue
		FROM orderproducts op
		JOIN orders o ON op.idorder = o.idorder
		JOIN productsizes ps ON op.idproductsize = ps.idproductsize
		JOIN products p ON ps.idproduct = p.idproduct
		WHERE o.orderdate >= CURRENT_DATE - INTERVAL '30 days'
		GROUP BY p.idproduct, p.name
		ORDER BY total_revenue DESC
		LIMIT 10`)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer rows.Close()
	
	for rows.Next() {
		var ps models.ProductSales
		rows.Scan(&ps.ProductName, &ps.TotalSold, &ps.TotalRevenue)
		salesData.ProductSales = append(salesData.ProductSales, ps)
	}
	
	salesData.Period = "Последние 30 дней"
	
	// Создаем Excel файл
	f := excelize.NewFile()
	sheetName := "Отчет по продажам"
	f.NewSheet(sheetName)
	f.DeleteSheet("Sheet1")
	
	// Стили
	headerStyle, _ := f.NewStyle(&excelize.Style{
		Font: &excelize.Font{Bold: true, Size: 14, Color: "#FFFFFF"},
		Fill: excelize.Fill{Type: "pattern", Color: []string{"#4472C4"}, Pattern: 1},
		Alignment: &excelize.Alignment{Horizontal: "center", Vertical: "center"},
	})
	
	titleStyle, _ := f.NewStyle(&excelize.Style{
		Font: &excelize.Font{Bold: true, Size: 16},
	})
	
	// Заголовок
	f.SetCellValue(sheetName, "A1", "ОТЧЕТ ПО ПРОДАЖАМ")
	f.MergeCell(sheetName, "A1", "D1")
	f.SetCellStyle(sheetName, "A1", "A1", titleStyle)
	f.SetRowHeight(sheetName, 1, 25)
	
	// Общая информация
	f.SetCellValue(sheetName, "A3", "Общая выручка:")
	f.SetCellValue(sheetName, "B3", fmt.Sprintf("%.2f RUB", salesData.TotalSales))
	f.SetCellValue(sheetName, "A4", "Всего заказов:")
	f.SetCellValue(sheetName, "B4", salesData.TotalOrders)
	f.SetCellValue(sheetName, "A5", "Период:")
	f.SetCellValue(sheetName, "B5", salesData.Period)
	
	// Заголовок таблицы
	f.SetCellValue(sheetName, "A7", "Товар")
	f.SetCellValue(sheetName, "B7", "Количество (шт.)")
	f.SetCellValue(sheetName, "C7", "Выручка (RUB)")
	f.SetCellStyle(sheetName, "A7", "C7", headerStyle)
	f.SetRowHeight(sheetName, 7, 20)
	
	// Данные продуктов
	for i, ps := range salesData.ProductSales {
		row := 8 + i
		f.SetCellValue(sheetName, fmt.Sprintf("A%d", row), ps.ProductName)
		f.SetCellValue(sheetName, fmt.Sprintf("B%d", row), ps.TotalSold)
		f.SetCellValue(sheetName, fmt.Sprintf("C%d", row), fmt.Sprintf("%.2f", ps.TotalRevenue))
	}
	
	// Автоширина колонок
	f.SetColWidth(sheetName, "A", "A", 40)
	f.SetColWidth(sheetName, "B", "B", 20)
	f.SetColWidth(sheetName, "C", "C", 20)
	
	// Сохраняем во временный файл
	tempFile := fmt.Sprintf("temp_sales_report_%d.xlsx", time.Now().Unix())
	if err := f.SaveAs(tempFile); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	
	// Читаем файл и отправляем
	fileContent, err := os.ReadFile(tempFile)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	
	// Удаляем временный файл
	os.Remove(tempFile)
	
	// Логируем генерацию отчета
	LogUserAction(r, "GENERATE_EXCEL", "report", 0, "Сгенерирован Excel отчет по продажам")
	
	w.Header().Set("Content-Type", "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet")
	w.Header().Set("Content-Disposition", "attachment; filename=sales_report.xlsx")
	w.Write(fileContent)
}

// @Summary Сгенерировать отчет по инвентарю в Excel
// @Tags Reports
// @Produce application/vnd.openxmlformats-officedocument.spreadsheetml.sheet
// @Security BearerAuth
// @Success 200 {file} file "Excel файл отчета"
// @Failure 500 {string} string "Internal Server Error"
// @Router /reports/inventory/excel [get]
func GenerateInventoryExcelHandler(w http.ResponseWriter, r *http.Request) {
	var inventoryData models.InventoryReportData
	
	// Общее количество продуктов
	err := db.Pool.QueryRow(context.Background(),
		`SELECT COUNT(DISTINCT p.idproduct) FROM products p`).Scan(&inventoryData.TotalProducts)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	
	// Продукты с низким остатком
	rows, err := db.Pool.Query(context.Background(),
		`SELECT p.idproduct, p.name, MIN(ps.quantity) as min_quantity, 5 as min_stock
		 FROM products p
		 JOIN productsizes ps ON p.idproduct = ps.idproduct
		 WHERE ps.quantity <= 5
		 GROUP BY p.idproduct, p.name
		 ORDER BY min_quantity ASC`)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer rows.Close()
	
	for rows.Next() {
		var lsp models.LowStockProduct
		rows.Scan(&lsp.ProductID, &lsp.ProductName, &lsp.CurrentStock, &lsp.MinStock)
		inventoryData.LowStockProducts = append(inventoryData.LowStockProducts, lsp)
	}
	
	// Создаем Excel файл
	f := excelize.NewFile()
	sheetName := "Отчет по инвентарю"
	f.NewSheet(sheetName)
	f.DeleteSheet("Sheet1")
	
	// Стили
	headerStyle, _ := f.NewStyle(&excelize.Style{
		Font: &excelize.Font{Bold: true, Size: 14, Color: "#FFFFFF"},
		Fill: excelize.Fill{Type: "pattern", Color: []string{"#E74C3C"}, Pattern: 1},
		Alignment: &excelize.Alignment{Horizontal: "center", Vertical: "center"},
	})
	
	titleStyle, _ := f.NewStyle(&excelize.Style{
		Font: &excelize.Font{Bold: true, Size: 16},
	})
	
	// Заголовок
	f.SetCellValue(sheetName, "A1", "ОТЧЕТ ПО ИНВЕНТАРЮ")
	f.MergeCell(sheetName, "A1", "D1")
	f.SetCellStyle(sheetName, "A1", "A1", titleStyle)
	f.SetRowHeight(sheetName, 1, 25)
	
	// Общая информация
	f.SetCellValue(sheetName, "A3", "Всего продуктов:")
	f.SetCellValue(sheetName, "B3", inventoryData.TotalProducts)
	f.SetCellValue(sheetName, "A4", "Продуктов с низким остатком:")
	f.SetCellValue(sheetName, "B4", len(inventoryData.LowStockProducts))
	
	// Заголовок таблицы
	f.SetCellValue(sheetName, "A6", "ID Товара")
	f.SetCellValue(sheetName, "B6", "Название товара")
	f.SetCellValue(sheetName, "C6", "Текущий остаток")
	f.SetCellValue(sheetName, "D6", "Минимальный остаток")
	f.SetCellStyle(sheetName, "A6", "D6", headerStyle)
	f.SetRowHeight(sheetName, 6, 20)
	
	// Данные продуктов
	for i, lsp := range inventoryData.LowStockProducts {
		row := 7 + i
		f.SetCellValue(sheetName, fmt.Sprintf("A%d", row), lsp.ProductID)
		f.SetCellValue(sheetName, fmt.Sprintf("B%d", row), lsp.ProductName)
		f.SetCellValue(sheetName, fmt.Sprintf("C%d", row), lsp.CurrentStock)
		f.SetCellValue(sheetName, fmt.Sprintf("D%d", row), lsp.MinStock)
	}
	
	// Автоширина колонок
	f.SetColWidth(sheetName, "A", "A", 12)
	f.SetColWidth(sheetName, "B", "B", 40)
	f.SetColWidth(sheetName, "C", "C", 18)
	f.SetColWidth(sheetName, "D", "D", 20)
	
	// Сохраняем во временный файл
	tempFile := fmt.Sprintf("temp_inventory_report_%d.xlsx", time.Now().Unix())
	if err := f.SaveAs(tempFile); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	
	// Читаем файл и отправляем
	fileContent, err := os.ReadFile(tempFile)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	
	// Удаляем временный файл
	os.Remove(tempFile)
	
	// Логируем генерацию отчета
	LogUserAction(r, "GENERATE_EXCEL", "report", 0, "Сгенерирован Excel отчет по инвентарю")
	
	w.Header().Set("Content-Type", "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet")
	w.Header().Set("Content-Disposition", "attachment; filename=inventory_report.xlsx")
	w.Write(fileContent)
}

// @Summary Сгенерировать отчет по клиентам в Excel
// @Tags Reports
// @Produce application/vnd.openxmlformats-officedocument.spreadsheetml.sheet
// @Security BearerAuth
// @Success 200 {file} file "Excel файл отчета"
// @Failure 500 {string} string "Internal Server Error"
// @Router /reports/customers/excel [get]
func GenerateCustomerExcelHandler(w http.ResponseWriter, r *http.Request) {
	var customerData models.CustomerReportData
	
	// Общее количество клиентов
	err := db.Pool.QueryRow(context.Background(),
		`SELECT COUNT(*) FROM users WHERE roleid = 3`).Scan(&customerData.TotalCustomers)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	
	// Топ клиенты
	rows, err := db.Pool.Query(context.Background(),
		`SELECT u.iduser, u.fullname, 
		        COUNT(o.idorder) as total_orders,
		        GetUserTotalSpent(u.iduser) as total_spent
		 FROM users u
		 LEFT JOIN orders o ON u.iduser = o.iduser
		 WHERE u.roleid = 3
		 GROUP BY u.iduser, u.fullname
		 ORDER BY GetUserTotalSpent(u.iduser) DESC
		 LIMIT 10`)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer rows.Close()
	
	for rows.Next() {
		var tc models.TopCustomer
		rows.Scan(&tc.UserID, &tc.UserName, &tc.TotalOrders, &tc.TotalSpent)
		customerData.TopCustomers = append(customerData.TopCustomers, tc)
	}
	
	// Создаем Excel файл
	f := excelize.NewFile()
	sheetName := "Отчет по клиентам"
	f.NewSheet(sheetName)
	f.DeleteSheet("Sheet1")
	
	// Стили
	headerStyle, _ := f.NewStyle(&excelize.Style{
		Font: &excelize.Font{Bold: true, Size: 14, Color: "#FFFFFF"},
		Fill: excelize.Fill{Type: "pattern", Color: []string{"#27AE60"}, Pattern: 1},
		Alignment: &excelize.Alignment{Horizontal: "center", Vertical: "center"},
	})
	
	titleStyle, _ := f.NewStyle(&excelize.Style{
		Font: &excelize.Font{Bold: true, Size: 16},
	})
	
	// Заголовок
	f.SetCellValue(sheetName, "A1", "ОТЧЕТ ПО КЛИЕНТАМ")
	f.MergeCell(sheetName, "A1", "D1")
	f.SetCellStyle(sheetName, "A1", "A1", titleStyle)
	f.SetRowHeight(sheetName, 1, 25)
	
	// Общая информация
	f.SetCellValue(sheetName, "A3", "Всего клиентов:")
	f.SetCellValue(sheetName, "B3", customerData.TotalCustomers)
	
	// Заголовок таблицы
	f.SetCellValue(sheetName, "A5", "ID Клиента")
	f.SetCellValue(sheetName, "B5", "Имя клиента")
	f.SetCellValue(sheetName, "C5", "Количество заказов")
	f.SetCellValue(sheetName, "D5", "Общая сумма (RUB)")
	f.SetCellStyle(sheetName, "A5", "D5", headerStyle)
	f.SetRowHeight(sheetName, 5, 20)
	
	// Данные клиентов
	for i, tc := range customerData.TopCustomers {
		row := 6 + i
		f.SetCellValue(sheetName, fmt.Sprintf("A%d", row), tc.UserID)
		f.SetCellValue(sheetName, fmt.Sprintf("B%d", row), tc.UserName)
		f.SetCellValue(sheetName, fmt.Sprintf("C%d", row), tc.TotalOrders)
		f.SetCellValue(sheetName, fmt.Sprintf("D%d", row), fmt.Sprintf("%.2f", tc.TotalSpent))
	}
	
	// Автоширина колонок
	f.SetColWidth(sheetName, "A", "A", 12)
	f.SetColWidth(sheetName, "B", "B", 30)
	f.SetColWidth(sheetName, "C", "C", 22)
	f.SetColWidth(sheetName, "D", "D", 22)
	
	// Сохраняем во временный файл
	tempFile := fmt.Sprintf("temp_customer_report_%d.xlsx", time.Now().Unix())
	if err := f.SaveAs(tempFile); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	
	// Читаем файл и отправляем
	fileContent, err := os.ReadFile(tempFile)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	
	// Удаляем временный файл
	os.Remove(tempFile)
	
	// Логируем генерацию отчета
	LogUserAction(r, "GENERATE_EXCEL", "report", 0, "Сгенерирован Excel отчет по клиентам")
	
	w.Header().Set("Content-Type", "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet")
	w.Header().Set("Content-Disposition", "attachment; filename=customer_report.xlsx")
	w.Write(fileContent)
}

// @Summary Сгенерировать отчет по категориям в Excel
// @Tags Reports
// @Produce application/vnd.openxmlformats-officedocument.spreadsheetml.sheet
// @Security BearerAuth
// @Success 200 {file} file "Excel файл отчета"
// @Failure 500 {string} string "Internal Server Error"
// @Router /reports/categories/excel [get]
func GenerateCategoriesExcelHandler(w http.ResponseWriter, r *http.Request) {
	var categoryData []models.CategorySales
	
	// Выручка по категориям за последние 30 дней
	rows, err := db.Pool.Query(context.Background(),
		`SELECT 
			c.categoryname as category_name,
			SUM(op.quantity * p.price) as total_revenue,
			SUM(op.quantity) as total_sold
		FROM orderproducts op
		JOIN orders o ON op.idorder = o.idorder
		JOIN productsizes ps ON op.idproductsize = ps.idproductsize
		JOIN products p ON ps.idproduct = p.idproduct
		JOIN categories c ON p.idcategory = c.idcategory
		WHERE o.orderdate >= CURRENT_DATE - INTERVAL '30 days'
		GROUP BY c.idcategory, c.categoryname
		ORDER BY total_revenue DESC`)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer rows.Close()
	
	for rows.Next() {
		var cs models.CategorySales
		err := rows.Scan(&cs.CategoryName, &cs.TotalRevenue, &cs.TotalSold)
		if err != nil {
			continue
		}
		categoryData = append(categoryData, cs)
	}
	
	// Создаем Excel файл
	f := excelize.NewFile()
	sheetName := "Отчет по категориям"
	f.NewSheet(sheetName)
	f.DeleteSheet("Sheet1")
	
	// Стили
	headerStyle, _ := f.NewStyle(&excelize.Style{
		Font: &excelize.Font{Bold: true, Size: 14, Color: "#FFFFFF"},
		Fill: excelize.Fill{Type: "pattern", Color: []string{"#9B59B6"}, Pattern: 1},
		Alignment: &excelize.Alignment{Horizontal: "center", Vertical: "center"},
	})
	
	titleStyle, _ := f.NewStyle(&excelize.Style{
		Font: &excelize.Font{Bold: true, Size: 16},
	})
	
	// Заголовок
	f.SetCellValue(sheetName, "A1", "ОТЧЕТ ПО КАТЕГОРИЯМ")
	f.MergeCell(sheetName, "A1", "D1")
	f.SetCellStyle(sheetName, "A1", "A1", titleStyle)
	f.SetRowHeight(sheetName, 1, 25)
	
	// Общая информация
	f.SetCellValue(sheetName, "A3", "Период:")
	f.SetCellValue(sheetName, "B3", "Последние 30 дней")
	
	// Заголовок таблицы
	f.SetCellValue(sheetName, "A5", "Категория")
	f.SetCellValue(sheetName, "B5", "Выручка (RUB)")
	f.SetCellValue(sheetName, "C5", "Продано (шт.)")
	f.SetCellStyle(sheetName, "A5", "C5", headerStyle)
	f.SetRowHeight(sheetName, 5, 20)
	
	// Данные категорий
	for i, cs := range categoryData {
		row := 6 + i
		f.SetCellValue(sheetName, fmt.Sprintf("A%d", row), cs.CategoryName)
		f.SetCellValue(sheetName, fmt.Sprintf("B%d", row), fmt.Sprintf("%.2f", cs.TotalRevenue))
		f.SetCellValue(sheetName, fmt.Sprintf("C%d", row), cs.TotalSold)
	}
	
	// Автоширина колонок
	f.SetColWidth(sheetName, "A", "A", 35)
	f.SetColWidth(sheetName, "B", "B", 20)
	f.SetColWidth(sheetName, "C", "C", 18)
	
	// Сохраняем во временный файл
	tempFile := fmt.Sprintf("temp_categories_report_%d.xlsx", time.Now().Unix())
	if err := f.SaveAs(tempFile); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	
	// Читаем файл и отправляем
	fileContent, err := os.ReadFile(tempFile)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	
	// Удаляем временный файл
	os.Remove(tempFile)
	
	// Логируем генерацию отчета
	LogUserAction(r, "GENERATE_EXCEL", "report", 0, "Сгенерирован Excel отчет по категориям")
	
	w.Header().Set("Content-Type", "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet")
	w.Header().Set("Content-Disposition", "attachment; filename=categories_report.xlsx")
	w.Write(fileContent)
}