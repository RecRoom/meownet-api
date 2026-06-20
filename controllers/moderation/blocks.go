package moderation

import (
	"encoding/json"
	"net/http"
	"time"

	"meow.net/controllers"
	"meow.net/db"
	"meow.net/models"
)

func emptyBlockDetails() map[string]interface{} {
	return map[string]interface{}{
		"Duration":         0,
		"GameSessionId":    0,
		"IsBan":            false,
		"IsHostKick":       false,
		"Message":          nil,
		"PlayerIdReporter": nil,
		"ReportCategory":   0,
	}
}

func blockFromBan(ban models.AccountBan) models.ModerationBlock {
	message := ban.Message
	if message == "" {
		message = ban.Reason
	}

	duration := 0
	if ban.ExpiresAt != nil {
		if remaining := int(time.Until(*ban.ExpiresAt).Seconds()); remaining > 0 {
			duration = remaining
		}
	}

	return models.ModerationBlock{
		IsBan:          ban.IsBan,
		IsHostKick:     false,
		Message:        &message,
		ReportCategory: int(models.ReportCategoryModerator),
		Duration:       duration,
		ExpiresAt:      ban.ExpiresAt,
	}
}

func BlockDetails(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	currentUserID, err := controllers.CurrentUserIDFromRequest(r)
	if err != nil {
		json.NewEncoder(w).Encode(emptyBlockDetails())
		return
	}

	now := time.Now()

	var ban models.AccountBan
	err = db.DB.Where("account_id = ? AND (expires_at IS NULL OR expires_at > ?)", currentUserID, now).
		Order("created_at DESC").First(&ban).Error
	if err == nil {
		json.NewEncoder(w).Encode(blockFromBan(ban))
		return
	}

	var block models.ModerationBlock
	err = db.DB.Where("account_id = ?", currentUserID).
		Where("expires_at IS NULL OR expires_at > ?", now).
		Order("created_at DESC").
		First(&block).Error
	if err != nil {
		json.NewEncoder(w).Encode(emptyBlockDetails())
		return
	}

	json.NewEncoder(w).Encode(block)
}
