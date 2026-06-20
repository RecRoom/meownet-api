package admin

import (
	"crypto/rand"
	"encoding/binary"
	"fmt"
	"net/http"
	"strconv"

	"meow.net/controllers/hub"
	"meow.net/db"
	"meow.net/models"
)

type adminInstanceEntry struct {
	RoomInstanceId int64  `json:"roomInstanceId"`
	RoomId         int64  `json:"roomId"`
	SubRoomId      int64  `json:"subRoomId"`
	PhotonRoomId   string `json:"photonRoomId"`
	PhotonRegionId string `json:"photonRegionId"`
	IsPrivate      bool   `json:"isPrivate"`
	IsFull         bool   `json:"isFull"`
	IsInProgress   bool   `json:"isInProgress"`
	PlayerIds      []int  `json:"playerIds"`
	PlayerCount    int    `json:"playerCount"`
}

func ListInstances(w http.ResponseWriter, r *http.Request) {
	if !requireAdmin(w, r) {
		return
	}

	var instances []models.RoomInstance
	db.DB.Find(&instances)

	out := make([]adminInstanceEntry, 0, len(instances))
	for _, inst := range instances {
		players := hub.PlayersInInstance(inst.Id)
		if players == nil {
			players = []int{}
		}
		out = append(out, adminInstanceEntry{
			RoomInstanceId: inst.Id,
			RoomId:         inst.RoomId,
			SubRoomId:      inst.SubRoomId,
			PhotonRoomId:   inst.PhotonRoomId,
			PhotonRegionId: inst.PhotonRegionId,
			IsPrivate:      inst.IsPrivate,
			IsFull:         inst.IsFull,
			IsInProgress:   inst.IsInProgress,
			PlayerIds:      players,
			PlayerCount:    len(players),
		})
	}

	writeJSON(w, http.StatusOK, out)
}

type killInstanceResult struct {
	Success          bool  `json:"success"`
	KilledInstanceId int64 `json:"killedInstanceId"`
	DestInstanceId   int64 `json:"destInstanceId"`
	MovedPlayerIds   []int `json:"movedPlayerIds"`
	MovedCount       int   `json:"movedCount"`
	KilledDeleted    bool  `json:"killedDeleted"`
}

func KillInstance(w http.ResponseWriter, r *http.Request) {
	if !requireAdmin(w, r) {
		return
	}

	instanceId, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
	if err != nil || instanceId <= 0 {
		http.Error(w, "bad instance id", http.StatusBadRequest)
		return
	}

	var killed models.RoomInstance
	if err := db.DB.First(&killed, instanceId).Error; err != nil {
		http.Error(w, "instance not found", http.StatusNotFound)
		return
	}

	if !killed.JoinDisabled {
		killed.JoinDisabled = true
		db.DB.Model(&killed).Update("join_disabled", true)
		hub.HubBroadcastRoomInstanceUpdate(killed.Id)
	}

	players := hub.PlayersInInstance(killed.Id)
	if len(players) == 0 {
		db.DB.Delete(&models.RoomInstance{}, killed.Id)
		writeJSON(w, http.StatusOK, killInstanceResult{
			Success:          true,
			KilledInstanceId: killed.Id,
			MovedPlayerIds:   []int{},
			KilledDeleted:    true,
		})
		return
	}

	dest, ok := findDestinationInstance(killed, len(players))
	if !ok {
		dest = models.RoomInstance{
			OwnerAccountId: killed.OwnerAccountId,
			RoomId:         killed.RoomId,
			SubRoomId:      killed.SubRoomId,
			Location:       killed.Location,
			PhotonRegionId: killed.PhotonRegionId,
			PhotonRoomId:   randomPhotonRoomId(),
			Name:           killed.Name,
			MaxCapacity:    destCapacity(killed, len(players)),
			IsPrivate:      false,
		}
		db.DB.Create(&dest)
	}

	moved := make([]int, 0, len(players))
	for _, pid := range players {
		hub.SetPlayerInstance(pid, dest.Id)
		hub.PruneOwnedInstances(pid, dest.Id)
		hub.HubSendRoomInstanceToPlayer(pid, dest.Id)
		moved = append(moved, pid)
	}

	killedDeleted := false
	if db.DB.First(&models.RoomInstance{}, killed.Id).Error != nil {
		killedDeleted = true
	} else if len(hub.PlayersInInstance(killed.Id)) == 0 {
		db.DB.Delete(&models.RoomInstance{}, killed.Id)
		killedDeleted = true
	}

	writeJSON(w, http.StatusOK, killInstanceResult{
		Success:          true,
		KilledInstanceId: killed.Id,
		DestInstanceId:   dest.Id,
		MovedPlayerIds:   moved,
		MovedCount:       len(moved),
		KilledDeleted:    killedDeleted,
	})
}

func destCapacity(killed models.RoomInstance, need int) int {
	capacity := killed.MaxCapacity
	var subRoom models.SubRoom
	if db.DB.First(&subRoom, killed.SubRoomId).Error == nil && subRoom.MaxPlayers > 0 {
		capacity = subRoom.MaxPlayers
	}
	if capacity < need {
		capacity = need
	}
	return capacity
}

func findDestinationInstance(killed models.RoomInstance, need int) (models.RoomInstance, bool) {
	var candidates []models.RoomInstance
	db.DB.
		Where("room_id = ? AND sub_room_id = ? AND is_private = false AND is_in_progress = false AND join_disabled = false AND id != ?",
			killed.RoomId, killed.SubRoomId, killed.Id).
		Order("created_at ASC").
		Find(&candidates)

	for i := range candidates {
		c := &candidates[i]
		capacity := c.MaxCapacity
		if capacity <= 0 {
			capacity = need
		}
		if capacity-hub.LivePlayerCountInInstance(c.Id, 0) < need {
			continue
		}
		if c.PhotonRoomId == "" {
			c.PhotonRoomId = randomPhotonRoomId()
			db.DB.Save(c)
		}
		return *c, true
	}
	return models.RoomInstance{}, false
}

func randomPhotonRoomId() string {
	var b [4]byte
	rand.Read(b[:])
	return fmt.Sprintf("%d", binary.BigEndian.Uint32(b[:])>>1)
}
