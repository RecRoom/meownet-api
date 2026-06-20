package utils

import (
	"encoding/json"
	"net/http"
	"strconv"
	"strings"
)

func JsonHandler(data interface{}) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(data)
	}
}

func QueryString(r *http.Request) string {
	if r.URL.RawQuery != "" {
		return "?" + r.URL.RawQuery
	}
	return ""
}

func GetAccountIDFromPath(r *http.Request) int64 {
	pathParts := strings.Split(r.URL.Path, "/")
	accountIdStr := pathParts[len(pathParts)-1]
	accountId, err := strconv.ParseInt(accountIdStr, 10, 64)
	if err != nil {
		return 0
	}
	return accountId
}

func GetBearerToken(r *http.Request) string {
	authHeader := r.Header.Get("Authorization")
	if len(authHeader) > 7 && strings.EqualFold(authHeader[0:7], "bearer ") {
		return strings.TrimSpace(authHeader[7:])
	}
	return ""
}
