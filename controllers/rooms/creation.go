package rooms

import (
	"encoding/json"
	"net/http"
	"strconv"
	"strings"

	"meow.net/controllers"
	"meow.net/db"
	"meow.net/models"
)

func writeRoomModerationRejection(w http.ResponseWriter, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusBadRequest)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"error":   message,
		"success": false,
		"value":   nil,
	})
}

func roomSuccessResponse(w http.ResponseWriter, room models.Room) {
	rooms := []models.Room{room}
	controllers.InitRoomSlices(rooms)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"error":   "",
		"success": true,
		"value":   rooms[0],
	})
}

func isRoomOwner(room models.Room, accountId int) bool {
	for _, role := range room.Roles {
		if role.AccountId == accountId && role.Role == 255 {
			return true
		}
	}
	return false
}

func canSaveRoom(room models.Room, accountId int) bool {
	for _, role := range room.Roles {
		if role.AccountId != accountId {
			continue
		}
		switch models.RoomRole(role.Role) {
		case models.RoomRoleCreator, models.RoomRoleCoOwner, models.RoomRoleTemporaryCoOwner:
			return true
		}
	}
	return false
}

var baseRoomIDs = []int{
	23,
}

func RoomsBase(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	var roomList []models.Room
	db.DB.Preload("SubRooms").Preload("Roles").Preload("Tags").
		Where("room_id IN (?) OR room_id IN (?)",
			db.DB.Table("room_tags").Select("room_id").Where("LOWER(tag) = ?", "base"),
			baseRoomIDs).
		Find(&roomList)
	controllers.InitRoomSlices(roomList)
	json.NewEncoder(w).Encode(roomList)
}

func RoomsOwnedBy(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	parts := strings.Split(strings.TrimRight(r.URL.Path, "/"), "/")
	userIdStr := parts[len(parts)-1]

	var accountId int
	var err error
	if userIdStr == "me" {
		id, err := controllers.CurrentUserIDFromRequest(r)
		if err != nil {
			json.NewEncoder(w).Encode([]interface{}{})
			return
		}
		accountId = int(id)
	} else {
		accountId, err = strconv.Atoi(userIdStr)
		if err != nil {
			http.Error(w, "Bad Request", http.StatusBadRequest)
			return
		}
	}

	viewerId, _ := controllers.CurrentUserIDFromRequest(r)
	isSelf := int(viewerId) == accountId

	query := db.DB.Preload("SubRooms").Preload("Roles").Preload("Tags").
		Where("room_id IN (?)",
			db.DB.Table("room_roles").Select("room_id").Where("account_id = ? AND role = ?", accountId, 255))
	if isSelf {
		query = query.Where("accessibility <> ?", int(models.RoomAccessibilityUnlisted))
	} else {
		query = query.Where("accessibility = ?", int(models.RoomAccessibilityPublic))
	}

	var roomList []models.Room
	query.Find(&roomList)
	controllers.InitRoomSlices(roomList)
	json.NewEncoder(w).Encode(roomList)
}

func RoomClone(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	parts := strings.Split(strings.Trim(r.URL.Path, "/"), "/")
	if len(parts) < 2 {
		http.Error(w, "Bad Request", http.StatusBadRequest)
		return
	}
	roomId, err := strconv.Atoi(parts[1])
	if err != nil {
		http.Error(w, "Bad Request", http.StatusBadRequest)
		return
	}

	accountId, err := controllers.CurrentUserIDFromRequest(r)
	if err != nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	name := r.FormValue("name")
	if msg, ok := validateRoomName(name, 0); !ok {
		writeRoomModerationRejection(w, msg)
		return
	}

	src, ok := loadRoom(roomId)
	if !ok {
		http.Error(w, "Not Found", http.StatusNotFound)
		return
	}

	if !src.CloningAllowed && !isRoomOwner(src, int(accountId)) {
		http.Error(w, "Forbidden", http.StatusForbidden)
		return
	}

	newRoom := models.Room{
		Name:                     name,
		Description:              src.Description,
		ImageName:                src.ImageName,
		CreatorAccountId:         int(accountId),
		State:                    0,
		Accessibility:            0,
		AutoLocalizeRoom:         src.AutoLocalizeRoom,
		CloningAllowed:           false,
		CustomWarning:            "",
		DisableMicAutoMute:       src.DisableMicAutoMute,
		DisableRoomComments:      src.DisableRoomComments,
		EncryptVoiceChat:         src.EncryptVoiceChat,
		IsDeveloperOwned:         false,
		IsDorm:                   false,
		IsRRO:                    false,
		LoadScreenLocked:         src.LoadScreenLocked,
		MaxPlayerCalculationMode: src.MaxPlayerCalculationMode,
		MaxPlayers:               src.MaxPlayers,
		MinLevel:                 src.MinLevel,
		PersistenceVersion:       src.PersistenceVersion,
		RankedEntityId:           "",
		RankingContext:           0,
		SupportsJuniors:          src.SupportsJuniors,
		SupportsLevelVoting:      src.SupportsLevelVoting,
		SupportsMobile:           src.SupportsMobile,
		SupportsQuest2:           src.SupportsQuest2,
		SupportsScreens:          src.SupportsScreens,
		SupportsTeleportVR:       src.SupportsTeleportVR,
		SupportsVRLow:            src.SupportsVRLow,
		SupportsWalkVR:           src.SupportsWalkVR,
		ToxmodEnabled:            src.ToxmodEnabled,
		UgcVersion:               src.UgcVersion,
		WarningMask:              0,
		DataBlob:                 nil,
		Roles: []models.RoomRoleEntry{
			{AccountId: int(accountId), InvitedRole: 0, Role: 255},
		},
	}

	for _, sr := range src.SubRooms {
		newRoom.SubRooms = append(newRoom.SubRooms, models.SubRoom{
			Accessibility:    sr.Accessibility,
			DataBlob:         "",
			IsSandbox:        sr.IsSandbox,
			MaxPlayers:       sr.MaxPlayers,
			Name:             sr.Name,
			SavedByAccountId: -1,
			UnitySceneId:     sr.UnitySceneId,
		})
	}

	if err := db.DB.Create(&newRoom).Error; err != nil {
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	loaded, ok := loadRoom(int(newRoom.RoomId))
	if !ok {
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	roomSuccessResponse(w, loaded)
}

func RoomDelete(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	parts := strings.Split(strings.Trim(r.URL.Path, "/"), "/")
	if len(parts) < 2 {
		http.Error(w, "Bad Request", http.StatusBadRequest)
		return
	}
	roomId, err := strconv.Atoi(parts[1])
	if err != nil {
		http.Error(w, "Bad Request", http.StatusBadRequest)
		return
	}

	accountId, err := controllers.CurrentUserIDFromRequest(r)
	if err != nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	room, ok := loadRoom(roomId)
	if !ok {
		http.Error(w, "Not Found", http.StatusNotFound)
		return
	}

	if !isRoomOwner(room, int(accountId)) {
		http.Error(w, "Forbidden", http.StatusForbidden)
		return
	}

	if err := db.DB.Delete(&models.Room{}, roomId).Error; err != nil {
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	json.NewEncoder(w).Encode(map[string]interface{}{
		"error":   "",
		"success": true,
		"value":   nil,
	})
}
