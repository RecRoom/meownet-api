package controllers

import (
	"encoding/json"
	"log"
	"math"
	"net/http"
	"strconv"
	"time"

	"meow.net/controllers/hub"
	"meow.net/controllers/player"
	"meow.net/db"
	"meow.net/models"
)

// GET /api/avatar/v2/gifts
func GiftsList(w http.ResponseWriter, r *http.Request) {
	log.Printf("[GIFTS] list")
	w.Header().Set("Content-Type", "application/json")

	accountID, ok := AccountIDFromRequest(r)
	if !ok {
		w.Write([]byte("[]"))
		return
	}

	var gifts []models.Gift
	db.DB.Where("account_id = ? AND consumed = ?", accountID, false).
		Order("created_at asc").
		Find(&gifts)
	if gifts == nil {
		gifts = []models.Gift{}
	}
	for i := range gifts {
		if gifts[i].Currency > math.MaxInt32 {
			gifts[i].Currency = math.MaxInt32
		}
	}
	json.NewEncoder(w).Encode(gifts)
}

// POST /api/avatar/v2/gifts/consume/
func GiftsConsume(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		log.Printf("[GIFTS] consume: parse form: %v", err)
	}
	id, _ := strconv.Atoi(r.FormValue("Id"))
	unlockedLevel, _ := strconv.Atoi(r.FormValue("UnlockedLevel"))
	log.Printf("[GIFTS] consume id=%d unlockedLevel=%d", id, unlockedLevel)

	accountID, ok := AccountIDFromRequest(r)
	if !ok || id == 0 {
		w.WriteHeader(http.StatusOK)
		return
	}

	var gift models.Gift
	if err := db.DB.Where("id = ? AND account_id = ? AND consumed = ?", id, accountID, false).
		First(&gift).Error; err != nil {
		w.WriteHeader(http.StatusOK)
		return
	}

	ucID, ucDuration := grantGift(accountID, gift)

	gift.Consumed = true
	db.DB.Save(&gift)

	if ucID > 0 {
		var preExisting int64
		db.DB.Model(&models.UserConsumable{}).
			Where("account_id = ? AND consumable_item_desc = ? AND id != ?", accountID, gift.ConsumableItemDesc, ucID).
			Count(&preExisting)
		hub.HubSendToPlayer(int(accountID), hub.NotifFrame(int(models.ConsumableMappingAdded), map[string]any{
			"Id":                    ucID,
			"ConsumableItemDesc":    gift.ConsumableItemDesc,
			"CreatedAt":             time.Now().UTC().Format(time.RFC3339Nano),
			"Count":                 1,
			"InitialCount":          int(preExisting),
			"IsActive":              false,
			"ActiveDurationMinutes": ucDuration,
			"IsTransferable":        false,
		}))
	}

	if gift.Currency > 0 {
		bal := player.GetOrCreateBalance(accountID, gift.CurrencyType)
		hub.HubSendToPlayer(int(accountID), hub.NotifFrame(int(models.StorefrontBalanceUpdate), map[string]any{
			"Balance":      bal.Amount,
			"CurrencyType": gift.CurrencyType,
			"BalanceType":  bal.BalanceType,
		}))
	}

	w.WriteHeader(http.StatusOK)
}

func grantGift(accountID uint, g models.Gift) (uint, int) {
	if g.ConsumableItemDesc != "" {
		uc := models.UserConsumable{
			AccountID:             accountID,
			ConsumableItemDesc:    g.ConsumableItemDesc,
			ActiveDurationMinutes: 5,
			InitialCount:          1,
			CreatedAt:             time.Now().UTC(),
		}
		if err := db.DB.Create(&uc).Error; err != nil {
			log.Printf("[GIFTS] grant consumable error: %v", err)
			return 0, 0
		}
		return uc.ID, uc.ActiveDurationMinutes
	}

	if g.EquipmentPrefabName != "" && g.EquipmentModificationGuid != "" {
		equip := models.UserEquipment{
			AccountID:        accountID,
			ModificationGuid: g.EquipmentModificationGuid,
			PrefabName:       g.EquipmentPrefabName,
			Rarity:           g.GiftRarity,
		}
		db.DB.Where(models.UserEquipment{
			AccountID:        equip.AccountID,
			ModificationGuid: equip.ModificationGuid,
		}).FirstOrCreate(&equip)
		uc := models.UserConsumable{
			AccountID:             accountID,
			ActiveDurationMinutes: 500,
			InitialCount:          1,
			CreatedAt:             time.Now().UTC(),
		}
		if err := db.DB.Create(&uc).Error; err != nil {
			log.Printf("[GIFTS] grant equipment uc error: %v", err)
			return 0, 0
		}
		return uc.ID, uc.ActiveDurationMinutes
	}

	if g.AvatarItemDesc != "" {
		catalog := models.AvatarItem{
			AvatarItemDesc: g.AvatarItemDesc,
			AvatarItemType: g.AvatarItemType,
			Rarity:         g.GiftRarity,
		}
		db.DB.Where(models.AvatarItem{AvatarItemDesc: catalog.AvatarItemDesc}).
			FirstOrCreate(&catalog)

		owned := models.UserAvatarItem{
			AccountID:      accountID,
			AvatarItemDesc: g.AvatarItemDesc,
			CreatedAt:      time.Now().UTC(),
		}
		db.DB.Where(models.UserAvatarItem{
			AccountID:      owned.AccountID,
			AvatarItemDesc: owned.AvatarItemDesc,
		}).FirstOrCreate(&owned)
	}

	if g.Currency > 0 {
		bal := player.GetOrCreateBalance(accountID, g.CurrencyType)
		bal.Amount += g.Currency
		db.DB.Save(&bal)
	}
	return 0, 0
}
