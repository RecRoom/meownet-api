package admin

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"strings"
	"time"

	"meow.net/controllers/hub"
	"meow.net/db"
	"meow.net/models"
)

type giftBody struct {
	AvatarItemDesc            string `json:"AvatarItemDesc"`
	AvatarItemType            int    `json:"AvatarItemType"`
	ConsumableItemDesc        string `json:"ConsumableItemDesc"`
	EquipmentPrefabName       string `json:"EquipmentPrefabName"`
	EquipmentModificationGuid string `json:"EquipmentModificationGuid"`
	Currency                  int    `json:"Currency"`
	CurrencyType              int    `json:"CurrencyType"`
	Level                     int    `json:"Level"`
	Xp                        int    `json:"Xp"`
	Message                   string `json:"Message"`
	GiftContext               int    `json:"GiftContext"`
	GiftRarity                int    `json:"GiftRarity"`
}

func (b giftBody) message() string {
	if b.Message == "" {
		return "A gift from the meow.net team! Enjoy :3"
	}
	return b.Message
}

func buildGift(accountID uint, b giftBody) models.Gift {
	return models.Gift{
		AccountID:                 accountID,
		FromPlayerId:              1,
		AvatarItemDesc:            b.AvatarItemDesc,
		AvatarItemType:            b.AvatarItemType,
		ConsumableItemDesc:        b.ConsumableItemDesc,
		EquipmentPrefabName:       b.EquipmentPrefabName,
		EquipmentModificationGuid: b.EquipmentModificationGuid,
		Currency:                  b.Currency,
		CurrencyType:              b.CurrencyType,
		Level:                     b.Level,
		Xp:                        b.Xp,
		Message:                   b.message(),
		BalanceType:               -2,
		GiftContext:               b.GiftContext,
		GiftRarity:                b.GiftRarity,
		CreatedAt:                 time.Now().UTC(),
	}
}

func giftNotifFrame(g models.Gift) []byte {
	return hub.NotifFrame(
		int(models.GiftPackageReceivedImmediate),
		map[string]any{
			"Id":                        g.ID,
			"FromGiftDropId":            0,
			"FromPlayerId":              g.FromPlayerId,
			"ConsumableItemDesc":        g.ConsumableItemDesc,
			"AvatarItemDesc":            g.AvatarItemDesc,
			"AvatarItemType":            g.AvatarItemType,
			"EquipmentPrefabName":       g.EquipmentPrefabName,
			"EquipmentModificationGuid": g.EquipmentModificationGuid,
			"CurrencyType":              g.CurrencyType,
			"Currency":                  g.Currency,
			"Xp":                        g.Xp,
			"Level":                     g.Level,
			"Platform":                  -1,
			"PlatformsToSpawnOn":        -1,
			"BalanceType":               g.BalanceType,
			"GiftContext":               g.GiftContext,
			"GiftRarity":                g.GiftRarity,
			"Message":                   g.Message,
		},
	)
}

func resolveAccountIDs(parts []string) ([]int, error) {
	ids := make([]int, 0, len(parts))
	for _, p := range parts {
		p = strings.TrimSpace(p)
		if p == "" {
			continue
		}
		id, err := strconv.Atoi(p)
		if err != nil {
			var acc models.Account
			if err := db.DB.Select("account_id").Where("username = ?", strings.ToLower(p)).First(&acc).Error; err != nil {
				return nil, fmt.Errorf("account not found: %s", p)
			}
			id = int(acc.AccountID)
		}
		ids = append(ids, id)
	}
	return ids, nil
}

func Gift(w http.ResponseWriter, r *http.Request) {
	if !requireAdmin(w, r) {
		return
	}
	accountIDs, err := resolveAccountIDs(strings.Split(r.PathValue("id"), ","))
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}
	if len(accountIDs) == 0 {
		http.Error(w, "no accounts specified", http.StatusBadRequest)
		return
	}

	var body giftBody
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		http.Error(w, "invalid body", http.StatusBadRequest)
		return
	}

	gifts := make([]models.Gift, 0, len(accountIDs))
	for _, accountID := range accountIDs {
		gift := buildGift(uint(accountID), body)
		if err := db.DB.Create(&gift).Error; err != nil {
			log.Printf("[ADMIN] gift: %v", err)
			http.Error(w, "db error", http.StatusInternalServerError)
			return
		}
		hub.HubSendToPlayer(accountID, giftNotifFrame(gift))
		gifts = append(gifts, gift)
	}

	writeJSON(w, http.StatusCreated, gifts)
}

func GiftBulk(w http.ResponseWriter, r *http.Request) {
	if !requireAdmin(w, r) {
		return
	}

	var body struct {
		giftBody
		All        bool     `json:"All"`
		AccountIds []string `json:"AccountIds"` // ignored when all is true
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		http.Error(w, "invalid body", http.StatusBadRequest)
		return
	}

	var accountIDs []int
	if body.All {
		if err := db.DB.Model(&models.Account{}).Pluck("account_id", &accountIDs).Error; err != nil {
			log.Printf("[ADMIN] gift bulk: list accounts: %v", err)
			http.Error(w, "db error", http.StatusInternalServerError)
			return
		}
	} else {
		ids, err := resolveAccountIDs(body.AccountIds)
		if err != nil {
			http.Error(w, err.Error(), http.StatusNotFound)
			return
		}
		accountIDs = ids
	}
	if len(accountIDs) == 0 {
		http.Error(w, "no accounts specified", http.StatusBadRequest)
		return
	}

	go runBulkGift(accountIDs, body.giftBody)

	writeJSON(w, http.StatusAccepted, map[string]any{"queued": len(accountIDs)})
}

func runBulkGift(accountIDs []int, b giftBody) {
	gifts := make([]models.Gift, 0, len(accountIDs))
	for _, id := range accountIDs {
		gifts = append(gifts, buildGift(uint(id), b))
	}

	if err := db.DB.CreateInBatches(&gifts, 500).Error; err != nil {
		log.Printf("[ADMIN] gift bulk: insert: %v", err)
		return
	}

	online := make(map[int]bool)
	for _, id := range hub.GetOnlinePlayers() {
		online[id] = true
	}

	pushed := 0
	for i := range gifts {
		aid := int(gifts[i].AccountID)
		if !online[aid] {
			continue
		}
		hub.HubSendToPlayer(aid, giftNotifFrame(gifts[i]))
		pushed++
	}

	log.Printf("[ADMIN] gift bulk done: created=%d pushed=%d", len(gifts), pushed)
}
