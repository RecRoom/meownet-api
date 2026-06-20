package moderation

import (
	"encoding/json"
	"math"
	"net/http"
	"time"

	"meow.net/controllers"
	"meow.net/controllers/hub"
	"meow.net/db"
	"meow.net/models"
)

const InstantKickBanDuration = 3 * time.Hour

type instantKickRequest struct {
	GameSessionId int64  `json:"GameSessionId"`
	PlayerIds     []uint `json:"PlayerIds"`
}

func InstantKick(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	currentUserID, err := controllers.CurrentUserIDFromRequest(r)
	if err != nil || currentUserID == 0 {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	var req instantKickRequest
	if r.Body != nil {
		_ = json.NewDecoder(r.Body).Decode(&req)
	}

	if req.GameSessionId == 0 {
		http.Error(w, "Bad Request", http.StatusBadRequest)
		return
	}

	var instance models.RoomInstance
	if err := db.DB.First(&instance, req.GameSessionId).Error; err != nil {
		http.Error(w, "Not Found", http.StatusNotFound)
		return
	}

	if !canModerateRoom(int64(instance.RoomId), int(currentUserID)) {
		http.Error(w, "Forbidden", http.StatusForbidden)
		return
	}

	expiresAt := time.Now().Add(InstantKickBanDuration)
	for _, pid := range req.PlayerIds {
		if pid == 0 {
			continue
		}
		db.DB.Create(&models.InstanceBan{
			InstanceID: req.GameSessionId,
			AccountID:  pid,
			IssuedBy:   currentUserID,
			ExpiresAt:  expiresAt,
		})
		frame := hub.NotifFrame(int(models.ModerationKick), map[string]interface{}{
			"ReportCategory":   -1,
			"Duration":         math.MinInt32,
			"GameSessionId":    -1,
			"IsHostKick":       true,
			"PlayerIdReporter": currentUserID,
			"Message":          "Kicked by a room moderator.",
			"IsBan":            false,
		})
		hub.HubSendToPlayer(int(pid), frame)
	}

	json.NewEncoder(w).Encode(map[string]interface{}{
		"Message": "",
		"Success": true,
	})
}

func canModerateRoom(roomId int64, accountId int) bool {
	var room models.Room
	if err := db.DB.Preload("Roles").Where("room_id = ?", roomId).First(&room).Error; err != nil {
		return false
	}
	if room.CreatorAccountId == accountId {
		return true
	}
	for _, role := range room.Roles {
		if role.AccountId == accountId && role.Role >= int(models.RoomRoleHost) {
			return true
		}
	}
	return false
}

func VoteToKickReasons(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode([]map[string]interface{}{
		{"Reason": "Cheating", "ReportCategory": 102},
	})
}
