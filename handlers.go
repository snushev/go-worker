package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

func (cfg *dbConfig) HandlerJobsList(w http.ResponseWriter, r *http.Request) {
	query := `
		SELECT id, payload, status, created_at
		FROM jobs
		WHERE status IN ('pending', 'failed')
		ORDER BY 
			CASE 
				WHEN status = 'failed' THEN 0
				ELSE 1
			END,
			created_at DESC
	`

	rows, err := cfg.DB.Query(r.Context(), query)
	if err != nil {
		http.Error(w, "Database error", http.StatusInternalServerError)
		fmt.Printf("DB error: %v\n", err)
		return
	}
	defer rows.Close()

	type JobResponse struct {
		ID        int             `json:"id"`
		Payload   json.RawMessage `json:"payload"`
		Status    string          `json:"status"`
		CreatedAt time.Time       `json:"created_at"`
	}

	var jobs []JobResponse

	for rows.Next() {
		var j JobResponse

		err := rows.Scan(&j.ID, &j.Payload, &j.Status, &j.CreatedAt)
		if err != nil {
			fmt.Printf("Scan error: %v\n", err)
			continue
		}

		jobs = append(jobs, j)
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(jobs)
}

func (cfg *dbConfig) HandlerJobCreate(w http.ResponseWriter, r *http.Request) {
	var payload JobPayload

	err := json.NewDecoder(r.Body).Decode(&payload)
	if err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}
	fmt.Printf("Created json: %v\n", payload)

	jsonData, err := json.Marshal(payload)
	if err != nil {
		http.Error(w, "Failed to encode JSON", http.StatusInternalServerError)
		return
	}

	query := `INSERT INTO jobs (payload) VALUES ($1)`
	_, err = cfg.DB.Exec(r.Context(), query, jsonData)
	if err != nil {
		http.Error(w, "Database error", http.StatusInternalServerError)
		fmt.Printf("Database error: %v\n", err)
		return
	}
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(payload)

}
