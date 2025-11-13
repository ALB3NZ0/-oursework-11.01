package handlers

import (
	"context"
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/gorilla/mux"
	"shoes-store-backend/db"
	"shoes-store-backend/models"
)

// ------------------
// Get all Logs
// ------------------
// @Summary Получить все логи действий
// @Tags Logs
// @Produce json
// @Success 200 {array} models.Log
// @Failure 500 {string} string "Internal Server Error"
// @Router /logs [get]
func GetLogsHandler(w http.ResponseWriter, r *http.Request) {
	rows, err := db.Pool.Query(context.Background(),
		"SELECT idlog, iduser, action, entity, entityid, details, createdat FROM logs ORDER BY createdat DESC")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	var logs []models.Log
	for rows.Next() {
		var l models.Log
		if err := rows.Scan(&l.Id, &l.UserID, &l.Action, &l.Entity, &l.EntityID, &l.Details, &l.CreatedAt); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		logs = append(logs, l)
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(logs)
}

// ------------------
// Get Log by ID
// ------------------
// @Summary Получить лог по ID
// @Tags Logs
// @Produce json
// @Param id path int true "Log ID"
// @Success 200 {object} models.Log
// @Failure 400 {string} string "Invalid ID"
// @Failure 404 {string} string "Log not found"
// @Router /logs/{id} [get]
func GetLogByIDHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id, err := strconv.Atoi(vars["id"])
	if err != nil {
		http.Error(w, "Invalid ID", http.StatusBadRequest)
		return
	}

	var log models.Log
	err = db.Pool.QueryRow(context.Background(),
		"SELECT idlog, iduser, action, entity, entityid, details, createdat FROM logs WHERE idlog=$1",
		id).Scan(&log.Id, &log.UserID, &log.Action, &log.Entity, &log.EntityID, &log.Details, &log.CreatedAt)
	if err != nil {
		http.Error(w, "Log not found", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(log)
}
