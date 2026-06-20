package store

import (
	"encoding/json"
	"log"
	"net/http"
	"time"

	"meow.net/controllers"
	"meow.net/controllers/hub"
	"meow.net/controllers/player"
	"meow.net/db"
	"meow.net/models"
)

type giftRequest struct {
	ToPlayerId  int    `json:"ToPlayerId"`
	Message     string `json:"Message"`
	GiftContext int    `json:"GiftContext"`
	Anonymous   bool   `json:"Anonymous"`
}

type buyItemRequest struct {
	PurchasableItemId               int          `json:"PurchasableItemId"`
	CurrencyType                    int          `json:"CurrencyType"`
	RequestedPrice                  int          `json:"RequestedPrice"`
	StorefrontType                  int          `json:"StorefrontType"`
	CouponConsumablePlayerMappingId *int         `json:"CouponConsumablePlayerMappingId"`
	Gift                            *giftRequest `json:"Gift"`
}

type buyItemDataEntry struct {
	AvatarItemDesc            string `json:"AvatarItemDesc"`
	AvatarItemType            int    `json:"AvatarItemType"`
	BalanceType               int    `json:"BalanceType"`
	ConsumableItemDesc        string `json:"ConsumableItemDesc"`
	Currency                  int    `json:"Currency"`
	CurrencyType              int    `json:"CurrencyType"`
	EquipmentModificationGuid string `json:"EquipmentModificationGuid"`
	EquipmentPrefabName       string `json:"EquipmentPrefabName"`
	FromPlayerId              int    `json:"FromPlayerId"`
	GiftContext               int    `json:"GiftContext"`
	GiftRarity                int    `json:"GiftRarity"`
	Id                        uint   `json:"Id"`
	Level                     int    `json:"Level"`
	Message                   string `json:"Message"`
	Platform                  int    `json:"Platform"`
	PlatformsToSpawnOn        int    `json:"PlatformsToSpawnOn"`
	Xp                        int    `json:"Xp"`
}

type buyItemBalanceUpdate struct {
	Data           []buyItemDataEntry `json:"Data"`
	UpdateResponse int                `json:"UpdateResponse"`
}

type buyItemResponse struct {
	Balance        int                    `json:"Balance"`
	BalanceType    int                    `json:"BalanceType"`
	BalanceUpdates []buyItemBalanceUpdate `json:"BalanceUpdates"`
	CurrencyType   int                    `json:"CurrencyType"`
}

func BuyItem(w http.ResponseWriter, r *http.Request) {
	log.Printf("[STOREFRONTS] buyItem")
	w.Header().Set("Content-Type", "application/json")

	accountID, ok := controllers.AccountIDFromRequest(r)
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	var req buyItemRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "bad request", http.StatusBadRequest)
		return
	}

	item, _, ok := LookupStorefrontItem(req.PurchasableItemId, req.StorefrontType)
	if !ok {
		log.Printf("[BUYITEM] item %d not found", req.PurchasableItemId)
		http.Error(w, "item not found", http.StatusNotFound)
		return
	}

	reg, regOK := priceFor(item.Prices, req.CurrencyType)
	sub, subOK := priceFor(item.SubscriberPrices, req.CurrencyType)
	if !regOK && !subOK {
		log.Printf("[BUYITEM] item %d has no price for currency %d", req.PurchasableItemId, req.CurrencyType)
		http.Error(w, "currency not accepted", http.StatusBadRequest)
		return
	}

	drop := firstDrop(item)

	if req.Gift != nil && recipientAlreadyOwns(uint(req.Gift.ToPlayerId), drop) {
		log.Printf("[BUYITEM] recipient %d already owns gifted item", req.Gift.ToPlayerId)
		http.Error(w, "recipient already owns this item", http.StatusConflict)
		return
	}

	if req.Gift == nil && recipientAlreadyOwns(accountID, drop) {
		log.Printf("[BUYITEM] account %d already owns item", accountID)
		http.Error(w, "already owned", http.StatusConflict)
		return
	}

	price := reg
	if subOK && req.RequestedPrice == sub {
		price = sub
	} else if !regOK {
		price = sub
	}
	if req.RequestedPrice != 0 && req.RequestedPrice != price {
		log.Printf("[BUYITEM] ignoring requested price %d, charging %d (regular=%d sub=%d)", req.RequestedPrice, price, reg, sub)
	}

	bal := player.GetOrCreateBalance(accountID, req.CurrencyType)

	if bal.Amount < price {
		http.Error(w, "insufficient funds", http.StatusPaymentRequired)
		return
	}
	bal.Amount -= price
	db.DB.Save(&bal)

	var toAccountID uint
	var fromPlayerID uint
	var message string
	var giftContext int

	if req.Gift != nil {
		toAccountID = uint(req.Gift.ToPlayerId)
		if req.Gift.Anonymous {
			fromPlayerID = 0
		} else {
			fromPlayerID = accountID
		}
		message = req.Gift.Message
		giftContext = req.Gift.GiftContext
	} else {
		toAccountID = accountID
		fromPlayerID = 1
		message = "A gift for you <3"
		giftContext = drop.Context
	}

	avatarType := 0
	if drop.AvatarItemType != nil {
		avatarType = *drop.AvatarItemType
	}

	gift := models.Gift{
		AccountID:                 toAccountID,
		FromPlayerId:              fromPlayerID,
		Message:                   message,
		AvatarItemDesc:            drop.AvatarItemDesc,
		AvatarItemType:            avatarType,
		ConsumableItemDesc:        drop.ConsumableItemDesc,
		EquipmentPrefabName:       drop.EquipmentPrefabName,
		EquipmentModificationGuid: drop.EquipmentModificationGuid,
		Currency:                  drop.Currency,
		CurrencyType:              drop.CurrencyType,
		BalanceType:               -2,
		Level:                     drop.Level,
		GiftContext:               giftContext,
		GiftRarity:                drop.Rarity,
		Platform:                  -1,
		PlatformsToSpawnOn:        -1,
	}
	if err := db.DB.Create(&gift).Error; err != nil {
		log.Printf("[BUYITEM] gift create error: %v", err)
	}

	if req.Gift != nil {
		hub.HubSendToPlayer(int(toAccountID), hub.NotifFrame(int(models.GiftPackageReceivedImmediate), map[string]any{
			"Id":                        gift.ID,
			"FromGiftDropId":            0,
			"FromPlayerId":              gift.FromPlayerId,
			"Message":                   gift.Message,
			"AvatarItemDesc":            gift.AvatarItemDesc,
			"AvatarItemType":            gift.AvatarItemType,
			"ConsumableItemDesc":        gift.ConsumableItemDesc,
			"EquipmentPrefabName":       gift.EquipmentPrefabName,
			"EquipmentModificationGuid": gift.EquipmentModificationGuid,
			"Currency":                  gift.Currency,
			"CurrencyType":              gift.CurrencyType,
			"BalanceType":               gift.BalanceType,
			"Level":                     gift.Level,
			"GiftContext":               gift.GiftContext,
			"GiftRarity":                gift.GiftRarity,
			"Platform":                  gift.Platform,
			"PlatformsToSpawnOn":        gift.PlatformsToSpawnOn,
			"Xp":                        gift.Xp,
		}))
	}

	resp := buyItemResponse{
		Balance:      bal.Amount,
		BalanceType:  bal.BalanceType,
		CurrencyType: req.CurrencyType,
		BalanceUpdates: []buyItemBalanceUpdate{{
			UpdateResponse: 0,
			Data: []buyItemDataEntry{{
				AvatarItemDesc:            drop.AvatarItemDesc,
				AvatarItemType:            avatarType,
				BalanceType:               bal.BalanceType,
				ConsumableItemDesc:        drop.ConsumableItemDesc,
				Currency:                  drop.Currency,
				CurrencyType:              drop.CurrencyType,
				EquipmentModificationGuid: drop.EquipmentModificationGuid,
				EquipmentPrefabName:       drop.EquipmentPrefabName,
				FromPlayerId:              int(fromPlayerID),
				GiftContext:               drop.Context,
				GiftRarity:                drop.Rarity,
				Id:                        gift.ID,
				Level:                     drop.Level,
				Message:                   message,
				Platform:                  -1,
				PlatformsToSpawnOn:        -1,
				Xp:                        0,
			}},
		}},
	}
	json.NewEncoder(w).Encode(resp)
}

func priceFor(prices []StorefrontPrice, currencyType int) (int, bool) {
	for _, p := range prices {
		if p.CurrencyType == currencyType {
			return effectivePrice(p), true
		}
	}
	return 0, false
}

func effectivePrice(p StorefrontPrice) int {
	sale := p.StorefrontSaleData
	if sale == nil || sale.SalePercent <= 0 || sale.SalePercent >= 100 {
		return p.Price
	}
	start, err1 := time.Parse(time.RFC3339, sale.SaleStartDate)
	end, err2 := time.Parse(time.RFC3339, sale.SaleEndDate)
	if err1 != nil || err2 != nil {
		return p.Price
	}
	now := time.Now().UTC()
	if now.Before(start) || now.After(end) {
		return p.Price
	}
	return p.Price * (100 - sale.SalePercent) / 100
}

func recipientAlreadyOwns(accountID uint, drop StorefrontDrop) bool {
	if accountID == 0 {
		return false
	}
	if drop.AvatarItemDesc != "" {
		var count int64
		db.DB.Model(&models.UserAvatarItem{}).
			Where("account_id = ? AND avatar_item_desc = ?", accountID, drop.AvatarItemDesc).
			Count(&count)
		if count > 0 {
			return true
		}
	}
	if drop.EquipmentModificationGuid != "" {
		var count int64
		db.DB.Model(&models.UserEquipment{}).
			Where("account_id = ? AND modification_guid = ?", accountID, drop.EquipmentModificationGuid).
			Count(&count)
		if count > 0 {
			return true
		}
	}
	return false
}

func firstDrop(item StorefrontItem) StorefrontDrop {
	if item.GiftDrop != nil {
		return *item.GiftDrop
	}
	if len(item.GiftDrops) > 0 {
		return item.GiftDrops[0]
	}
	return StorefrontDrop{}
}
