package hub

import (
	"encoding/json"
	"errors"
	"log"
	"runtime/debug"
	"sync"
	"time"

	"github.com/gorilla/websocket"
	"meow.net/controllers/reputation"
	"meow.net/db"
	"meow.net/models"
)

func recoverGoroutine(label string) {
	if r := recover(); r != nil {
		log.Printf("[HUB] recovered panic in %s: %v\n%s", label, r, debug.Stack())
	}
}

const lastOnlineFormat = "2006-01-02T15:04:05"
const lastOnlineZero = "0001-01-01T00:00:00"

func LastOnlineString(playerId int) string {
	if hubIsOnline(playerId) {
		return time.Now().UTC().Format(lastOnlineFormat)
	}
	var row struct {
		LastOnline *time.Time
	}
	if err := db.DB.Model(&models.Account{}).
		Select("last_online").
		Where("account_id = ?", playerId).
		Scan(&row).Error; err != nil || row.LastOnline == nil {
		return lastOnlineZero
	}
	return row.LastOnline.UTC().Format(lastOnlineFormat)
}

func ClearLoginLock(playerId int) {
	if playerId == 0 {
		return
	}
	if err := db.DB.Model(&models.PlayerState{}).
		Where("account_id = ?", playerId).
		UpdateColumn("login_lock_token", nil).Error; err != nil {
		log.Printf("[HUB] failed to clear login lock pid=%d: %v", playerId, err)
	}
}

func MarkPlayerOffline(playerId int) {
	now := time.Now().UTC()
	if err := db.DB.Model(&models.Account{}).
		Where("account_id = ?", playerId).
		UpdateColumn("last_online", now).Error; err != nil {
		log.Printf("[HUB] failed to write last_online pid=%d: %v", playerId, err)
	}
}

const (
	maxConnsPerPlayer = 5
	DefaultAppVersion = "20210827"
	wsWriteTimeout    = 10 * time.Second
	sendQueueSize     = 64
	pingInterval      = 15 * time.Second
)

var (
	errConnClosed   = errors.New("hub: connection closed")
	errSlowConsumer = errors.New("hub: send queue full")
	pingFrame       = []byte(`{"type":6}` + "\x1e")
)

type connState struct {
	playerId  int
	conn      *websocket.Conn
	playerIds map[int]bool
	createdAt time.Time

	send      chan []byte
	done      chan struct{}
	closeOnce sync.Once
}

func (s *connState) writeFrame(data []byte) error {
	select {
	case <-s.done:
		return errConnClosed
	case s.send <- data:
		return nil
	default:
		log.Printf("[HUB] slow consumer pid=%d, closing connection", s.playerId)
		s.close()
		return errSlowConsumer
	}
}

func (s *connState) close() {
	s.closeOnce.Do(func() {
		close(s.done)
		s.conn.Close()
	})
}

func (s *connState) writePump() {
	defer func() {
		if r := recover(); r != nil {
			log.Printf("[HUB] writePump panic pid=%d: %v\n%s", s.playerId, r, debug.Stack())
			s.close()
		}
	}()
	ticker := time.NewTicker(pingInterval)
	defer ticker.Stop()
	for {
		select {
		case <-s.done:
			return
		case data := <-s.send:
			s.conn.SetWriteDeadline(time.Now().Add(wsWriteTimeout))
			if err := s.conn.WriteMessage(websocket.TextMessage, data); err != nil {
				log.Printf("[HUB] write err pid=%d: %v", s.playerId, err)
				s.close()
				return
			}
		case <-ticker.C:
			s.conn.SetWriteDeadline(time.Now().Add(wsWriteTimeout))
			if err := s.conn.WriteMessage(websocket.TextMessage, pingFrame); err != nil {
				log.Printf("[HUB] ping write err pid=%d: %v", s.playerId, err)
				s.close()
				return
			}
		}
	}
}

var hub = struct {
	sync.RWMutex
	conns map[int]map[*connState]struct{}
	subs  map[int]map[*connState]struct{}
}{
	conns: map[int]map[*connState]struct{}{},
	subs:  map[int]map[*connState]struct{}{},
}

func hubRegister(s *connState) bool {
	s.createdAt = time.Now()

	hub.Lock()
	set := hub.conns[s.playerId]
	var evict *connState
	if s.playerId != 0 && len(set) >= maxConnsPerPlayer {
		for c := range set {
			if evict == nil || c.createdAt.Before(evict.createdAt) {
				evict = c
			}
		}
		delete(set, evict)
	}
	if hub.conns[s.playerId] == nil {
		hub.conns[s.playerId] = map[*connState]struct{}{}
	}
	hub.conns[s.playerId][s] = struct{}{}
	hub.Unlock()

	if evict != nil {
		log.Printf("[HUB] evicting oldest conn pid=%d to admit newer connection", s.playerId)
		evict.conn.WriteControl(websocket.CloseMessage,
			websocket.FormatCloseMessage(websocket.CloseTryAgainLater, "replaced by newer connection"),
			time.Now().Add(time.Second))
		evict.close()
	}
	return true
}

func hubUnregister(s *connState) (wentOffline bool) {
	hub.Lock()
	defer hub.Unlock()
	if set := hub.conns[s.playerId]; set != nil {
		delete(set, s)
		if len(set) == 0 {
			delete(hub.conns, s.playerId)
			wentOffline = true
		}
	}
	for id := range s.playerIds {
		if set := hub.subs[id]; set != nil {
			delete(set, s)
			if len(set) == 0 {
				delete(hub.subs, id)
			}
		}
	}
	return
}

func hubIsOnline(playerId int) bool {
	hub.RLock()
	defer hub.RUnlock()
	return len(hub.conns[playerId]) > 0
}

func HubIsOnline(playerId int) bool { return hubIsOnline(playerId) }

func GetOnlinePlayers() []int {
	hub.RLock()
	defer hub.RUnlock()
	var players []int
	for playerId, conns := range hub.conns {
		if playerId != 0 && len(conns) > 0 {
			players = append(players, playerId)
		}
	}
	return players
}

func hubSubscribers(playerId int) []*connState {
	hub.RLock()
	defer hub.RUnlock()
	set := hub.subs[playerId]
	out := make([]*connState, 0, len(set))
	for s := range set {
		out = append(out, s)
	}
	return out
}

func NotifFrame(id interface{}, msg interface{}) []byte {
	inner, _ := json.Marshal(map[string]interface{}{
		"Id":  id,
		"Msg": msg,
	})
	frame, _ := json.Marshal(map[string]interface{}{
		"type":      1,
		"target":    "Notification",
		"arguments": []interface{}{string(inner)},
	})
	return append(frame, 0x1e)
}

func HubSendToPlayer(playerId int, frame []byte) {
	hub.RLock()
	set := hub.conns[playerId]
	targets := make([]*connState, 0, len(set))
	for s := range set {
		targets = append(targets, s)
	}
	hub.RUnlock()
	for _, s := range targets {
		if err := s.writeFrame(frame); err != nil {
			log.Printf("[HUB] send err pid=%d: %v", playerId, err)
		}
	}
}

func HubBroadcastToAll(frame []byte) int {
	hub.RLock()
	targets := make([]*connState, 0, len(hub.conns))
	for playerId, set := range hub.conns {
		if playerId == 0 {
			continue
		}
		for s := range set {
			targets = append(targets, s)
		}
	}
	hub.RUnlock()

	for _, s := range targets {
		if err := s.writeFrame(frame); err != nil {
			log.Printf("[HUB] broadcast write err pid=%d: %v", s.playerId, err)
		}
	}
	return len(targets)
}

type playerStatus struct {
	StatusVisibility int
	VrMovementMode   int
}

func loadPlayerStatus(accountId int) playerStatus {
	ps := playerStatus{StatusVisibility: 0, VrMovementMode: 1}
	if accountId == 0 {
		return ps
	}
	var st models.PlayerState
	if err := db.DB.Select("status_visibility, vr_movement_mode").
		First(&st, accountId).Error; err == nil {
		ps.StatusVisibility = st.StatusVisibility
		ps.VrMovementMode = st.VrMovementMode
	}
	return ps
}

func playerStatusBase(accountId int, isOnline bool, roomInstance interface{}, errorCode int) map[string]interface{} {
	ps := loadPlayerStatus(accountId)
	return map[string]interface{}{
		"playerId":         accountId,
		"statusVisibility": ps.StatusVisibility,
		"deviceClass":      0,
		"vrMovementMode":   ps.VrMovementMode,
		"roomInstance":     roomInstance,
		"isOnline":         isOnline,
		"appVersion":       DefaultAppVersion,
		"errorCode":        errorCode,
		"lastOnline":       LastOnlineString(accountId),
		"clientJoinData":   ``,
	}
}

func BuildSelfStatus(accountId int, roomInstance interface{}, errorCode int) map[string]interface{} {
	return playerStatusBase(accountId, true, roomInstance, errorCode)
}

func BuildPresence(playerId int) map[string]interface{} {
	isOnline := hubIsOnline(playerId)
	var roomInstance interface{} = nil
	if isOnline {
		if instanceId, ok := GetPlayerInstance(playerId); ok && instanceId > 0 {
			var instance models.RoomInstance
			if err := db.DB.First(&instance, instanceId).Error; err == nil {
				roomInstance = instance
			}
		}
	}
	return playerStatusBase(playerId, isOnline, roomInstance, 0)
}

func BuildPresenceFor(viewerId, playerId int) map[string]interface{} {
	return BuildPresence(playerId)
}

func BuildPresenceForBatch(viewerId int, ids []int) []map[string]interface{} {
	results := make([]map[string]interface{}, 0, len(ids))
	if len(ids) == 0 {
		return results
	}

	uniq := make(map[int]struct{}, len(ids))
	for _, id := range ids {
		uniq[id] = struct{}{}
	}
	idList := make([]int, 0, len(uniq))
	for id := range uniq {
		idList = append(idList, id)
	}

	online := make(map[int]bool, len(uniq))
	instanceOf := make(map[int]int64, len(uniq))
	instanceIdSet := make(map[int64]struct{})
	for id := range uniq {
		if hubIsOnline(id) {
			online[id] = true
			if inst, ok := GetPlayerInstance(id); ok && inst > 0 {
				instanceOf[id] = inst
				instanceIdSet[inst] = struct{}{}
			}
		}
	}

	states := make(map[int]playerStatus, len(idList))
	var stateRows []models.PlayerState
	db.DB.Select("account_id, status_visibility, vr_movement_mode").
		Where("account_id IN ?", idList).Find(&stateRows)
	for _, r := range stateRows {
		states[int(r.AccountID)] = playerStatus{StatusVisibility: r.StatusVisibility, VrMovementMode: r.VrMovementMode}
	}

	lastOnline := make(map[int]string, len(idList))
	offline := make([]int, 0, len(idList))
	for _, id := range idList {
		if !online[id] {
			offline = append(offline, id)
		}
	}
	if len(offline) > 0 {
		var rows []struct {
			AccountID  uint
			LastOnline *time.Time
		}
		db.DB.Model(&models.Account{}).
			Select("account_id, last_online").
			Where("account_id IN ?", offline).Scan(&rows)
		for _, r := range rows {
			if r.LastOnline != nil {
				lastOnline[int(r.AccountID)] = r.LastOnline.UTC().Format(lastOnlineFormat)
			}
		}
	}

	instances := make(map[int64]models.RoomInstance, len(instanceIdSet))
	if len(instanceIdSet) > 0 {
		instIds := make([]int64, 0, len(instanceIdSet))
		for iid := range instanceIdSet {
			instIds = append(instIds, iid)
		}
		var rows []models.RoomInstance
		db.DB.Where("id IN ?", instIds).Find(&rows)
		for _, r := range rows {
			instances[r.Id] = r
		}
	}

	for _, id := range ids {
		isOnline := online[id]
		ps, ok := states[id]
		if !ok {
			ps = playerStatus{StatusVisibility: 0, VrMovementMode: 1}
		}
		var roomInstance interface{} = nil
		if isOnline {
			if inst, ok := instanceOf[id]; ok {
				if ri, ok2 := instances[inst]; ok2 {
					roomInstance = ri
				}
			}
		}
		var last string
		switch {
		case isOnline:
			last = time.Now().UTC().Format(lastOnlineFormat)
		case lastOnline[id] != "":
			last = lastOnline[id]
		default:
			last = lastOnlineZero
		}
		p := map[string]interface{}{
			"playerId":         id,
			"statusVisibility": ps.StatusVisibility,
			"deviceClass":      0,
			"vrMovementMode":   ps.VrMovementMode,
			"roomInstance":     roomInstance,
			"isOnline":         isOnline,
			"appVersion":       DefaultAppVersion,
			"errorCode":        0,
			"lastOnline":       last,
			"clientJoinData":   ``,
		}
		results = append(results, p)
	}
	return results
}

func FriendIDs(playerId int) []int {
	if playerId == 0 {
		return nil
	}
	var rels []models.Relationship
	if err := db.DB.Select("requester_id, target_id").
		Where("relationship_type = ? AND (requester_id = ? OR target_id = ?)",
			models.RelationshipFriend, playerId, playerId).
		Find(&rels).Error; err != nil {
		log.Printf("[HUB] friend lookup err pid=%d: %v", playerId, err)
		return nil
	}
	ids := make([]int, 0, len(rels))
	for _, rel := range rels {
		if int(rel.RequesterID) == playerId {
			ids = append(ids, int(rel.TargetID))
		} else {
			ids = append(ids, int(rel.RequesterID))
		}
	}
	return ids
}

func HubBroadcastPresence(playerId int) {
	frame := NotifFrame("PresenceUpdate", BuildPresence(playerId))

	seen := map[*connState]struct{}{}
	send := func(conns []*connState) {
		for _, s := range conns {
			if _, dup := seen[s]; dup {
				continue
			}
			seen[s] = struct{}{}
			if err := s.writeFrame(frame); err != nil {
				log.Printf("[HUB] presence write err pid=%d: %v", playerId, err)
			}
		}
	}

	send(hubSubscribers(playerId))
	for _, fid := range FriendIDs(playerId) {
		send(playerConns(fid))
	}
}

func playerConns(playerId int) []*connState {
	hub.RLock()
	set := hub.conns[playerId]
	out := make([]*connState, 0, len(set))
	for s := range set {
		out = append(out, s)
	}
	hub.RUnlock()
	return out
}

func HubBroadcastRoomInstance(instanceId int64) {
	if instanceId <= 0 {
		return
	}
	var instance models.RoomInstance
	if err := db.DB.First(&instance, instanceId).Error; err != nil {
		return
	}
	if instance.MaxCapacity > 0 {
		instance.IsFull = LivePlayerCountInInstance(instanceId, 0) >= instance.MaxCapacity
	}

	frame := NotifFrame("RoomInstanceUpdate", instance)
	seen := map[*connState]struct{}{}
	send := func(conns []*connState) {
		for _, s := range conns {
			if _, dup := seen[s]; dup {
				continue
			}
			seen[s] = struct{}{}
			if err := s.writeFrame(frame); err != nil {
				log.Printf("[HUB] room instance write err inst=%d: %v", instanceId, err)
			}
		}
	}

	send(playerConns(instance.OwnerAccountId))
	for _, pid := range PlayersInInstance(instanceId) {
		send(playerConns(pid))
		send(hubSubscribers(pid))
	}
}

func HubBroadcastRoomInstanceUpdate(instanceId int64) {
	var instance models.RoomInstance
	if err := db.DB.First(&instance, instanceId).Error; err != nil {
		return
	}
	frame := NotifFrame("RoomInstanceUpdate", instance)
	HubSendToPlayer(instance.OwnerAccountId, frame)

	var room models.Room
	if err := db.DB.Preload("SubRooms").Preload("Roles").Preload("Tags").
		First(&room, instance.RoomId).Error; err != nil {
		return
	}
	rooms := []models.Room{room}
	initRoomSlices(rooms)
	HubSendToPlayer(instance.OwnerAccountId, NotifFrame("RoomUpdate", rooms[0]))
}

func HubSendRoomInstanceToPlayer(playerId int, instanceId int64) {
	var instance models.RoomInstance
	if err := db.DB.First(&instance, instanceId).Error; err != nil {
		return
	}
	HubSendToPlayer(playerId, NotifFrame("RoomInstanceUpdate", instance))

	var room models.Room
	if err := db.DB.Preload("SubRooms").Preload("Roles").Preload("Tags").
		First(&room, instance.RoomId).Error; err != nil {
		return
	}
	rooms := []models.Room{room}
	initRoomSlices(rooms)
	HubSendToPlayer(playerId, NotifFrame("RoomUpdate", rooms[0]))
}

func (s *connState) updateSubscriptions(ids []int) []int {
	newSet := make(map[int]bool, len(ids))
	for _, id := range ids {
		newSet[id] = true
	}

	var added, removed []int
	for id := range newSet {
		if !s.playerIds[id] {
			added = append(added, id)
		}
	}
	for id := range s.playerIds {
		if !newSet[id] {
			removed = append(removed, id)
		}
	}
	s.playerIds = newSet

	if len(added) == 0 && len(removed) == 0 {
		return nil
	}

	hub.Lock()
	for _, id := range added {
		if hub.subs[id] == nil {
			hub.subs[id] = map[*connState]struct{}{}
		}
		hub.subs[id][s] = struct{}{}
	}
	for _, id := range removed {
		if set := hub.subs[id]; set != nil {
			delete(set, s)
			if len(set) == 0 {
				delete(hub.subs, id)
			}
		}
	}
	hub.Unlock()
	return added
}

func (s *connState) sendInitialState() {
	pid := s.playerId
	if pid == 0 {
		return
	}

	var selfAccount models.SelfAccount
	if db.DB.First(&selfAccount, pid).Error == nil {
		selfAccount.AvailableUsernameChanges = 3
		s.writeFrame(NotifFrame("SelfAccountUpdate", selfAccount))
		s.writeFrame(NotifFrame("AccountUpdate", selfAccount.Account))
	}

	var prog models.Progression
	db.DB.Where(models.Progression{AccountID: uint(pid)}).
		Attrs(models.Progression{Level: 1, XP: 0}).
		FirstOrCreate(&prog)
	s.writeFrame(NotifFrame("PlayerProgressionLevelUpdate", map[string]interface{}{
		"PlayerId": pid,
		"Level":    prog.Level,
		"XP":       prog.XP,
	}))

	s.writeFrame(NotifFrame("ReputationUpdate", reputation.Build(uint(pid))))

	if instanceId, ok := GetPlayerInstance(pid); ok && instanceId > 0 {
		var instance models.RoomInstance
		if db.DB.First(&instance, instanceId).Error == nil {
			s.writeFrame(NotifFrame("RoomInstanceUpdate", instance))
			var room models.Room
			if db.DB.Preload("SubRooms").Preload("Roles").Preload("Tags").
				First(&room, instance.RoomId).Error == nil {
				rooms := []models.Room{room}
				initRoomSlices(rooms)
				s.writeFrame(NotifFrame("RoomUpdate", rooms[0]))
			}
		}
	}

	s.writeFrame(NotifFrame("PresenceUpdate", BuildPresence(pid)))
	for _, fid := range FriendIDs(pid) {
		if hubIsOnline(fid) {
			s.writeFrame(NotifFrame("PresenceUpdate", BuildPresence(fid)))
		}
	}
}

func HubSendProgressionUpdate(playerId int, level int, xp int) {
	frame := NotifFrame("PlayerProgressionLevelUpdate", map[string]any{
		"PlayerId": playerId,
		"Level":    level,
		"XP":       xp,
	})
	HubSendToPlayer(playerId, frame)
	for _, s := range hubSubscribers(playerId) {
		if err := s.writeFrame(frame); err != nil {
			log.Printf("[HUB] progression write err pid=%d: %v", playerId, err)
		}
	}
}

func HubBroadcastReputationUpdate(playerId int) {
	frame := NotifFrame("ReputationUpdate", reputation.Build(uint(playerId)))
	HubSendToPlayer(playerId, frame)
	for _, s := range hubSubscribers(playerId) {
		if err := s.writeFrame(frame); err != nil {
			log.Printf("[HUB] reputation write err pid=%d: %v", playerId, err)
		}
	}
}

func HubKickPlayer(accountID int) {
	hub.RLock()
	set := hub.conns[accountID]
	targets := make([]*connState, 0, len(set))
	for s := range set {
		targets = append(targets, s)
	}
	hub.RUnlock()
	for _, s := range targets {
		s.conn.WriteControl(websocket.CloseMessage,
			websocket.FormatCloseMessage(websocket.ClosePolicyViolation, "banned"),
			time.Now().Add(time.Second))
		s.close()
	}
}

func HubBroadcastAccountUpdate(accountId int) {
	var acc models.Account
	if err := db.DB.First(&acc, accountId).Error; err != nil {
		return
	}
	frame := NotifFrame("AccountUpdate", acc)
	for _, s := range hubSubscribers(accountId) {
		if err := s.writeFrame(frame); err != nil {
			log.Printf("[HUB] account update write err pid=%d: %v", accountId, err)
		}
	}
}

func initRoomSlices(rooms []models.Room) {
	for i := range rooms {
		if rooms[i].Roles == nil {
			rooms[i].Roles = []models.RoomRoleEntry{}
		}
		if rooms[i].Tags == nil {
			rooms[i].Tags = []models.RoomTag{}
		}
		if rooms[i].SubRooms == nil {
			rooms[i].SubRooms = []models.SubRoom{}
		}
		if rooms[i].LoadScreens == nil {
			rooms[i].LoadScreens = []interface{}{}
		}
		if rooms[i].PromoImages == nil {
			rooms[i].PromoImages = []interface{}{}
		}
		if rooms[i].PromoExternalContent == nil {
			rooms[i].PromoExternalContent = []interface{}{}
		}
	}
}
