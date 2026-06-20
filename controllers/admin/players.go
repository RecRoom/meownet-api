package admin

import (
	"encoding/json"
	"net/http"
	"strconv"
	"strings"

	"meow.net/controllers/hub"
	"meow.net/db"
	"meow.net/models"
	"meow.net/utils"
)

func parseAccountID(w http.ResponseWriter, r *http.Request) (int, bool) {
	id, err := strconv.Atoi(r.PathValue("id"))
	if err != nil || id <= 0 {
		http.Error(w, "bad account id", http.StatusBadRequest)
		return 0, false
	}
	return id, true
}

func GetAccountDetail(w http.ResponseWriter, r *http.Request) {
	if !requireAdmin(w, r) {
		return
	}
	accountID, ok := parseAccountID(w, r)
	if !ok {
		return
	}

	var account models.SelfAccount
	if err := db.DB.First(&account, accountID).Error; err != nil {
		http.Error(w, "account not found", http.StatusNotFound)
		return
	}

	var balances []models.Balance
	db.DB.Where("account_id = ?", accountID).Find(&balances)

	var bans []models.AccountBan
	db.DB.Where("account_id = ?", accountID).Find(&bans)

	var deviceLogins []models.DeviceLogin
	db.DB.Where("account_id = ?", accountID).Find(&deviceLogins)

	var prog models.Progression
	db.DB.Where(models.Progression{AccountID: uint(accountID)}).
		Attrs(models.Progression{Level: 1, XP: 0}).
		FirstOrCreate(&prog)

	writeJSON(w, http.StatusOK, map[string]any{
		"account":      account,
		"balances":     balances,
		"bans":         bans,
		"progression":  prog,
		"isOnline":     hub.HubIsOnline(accountID),
		"presence":     hub.BuildPresence(accountID),
		"deviceLogins": deviceLogins,
	})
}

func UpdateAccount(w http.ResponseWriter, r *http.Request) {
	if !requireAdmin(w, r) {
		return
	}
	accountID, ok := parseAccountID(w, r)
	if !ok {
		return
	}

	var body struct {
		DisplayName   *string `json:"display_name"`
		Username      *string `json:"username"`
		ProfileImage  *string `json:"profile_image"`
		IsJunior      *bool   `json:"is_junior"`
		TreatAsJunior *bool   `json:"treat_as_junior"`
		IsDeveloper   *bool   `json:"is_developer"`
		IsModerator   *bool   `json:"is_moderator"`
		Platforms     *int    `json:"platforms"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		http.Error(w, "invalid body", http.StatusBadRequest)
		return
	}

	var account models.Account
	if err := db.DB.First(&account, accountID).Error; err != nil {
		http.Error(w, "account not found", http.StatusNotFound)
		return
	}

	updates := map[string]any{}
	if body.DisplayName != nil {
		updates["display_name"] = *body.DisplayName
	}
	if body.Username != nil {
		raw := strings.TrimSpace(*body.Username)
		if raw == "" {
			http.Error(w, "username cannot be empty", http.StatusBadRequest)
			return
		}
		lower := strings.ToLower(raw)
		var existing models.Account
		if err := db.DB.Where("LOWER(username) = ?", lower).First(&existing).Error; err == nil && existing.AccountID != account.AccountID {
			http.Error(w, "username already taken", http.StatusConflict)
			return
		}
		updates["username"] = lower
		updates["raw_username"] = raw
	}
	if body.ProfileImage != nil {
		updates["profile_image"] = *body.ProfileImage
	}
	if body.IsJunior != nil {
		updates["is_junior"] = *body.IsJunior
	}
	if body.TreatAsJunior != nil {
		updates["treat_as_junior"] = *body.TreatAsJunior
	}
	if body.IsDeveloper != nil {
		updates["is_developer"] = *body.IsDeveloper
	}
	if body.IsModerator != nil {
		updates["is_moderator"] = *body.IsModerator
	}
	if body.Platforms != nil {
		updates["platforms"] = *body.Platforms
	}

	if len(updates) == 0 {
		http.Error(w, "no fields to update", http.StatusBadRequest)
		return
	}

	if err := db.DB.Model(&account).Updates(updates).Error; err != nil {
		http.Error(w, "db error", http.StatusInternalServerError)
		return
	}

	var selfAccount models.SelfAccount
	if db.DB.First(&selfAccount, accountID).Error == nil {
		hub.HubSendToPlayer(accountID, hub.NotifFrame("SelfAccountUpdate", selfAccount))
	}
	hub.HubBroadcastAccountUpdate(accountID)

	writeJSON(w, http.StatusOK, selfAccount)
}

func SetProgression(w http.ResponseWriter, r *http.Request) {
	if !requireAdmin(w, r) {
		return
	}
	accountID, ok := parseAccountID(w, r)
	if !ok {
		return
	}

	var body struct {
		Level      *int `json:"level"`
		XP         *int `json:"xp"`
		LevelDelta int  `json:"level_delta"`
		XPDelta    int  `json:"xp_delta"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		http.Error(w, "invalid body", http.StatusBadRequest)
		return
	}

	var prog models.Progression
	db.DB.Where(models.Progression{AccountID: uint(accountID)}).
		Attrs(models.Progression{Level: 1, XP: 0}).
		FirstOrCreate(&prog)

	if body.Level != nil {
		prog.Level = *body.Level
	}
	if body.XP != nil {
		prog.XP = *body.XP
	}
	prog.Level += body.LevelDelta
	prog.XP += body.XPDelta

	if prog.Level < 1 {
		prog.Level = 1
	}
	if prog.XP < 0 {
		prog.XP = 0
	}

	if err := db.DB.Save(&prog).Error; err != nil {
		http.Error(w, "db error", http.StatusInternalServerError)
		return
	}

	hub.HubSendProgressionUpdate(accountID, prog.Level, prog.XP)
	writeJSON(w, http.StatusOK, prog)
}

func KickPlayer(w http.ResponseWriter, r *http.Request) {
	if !requireAdmin(w, r) {
		return
	}
	accountID, ok := parseAccountID(w, r)
	if !ok {
		return
	}

	utils.RevokeAccessTokens(strconv.Itoa(accountID))

	res := db.DB.Where("account_id = ?", accountID).Delete(&models.RefreshToken{})
	if res.Error != nil {
		http.Error(w, "db error", http.StatusInternalServerError)
		return
	}

	wasOnline := hub.HubIsOnline(accountID)
	hub.HubKickPlayer(accountID)

	writeJSON(w, http.StatusOK, map[string]any{
		"success":                true,
		"was_online":             wasOnline,
		"refresh_tokens_deleted": res.RowsAffected,
	})
}

func RefreshPresence(w http.ResponseWriter, r *http.Request) {
	if !requireAdmin(w, r) {
		return
	}
	accountID, ok := parseAccountID(w, r)
	if !ok {
		return
	}

	hub.HubBroadcastPresence(accountID)

	writeJSON(w, http.StatusOK, map[string]any{
		"success":  true,
		"presence": hub.BuildPresence(accountID),
	})
}
