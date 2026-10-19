package httpx

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
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

func GetIntQueryParam(r *http.Request, name string, defaultValue int, minValue int) (int, error) {
	raw := r.URL.Query().Get(name)

	if raw == "" {
		return defaultValue, nil
	}

	value, err := strconv.Atoi(raw)
	if err != nil || value < minValue {
		return 0, fmt.Errorf("invalid query parameter: %s", name)
	}

	return value, nil
}
