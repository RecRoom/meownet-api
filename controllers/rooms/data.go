package rooms

import (
	"encoding/json"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"meow.net/controllers"
	"meow.net/controllers/hub"
	"meow.net/db"
	"meow.net/discord"
	"meow.net/models"
)

func RoomData(w http.ResponseWriter, r *http.Request) {
	parts := strings.Split(strings.TrimRight(r.URL.Path, "/"), "/")
	if len(parts) >= 4 && parts[len(parts)-1] == "instances" {
		RoomInstancesList(w, r)
		return
	}

	filename := parts[len(parts)-1]
	if strings.Contains(filename, "..") || filename == "" {
		http.Error(w, "Bad Request", http.StatusBadRequest)
		return
	}

	data, err := os.ReadFile("data/room/" + filename)
	if err != nil {
		http.Error(w, "Not Found", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/octet-stream")
	w.Write(data)
}

func RoomPlayerDataMe(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusNotFound)
	json.NewEncoder(w).Encode(map[string]interface{}{"Data": ""})
}

func RoomSaveSubRoomData(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	accountId, err := controllers.CurrentUserIDFromRequest(r)
	if err != nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	parts := strings.Split(strings.Trim(r.URL.Path, "/"), "/")
	roomIdx, subIdx := -1, -1
	for i, seg := range parts {
		switch seg {
		case "rooms":
			roomIdx = i
		case "subrooms":
			subIdx = i
		}
	}
	if roomIdx < 0 || subIdx < 0 || roomIdx+1 >= len(parts) || subIdx+1 >= len(parts) {
		http.Error(w, "Bad Request", http.StatusBadRequest)
		return
	}
	roomId, err := strconv.Atoi(parts[roomIdx+1])
	if err != nil {
		http.Error(w, "Bad Request", http.StatusBadRequest)
		return
	}
	subRoomId, err := strconv.Atoi(parts[subIdx+1])
	if err != nil {
		http.Error(w, "Bad Request", http.StatusBadRequest)
		return
	}

	if err := r.ParseForm(); err != nil {
		http.Error(w, "Bad Request", http.StatusBadRequest)
		return
	}
	subRoomBlob := r.FormValue("filename")
	roomBlob := r.FormValue("roomDataFilename")

	var room models.Room
	if err := db.DB.Preload("SubRooms").Preload("Roles").Preload("Tags").
		First(&room, roomId).Error; err != nil {
		http.Error(w, "Not Found", http.StatusNotFound)
		return
	}

	if !canSaveRoom(room, int(accountId)) {
		http.Error(w, "Forbidden", http.StatusForbidden)
		return
	}

	if roomBlob != "" {
		db.DB.Model(&models.Room{}).Where("room_id = ?", roomId).
			Update("data_blob", roomBlob)
		room.DataBlob = &roomBlob
	}

	subRoomFound := false
	for i := range room.SubRooms {
		if int(room.SubRooms[i].SubRoomId) == subRoomId {
			updates := map[string]interface{}{}
			if subRoomBlob != "" {
				updates["data_blob"] = subRoomBlob
				room.SubRooms[i].DataBlob = subRoomBlob
			}
			updates["saved_by_account_id"] = int(accountId)
			room.SubRooms[i].SavedByAccountId = int(accountId)
			if len(updates) > 0 {
				db.DB.Model(&models.SubRoom{}).
					Where("sub_room_id = ?", subRoomId).
					Updates(updates)
			}
			if subRoomBlob != "" {
				db.DB.Create(&models.SubRoomDataHistory{
					SubRoomId:        subRoomId,
					DataBlob:         subRoomBlob,
					SavedByAccountId: int(accountId),
				})
			}
			subRoomFound = true
			break
		}
	}
	if !subRoomFound {
		http.Error(w, "SubRoom not found", http.StatusNotFound)
		return
	}

	roomList := []models.Room{room}
	controllers.InitRoomSlices(roomList)

	json.NewEncoder(w).Encode(map[string]interface{}{
		"error":   "",
		"success": true,
		"value":   roomList[0],
	})
}

func RoomSubRoomDataHistory(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	accountId, err := controllers.CurrentUserIDFromRequest(r)
	if err != nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	roomId, subRoomId, ok := parseRoomSubRoomPath(r)
	if !ok {
		http.Error(w, "Bad Request", http.StatusBadRequest)
		return
	}

	room, ok := loadRoom(roomId)
	if !ok {
		http.Error(w, "Not Found", http.StatusNotFound)
		return
	}
	if !canSaveRoom(room, int(accountId)) {
		http.Error(w, "Forbidden", http.StatusForbidden)
		return
	}

	subRoomFound := false
	for i := range room.SubRooms {
		if int(room.SubRooms[i].SubRoomId) == subRoomId {
			subRoomFound = true
			break
		}
	}
	if !subRoomFound {
		http.Error(w, "Not Found", http.StatusNotFound)
		return
	}

	history := []models.SubRoomDataHistory{}
	db.DB.Where("sub_room_id = ?", subRoomId).
		Order("created_at DESC").
		Find(&history)

	json.NewEncoder(w).Encode(history)
}

func RoomRestoreSubRoomData(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	accountId, err := controllers.CurrentUserIDFromRequest(r)
	if err != nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	roomId, subRoomId, ok := parseRoomSubRoomPath(r)
	if !ok {
		http.Error(w, "Bad Request", http.StatusBadRequest)
		return
	}

	if err := r.ParseForm(); err != nil {
		http.Error(w, "Bad Request", http.StatusBadRequest)
		return
	}
	blob := r.FormValue("filename")
	if blob == "" {
		http.Error(w, "Bad Request", http.StatusBadRequest)
		return
	}

	room, ok := loadRoom(roomId)
	if !ok {
		http.Error(w, "Not Found", http.StatusNotFound)
		return
	}
	if !canSaveRoom(room, int(accountId)) {
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
		http.Error(w, "SubRoom not found", http.StatusNotFound)
		return
	}

	var count int64
	db.DB.Model(&models.SubRoomDataHistory{}).
		Where("sub_room_id = ? AND data_blob = ?", subRoomId, blob).
		Count(&count)
	if count == 0 {
		http.Error(w, "Not Found", http.StatusNotFound)
		return
	}

	db.DB.Model(&models.SubRoom{}).
		Where("sub_room_id = ?", subRoomId).
		Updates(map[string]interface{}{
			"data_blob":           blob,
			"saved_by_account_id": int(accountId),
		})
	target.DataBlob = blob
	target.SavedByAccountId = int(accountId)

	db.DB.Create(&models.SubRoomDataHistory{
		SubRoomId:        subRoomId,
		DataBlob:         blob,
		SavedByAccountId: int(accountId),
	})

	roomSuccessResponse(w, room)
}

func RoomInstanceDispatch(w http.ResponseWriter, r *http.Request) {
	parts := strings.Split(strings.Trim(r.URL.Path, "/"), "/")
	if len(parts) < 3 {
		http.Error(w, "Bad Request", http.StatusBadRequest)
		return
	}
	switch parts[2] {
	case "markprivate":
		RoomInstanceMarkPrivate(w, r)
	case "reportjoinresult":
		RoomInstanceReportJoinResult(w, r)
	default:
		RoomInstanceInProgress(w, r)
	}
}

func RoomInstanceInProgress(w http.ResponseWriter, r *http.Request) {
	accountId, err := controllers.CurrentUserIDFromRequest(r)
	if err != nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	parts := strings.Split(strings.Trim(r.URL.Path, "/"), "/")
	if len(parts) < 3 {
		http.Error(w, "Bad Request", http.StatusBadRequest)
		return
	}
	instanceId, err := strconv.ParseInt(parts[1], 10, 64)
	if err != nil {
		http.Error(w, "Bad Request", http.StatusBadRequest)
		return
	}

	if err := r.ParseForm(); err != nil {
		http.Error(w, "Bad Request", http.StatusBadRequest)
		return
	}
	inProgress := strings.EqualFold(r.FormValue("inProgress"), "true")

	var instance models.RoomInstance
	if err := db.DB.First(&instance, instanceId).Error; err != nil {
		http.Error(w, "Not Found", http.StatusNotFound)
		return
	}

	currentInstance, _ := hub.GetPlayerInstance(int(accountId))
	if instance.OwnerAccountId != int(accountId) && currentInstance != instanceId {
		http.Error(w, "Forbidden", http.StatusForbidden)
		return
	}
	instance.IsInProgress = inProgress
	db.DB.Save(&instance)

	hub.HubBroadcastRoomInstanceUpdate(instanceId)

	w.WriteHeader(http.StatusOK)
}

func RoomInstanceMarkPrivate(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	accountId, err := controllers.CurrentUserIDFromRequest(r)
	if err != nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	parts := strings.Split(strings.Trim(r.URL.Path, "/"), "/")
	if len(parts) < 3 {
		http.Error(w, "Bad Request", http.StatusBadRequest)
		return
	}
	instanceId, err := strconv.ParseInt(parts[1], 10, 64)
	if err != nil {
		http.Error(w, "Bad Request", http.StatusBadRequest)
		return
	}

	var instance models.RoomInstance
	if err := db.DB.First(&instance, instanceId).Error; err != nil {
		http.Error(w, "Not Found", http.StatusNotFound)
		return
	}
	currentInstance, _ := hub.GetPlayerInstance(int(accountId))
	if instance.OwnerAccountId != int(accountId) && currentInstance != instanceId {
		http.Error(w, "Forbidden", http.StatusForbidden)
		return
	}

	instance.IsPrivate = true
	db.DB.Save(&instance)

	hub.HubBroadcastRoomInstanceUpdate(instanceId)
	json.NewEncoder(w).Encode(instance)
}

const roomInstanceJoinResultFailed = 6

func RoomInstanceReportJoinResult(w http.ResponseWriter, r *http.Request) {
	parts := strings.Split(strings.Trim(r.URL.Path, "/"), "/")
	if len(parts) < 3 {
		http.Error(w, "Bad Request", http.StatusBadRequest)
		return
	}
	instanceId, err := strconv.ParseInt(parts[1], 10, 64)
	if err != nil {
		http.Error(w, "Bad Request", http.StatusBadRequest)
		return
	}

	if err := r.ParseForm(); err != nil {
		http.Error(w, "Bad Request", http.StatusBadRequest)
		return
	}
	result, _ := strconv.Atoi(r.FormValue("result"))

	if result != roomInstanceJoinResultFailed {
		w.WriteHeader(http.StatusOK)
		return
	}

	var instance models.RoomInstance
	if err := db.DB.First(&instance, instanceId).Error; err != nil {
		http.Error(w, "Not Found", http.StatusNotFound)
		return
	}

	if !instance.JoinDisabled {
		instance.JoinDisabled = true
		db.DB.Save(&instance)
		hub.HubBroadcastRoomInstanceUpdate(instanceId)

		discord.SendInstanceClosed(discord.InstanceClosedInfo{
			InstanceID: instance.Id,
			RoomName:   instance.Name,
			Result:     result,
		})
	}

	w.WriteHeader(http.StatusOK)
}

type roomInstanceListEntry struct {
	CreatedAt      time.Time `json:"createdAt"`
	IsFull         bool      `json:"isFull"`
	PlayerIds      []int     `json:"playerIds"`
	RoomId         int64     `json:"roomId"`
	RoomInstanceId int64     `json:"roomInstanceId"`
	SubRoomId      int64     `json:"subRoomId"`
}

func RoomAllowNewUsers(w http.ResponseWriter, r *http.Request) {
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

	allow := strings.EqualFold(r.FormValue("allowNewUsers"), "true")

	var instances []models.RoomInstance
	db.DB.Where("room_id = ? AND owner_account_id = ?", roomId, int(accountId)).Find(&instances)
	for _, inst := range instances {
		inst.AllowNewUsers = allow
		db.DB.Save(&inst)
		hub.HubBroadcastRoomInstanceUpdate(inst.Id)
	}

	roomSuccessResponse(w, room)
}

func RoomInstancesList(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	parts := strings.Split(strings.Trim(r.URL.Path, "/"), "/")
	if len(parts) < 3 {
		http.Error(w, "Bad Request", http.StatusBadRequest)
		return
	}
	roomId, err := strconv.ParseInt(parts[1], 10, 64)
	if err != nil {
		http.Error(w, "Bad Request", http.StatusBadRequest)
		return
	}

	var instances []models.RoomInstance
	db.DB.Where("room_id = ? AND is_private = ?", roomId, false).Find(&instances)

	out := make([]roomInstanceListEntry, 0, len(instances))
	for _, inst := range instances {
		players := hub.PlayersInInstance(inst.Id)
		if players == nil {
			players = []int{}
		}
		out = append(out, roomInstanceListEntry{
			CreatedAt:      inst.CreatedAt,
			IsFull:         inst.IsFull,
			PlayerIds:      players,
			RoomId:         inst.RoomId,
			RoomInstanceId: inst.Id,
			SubRoomId:      inst.SubRoomId,
		})
	}

	json.NewEncoder(w).Encode(out)
}
