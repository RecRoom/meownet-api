package moderation

import (
	"encoding/json"
	"io"
	"net/http"
	"strings"

	"meow.net/utils"
)

func SanitizeIsPure(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	value := extractSanitizeValue(r)
	isPure := !utils.IsTextFlagged(value)

	json.NewEncoder(w).Encode(map[string]interface{}{
		"isPure": isPure,
	})
}

func extractSanitizeValue(r *http.Request) string {
	if r.Body == nil {
		return ""
	}
	body, _ := io.ReadAll(r.Body)
	if len(body) == 0 {
		return ""
	}

	var req struct {
		Value           string `json:"Value"`
		ReplacementChar int    `json:"ReplacementChar"`
	}
	if err := json.Unmarshal(body, &req); err == nil && req.Value != "" {
		return req.Value
	}

	raw := strings.TrimSpace(string(body))
	if len(raw) >= 2 && raw[0] == '"' && raw[len(raw)-1] == '"' {
		var s string
		if err := json.Unmarshal([]byte(raw), &s); err == nil {
			return s
		}
	}
	return raw
}
