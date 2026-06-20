package rooms

import (
	"encoding/json"
	"net/http"
	"strconv"
	"strings"

	"meow.net/controllers"
	"meow.net/db"
	"meow.net/models"
	"meow.net/utils"
)

func roomByName(name string) (models.Room, error) {
	var room models.Room
	err := db.DB.Preload("SubRooms").Preload("Roles").Preload("Tags").
		Where("LOWER(name) = ?", strings.ToLower(name)).First(&room).Error
	return room, err
}

func validateRoomName(name string, excludeRoomId int) (string, bool) {
	trimmed := strings.TrimSpace(name)
	if trimmed == "" {
		return "This room name is taken", false
	}
	if !utils.IsValidName(trimmed) {
		return "Room names can only use letters, numbers, and basic punctuation", false
	}
	if !utils.IsValidNameLength(trimmed) {
		return "Room names can be at most 16 characters", false
	}
	if utils.IsTextFlagged(trimmed) {
		return "Try another room name", false
	}
	existing, err := roomByName(trimmed)
	if err == nil && int(existing.RoomId) != excludeRoomId {
		return "This room name is taken", false
	}
	return "", true
}

func isAssignableRoomRole(role int) bool {
	switch models.RoomRole(role) {
	case models.RoomRoleNone,
		models.RoomRoleModerator,
		models.RoomRoleCoOwner,
		models.RoomRoleTemporaryCoOwner:
		return true
	}
	return false
}

func serializeSingleRoom(w http.ResponseWriter, room models.Room) {
	rooms := []models.Room{room}
	controllers.InitRoomSlices(rooms)
	json.NewEncoder(w).Encode(rooms[0])
}

func loadRoom(roomId int) (models.Room, bool) {
	var room models.Room
	if err := db.DB.Preload("SubRooms").Preload("Roles").Preload("Tags").
		First(&room, roomId).Error; err != nil {
		return room, false
	}
	rooms := []models.Room{room}
	controllers.InitRoomSlices(rooms)
	return rooms[0], true
}

func parseRoomBool(r *http.Request, key string) bool {
	v := r.FormValue(key)
	return strings.EqualFold(v, "true") || v == "1"
}

func parseRoleInvitePath(r *http.Request, currentUserId int) (roomId, inviteeId int, ok bool) {
	parts := strings.Split(strings.Trim(r.URL.Path, "/"), "/")
	roomIdx, rolesIdx := -1, -1
	for i, seg := range parts {
		switch seg {
		case "rooms":
			roomIdx = i
		case "roles":
			rolesIdx = i
		}
	}
	if roomIdx < 0 || rolesIdx < 0 || roomIdx+1 >= len(parts) || rolesIdx+1 >= len(parts) {
		return 0, 0, false
	}
	rid, err := strconv.Atoi(parts[roomIdx+1])
	if err != nil {
		return 0, 0, false
	}
	idSeg := parts[rolesIdx+1]
	if strings.EqualFold(idSeg, "me") {
		return rid, currentUserId, currentUserId != 0
	}
	iid, err := strconv.Atoi(idSeg)
	if err != nil {
		return 0, 0, false
	}
	return rid, iid, true
}
