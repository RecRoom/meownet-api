package player

import (
	"encoding/json"
	"log"
	"net/http"
	"strconv"

	"meow.net/controllers/hub"
	"meow.net/db"
	"meow.net/models"
)

const consumableTimeFormat = "2006-01-02T15:04:05.0000000"

type consumableGroup struct {
	ActiveDurationMinutes int      `json:"ActiveDurationMinutes"`
	ConsumableItemDesc    string   `json:"ConsumableItemDesc"`
	Count                 int      `json:"Count"`
	CreatedAts            []string `json:"CreatedAts"`
	Ids                   []uint   `json:"Ids"`
	InitialCount          int      `json:"InitialCount"`
	IsActive              bool     `json:"IsActive"`
	IsTransferable        bool     `json:"IsTransferable"`
}

func Consumables(w http.ResponseWriter, r *http.Request) {
	log.Printf("[CONSUMABLES] getUnlocked")
	w.Header().Set("Content-Type", "application/json")

	accountID, ok := accountIDFromRequest(r)
	if !ok {
		w.Write([]byte("[]"))
		return
	}

	var rows []models.UserConsumable
	if err := db.DB.
		Where("account_id = ?", accountID).
		Order("created_at asc").
		Find(&rows).Error; err != nil {
		log.Printf("[CONSUMABLES] query error: %v", err)
		w.Write([]byte("[]"))
		return
	}

	out := groupConsumables(rows)
	json.NewEncoder(w).Encode(out)
}

func groupConsumables(rows []models.UserConsumable) []consumableGroup {
	idx := map[string]int{}
	out := []consumableGroup{}
	for _, c := range rows {
		key := c.ConsumableItemDesc + "|" + strconv.Itoa(c.ActiveDurationMinutes)
		ts := c.CreatedAt.Format(consumableTimeFormat)
		if i, found := idx[key]; found {
			out[i].Count++
			out[i].Ids = append(out[i].Ids, c.ID)
			out[i].CreatedAts = append(out[i].CreatedAts, ts)
			continue
		}
		idx[key] = len(out)
		out = append(out, consumableGroup{
			ActiveDurationMinutes: c.ActiveDurationMinutes,
			ConsumableItemDesc:    c.ConsumableItemDesc,
			Count:                 1,
			CreatedAts:            []string{ts},
			Ids:                   []uint{c.ID},
			InitialCount:          c.InitialCount,
			IsActive:              c.IsActive,
			IsTransferable:        c.IsTransferable,
		})
	}
	return out
}

func ConsumableConsume(w http.ResponseWriter, r *http.Request) {
	log.Printf("[CONSUMABLES] v1/consume")
	w.Header().Set("Content-Type", "application/json")

	accountID, ok := accountIDFromRequest(r)
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	var body struct {
		Id         uint `json:"Id"`
		DeltaCount int  `json:"DeltaCount"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil || body.Id == 0 {
		w.WriteHeader(http.StatusOK)
		return
	}
	if body.DeltaCount <= 0 {
		body.DeltaCount = 1
	}

	var target models.UserConsumable
	if err := db.DB.
		Where("id = ? AND account_id = ?", body.Id, accountID).
		First(&target).Error; err != nil {
		log.Printf("[CONSUMABLES] consume: %d not owned by %d (%v)", body.Id, accountID, err)
		w.WriteHeader(http.StatusOK)
		return
	}

	newInitialCount := 0
	if target.InitialCount > body.DeltaCount {
		newInitialCount = target.InitialCount - body.DeltaCount
		db.DB.Model(&target).Update("initial_count", newInitialCount)
	} else {
		db.DB.Delete(&target)
		remainder := body.DeltaCount - target.InitialCount
		if remainder > 0 {
			var siblings []models.UserConsumable
			db.DB.Where(
				"account_id = ? AND consumable_item_desc = ? AND active_duration_minutes = ? AND id != ?",
				accountID, target.ConsumableItemDesc, target.ActiveDurationMinutes, target.ID,
			).
				Order("created_at asc").
				Find(&siblings)

			toDelete := []uint{}
			for _, s := range siblings {
				if remainder <= 0 {
					break
				}
				if s.InitialCount <= remainder {
					toDelete = append(toDelete, s.ID)
					remainder -= s.InitialCount
				} else {
					db.DB.Model(&s).Update("initial_count", s.InitialCount-remainder)
					remainder = 0
				}
			}
			if len(toDelete) > 0 {
				db.DB.Where("id IN ?", toDelete).Delete(&models.UserConsumable{})
			}
		}
	}

	var remaining int64
	db.DB.Model(&models.UserConsumable{}).
		Where("account_id = ? AND consumable_item_desc = ? AND active_duration_minutes = ?",
			accountID, target.ConsumableItemDesc, target.ActiveDurationMinutes).
		Count(&remaining)

	hub.HubSendToPlayer(int(accountID), hub.NotifFrame(int(models.ConsumableMappingRemoved), map[string]any{
		"Id":                    target.ID,
		"ConsumableItemDesc":    target.ConsumableItemDesc,
		"CreatedAt":             target.CreatedAt.Format(consumableTimeFormat),
		"Count":                 int(remaining),
		"InitialCount":          newInitialCount,
		"IsActive":              target.IsActive,
		"ActiveDurationMinutes": target.ActiveDurationMinutes,
		"IsTransferable":        target.IsTransferable,
	}))

	w.WriteHeader(http.StatusOK)
}
