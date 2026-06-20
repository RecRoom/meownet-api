package account

import (
	"encoding/json"
	"net/http"
)

func writeModerationRejection(w http.ResponseWriter, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusBadRequest)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"error":   message,
		"success": false,
		"value":   nil,
	})
}
