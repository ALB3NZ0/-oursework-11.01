// @title Shoes Store API
// @version 1.0
// @description API для интернет-магазина обуви

// @host localhost:8080
// @BasePath /

// @securityDefinitions.apikey BearerAuth
// @in header
// @name Authorization

package main

import (
	"fmt"
	"net/http"

	"github.com/gorilla/mux"
	httpSwagger "github.com/swaggo/http-swagger"
	_ "shoes-store-backend/docs"
	"shoes-store-backend/db"
	"shoes-store-backend/handlers"
	"shoes-store-backend/handlers/admin"
	"shoes-store-backend/middlewares"


)

func main() {
	db.InitDB()

	r := mux.NewRouter()

	// CORS для фронтенда
	r.Use(func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			origin := r.Header.Get("Origin")
			
			// Всегда отдаем CORS headers
			w.Header().Set("Access-Control-Allow-Origin", "*")
			w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS, PATCH")
			w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization, X-Requested-With")
			w.Header().Set("Access-Control-Max-Age", "3600")
			
			fmt.Printf("🌐 CORS: Method=%s, Origin=%s, Path=%s\n", r.Method, origin, r.URL.Path)
			
			// Если OPTIONS запрос - сразу отдаем успешный ответ
			if r.Method == "OPTIONS" {
				fmt.Printf("✅ CORS preflight OK for %s\n", r.URL.Path)
				w.WriteHeader(http.StatusOK)
				return
			}
			
			next.ServeHTTP(w, r)
		})
	})

	r.Use(middlewares.JWTMiddleware)
	r.Use(middlewares.LoggerMiddleware)


	// General
	r.HandleFunc("/", handlers.HelloHandler).Methods("GET")

	// Auth
	r.HandleFunc("/register", handlers.RegisterHandler).Methods("POST", "OPTIONS")
	r.HandleFunc("/login", handlers.LoginHandler).Methods("POST", "OPTIONS")

	// Users CRUD (только админы)
	r.Handle("/users", middlewares.RequireRole("admin")(http.HandlerFunc(handlers.CreateUserHandler))).Methods("POST", "OPTIONS")   // Создать пользователя
	r.Handle("/users", middlewares.RequireRole("admin")(http.HandlerFunc(handlers.GetUsersHandler))).Methods("GET", "OPTIONS")      // Получить всех
	r.Handle("/users/{id}", middlewares.RequireRole("admin")(http.HandlerFunc(handlers.GetUserByIDHandler))).Methods("GET", "OPTIONS") // Получить по ID
	r.Handle("/users/{id}", middlewares.RequireRole("admin")(http.HandlerFunc(handlers.UpdateUserHandler))).Methods("PUT", "OPTIONS")  // Обновить
	r.Handle("/users/{id}", middlewares.RequireRole("admin")(http.HandlerFunc(handlers.DeleteUserHandler))).Methods("DELETE", "OPTIONS") // Удалить
	// Product CRUD (все могут читать, только админы могут изменять)
	r.Handle("/products", middlewares.RequireRole("admin")(http.HandlerFunc(handlers.CreateProductHandler))).Methods("POST", "OPTIONS")
	r.HandleFunc("/products", handlers.GetProductsHandler).Methods("GET", "OPTIONS") // Все могут читать
	r.HandleFunc("/products/{id}", handlers.GetProductByIDHandler).Methods("GET", "OPTIONS") // Все могут читать
	r.Handle("/products/{id}", middlewares.RequireRole("admin")(http.HandlerFunc(handlers.UpdateProductHandler))).Methods("PUT", "OPTIONS")
	r.Handle("/products/{id}", middlewares.RequireRole("admin")(http.HandlerFunc(handlers.DeleteProductHandler))).Methods("DELETE", "OPTIONS")
	// Brands (все могут читать, только админы могут изменять)
	r.Handle("/brands", middlewares.RequireRole("admin")(http.HandlerFunc(handlers.CreateBrandHandler))).Methods("POST", "OPTIONS")
	r.HandleFunc("/brands", handlers.GetBrandsHandler).Methods("GET", "OPTIONS") // Все могут читать
	r.HandleFunc("/brands/{id}", handlers.GetBrandByIDHandler).Methods("GET", "OPTIONS") // Все могут читать
	r.Handle("/brands/{id}", middlewares.RequireRole("admin")(http.HandlerFunc(handlers.UpdateBrandHandler))).Methods("PUT", "OPTIONS")
	r.Handle("/brands/{id}", middlewares.RequireRole("admin")(http.HandlerFunc(handlers.DeleteBrandHandler))).Methods("DELETE", "OPTIONS")
	// Categories CRUD (все могут читать, только админы могут изменять)
	r.Handle("/categories", middlewares.RequireRole("admin")(http.HandlerFunc(handlers.CreateCategoryHandler))).Methods("POST", "OPTIONS")
	r.HandleFunc("/categories", handlers.GetCategoriesHandler).Methods("GET", "OPTIONS") // Все могут читать
	r.HandleFunc("/categories/{id}", handlers.GetCategoryByIDHandler).Methods("GET", "OPTIONS") // Все могут читать
	r.Handle("/categories/{id}", middlewares.RequireRole("admin")(http.HandlerFunc(handlers.UpdateCategoryHandler))).Methods("PUT", "OPTIONS")
	r.Handle("/categories/{id}", middlewares.RequireRole("admin")(http.HandlerFunc(handlers.DeleteCategoryHandler))).Methods("DELETE", "OPTIONS")
	// ProductSizes (все могут читать, только админы могут обновлять количество)
	r.HandleFunc("/products/{product_id}/sizes", handlers.GetSizesByProductHandler).Methods("GET", "OPTIONS") // Все могут читать
	r.Handle("/productsizes/{id}", middlewares.RequireRole("admin")(http.HandlerFunc(handlers.UpdateProductSizeHandler))).Methods("PUT", "OPTIONS")
	// Basket (пользователи могут работать со своими корзинами)
	r.Handle("/basket/{user_id}", middlewares.RequireRole("user")(http.HandlerFunc(handlers.GetBasketHandler))).Methods("GET", "OPTIONS")
	r.Handle("/basket", middlewares.RequireRole("user")(http.HandlerFunc(handlers.AddToBasketHandler))).Methods("POST", "OPTIONS")
	r.Handle("/basket/{id}", middlewares.RequireRole("user")(http.HandlerFunc(handlers.UpdateBasketHandler))).Methods("PUT", "OPTIONS")
	r.Handle("/basket/{id}", middlewares.RequireRole("user")(http.HandlerFunc(handlers.DeleteBasketHandler))).Methods("DELETE", "OPTIONS")
	// Favorites (пользователи могут работать со своими избранными)
	r.Handle("/favorites/{user_id}", middlewares.RequireRole("user")(http.HandlerFunc(handlers.GetFavoritesHandler))).Methods("GET", "OPTIONS")
	r.Handle("/favorites", middlewares.RequireRole("user")(http.HandlerFunc(handlers.AddToFavoritesHandler))).Methods("POST", "OPTIONS")
	r.Handle("/favorites/{id}", middlewares.RequireRole("user")(http.HandlerFunc(handlers.DeleteFavoriteHandler))).Methods("DELETE", "OPTIONS")
	// Reviews (все могут читать, пользователи могут создавать свои)
	r.HandleFunc("/reviews/product/{id}", handlers.GetReviewsByProductHandler).Methods("GET", "OPTIONS") // Все могут читать
	r.Handle("/reviews/user/{id}", middlewares.RequireRole("user")(http.HandlerFunc(handlers.GetReviewsByUserHandler))).Methods("GET", "OPTIONS")
	r.Handle("/reviews", middlewares.RequireRole("user")(http.HandlerFunc(handlers.CreateReviewHandler))).Methods("POST", "OPTIONS")
	r.Handle("/reviews/{id}", middlewares.RequireRole("user")(http.HandlerFunc(handlers.UpdateReviewHandler))).Methods("PUT", "OPTIONS")
	r.Handle("/reviews/{id}", middlewares.RequireRole("user")(http.HandlerFunc(handlers.DeleteReviewHandler))).Methods("DELETE", "OPTIONS")
	// Orders (только админы могут видеть все, пользователи создают свои)
	r.Handle("/orders", middlewares.RequireRole("admin")(http.HandlerFunc(handlers.GetOrdersHandler))).Methods("GET", "OPTIONS")
	r.Handle("/orders/user/{user_id}", middlewares.RequireRole("user")(http.HandlerFunc(handlers.GetOrdersByUserHandler))).Methods("GET", "OPTIONS")
	r.Handle("/orders", middlewares.RequireRole("user")(http.HandlerFunc(handlers.CreateOrderHandler))).Methods("POST", "OPTIONS")
	// OrdersDetails (только админы могут видеть все, пользователи создают свои)
	r.Handle("/order-products", middlewares.RequireRole("user")(http.HandlerFunc(handlers.CreateOrderProductHandler))).Methods("POST", "OPTIONS")
	r.Handle("/order-products", middlewares.RequireRole("admin")(http.HandlerFunc(handlers.GetOrderProductsHandler))).Methods("GET", "OPTIONS")
	r.Handle("/order-products/order/{order_id}", middlewares.RequireRole("user")(http.HandlerFunc(handlers.GetOrderProductsByOrderIDHandler))).Methods("GET", "OPTIONS")
	r.Handle("/order-products/{id}", middlewares.RequireRole("admin")(http.HandlerFunc(handlers.UpdateOrderProductHandler))).Methods("PUT", "OPTIONS")
	// Reports (только менеджеры)
	r.Handle("/reports", middlewares.RequireRole("manager")(http.HandlerFunc(handlers.CreateReportHandler))).Methods("POST", "OPTIONS")
	r.Handle("/reports", middlewares.RequireRole("manager")(http.HandlerFunc(handlers.GetReportsHandler))).Methods("GET", "OPTIONS")
	r.Handle("/reports/{id}", middlewares.RequireRole("manager")(http.HandlerFunc(handlers.GetReportByIDHandler))).Methods("GET", "OPTIONS")
	
	// PDF Reports (только менеджеры)
	r.Handle("/reports/sales/pdf", middlewares.RequireRole("manager")(http.HandlerFunc(handlers.GenerateSalesPDFHandler))).Methods("GET", "OPTIONS")
	r.Handle("/reports/inventory/pdf", middlewares.RequireRole("manager")(http.HandlerFunc(handlers.GenerateInventoryPDFHandler))).Methods("GET", "OPTIONS")
	r.Handle("/reports/customers/pdf", middlewares.RequireRole("manager")(http.HandlerFunc(handlers.GenerateCustomerPDFHandler))).Methods("GET", "OPTIONS")
	r.Handle("/reports/categories/pdf", middlewares.RequireRole("manager")(http.HandlerFunc(handlers.GenerateCategoriesPDFHandler))).Methods("GET", "OPTIONS")
	
	// Excel Reports (только менеджеры)
	r.Handle("/reports/sales/excel", middlewares.RequireRole("manager")(http.HandlerFunc(handlers.GenerateSalesExcelHandler))).Methods("GET", "OPTIONS")
	r.Handle("/reports/inventory/excel", middlewares.RequireRole("manager")(http.HandlerFunc(handlers.GenerateInventoryExcelHandler))).Methods("GET", "OPTIONS")
	r.Handle("/reports/customers/excel", middlewares.RequireRole("manager")(http.HandlerFunc(handlers.GenerateCustomerExcelHandler))).Methods("GET", "OPTIONS")
	r.Handle("/reports/categories/excel", middlewares.RequireRole("manager")(http.HandlerFunc(handlers.GenerateCategoriesExcelHandler))).Methods("GET", "OPTIONS")
	
	// Text Reports (только менеджеры) - UTF-8 совместимые
	r.Handle("/reports/customers/text", middlewares.RequireRole("manager")(http.HandlerFunc(handlers.GenerateCustomerTextHandler))).Methods("GET", "OPTIONS")
	r.Handle("/reports/inventory/text", middlewares.RequireRole("manager")(http.HandlerFunc(handlers.GenerateInventoryTextHandler))).Methods("GET", "OPTIONS")
	//Logs (только админы)
	r.Handle("/logs", middlewares.RequireRole("admin")(http.HandlerFunc(handlers.GetLogsHandler))).Methods("GET", "OPTIONS")
	r.Handle("/logs/{id}", middlewares.RequireRole("admin")(http.HandlerFunc(handlers.GetLogByIDHandler))).Methods("GET", "OPTIONS")

	// Admin Product (только для админов)
	r.Handle("/admin/products", middlewares.RequireRole("admin")(http.HandlerFunc(admin.AdminGetProductsHandler))).Methods("GET", "OPTIONS")
	r.Handle("/admin/products/{id}", middlewares.RequireRole("admin")(http.HandlerFunc(admin.AdminGetProductByIDHandler))).Methods("GET", "OPTIONS")
	r.Handle("/admin/products", middlewares.RequireRole("admin")(http.HandlerFunc(admin.AdminCreateProductHandler))).Methods("POST", "OPTIONS")
	r.Handle("/admin/products/{id}", middlewares.RequireRole("admin")(http.HandlerFunc(admin.AdminUpdateProductHandler))).Methods("PUT", "OPTIONS")
	r.Handle("/admin/products/{id}", middlewares.RequireRole("admin")(http.HandlerFunc(admin.AdminDeleteProductHandler))).Methods("DELETE", "OPTIONS")

	// Admin Brands (только для админов)
	r.Handle("/admin/brands", middlewares.RequireRole("admin")(http.HandlerFunc(admin.AdminGetBrandsHandler))).Methods("GET", "OPTIONS")
	r.Handle("/admin/brands/{id}", middlewares.RequireRole("admin")(http.HandlerFunc(admin.AdminGetBrandByIDHandler))).Methods("GET", "OPTIONS")
	r.Handle("/admin/brands", middlewares.RequireRole("admin")(http.HandlerFunc(admin.AdminCreateBrandHandler))).Methods("POST", "OPTIONS")
	r.Handle("/admin/brands/{id}", middlewares.RequireRole("admin")(http.HandlerFunc(admin.AdminUpdateBrandHandler))).Methods("PUT", "OPTIONS")
	r.Handle("/admin/brands/{id}", middlewares.RequireRole("admin")(http.HandlerFunc(admin.AdminDeleteBrandHandler))).Methods("DELETE", "OPTIONS")
	// Admin Categories (только для админов)
	r.Handle("/admin/categories", middlewares.RequireRole("admin")(http.HandlerFunc(admin.AdminGetCategoriesHandler))).Methods("GET", "OPTIONS")
	r.Handle("/admin/categories/{id}", middlewares.RequireRole("admin")(http.HandlerFunc(admin.AdminGetCategoryByIDHandler))).Methods("GET", "OPTIONS")
	r.Handle("/admin/categories", middlewares.RequireRole("admin")(http.HandlerFunc(admin.AdminCreateCategoryHandler))).Methods("POST", "OPTIONS")
	r.Handle("/admin/categories/{id}", middlewares.RequireRole("admin")(http.HandlerFunc(admin.AdminUpdateCategoryHandler))).Methods("PUT", "OPTIONS")
	r.Handle("/admin/categories/{id}", middlewares.RequireRole("admin")(http.HandlerFunc(admin.AdminDeleteCategoryHandler))).Methods("DELETE", "OPTIONS")
	// Admin Users (только для админов)
	r.Handle("/admin/users", middlewares.RequireRole("admin")(http.HandlerFunc(admin.AdminGetUsersHandler))).Methods("GET", "OPTIONS")
	r.Handle("/admin/users/{id}", middlewares.RequireRole("admin")(http.HandlerFunc(admin.AdminGetUserByIDHandler))).Methods("GET", "OPTIONS")
	r.Handle("/admin/users", middlewares.RequireRole("admin")(http.HandlerFunc(admin.AdminCreateUserHandler))).Methods("POST", "OPTIONS")
	r.Handle("/admin/users/{id}", middlewares.RequireRole("admin")(http.HandlerFunc(admin.AdminUpdateUserHandler))).Methods("PUT", "OPTIONS")
	r.Handle("/admin/users/{id}", middlewares.RequireRole("admin")(http.HandlerFunc(admin.AdminDeleteUserHandler))).Methods("DELETE", "OPTIONS")
	// Admin Reviews (только для админов)
	r.Handle("/admin/reviews", middlewares.RequireRole("admin")(http.HandlerFunc(admin.AdminGetReviewsHandler))).Methods("GET", "OPTIONS")
	r.Handle("/admin/reviews/{id}", middlewares.RequireRole("admin")(http.HandlerFunc(admin.AdminGetReviewByIDHandler))).Methods("GET", "OPTIONS")
	r.Handle("/admin/reviews/{id}", middlewares.RequireRole("admin")(http.HandlerFunc(admin.AdminUpdateReviewHandler))).Methods("PUT", "OPTIONS")
	r.Handle("/admin/reviews/{id}", middlewares.RequireRole("admin")(http.HandlerFunc(admin.AdminDeleteReviewHandler))).Methods("DELETE", "OPTIONS")

	// Admin Logs (только для админов)
	r.Handle("/admin/logs", middlewares.RequireRole("admin")(http.HandlerFunc(admin.AdminGetLogsHandler))).Methods("GET", "OPTIONS")
	r.Handle("/admin/logs/{id}", middlewares.RequireRole("admin")(http.HandlerFunc(admin.AdminGetLogByIDHandler))).Methods("GET", "OPTIONS")
	r.Handle("/admin/logs/{id}", middlewares.RequireRole("admin")(http.HandlerFunc(admin.AdminDeleteLogHandler))).Methods("DELETE", "OPTIONS")

	// Support (все могут отправлять сообщения)
	r.HandleFunc("/support", handlers.SendSupportMessageHandler).Methods("POST", "OPTIONS")

	// Admin Backup (только для админов)
	r.Handle("/admin/backup", middlewares.RequireRole("admin")(http.HandlerFunc(handlers.CreateBackupHandler))).Methods("POST", "OPTIONS")
	r.Handle("/admin/backup/info", middlewares.RequireRole("admin")(http.HandlerFunc(handlers.GetBackupInfoHandler))).Methods("GET", "OPTIONS")
	r.Handle("/admin/backup/restore", middlewares.RequireRole("admin")(http.HandlerFunc(handlers.RestoreBackupHandler))).Methods("POST", "OPTIONS")
	r.Handle("/admin/backup/download/{filename}", middlewares.RequireRole("admin")(http.HandlerFunc(handlers.DownloadBackupHandler))).Methods("GET", "OPTIONS")
	r.Handle("/admin/backup/{filename}", middlewares.RequireRole("admin")(http.HandlerFunc(handlers.DeleteBackupHandler))).Methods("DELETE", "OPTIONS")

	// Password Management
	r.HandleFunc("/password/reset", handlers.RequestPasswordResetHandler).Methods("POST", "OPTIONS")                    // Запрос восстановления (без авторизации)
	r.HandleFunc("/password/reset/confirm", handlers.ConfirmPasswordResetHandler).Methods("POST", "OPTIONS")            // Подтверждение восстановления (без авторизации)
	r.Handle("/password/change", middlewares.RequireRole("user")(http.HandlerFunc(handlers.ChangePasswordHandler))).Methods("POST", "OPTIONS")        // Смена пароля (требует авторизации)
	r.Handle("/password/change/confirm", middlewares.RequireRole("user")(http.HandlerFunc(handlers.ConfirmPasswordChangeHandler))).Methods("POST", "OPTIONS") // Подтверждение смены (требует авторизации)

	// Swagger UI
	r.PathPrefix("/swagger/").Handler(httpSwagger.WrapHandler)

	fmt.Println("✅ Server running on http://localhost:8080")
	http.ListenAndServe(":8080", r)
}
