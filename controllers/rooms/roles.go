package rooms

import (
	"encoding/json"
	"net/http"
	"strconv"

	"meow.net/controllers"
	"meow.net/controllers/hub"
	"meow.net/db"
	"meow.net/models"
	"meow.net/utils"
)

func RoomVerifyRole(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	if r.Method != http.MethodPost {
		w.Write([]byte("false"))
		return
	}

	roomIdStr := r.URL.Query().Get("roomId")
	roleStr := r.URL.Query().Get("role")

	if r.Method == http.MethodPost {
		if err := r.ParseForm(); err == nil {
			if r.FormValue("roomId") != "" {
				roomIdStr = r.FormValue("roomId")
			}
			if r.FormValue("role") != "" {
				roleStr = r.FormValue("role")
			}
		}
	}

	roomId, _ := strconv.Atoi(roomIdStr)
	role, _ := strconv.Atoi(roleStr)

	tokenStr := utils.GetBearerToken(r)
	if tokenStr == "" {
		w.Write([]byte("false"))
		return
	}

	accountIdStr, err := utils.ParseSubFromJWT(tokenStr)
	if err != nil {
		w.Write([]byte("false"))
		return
	}
	accountId, _ := strconv.Atoi(accountIdStr)

	var room models.Room
	if err := db.DB.Where("room_id = ?", roomId).First(&room).Error; err != nil {
		w.Write([]byte("false"))
		return
	}

	if room.CreatorAccountId == accountId {
		w.Write([]byte("true"))
		return
	}

	var roomRole models.RoomRoleEntry
	if err := db.DB.Where("room_id = ? AND account_id = ? AND role >= ?", roomId, accountId, role).First(&roomRole).Error; err == nil {
		w.Write([]byte("true"))
		return
	}

	w.Write([]byte("false"))
}

func RoomRoleInvite(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	senderId, err := controllers.CurrentUserIDFromRequest(r)
	if err != nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	roomId, inviteeId, ok := parseRoleInvitePath(r, int(senderId))
	if !ok || inviteeId == 0 {
		http.Error(w, "Bad Request", http.StatusBadRequest)
		return
	}

	r.ParseForm()
	role, _ := strconv.Atoi(r.FormValue("role"))
	if role == 0 {
		role, _ = strconv.Atoi(r.URL.Query().Get("role"))
	}
	if !isAssignableRoomRole(role) {
		http.Error(w, "Bad Request", http.StatusBadRequest)
		return
	}

	room, ok := loadRoom(roomId)
	if !ok {
		http.Error(w, "Not Found", http.StatusNotFound)
		return
	}

	isOwner := isRoomOwner(room, int(senderId))
	isSelfDecline := int(senderId) == inviteeId && role == 0
	if !isOwner && !isSelfDecline {
		http.Error(w, "Forbidden", http.StatusForbidden)
		return
	}

	var entry models.RoomRoleEntry
	err = db.DB.Where("room_id = ? AND account_id = ?", roomId, inviteeId).First(&entry).Error
	if err != nil {
		entry = models.RoomRoleEntry{
			RoomId:      uint(roomId),
			AccountId:   inviteeId,
			Role:        0,
			InvitedRole: role,
		}
		db.DB.Create(&entry)
	} else {
		entry.InvitedRole = role
		db.DB.Save(&entry)
	}

	room, ok = loadRoom(roomId)
	if !ok {
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	frame := hub.NotifFrame("RoomUpdate", room)
	hub.HubSendToPlayer(int(senderId), frame)
	if inviteeId != int(senderId) {
		hub.HubSendToPlayer(inviteeId, frame)
	}

	roomIdU := uint(roomId)
	msg := models.Message{
		FromPlayerId: senderId,
		ToPlayerId:   uint(inviteeId),
		Type:         62,
		RoomId:       &roomIdU,
		Data:         "",
	}
	db.DB.Create(&msg)
	hub.HubSendToPlayer(inviteeId, hub.NotifFrame(models.MessageReceived, msg))

	json.NewEncoder(w).Encode(map[string]interface{}{
		"error":   "",
		"success": true,
		"value":   room,
	})
}

func RoomRoleSet(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	senderId, err := controllers.CurrentUserIDFromRequest(r)
	if err != nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	roomId, targetId, ok := parseRoleInvitePath(r, int(senderId))
	if !ok || targetId == 0 {
		http.Error(w, "Bad Request", http.StatusBadRequest)
		return
	}

	r.ParseForm()
	roleStr := r.FormValue("role")
	if roleStr == "" {
		roleStr = r.URL.Query().Get("role")
	}
	if roleStr == "" {
		var body struct {
			Role *int `json:"role"`
		}
		if r.Body != nil {
			_ = json.NewDecoder(r.Body).Decode(&body)
		}
		if body.Role != nil {
			roleStr = strconv.Itoa(*body.Role)
		}
	}
	role, _ := strconv.Atoi(roleStr)
	if !isAssignableRoomRole(role) {
		http.Error(w, "Bad Request", http.StatusBadRequest)
		return
	}

	room, ok := loadRoom(roomId)
	if !ok {
		http.Error(w, "Not Found", http.StatusNotFound)
		return
	}

	var entry models.RoomRoleEntry
	entryErr := db.DB.Where("room_id = ? AND account_id = ?", roomId, targetId).First(&entry).Error

	isOwner := isRoomOwner(room, int(senderId))
	isSelfAccept := int(senderId) == targetId && entryErr == nil && entry.InvitedRole != 0 && role == entry.InvitedRole
	if !isOwner && !isSelfAccept {
		http.Error(w, "Forbidden", http.StatusForbidden)
		return
	}

	if entryErr != nil {
		entry = models.RoomRoleEntry{
			RoomId:    uint(roomId),
			AccountId: targetId,
			Role:      role,
		}
		db.DB.Create(&entry)
	} else {
		entry.Role = role
		entry.InvitedRole = 0
		db.DB.Save(&entry)
	}

	room, ok = loadRoom(roomId)
	if !ok {
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	frame := hub.NotifFrame("RoomUpdate", room)
	notified := map[int]bool{}
	for _, rr := range room.Roles {
		if rr.AccountId != 0 && !notified[rr.AccountId] {
			hub.HubSendToPlayer(rr.AccountId, frame)
			notified[rr.AccountId] = true
		}
	}
	if room.CreatorAccountId != 0 && !notified[room.CreatorAccountId] {
		hub.HubSendToPlayer(room.CreatorAccountId, frame)
	}
	if !notified[targetId] {
		hub.HubSendToPlayer(targetId, frame)
	}

	json.NewEncoder(w).Encode(map[string]interface{}{
		"error":   "",
		"success": true,
		"value":   room,
	})
}

func RoomRoleAcceptInvite(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	accepterId, err := controllers.CurrentUserIDFromRequest(r)
	if err != nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	roomId, inviteeId, ok := parseRoleInvitePath(r, int(accepterId))
	if !ok || inviteeId == 0 {
		http.Error(w, "Bad Request", http.StatusBadRequest)
		return
	}
	if int(accepterId) != inviteeId {
		http.Error(w, "Forbidden", http.StatusForbidden)
		return
	}

	var entry models.RoomRoleEntry
	if err := db.DB.Where("room_id = ? AND account_id = ?", roomId, inviteeId).First(&entry).Error; err != nil {
		http.Error(w, "Not Found", http.StatusNotFound)
		return
	}
	if entry.InvitedRole != 0 {
		entry.Role = entry.InvitedRole
		entry.InvitedRole = 0
		db.DB.Save(&entry)
	}

	room, ok := loadRoom(roomId)
	if !ok {
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	frame := hub.NotifFrame("RoomUpdate", room)
	notified := map[int]bool{}
	for _, role := range room.Roles {
		if role.AccountId != 0 && !notified[role.AccountId] {
			hub.HubSendToPlayer(role.AccountId, frame)
			notified[role.AccountId] = true
		}
	}
	if room.CreatorAccountId != 0 && !notified[room.CreatorAccountId] {
		hub.HubSendToPlayer(room.CreatorAccountId, frame)
	}

	json.NewEncoder(w).Encode(map[string]interface{}{
		"error":   "",
		"success": true,
		"value":   room,
	})
}
