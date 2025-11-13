package middlewares

import (
	"context"
	"net/http"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

var jwtKey = []byte("secret_key") // лучше хранить в env

type Claims struct {
	UserID int    `json:"user_id"`
	Role   string `json:"role"`
	jwt.RegisteredClaims
}

// --------------------
// Генерация токена
// --------------------
func GenerateJWT(userID int, role string) (string, error) {
	expirationTime := time.Now().Add(24 * time.Hour)

	claims := &Claims{
		UserID: userID,
		Role:   role,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(expirationTime),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(jwtKey)
}

// --------------------
// Middleware для проверки JWT
// --------------------
func JWTMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		// Пропускаем OPTIONS запросы для CORS
		if r.Method == "OPTIONS" {
			next.ServeHTTP(w, r)
			return
		}

		path := r.URL.Path

		// Разрешённые пути без токена
		if strings.HasPrefix(path, "/swagger/") ||
			path == "/login" ||
			path == "/register" ||
			path == "/" ||
			path == "/support" ||
			path == "/products" ||
			path == "/brands" ||
			path == "/categories" ||
			strings.HasPrefix(path, "/products/") ||
			strings.HasPrefix(path, "/brands/") ||
			strings.HasPrefix(path, "/categories/") ||
			strings.HasPrefix(path, "/products/") && strings.Contains(path, "/sizes") ||
			strings.HasPrefix(path, "/reviews/product/") ||
			path == "/password/reset" ||
			path == "/password/reset/confirm" {
			next.ServeHTTP(w, r)
			return
		}

		authHeader := r.Header.Get("Authorization")
		
		if authHeader == "" {
			http.Error(w, "Missing token", http.StatusUnauthorized)
			return
		}

		tokenStr := strings.TrimPrefix(authHeader, "Bearer ")

		claims := &Claims{}
		token, err := jwt.ParseWithClaims(tokenStr, claims, func(token *jwt.Token) (interface{}, error) {
			return jwtKey, nil
		})

		if err != nil {
			http.Error(w, "Invalid token", http.StatusUnauthorized)
			return
		}

		if !token.Valid {
			http.Error(w, "Invalid token", http.StatusUnauthorized)
			return
		}

		ctx := context.WithValue(r.Context(), "userID", claims.UserID)
		ctx = context.WithValue(ctx, "role", claims.Role)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// --------------------
// Проверка ролей
// --------------------
func RequireRole(requiredRole string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			role := r.Context().Value("role").(string)
			
			// Админ может все
			if role == "admin" {
				next.ServeHTTP(w, r)
				return
			}
			
			// Менеджер может только отчеты
			if role == "manager" && requiredRole == "manager" {
				next.ServeHTTP(w, r)
				return
			}
			
			// Если требуемая роль не "manager", то менеджеру доступ запрещен
			if role == "manager" && requiredRole != "manager" {
				http.Error(w, "Access denied", http.StatusForbidden)
				return
			}
			
			// Пользователь может только свои данные
			if role == "user" && requiredRole == "user" {
				next.ServeHTTP(w, r)
				return
			}
			
			http.Error(w, "Access denied", http.StatusForbidden)
		})
	}
}