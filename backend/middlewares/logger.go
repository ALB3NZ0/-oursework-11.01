package middlewares

import (
	"context"
	"net/http"
	"strings"
	"time"

	"shoes-store-backend/db"
)

// LoggerMiddleware — логгер всех действий
func LoggerMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		next.ServeHTTP(w, r)

		if r.Method != http.MethodPost && r.Method != http.MethodPut && r.Method != http.MethodDelete {
			return
		}

		path := strings.TrimPrefix(r.URL.Path, "/")
		parts := strings.Split(path, "/")

		role := "user"
		if len(parts) > 0 {
			if parts[0] == "admin" {
				role = "admin"
			} else if parts[0] == "manager" {
				role = "manager"
			}
		}

		entity := ""
		if len(parts) > 1 {
			entity = parts[1]
		} else {
			entity = parts[0]
		}

		action := map[string]string{
			http.MethodPost:   "CREATE",
			http.MethodPut:    "UPDATE",
			http.MethodDelete: "DELETE",
		}[r.Method]

		userID, ok := r.Context().Value("userID").(int)
		if !ok {
			// Не логируем если нет userID (например, для публичных endpoints)
			return
		}

		details := role + " " + action + " " + entity

		go func() {
			_, err := db.Pool.Exec(context.Background(),
				`INSERT INTO logs (iduser, action, entity, details, createdat)
				VALUES ($1, $2, $3, $4, $5)`,
				userID, action, role+"/"+entity, details, start)
			if err != nil { println("Ошибка при логировании:", err.Error()) }
		}()

	})
}

