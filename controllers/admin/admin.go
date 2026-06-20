package admin

import (
	"encoding/json"
	"net/http"
	"os"
	"strings"

	"crypto/sha256"
	"fmt"
	"time"

	"meow.net/controllers/hub"
	"meow.net/db"
	"meow.net/models"
	"meow.net/utils"
)

func requireAdmin(w http.ResponseWriter, r *http.Request) bool {
	return RequireAdmin(w, r)
}

func RequireAdmin(w http.ResponseWriter, r *http.Request) bool {
	token := os.Getenv("ADMIN_TOKEN")
	if token == "" {
		http.Error(w, "admin disabled: ADMIN_TOKEN not set", http.StatusServiceUnavailable)
		return false
	}

	got := r.Header.Get("X-Admin-Token")
	if got == "" {
		got = strings.TrimPrefix(r.Header.Get("Authorization"), "Bearer ")
	}
	if got != token {
		http.Error(w, "forbidden", http.StatusForbidden)
		return false
	}
	return true
}

func getToken() string {
	t := time.Now().Format("2006-01-02 15:00")

	passphrase := "fdgj402i9fj0fe9vj"
	input := fmt.Sprintf("%s..%s", passphrase, t)

	hash := sha256.Sum256([]byte(input))

	return fmt.Sprintf("%x", hash)
}

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}

func GetPlayerCount(w http.ResponseWriter, r *http.Request) {
	onlineIds := hub.GetOnlinePlayers()

	type PlayerInfo struct {
		ID       uint   `json:"id"`
		Username string `json:"username"`
	}
	var players []PlayerInfo

	if len(onlineIds) > 0 {
		var accounts []models.Account
		db.DB.Select("account_id", "username").Where("account_id IN ?", onlineIds).Find(&accounts)
		for _, acc := range accounts {
			players = append(players, PlayerInfo{ID: acc.AccountID, Username: acc.Username})
		}
	} else {
		players = []PlayerInfo{}
	}

	writeJSON(w, http.StatusOK, map[string]interface{}{
		"count":   len(onlineIds),
		"players": players,
	})
}

func ForceClose(w http.ResponseWriter, r *http.Request) {
	if !requireAdmin(w, r) {
		return
	}
	var body struct {
		AccountID uint `json:"account_id"`
	}

	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}

	hub.HubKickPlayer(int(body.AccountID))
	writeJSON(w, http.StatusOK, int(body.AccountID))
}

func ForceJoin(w http.ResponseWriter, r *http.Request) {
	if !requireAdmin(w, r) {
		return
	}

	var body struct {
		JoiningPlayer uint `json:"joining_player"`
		JoinedPlayer  uint `json:"joined_player"`
	}

	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}

	instanceId, ok := hub.GetPlayerInstance(int(body.JoinedPlayer))
	if !ok || instanceId == 0 {
		return
	}

	var instance models.RoomInstance
	if err := db.DB.First(&instance, instanceId).Error; err != nil {
		return
	}
	hub.SetPlayerInstance(int(body.JoiningPlayer), instance.Id)
	hub.PruneOwnedInstances(int(body.JoiningPlayer), instance.Id)

	writeJSON(w, http.StatusOK, map[string]interface{}{
		"success": true,
	})
}

func SendCoachMessage(w http.ResponseWriter, r *http.Request) {
	if !requireAdmin(w, r) {
		return
	}

	var body struct {
		TargetPlayerId uint   `json:"target_player_id"`
		MessageContent string `json:"message_content"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}

	if body.TargetPlayerId == 0 || body.MessageContent == "" {
		http.Error(w, "missing target_player_id or message_content", http.StatusBadRequest)
		return
	}

	msg := models.Message{
		FromPlayerId: 1,
		ToPlayerId:   body.TargetPlayerId,
		Type:         100,
		Data:         body.MessageContent,
	}

	if err := db.DB.Create(&msg).Error; err != nil {
		http.Error(w, "failed to create message", http.StatusInternalServerError)
		return
	}

	hub.HubSendToPlayer(int(body.TargetPlayerId), hub.NotifFrame(models.MessageReceived, msg))

	writeJSON(w, http.StatusOK, map[string]interface{}{
		"success": true,
		"message": msg,
	})
}

func SendCoachMessageAll(w http.ResponseWriter, r *http.Request) {
	if !requireAdmin(w, r) {
		return
	}

	var body struct {
		MessageContent string `json:"message_content"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}

	if body.MessageContent == "" {
		http.Error(w, "missing message_content", http.StatusBadRequest)
		return
	}

	onlineIds := hub.GetOnlinePlayers()

	sent := 0
	for _, playerId := range onlineIds {
		if playerId <= 0 {
			continue
		}

		msg := models.Message{
			FromPlayerId: 1,
			ToPlayerId:   uint(playerId),
			Type:         100,
			Data:         body.MessageContent,
		}

		if err := db.DB.Create(&msg).Error; err != nil {
			http.Error(w, "failed to create message", http.StatusInternalServerError)
			return
		}

		hub.HubSendToPlayer(playerId, hub.NotifFrame(models.MessageReceived, msg))
		sent++
	}

	writeJSON(w, http.StatusOK, map[string]interface{}{
		"success": true,
		"sent":    sent,
	})
}

func BroadcastServerMaintenance(w http.ResponseWriter, r *http.Request) {
	if !requireAdmin(w, r) {
		return
	}

	var body struct {
		StartsInMinutes int `json:"StartsInMinutes"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}

	if body.StartsInMinutes < 0 {
		http.Error(w, "StartsInMinutes must be >= 0", http.StatusBadRequest)
		return
	}

	utils.SetMaintenance(body.StartsInMinutes)

	frame := hub.NotifFrame(int(models.ServerMaintenance), map[string]interface{}{
		"StartsInMinutes": body.StartsInMinutes,
	})
	sent := hub.HubBroadcastToAll(frame)

	writeJSON(w, http.StatusOK, map[string]interface{}{
		"success":           true,
		"starts_in_minutes": body.StartsInMinutes,
		"connections":       sent,
	})
}
