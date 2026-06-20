package rooms

import (
	"net/http"
	"strconv"
	"strings"

	"meow.net/controllers"
	"meow.net/controllers/hub"
	"meow.net/db"
	"meow.net/models"
	"meow.net/utils"
)

func RoomCreateSubRoom(w http.ResponseWriter, r *http.Request) {
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

	r.ParseForm()
	name := r.FormValue("name")
	if name == "" {
		http.Error(w, "Bad Request", http.StatusBadRequest)
		return
	}

	if utils.IsTextFlagged(name) {
		writeRoomModerationRejection(w, "Sub-room name violates the community guidelines.")
		return
	}

	for _, sr := range room.SubRooms {
		if strings.EqualFold(sr.Name, name) {
			writeRoomModerationRejection(w, "A sub-room with that name already exists.")
			return
		}
	}

	unitySceneId := ""
	for _, sr := range room.SubRooms {
		if !sr.IsSandbox {
			unitySceneId = sr.UnitySceneId
			break
		}
	}
	if unitySceneId == "" && len(room.SubRooms) > 0 {
		unitySceneId = room.SubRooms[0].UnitySceneId
	}

	sub := models.SubRoom{
		RoomId:           room.RoomId,
		Name:             name,
		DataBlob:         "",
		IsSandbox:        true,
		MaxPlayers:       20,
		Accessibility:    int(models.RoomAccessibilityPublic),
		UnitySceneId:     unitySceneId,
		SavedByAccountId: int(accountId),
	}
	if err := db.DB.Create(&sub).Error; err != nil {
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	updated, ok := loadRoom(roomId)
	if !ok {
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	frame := hub.NotifFrame("RoomUpdate", updated)
	notified := map[int]bool{}
	for _, role := range updated.Roles {
		if role.AccountId != 0 && !notified[role.AccountId] {
			hub.HubSendToPlayer(role.AccountId, frame)
			notified[role.AccountId] = true
		}
	}
	if updated.CreatorAccountId != 0 && !notified[updated.CreatorAccountId] {
		hub.HubSendToPlayer(updated.CreatorAccountId, frame)
	}

	roomSuccessResponse(w, updated)
}

func parseRoomSubRoomPath(r *http.Request) (roomId, subRoomId int, ok bool) {
	parts := strings.Split(strings.Trim(r.URL.Path, "/"), "/")
	rIdx, sIdx := -1, -1
	for i, seg := range parts {
		switch seg {
		case "rooms":
			rIdx = i
		case "subrooms":
			sIdx = i
		}
	}
	if rIdx < 0 || sIdx < 0 || rIdx+1 >= len(parts) || sIdx+1 >= len(parts) {
		return 0, 0, false
	}
	rid, err := strconv.Atoi(parts[rIdx+1])
	if err != nil {
		return 0, 0, false
	}
	sid, err := strconv.Atoi(parts[sIdx+1])
	if err != nil {
		return 0, 0, false
	}
	return rid, sid, true
}

func RoomSubRoomModify(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	roomId, subRoomId, ok := parseRoomSubRoomPath(r)
	if !ok {
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

	var target *models.SubRoom
	for i := range room.SubRooms {
		if int(room.SubRooms[i].SubRoomId) == subRoomId {
			target = &room.SubRooms[i]
			break
		}
	}
	if target == nil {
		http.Error(w, "Not Found", http.StatusNotFound)
		return
	}

	r.ParseForm()
	updates := map[string]interface{}{}

	if name := r.FormValue("name"); name != "" {
		if utils.IsTextFlagged(name) {
			writeRoomModerationRejection(w, "Sub-room name violates the community guidelines.")
			return
		}
		for i := range room.SubRooms {
			if int(room.SubRooms[i].SubRoomId) != subRoomId && strings.EqualFold(room.SubRooms[i].Name, name) {
				writeRoomModerationRejection(w, "A sub-room with that name already exists.")
				return
			}
		}
		updates["name"] = name
		target.Name = name
	}

	if accStr := r.FormValue("accessibility"); accStr != "" {
		acc, err := strconv.Atoi(accStr)
		if err != nil {
			http.Error(w, "Bad Request", http.StatusBadRequest)
			return
		}
		updates["accessibility"] = acc
		target.Accessibility = acc
	}

	if mpStr := r.FormValue("maxPlayers"); mpStr != "" {
		mp, err := strconv.Atoi(mpStr)
		if err != nil {
			http.Error(w, "Bad Request", http.StatusBadRequest)
			return
		}
		updates["max_players"] = mp
		target.MaxPlayers = mp
	}

	if len(updates) > 0 {
		db.DB.Model(&models.SubRoom{}).Where("sub_room_id = ?", subRoomId).Updates(updates)
	}

	roomSuccessResponse(w, room)
}

func RoomSubRoomAccessibility(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	roomId, subRoomId, ok := parseRoomSubRoomPath(r)
	if !ok {
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

	accRaw := strings.TrimSpace(r.FormValue("accessibility"))
	var accVal int
	switch strings.ToLower(accRaw) {
	case "private":
		accVal = int(models.RoomAccessibilityPrivate)
	case "public":
		accVal = int(models.RoomAccessibilityPublic)
	case "unlisted":
		accVal = int(models.RoomAccessibilityUnlisted)
	default:
		n, err := strconv.Atoi(accRaw)
		if err != nil {
			http.Error(w, "Bad Request", http.StatusBadRequest)
			return
		}
		accVal = n
	}

	found := false
	for i := range room.SubRooms {
		if int(room.SubRooms[i].SubRoomId) == subRoomId {
			room.SubRooms[i].Accessibility = accVal
			found = true
			break
		}
	}
	if !found {
		http.Error(w, "Not Found", http.StatusNotFound)
		return
	}

	db.DB.Model(&models.SubRoom{}).Where("sub_room_id = ?", subRoomId).Update("accessibility", accVal)
	roomSuccessResponse(w, room)
}
