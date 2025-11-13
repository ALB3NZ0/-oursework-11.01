package admin

import (
    "context"
    "encoding/json"
    "net/http"
    "strconv"

    "github.com/gorilla/mux"
    "shoes-store-backend/db"
    "shoes-store-backend/handlers"
    "shoes-store-backend/models"
)


// ---------------------------
// Get All Logs (Admin)
// ---------------------------

// @Summary Получить все логи (Admin, с пагинацией)
// @Tags Admin Logs
// @Produce json
// @Security BearerAuth
// @Param page query int false "Номер страницы (по умолчанию 1)"
// @Param limit query int false "Количество элементов на странице (по умолчанию 20, максимум 100)"
// @Success 200 {object} handlers.PaginatedResponse
// @Failure 500 {string} string "Internal Server Error"
// @Router /admin/logs [get]
func AdminGetLogsHandler(w http.ResponseWriter, r *http.Request) {
    params := handlers.ParsePaginationParams(r)

    // Получаем общее количество логов
    var total int
    err := db.Pool.QueryRow(context.Background(),
        `SELECT COUNT(*) FROM logs`).Scan(&total)
    if err != nil {
        http.Error(w, err.Error(), http.StatusInternalServerError)
        return
    }

    // Получаем логи с пагинацией
    rows, err := db.Pool.Query(context.Background(),
        `SELECT idlog, iduser, action, entity, entityid, details, createdat
         FROM logs ORDER BY createdat DESC LIMIT $1 OFFSET $2`,
        params.Limit, params.GetOffset())
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

    totalPages := handlers.CalculateTotalPages(total, params.Limit)
    response := handlers.PaginatedResponse{
        Data:       logs,
        Page:       params.Page,
        Limit:      params.Limit,
        Total:      total,
        TotalPages: totalPages,
    }

    w.Header().Set("Content-Type", "application/json")
    json.NewEncoder(w).Encode(response)
}


// ---------------------------
// Get Log by ID (Admin)
// ---------------------------

// @Summary Получить лог по ID (Admin)
// @Tags Admin Logs
// @Produce json
// @Security BearerAuth
// @Param id path int true "Log ID"
// @Success 200 {object} models.Log
// @Failure 400 {string} string "Invalid ID"
// @Failure 404 {string} string "Log not found"
// @Router /admin/logs/{id} [get]
func AdminGetLogByIDHandler(w http.ResponseWriter, r *http.Request) {
    vars := mux.Vars(r)
    id, err := strconv.Atoi(vars["id"])
    if err != nil {
        http.Error(w, "Invalid ID", http.StatusBadRequest)
        return
    }

    var log models.Log
    err = db.Pool.QueryRow(context.Background(),
        `SELECT idlog, iduser, action, entity, entityid, details, createdat
         FROM logs WHERE idlog=$1`, id).
        Scan(&log.Id, &log.UserID, &log.Action, &log.Entity, &log.EntityID, &log.Details, &log.CreatedAt)
    if err != nil {
        http.Error(w, "Log not found", http.StatusNotFound)
        return
    }

    w.Header().Set("Content-Type", "application/json")
    json.NewEncoder(w).Encode(log)
}


// ---------------------------
// Delete Log (Admin)
// ---------------------------

// @Summary Удалить лог (Admin)
// @Tags Admin Logs
// @Produce json
// @Security BearerAuth
// @Param id path int true "Log ID"
// @Success 204 "No Content"
// @Failure 400 {string} string "Invalid ID"
// @Failure 404 {string} string "Log not found"
// @Router /admin/logs/{id} [delete]
func AdminDeleteLogHandler(w http.ResponseWriter, r *http.Request) {
    vars := mux.Vars(r)
    id, err := strconv.Atoi(vars["id"])
    if err != nil {
        http.Error(w, "Invalid ID", http.StatusBadRequest)
        return
    }

    res, err := db.Pool.Exec(context.Background(),
        `DELETE FROM logs WHERE idlog=$1`, id)
    if err != nil {
        http.Error(w, err.Error(), http.StatusInternalServerError)
        return
    }

    rowsAffected := res.RowsAffected()
    if rowsAffected == 0 {
        http.Error(w, "Log not found", http.StatusNotFound)
        return
    }

    w.WriteHeader(http.StatusNoContent)
}
