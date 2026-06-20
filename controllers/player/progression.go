package player

import (
	"encoding/json"
	"net/http"
	"strconv"

	"meow.net/controllers/hub"
	"meow.net/db"
	"meow.net/models"
	"meow.net/utils"
)

func tokenAccountID(r *http.Request) (int64, bool) {
	token := utils.GetBearerToken(r)
	if token == "" {
		return 0, false
	}
	sub, err := utils.ParseSubFromJWT(token)
	if err != nil || sub == "" {
		return 0, false
	}
	id, err := strconv.ParseInt(sub, 10, 64)
	if err != nil {
		return 0, false
	}
	return id, true
}

func PlayerProgression(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	authId, ok := tokenAccountID(r)
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}
	pathId := utils.GetAccountIDFromPath(r)
	if pathId != authId {
		http.Error(w, "Forbidden", http.StatusForbidden)
		return
	}

	var prog models.Progression
	db.DB.Where(models.Progression{AccountID: uint(authId)}).
		Attrs(models.Progression{Level: 1, XP: 0}).
		FirstOrCreate(&prog)

	hub.HubSendProgressionUpdate(int(authId), prog.Level, prog.XP)

	json.NewEncoder(w).Encode(map[string]interface{}{
		"PlayerId": authId,
		"Level":    prog.Level,
		"XP":       prog.XP,
	})
}

func PlayerProgressionBulk(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	if _, ok := tokenAccountID(r); !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	ids := r.URL.Query()["id"]
	results := make([]map[string]interface{}, 0)
	for _, idStr := range ids {
		id, err := strconv.Atoi(idStr)
		if err != nil {
			continue
		}
		var prog models.Progression
		if db.DB.Where(models.Progression{AccountID: uint(id)}).First(&prog).Error != nil {
			prog = models.Progression{Level: 1, XP: 0}
		}
		results = append(results, map[string]interface{}{
			"PlayerId": id,
			"Level":    prog.Level,
			"XP":       prog.XP,
		})
	}
	json.NewEncoder(w).Encode(results)
}
