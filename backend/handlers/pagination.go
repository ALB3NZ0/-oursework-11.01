package handlers

import (
	"net/http"
	"strconv"
)

// PaginationParams содержит параметры пагинации
type PaginationParams struct {
	Page  int
	Limit int
}

// PaginatedResponse содержит данные и метаинформацию о пагинации
type PaginatedResponse struct {
	Data       interface{} `json:"data"`
	Page       int         `json:"page"`
	Limit      int         `json:"limit"`
	Total      int         `json:"total"`
	TotalPages int         `json:"total_pages"`
}

// ParsePaginationParams извлекает параметры пагинации из запроса
// Возвращает page (по умолчанию 1) и limit (по умолчанию 20)
func ParsePaginationParams(r *http.Request) PaginationParams {
	page := 1
	limit := 20

	if pageStr := r.URL.Query().Get("page"); pageStr != "" {
		if p, err := strconv.Atoi(pageStr); err == nil && p > 0 {
			page = p
		}
	}

	if limitStr := r.URL.Query().Get("limit"); limitStr != "" {
		if l, err := strconv.Atoi(limitStr); err == nil && l > 0 {
			// Максимальный лимит 100
			if l > 100 {
				l = 100
			}
			limit = l
		}
	}

	return PaginationParams{
		Page:  page,
		Limit: limit,
	}
}

// GetOffset вычисляет offset для SQL запроса
func (p PaginationParams) GetOffset() int {
	return (p.Page - 1) * p.Limit
}

// CalculateTotalPages вычисляет общее количество страниц
func CalculateTotalPages(total int, limit int) int {
	if limit <= 0 {
		return 1
	}
	pages := total / limit
	if total%limit > 0 {
		pages++
	}
	return pages
}














