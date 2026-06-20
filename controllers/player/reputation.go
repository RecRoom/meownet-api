package player

import (
	"encoding/json"
	"net/http"
	"strconv"

	"meow.net/controllers/reputation"
	"meow.net/utils"
)

func PlayerReputation(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	accountId := utils.GetAccountIDFromPath(r)
	json.NewEncoder(w).Encode(reputation.Build(uint(accountId)))
}

func PlayerReputationBulk(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	ids := r.URL.Query()["id"]
	results := make([]reputation.Reputation, 0, len(ids))
	for _, idStr := range ids {
		if id, err := strconv.Atoi(idStr); err == nil {
			results = append(results, reputation.Build(uint(id)))
		}
	}
	json.NewEncoder(w).Encode(results)
}
