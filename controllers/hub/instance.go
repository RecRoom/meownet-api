package hub

import (
	"log"
	"sync"
	"time"

	"meow.net/db"
	"meow.net/models"
)

var (
	playerCurrentInstance   = map[int]int64{}
	playerInstanceSince     = map[int]time.Time{}
	playerCurrentInstanceMu sync.RWMutex
)

const (
	instanceReapInterval = 30 * time.Second
	instanceReapGrace    = 90 * time.Second
)

func SetPlayerInstance(accountId int, instanceId int64) {
	playerCurrentInstanceMu.Lock()
	oldId := playerCurrentInstance[accountId]
	playerCurrentInstance[accountId] = instanceId
	playerInstanceSince[accountId] = time.Now()
	playerCurrentInstanceMu.Unlock()
	if oldId > 0 && oldId != instanceId {
		leftInstance(oldId, accountId)
	}
	HubBroadcastPresence(accountId)
	if instanceId > 0 {
		HubBroadcastRoomInstanceUpdate(instanceId)
		var prog models.Progression
		db.DB.Where(models.Progression{AccountID: uint(accountId)}).
			Attrs(models.Progression{Level: 1, XP: 0}).
			FirstOrCreate(&prog)
		HubSendProgressionUpdate(accountId, prog.Level, prog.XP)
	}
}

func GetPlayerInstance(accountId int) (int64, bool) {
	playerCurrentInstanceMu.RLock()
	id, ok := playerCurrentInstance[accountId]
	playerCurrentInstanceMu.RUnlock()
	return id, ok
}

func PlayerCurrentRoomID(accountId int) (int64, bool) {
	instanceId, ok := GetPlayerInstance(accountId)
	if !ok || instanceId <= 0 {
		return 0, false
	}
	var inst models.RoomInstance
	if err := db.DB.Select("room_id").First(&inst, instanceId).Error; err != nil {
		return 0, false
	}
	return inst.RoomId, true
}

func ClearPlayerInstance(accountId int) {
	playerCurrentInstanceMu.Lock()
	oldId := playerCurrentInstance[accountId]
	playerCurrentInstance[accountId] = 0
	delete(playerInstanceSince, accountId)
	playerCurrentInstanceMu.Unlock()
	if oldId > 0 {
		leftInstance(oldId, accountId)
	}
	HubBroadcastPresence(accountId)
}

func leftInstance(instanceId int64, leaverId int) {
	if !deleteInstanceIfEmpty(instanceId, leaverId) {
		HubBroadcastRoomInstance(instanceId)
	}
}

func deleteInstanceIfEmpty(instanceId int64, exceptAccountId int) bool {
	if instanceId <= 0 {
		return false
	}
	if instanceInUseByOthers(instanceId, exceptAccountId) {
		return false
	}
	db.DB.Delete(&models.RoomInstance{}, instanceId)
	return true
}

func instanceInUseByOthers(instanceId int64, exceptId int) bool {
	playerCurrentInstanceMu.RLock()
	defer playerCurrentInstanceMu.RUnlock()
	for pid, id := range playerCurrentInstance {
		if pid != exceptId && id == instanceId {
			return true
		}
	}
	return false
}

func PlayersInInstance(instanceId int64) []int {
	playerCurrentInstanceMu.RLock()
	defer playerCurrentInstanceMu.RUnlock()
	var ids []int
	for pid, id := range playerCurrentInstance {
		if id == instanceId {
			ids = append(ids, pid)
		}
	}
	return ids
}

func PlayerCountInInstance(instanceId int64, ignoreAccountId int) int {
	playerCurrentInstanceMu.RLock()
	defer playerCurrentInstanceMu.RUnlock()
	count := 0
	for pid, id := range playerCurrentInstance {
		if id == instanceId && pid != ignoreAccountId {
			count++
		}
	}
	return count
}

func LivePlayerCountInInstance(instanceId int64, ignoreAccountId int) int {
	count := 0
	for _, pid := range PlayersInInstance(instanceId) {
		if pid != ignoreAccountId && hubIsOnline(pid) {
			count++
		}
	}
	return count
}

func PruneOwnedInstances(accountId int, keepInstanceId int64) {
	var instances []models.RoomInstance
	db.DB.Where("owner_account_id = ? AND id != ?", accountId, keepInstanceId).Find(&instances)
	for _, inst := range instances {
		if !instanceInUseByOthers(inst.Id, accountId) {
			db.DB.Delete(&inst)
		}
	}
}

func StartInstanceReaper() {
	go func() {
		defer recoverGoroutine("instanceReaper")
		ticker := time.NewTicker(instanceReapInterval)
		defer ticker.Stop()
		for range ticker.C {
			reapStaleInstanceMembers()
		}
	}()
}

func reapStaleInstanceMembers() {
	now := time.Now()
	type member struct {
		pid  int
		inst int64
	}
	var candidates []member
	playerCurrentInstanceMu.RLock()
	for pid, id := range playerCurrentInstance {
		if id <= 0 {
			continue
		}
		if since, ok := playerInstanceSince[pid]; ok && now.Sub(since) < instanceReapGrace {
			continue
		}
		candidates = append(candidates, member{pid, id})
	}
	playerCurrentInstanceMu.RUnlock()

	for _, m := range candidates {
		if hubIsOnline(m.pid) {
			continue
		}
		log.Printf("[REAP] clearing ghost pid=%d inst=%d (presence ws gone)", m.pid, m.inst)
		ClearPlayerInstance(m.pid)
	}
}
