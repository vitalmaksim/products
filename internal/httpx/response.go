package httpx

import (
	"encoding/json"
	"net/http"
)

func JSON(w http.ResponseWriter, status int, data any) {
	body, err := json.Marshal(data)
	if err != nil {
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_, _ = w.Write(body)
}

func Error(w http.ResponseWriter, status int, message string) {
	JSON(w, status, map[string]string{
		"error": message,
	})
}
