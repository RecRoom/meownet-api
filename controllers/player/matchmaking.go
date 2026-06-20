package player

import (
	"crypto/rand"
	"encoding/binary"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"strings"
	"time"

	"gorm.io/gorm"
	"meow.net/controllers/hub"
	"meow.net/db"
	"meow.net/models"
	"meow.net/utils"
)

func isInstanceBanned(instanceID int64, accountID int) bool {
	if instanceID == 0 || accountID == 0 {
		return false
	}
	var count int64
	db.DB.Model(&models.InstanceBan{}).
		Where("instance_id = ? AND account_id = ? AND expires_at > ?", instanceID, uint(accountID), time.Now()).
		Count(&count)
	return count > 0
}

func isInstanceInvited(instanceId int64, accountId int) bool {
	if instanceId == 0 || accountId == 0 {
		return false
	}
	var count int64
	db.DB.Model(&models.InstanceInvite{}).
		Where("instance_id = ? AND account_id = ?", instanceId, accountId).
		Count(&count)
	return count > 0
}

func isAccountBanned(accountID int) bool {
	if accountID == 0 {
		return false
	}
	var count int64
	db.DB.Model(&models.AccountBan{}).
		Where("account_id = ? AND (expires_at IS NULL OR expires_at > ?)", accountID, time.Now()).
		Count(&count)
	return count > 0
}

const (
	defaultPhotonRegion = "us"
	defaultMaxPlayers   = 4
)

func writeMatchmakingError(w http.ResponseWriter, accountId int, errorCode models.MatchmakingErrorCode) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(hub.BuildSelfStatus(accountId, nil, int(errorCode)))
}

func instanceIsFull(instance *models.RoomInstance, accountId int, fallbackCapacity int) bool {
	capacity := instance.MaxCapacity
	if capacity == 0 {
		capacity = fallbackCapacity
	}
	if capacity == 0 {
		capacity = defaultMaxPlayers
	}
	return hub.LivePlayerCountInInstance(instance.Id, accountId) >= capacity
}

func randomPhotonRoomId() string {
	var b [4]byte
	rand.Read(b[:])
	return fmt.Sprintf("%d", binary.BigEndian.Uint32(b[:])>>1)
}

func GotoRoom(w http.ResponseWriter, r *http.Request) {
	tokenStr := utils.GetBearerToken(r)
	if tokenStr == "" {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}
	accountIdStr, err := utils.ParseSubFromJWT(tokenStr)
	if err != nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}
	accountId, err := strconv.Atoi(accountIdStr)
	if err != nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	parts := strings.Split(strings.TrimRight(r.URL.Path, "/"), "/")
	if len(parts) < 4 {
		http.Error(w, "Bad Request", http.StatusBadRequest)
		return
	}
	roomParam := parts[3]
	subRoomParam := ""
	if len(parts) >= 5 {
		subRoomParam = parts[4]
	}

	var roomData models.Room
	roomLower := strings.ToLower(roomParam)

	if roomLower != "dormroom" && isAccountBanned(accountId) {
		http.Error(w, "Account banned", http.StatusForbidden)
		return
	}

	if roomLower == "dormroom" {
		if err := db.DB.Where(
			"creator_account_id = ? AND (is_dorm = ? OR LOWER(name) = ?)",
			accountId, true, "dormroom",
		).First(&roomData).Error; err != nil {
			log.Printf("[GOTO] no dorm row found for accountId=%d, generating a new one...", accountId)

			var acc models.Account
			db.DB.First(&acc, accountId)

			roomData = models.Room{
				Name:                "@" + acc.Username + "'s Dorm",
				Description:         "Your personal room",
				CreatorAccountId:    accountId,
				ImageName:           "",
				State:               0,
				Accessibility:       0,
				SupportsLevelVoting: false,
				IsRRO:               false,
				IsDorm:              true,
				CloningAllowed:      false,
				SupportsVRLow:       true,
				SupportsMobile:      true,
				SupportsScreens:     true,
				SupportsWalkVR:      true,
				SupportsTeleportVR:  true,
				SupportsJuniors:     true,
				MaxPlayers:          4,
				PersistenceVersion:  1,
				UgcVersion:          1,
				WarningMask:         0,
				DisableMicAutoMute:  false,
				CreatedAt:           time.Now(),
			}
			db.DB.Create(&roomData)

			dormSubRoom := models.SubRoom{
				RoomId:           roomData.RoomId,
				Name:             "Home",
				Accessibility:    0,
				MaxPlayers:       4,
				SavedByAccountId: accountId,
				UnitySceneId:     "76d98498-60a1-430c-ab76-b54a29b7a163",
			}
			db.DB.Create(&dormSubRoom)

			db.DB.Create(&models.RoomRoleEntry{
				RoomId:      roomData.RoomId,
				AccountId:   accountId,
				InvitedRole: 0,
				Role:        255,
			})
		}
	} else if roomId, err := strconv.Atoi(roomParam); err == nil {
		if db.DB.First(&roomData, roomId).Error != nil {
			http.Error(w, "Room not found", http.StatusNotFound)
			return
		}
	} else {
		if db.DB.Where("LOWER(name) = ?", roomLower).First(&roomData).Error != nil {
			http.Error(w, "Room not found", http.StatusNotFound)
			return
		}
	}

	enterRoom(w, r, accountId, roomData, subRoomParam)
}

func enterRoom(w http.ResponseWriter, r *http.Request, accountId int, roomData models.Room, subRoomParam string) {
	var subRoom models.SubRoom
	subRoomQuery := db.DB.Where("room_id = ?", roomData.RoomId)
	if subRoomParam != "" {
		subRoomQuery = subRoomQuery.Where("LOWER(name) = ?", strings.ToLower(subRoomParam))
	}
	if err := subRoomQuery.First(&subRoom).Error; err != nil {
		log.Printf("[GOTO] no sub_room found for room_id=%d subroom=%s", roomData.RoomId, subRoomParam)
		http.Error(w, "Sub-room not found", http.StatusNotFound)
		return
	}

	location := strings.TrimSpace(subRoom.UnitySceneId)
	subRoomId := int64(subRoom.SubRoomId)
	maxCapacity := subRoom.MaxPlayers
	if maxCapacity == 0 {
		maxCapacity = defaultMaxPlayers
	}

	r.ParseForm()
	joinMode, _ := strconv.Atoi(r.FormValue("JoinMode"))
	wantPrivate := joinMode == 2 || roomData.IsDorm

	currentInstanceId, _ := hub.GetPlayerInstance(accountId)

	var instance models.RoomInstance

	if !wantPrivate {
		var candidates []models.RoomInstance
		db.DB.
			Where("room_id = ? AND is_private = false AND is_in_progress = false AND join_disabled = false AND id != ?",
				roomData.RoomId, currentInstanceId).
			Where("NOT EXISTS (SELECT 1 FROM instance_bans b WHERE b.instance_id = room_instances.id AND b.account_id = ? AND b.expires_at > ?)",
				uint(accountId), time.Now()).
			Order("created_at ASC").
			Find(&candidates)

		found := false
		for i := range candidates {
			c := &candidates[i]
			if hub.LivePlayerCountInInstance(c.Id, accountId) == 0 {
				continue
			}
			if instanceIsFull(c, accountId, maxCapacity) {
				continue
			}
			instance = *c
			found = true
			if instance.PhotonRoomId == "" {
				instance.PhotonRoomId = randomPhotonRoomId()
				db.DB.Save(&instance)
			}
			break
		}
		if !found {
			instance = models.RoomInstance{
				OwnerAccountId: accountId,
				RoomId:         int64(roomData.RoomId),
				SubRoomId:      subRoomId,
				Location:       location,
				PhotonRegionId: defaultPhotonRegion,
				PhotonRoomId:   randomPhotonRoomId(),
				Name:           roomData.Name,
				MaxCapacity:    maxCapacity,
				IsPrivate:      false,
			}
			db.DB.Create(&instance)
		}
	} else {
		instance = models.RoomInstance{
			OwnerAccountId: accountId,
			RoomId:         int64(roomData.RoomId),
			SubRoomId:      subRoomId,
			Location:       location,
			PhotonRegionId: defaultPhotonRegion,
			PhotonRoomId:   randomPhotonRoomId(),
			Name:           roomData.Name,
			MaxCapacity:    maxCapacity,
			IsPrivate:      true,
		}
		db.DB.Create(&instance)
	}

	hub.SetPlayerInstance(accountId, instance.Id)
	hub.PruneOwnedInstances(accountId, instance.Id)

	var visit models.RoomInteraction
	db.DB.Where("room_id = ? AND account_id = ?", roomData.RoomId, uint(accountId)).
		FirstOrCreate(&visit, models.RoomInteraction{RoomId: roomData.RoomId, AccountId: uint(accountId)})
	if !visit.Visited {
		visit.Visited = true
		db.DB.Save(&visit)
		db.DB.Model(&models.Room{}).Where("room_id = ?", roomData.RoomId).UpdateColumn("visitor_count", gorm.Expr("visitor_count + ?", 1))
	}
	db.DB.Model(&models.Room{}).Where("room_id = ?", roomData.RoomId).UpdateColumn("visit_count", gorm.Expr("visit_count + ?", 1))

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(hub.BuildSelfStatus(accountId, instance, 0))
}

func GotoEvent(w http.ResponseWriter, r *http.Request) {
	tokenStr := utils.GetBearerToken(r)
	if tokenStr == "" {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}
	accountIdStr, err := utils.ParseSubFromJWT(tokenStr)
	if err != nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}
	accountId, err := strconv.Atoi(accountIdStr)
	if err != nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	parts := strings.Split(strings.TrimRight(r.URL.Path, "/"), "/")
	if len(parts) < 4 {
		http.Error(w, "Bad Request", http.StatusBadRequest)
		return
	}
	eventId, err := strconv.ParseUint(parts[3], 10, 64)
	if err != nil {
		http.Error(w, "Bad Request", http.StatusBadRequest)
		return
	}

	var event models.PlayerEvent
	if err := db.DB.First(&event, eventId).Error; err != nil {
		http.Error(w, "Not Found", http.StatusNotFound)
		return
	}

	var roomData models.Room
	if err := db.DB.First(&roomData, event.RoomId).Error; err != nil {
		http.Error(w, "Room not found", http.StatusNotFound)
		return
	}

	if isAccountBanned(accountId) {
		http.Error(w, "Account banned", http.StatusForbidden)
		return
	}

	subRoomParam := ""
	if event.SubRoomId != nil {
		var subRoom models.SubRoom
		if db.DB.First(&subRoom, *event.SubRoomId).Error == nil {
			subRoomParam = subRoom.Name
		}
	}

	enterRoom(w, r, accountId, roomData, subRoomParam)
}

func GotoClub(w http.ResponseWriter, r *http.Request) {
	tokenStr := utils.GetBearerToken(r)
	if tokenStr == "" {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}
	accountIdStr, err := utils.ParseSubFromJWT(tokenStr)
	if err != nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}
	accountId, err := strconv.Atoi(accountIdStr)
	if err != nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	if isAccountBanned(accountId) {
		http.Error(w, "Account banned", http.StatusForbidden)
		return
	}

	parts := strings.Split(strings.TrimRight(r.URL.Path, "/"), "/")
	if len(parts) < 4 {
		http.Error(w, "Bad Request", http.StatusBadRequest)
		return
	}
	clubId, err := strconv.ParseInt(parts[3], 10, 64)
	if err != nil {
		http.Error(w, "Bad Request", http.StatusBadRequest)
		return
	}

	var club models.Club
	if err := db.DB.First(&club, clubId).Error; err != nil {
		http.Error(w, "Club not found", http.StatusNotFound)
		return
	}
	if club.ClubhouseRoomId == nil || *club.ClubhouseRoomId == 0 {
		http.Error(w, "Club has no clubhouse", http.StatusNotFound)
		return
	}

	var roomData models.Room
	if err := db.DB.First(&roomData, *club.ClubhouseRoomId).Error; err != nil {
		http.Error(w, "Clubhouse room not found", http.StatusNotFound)
		return
	}

	subRoomParam := ""
	if len(parts) >= 5 {
		subRoomParam = parts[4]
	}

	enterRoom(w, r, accountId, roomData, subRoomParam)
}

func GotoPlayer(w http.ResponseWriter, r *http.Request) {
	tokenStr := utils.GetBearerToken(r)
	if tokenStr == "" {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}
	accountIdStr, err := utils.ParseSubFromJWT(tokenStr)
	if err != nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}
	accountId, err := strconv.Atoi(accountIdStr)
	if err != nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	if isAccountBanned(accountId) {
		writeMatchmakingError(w, accountId, models.MMBanned)
		return
	}

	parts := strings.Split(strings.TrimRight(r.URL.Path, "/"), "/")
	targetPlayerIdStr := parts[len(parts)-1]
	targetPlayerId, err := strconv.Atoi(targetPlayerIdStr)
	if err != nil || targetPlayerId == 0 {
		http.Error(w, "Bad Request", http.StatusBadRequest)
		return
	}

	log.Printf("[GOTO] player=%d accountId=%d", targetPlayerId, accountId)

	instanceId, ok := hub.GetPlayerInstance(targetPlayerId)
	if !ok || instanceId == 0 {
		writeMatchmakingError(w, accountId, models.MMPlayerNotOnline)
		return
	}

	var instance models.RoomInstance
	if err := db.DB.First(&instance, instanceId).Error; err != nil {
		writeMatchmakingError(w, accountId, models.MMNoSuchGame)
		return
	}

	if instance.JoinDisabled {
		writeMatchmakingError(w, accountId, models.MMInstanceJoinNotPermitted)
		return
	}

	if isInstanceBanned(instance.Id, accountId) {
		writeMatchmakingError(w, accountId, models.MMBannedFromRoom)
		return
	}

	if instance.IsPrivate &&
		accountId != instance.OwnerAccountId &&
		!isInstanceInvited(instance.Id, accountId) {
		writeMatchmakingError(w, accountId, models.MMRoomInstanceIsPrivate)
		return
	}

	if !instance.AllowNewUsers && accountId != instance.OwnerAccountId {
		writeMatchmakingError(w, accountId, models.MMInstanceJoinNotPermitted)
		return
	}

	if instanceIsFull(&instance, accountId, defaultMaxPlayers) {
		writeMatchmakingError(w, accountId, models.MMInsufficientSpace)
		return
	}

	hub.SetPlayerInstance(accountId, instance.Id)
	hub.PruneOwnedInstances(accountId, instance.Id)

	if instance.RoomId > 0 {
		var visit models.RoomInteraction
		db.DB.Where("room_id = ? AND account_id = ?", instance.RoomId, uint(accountId)).
			FirstOrCreate(&visit, models.RoomInteraction{RoomId: uint(instance.RoomId), AccountId: uint(accountId)})
		if !visit.Visited {
			visit.Visited = true
			db.DB.Save(&visit)
		}
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(hub.BuildSelfStatus(accountId, instance, 0))
}

func GotoNone(w http.ResponseWriter, r *http.Request) {
	accountId := 0
	if tokenStr := utils.GetBearerToken(r); tokenStr != "" {
		if idStr, err := utils.ParseSubFromJWT(tokenStr); err == nil {
			accountId, _ = strconv.Atoi(idStr)
		}
	}

	hub.ClearPlayerInstance(accountId)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(hub.BuildSelfStatus(accountId, nil, 0))
}

func QuickPlay(w http.ResponseWriter, r *http.Request) {
	log.Printf("[QUICKPLAY] getandclear")
}

func NotifyDisconnect(w http.ResponseWriter, r *http.Request) {
	r.ParseForm()
	log.Printf("[NOTIFYDISCONNECT] playerId=%s roomInstanceId=%s",
		r.FormValue("PlayerId"), r.FormValue("RoomInstanceId"))
	w.WriteHeader(http.StatusOK)
}
