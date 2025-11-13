package models

import "time"

type Log struct {
	Id       int       `json:"id"`
	UserID   int       `json:"user_id"`
	Action   string    `json:"action"`
	Entity   *string   `json:"entity,omitempty"`
	EntityID *int      `json:"entity_id,omitempty"`
	Details  *string   `json:"details,omitempty"`
	CreatedAt time.Time `json:"created_at"`
}
