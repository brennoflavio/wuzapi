package main

import (
	"encoding/json"
	"net/http"
	"strconv"
	"time"

	"github.com/gorilla/mux"
)

type Event struct {
	ID        int       `json:"id" db:"id"`
	UserID    string    `json:"user_id" db:"user_id"`
	EventType string    `json:"event_type" db:"event_type"`
	Payload   string    `json:"payload" db:"payload"`
	CreatedAt time.Time `json:"created_at" db:"created_at"`
}

func (s *server) GetEvents() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		limit := 10
		if limitParam := r.URL.Query().Get("limit"); limitParam != "" {
			if parsed, err := strconv.Atoi(limitParam); err == nil && parsed > 0 {
				limit = parsed
			}
		}

		var events []Event
		err := s.db.Select(&events, "SELECT id, user_id, event_type, payload, created_at FROM events ORDER BY created_at ASC LIMIT ?", limit)
		if err != nil {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusInternalServerError)
			json.NewEncoder(w).Encode(map[string]interface{}{
				"success": false,
				"error":   err.Error(),
			})
			return
		}

		if events == nil {
			events = []Event{}
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"success": true,
			"events":  events,
		})
	}
}

func (s *server) DeleteEvent() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		vars := mux.Vars(r)
		id := vars["id"]

		s.db.Exec("DELETE FROM events WHERE id = ?", id)

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"success": true,
		})
	}
}
