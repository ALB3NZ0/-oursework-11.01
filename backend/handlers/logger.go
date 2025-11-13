package handlers

import (
	"context"
	"fmt"
	"net/http"

	"shoes-store-backend/db"
)

// ---------------------------
// Helper function to log user actions
// ---------------------------
func LogUserAction(r *http.Request, action, entity string, entityID int, details string) {
	userID := r.Context().Value("userID")
	if userID == nil {
		return // Если нет userID, не логируем
	}

	_, err := db.Pool.Exec(context.Background(),
		`INSERT INTO logs (iduser, action, entity, entityid, details, createdat)
		 VALUES ($1, $2, $3, $4, $5, NOW())`,
		userID, action, entity, entityID, details)
	if err != nil {
		fmt.Printf("Error logging user action: %v\n", err)
	}
}

