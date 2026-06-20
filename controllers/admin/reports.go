package admin

import (
	"encoding/json"
	"log"
	"net/http"
	"time"

	"meow.net/controllers/hub"
	"meow.net/db"
	"meow.net/models"
)

const (
	reportActionJunior = "junior"
	reportActionBan    = "ban"
	reportActionWarn   = "warn"

	reportRewardTokens   = 500
	reportRewardCurrency = 2
)

func ActOnReport(w http.ResponseWriter, r *http.Request) {
	if !requireAdmin(w, r) {
		return
	}

	var body struct {
		TargetPlayerID uint       `json:"target_player_id"`
		ReporterID     uint       `json:"reporter_id"`
		Action         string     `json:"action"`
		ActedBy        string     `json:"acted_by"`
		BanExpiresAt   *time.Time `json:"ban_expires_at"`
		BanReason      string     `json:"ban_reason"`
		Warning        string     `json:"warning"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		http.Error(w, "invalid body", http.StatusBadRequest)
		return
	}
	if body.TargetPlayerID == 0 || body.ReporterID == 0 {
		http.Error(w, "missing target_player_id or reporter_id", http.StatusBadRequest)
		return
	}

	switch body.Action {
	case reportActionJunior:
		if err := db.DB.Model(&models.Account{}).
			Where("account_id = ?", body.TargetPlayerID).
			Update("treat_as_junior", true).Error; err != nil {
			http.Error(w, "failed to set junior", http.StatusInternalServerError)
			return
		}
		hub.HubKickPlayer(int(body.TargetPlayerID))

	case reportActionBan:
		ban := models.AccountBan{
			AccountID: body.TargetPlayerID,
			Reason:    body.BanReason,
			BannedBy:  body.ActedBy,
			ExpiresAt: body.BanExpiresAt,
		}
		if err := db.DB.Save(&ban).Error; err != nil {
			http.Error(w, "failed to create ban", http.StatusInternalServerError)
			return
		}
		hub.HubKickPlayer(int(body.TargetPlayerID))

	case reportActionWarn:
		if body.Warning == "" {
			http.Error(w, "missing warning", http.StatusBadRequest)
			return
		}
		warnMsg := models.Message{
			FromPlayerId: 1,
			ToPlayerId:   body.TargetPlayerID,
			Type:         int(models.MessageTypeCoachMessage),
			Data:         body.Warning,
		}
		if err := db.DB.Create(&warnMsg).Error; err != nil {
			http.Error(w, "failed to send warning", http.StatusInternalServerError)
			return
		}
		hub.HubSendToPlayer(int(body.TargetPlayerID), hub.NotifFrame(models.MessageReceived, warnMsg))

	default:
		http.Error(w, "invalid action", http.StatusBadRequest)
		return
	}

	thanksMsg := models.Message{
		FromPlayerId: 1,
		ToPlayerId:   body.ReporterID,
		Type:         int(models.MessageTypeCoachMessage),
		Data:         "Hello! We have acted on one of your reports, and we've sent you 500 tokens as a gift",
	}
	if err := db.DB.Create(&thanksMsg).Error; err != nil {
		log.Printf("[ADMIN] act-on-report: thanks msg: %v", err)
	} else {
		hub.HubSendToPlayer(int(body.ReporterID), hub.NotifFrame(models.MessageReceived, thanksMsg))
	}

	gift := models.Gift{
		AccountID:          body.ReporterID,
		FromPlayerId:       1,
		Message:            "Thanks for helping keep meow.net safe :3",
		Currency:           reportRewardTokens,
		CurrencyType:       reportRewardCurrency,
		BalanceType:        -2,
		GiftContext:        int(models.GiftContextToken),
		Platform:           -1,
		PlatformsToSpawnOn: -1,
		CreatedAt:          time.Now().UTC(),
	}
	if err := db.DB.Create(&gift).Error; err != nil {
		log.Printf("[ADMIN] act-on-report: gift create: %v", err)
	} else {
		hub.HubSendToPlayer(int(body.ReporterID), hub.NotifFrame(
			int(models.GiftPackageReceivedImmediate),
			map[string]any{
				"Id":                        gift.ID,
				"FromGiftDropId":            0,
				"FromPlayerId":              gift.FromPlayerId,
				"Message":                   gift.Message,
				"ConsumableItemDesc":        gift.ConsumableItemDesc,
				"AvatarItemDesc":            gift.AvatarItemDesc,
				"AvatarItemType":            gift.AvatarItemType,
				"EquipmentPrefabName":       gift.EquipmentPrefabName,
				"EquipmentModificationGuid": gift.EquipmentModificationGuid,
				"Currency":                  gift.Currency,
				"CurrencyType":              gift.CurrencyType,
				"BalanceType":               gift.BalanceType,
				"Level":                     gift.Level,
				"Xp":                        gift.Xp,
				"GiftContext":               gift.GiftContext,
				"GiftRarity":                gift.GiftRarity,
				"Platform":                  gift.Platform,
				"PlatformsToSpawnOn":        gift.PlatformsToSpawnOn,
			},
		))
	}

	writeJSON(w, http.StatusOK, map[string]any{
		"success": true,
		"action":  body.Action,
	})
}
