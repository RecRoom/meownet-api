package admin

import (
	"encoding/json"
	"net/http"
	"strconv"
	"time"

	"meow.net/controllers/hub"
	"meow.net/db"
	"meow.net/models"
)

func ListBans(w http.ResponseWriter, r *http.Request) {
	if !requireAdmin(w, r) {
		return
	}
	var bans []models.AccountBan
	db.DB.Find(&bans)
	writeJSON(w, http.StatusOK, bans)
}

func CreateBan(w http.ResponseWriter, r *http.Request) {
	if !requireAdmin(w, r) {
		return
	}
	var body struct {
		AccountID       uint   `json:"account_id"`
		Reason          string `json:"reason"`
		Message         string `json:"message"`
		IsBan           *bool  `json:"is_ban"`
		BannedBy        string `json:"banned_by"`
		DurationMinutes *int   `json:"duration_minutes"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil || body.AccountID == 0 {
		http.Error(w, "bad request", http.StatusBadRequest)
		return
	}
	isBan := body.IsBan == nil || *body.IsBan
	var expiresAt *time.Time
	if body.DurationMinutes != nil {
		t := time.Now().UTC().Add(time.Duration(*body.DurationMinutes) * time.Minute)
		expiresAt = &t
	}
	ban := models.AccountBan{
		AccountID: body.AccountID,
		Reason:    body.Reason,
		Message:   body.Message,
		IsBan:     isBan,
		BannedBy:  body.BannedBy,
		ExpiresAt: expiresAt,
	}
	if err := applyAccountBan(&ban); err != nil {
		http.Error(w, "failed to create ban", http.StatusInternalServerError)
		return
	}
	writeJSON(w, http.StatusCreated, ban)
}

func applyAccountBan(ban *models.AccountBan) error {
	if err := db.DB.Create(ban).Error; err != nil {
		return err
	}
	if ban.IsBan {
		var logins []models.DeviceLogin
		db.DB.Where("account_id = ? AND device_id != ''", ban.AccountID).Find(&logins)
		for _, dl := range logins {
			deviceBan := models.DeviceBan{
				DeviceID:  dl.DeviceID,
				AccountID: ban.AccountID,
				BanID:     ban.ID,
				Reason:    ban.Reason,
				BannedBy:  ban.BannedBy,
				ExpiresAt: ban.ExpiresAt,
			}
			db.DB.Where(models.DeviceBan{DeviceID: dl.DeviceID}).FirstOrCreate(&deviceBan)
		}
	}
	hub.HubKickPlayer(int(ban.AccountID))
	return nil
}

func AnticheatBan(w http.ResponseWriter, r *http.Request) {
	if !requireAdmin(w, r) {
		return
	}
	var body struct {
		UserID        uint   `json:"user_id"`
		AccountID     uint   `json:"account_id"`
		DurationHours *int   `json:"duration_hours"`
		Message       string `json:"message"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		http.Error(w, "bad request", http.StatusBadRequest)
		return
	}

	accountID := body.UserID
	if accountID == 0 {
		accountID = body.AccountID
	}
	if accountID == 0 {
		http.Error(w, "missing user_id", http.StatusBadRequest)
		return
	}

	var expiresAt *time.Time
	if body.DurationHours != nil && *body.DurationHours > 0 {
		t := time.Now().UTC().Add(time.Duration(*body.DurationHours) * time.Hour)
		expiresAt = &t
	}

	ban := models.AccountBan{
		AccountID: accountID,
		Reason:    "anticheat",
		Message:   body.Message,
		IsBan:     true,
		BannedBy:  "anticheat",
		ExpiresAt: expiresAt,
	}
	if err := applyAccountBan(&ban); err != nil {
		http.Error(w, "failed to create ban", http.StatusInternalServerError)
		return
	}
	writeJSON(w, http.StatusCreated, ban)
}

func DeleteBan(w http.ResponseWriter, r *http.Request) {
	if !requireAdmin(w, r) {
		return
	}
	id, err := strconv.ParseUint(r.PathValue("account_id"), 10, 64)
	if err != nil || id == 0 {
		http.Error(w, "bad request", http.StatusBadRequest)
		return
	}
	result := db.DB.Where("account_id = ?", id).Delete(&models.AccountBan{})
	if result.RowsAffected == 0 {
		http.Error(w, "not found", http.StatusNotFound)
		return
	}

	db.DB.Where("account_id = ?", id).Delete(&models.DeviceBan{})

	w.WriteHeader(http.StatusNoContent)
}
