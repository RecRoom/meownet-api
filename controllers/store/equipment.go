package store

import (
	"encoding/json"
	"log"
	"net/http"

	"meow.net/controllers"
	"meow.net/db"
	"meow.net/models"
)

type equipmentOut struct {
	Favorited        bool   `json:"Favorited"`
	FriendlyName     string `json:"FriendlyName"`
	ModificationGuid string `json:"ModificationGuid"`
	PrefabName       string `json:"PrefabName"`
	Rarity           int    `json:"Rarity"`
	Tooltip          string `json:"Tooltip"`
}

func Equipment(w http.ResponseWriter, r *http.Request) {
	log.Printf("[EQUIPMENT] getUnlocked")
	w.Header().Set("Content-Type", "application/json")

	accountID, ok := controllers.AccountIDFromRequest(r)
	if !ok {
		w.Write([]byte("[]"))
		return
	}

	var items []models.UserEquipment
	db.DB.Where("account_id = ?", accountID).Find(&items)

	out := make([]equipmentOut, len(items))
	for i, e := range items {
		out[i] = equipmentOut{
			Favorited:        e.Favorited,
			FriendlyName:     e.FriendlyName,
			ModificationGuid: e.ModificationGuid,
			PrefabName:       e.PrefabName,
			Rarity:           e.Rarity,
			Tooltip:          e.Tooltip,
		}
	}
	json.NewEncoder(w).Encode(out)
}

type equipmentUpdateEntry struct {
	Favorited        bool   `json:"Favorited"`
	ModificationGuid string `json:"ModificationGuid"`
	PrefabName       string `json:"PrefabName"`
}

func EquipmentUpdate(w http.ResponseWriter, r *http.Request) {
	log.Printf("[EQUIPMENT] update")

	accountID, ok := controllers.AccountIDFromRequest(r)
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	var entries []equipmentUpdateEntry
	if err := json.NewDecoder(r.Body).Decode(&entries); err != nil {
		http.Error(w, "bad request", http.StatusBadRequest)
		return
	}

	for _, e := range entries {
		db.DB.Model(&models.UserEquipment{}).
			Where("account_id = ? AND modification_guid = ?", accountID, e.ModificationGuid).
			Update("favorited", e.Favorited)
	}

	w.WriteHeader(http.StatusOK)
}
