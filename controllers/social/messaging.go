package social

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"strings"
	"time"

	"gorm.io/gorm/clause"
	"meow.net/controllers"
	"meow.net/controllers/hub"
	"meow.net/db"
	"meow.net/models"
	"meow.net/utils"
)

func isReservedMessageType(t int) bool {
	switch models.MessageType(t) {
	case models.MessageTypeTextMessage,
		models.MessageTypeGameInvite,
		models.MessageTypeGameInviteV2,
		models.MessageTypeRequestGameInvite,
		models.MessageTypePartyUpRequest,
		models.MessageTypeFriendIntroduction:
		return false
	}
	return true
}

var liveOnlyMessageTypes = []int{
	int(models.MessageTypeGameInvite),
	int(models.MessageTypeGameInviteDeclined),
	int(models.MessageTypePartyActivitySwitch),
	int(models.MessageTypePartyActivitySwitchV2),
	int(models.MessageTypeGameInviteV2),
	int(models.MessageTypeRequestGameInvite),
	int(models.MessageTypeRequestGameInviteDeclined),
	int(models.MessageTypePartyUpRequest),
}

func isLiveOnlyMessageType(t int) bool {
	for _, et := range liveOnlyMessageTypes {
		if et == t {
			return true
		}
	}
	return false
}

func writeRateLimited(w http.ResponseWriter) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusTooManyRequests)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"error":   "rate limit exceeded",
		"success": false,
	})
}

func writeJsonEnvelope(w http.ResponseWriter, value interface{}) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"error":   "",
		"success": true,
		"value":   value,
	})
}

func Messages(w http.ResponseWriter, r *http.Request) {
	currentUserID, err := CurrentUserIDFromRequest(r)
	if err != nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	var messages []models.Message
	db.DB.Where("to_player_id = ? AND type NOT IN ?", currentUserID, liveOnlyMessageTypes).
		Order("sent_time asc").Find(&messages)
	if messages == nil {
		messages = []models.Message{}
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(messages)
}

func SendMessage(w http.ResponseWriter, r *http.Request) {
	currentUserID, err := CurrentUserIDFromRequest(r)
	if err != nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	r.ParseForm()
	toPlayerIDStr := r.FormValue("ToPlayerId")
	toPlayerID, err := strconv.Atoi(toPlayerIDStr)
	if err != nil || toPlayerID == 0 {
		http.Error(w, "Bad Request", http.StatusBadRequest)
		return
	}
	msgType, _ := strconv.Atoi(r.FormValue("Type"))
	data := r.FormValue("Data")

	if isReservedMessageType(msgType) {
		http.Error(w, "Forbidden", http.StatusForbidden)
		return
	}

	if !utils.AccountActionAllowBurst("send_message", currentUserID, 2*time.Second, 5) {
		writeRateLimited(w)
		return
	}

	if TargetIgnoresSender(currentUserID, uint(toPlayerID)) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(models.Message{
			FromPlayerId: currentUserID,
			ToPlayerId:   uint(toPlayerID),
			Type:         msgType,
			Data:         data,
		})
		return
	}

	if utils.IsTextFlagged(data) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"error":   "Message violates the community guidelines.",
			"success": false,
		})
		return
	}

	msg := models.Message{
		FromPlayerId: currentUserID,
		ToPlayerId:   uint(toPlayerID),
		Type:         msgType,
		Data:         data,
	}
	if isLiveOnlyMessageType(msgType) {
		msg.Id = utils.NextLiveMessageID()
		msg.SentTime = time.Now()
	} else {
		db.DB.Create(&msg)
	}

	frame := hub.NotifFrame(models.MessageReceived, msg)
	hub.HubSendToPlayer(toPlayerID, frame)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(msg)
}

func DeleteMessages(w http.ResponseWriter, r *http.Request) {
	log.Printf("[MESSAGES] delete")
	currentUserID, err := CurrentUserIDFromRequest(r)
	if err != nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}
	var body struct {
		MessageIds []uint `json:"MessageIds"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil || len(body.MessageIds) == 0 {
		w.WriteHeader(http.StatusOK)
		return
	}
	db.DB.Where("id IN ? AND (from_player_id = ? OR to_player_id = ?)", body.MessageIds, currentUserID, currentUserID).
		Delete(&models.Message{})
	w.WriteHeader(http.StatusOK)
}

func Invite(w http.ResponseWriter, r *http.Request) {
	log.Printf("[INVITE] %s", r.URL.Path)
	currentUserID, err := CurrentUserIDFromRequest(r)
	if err != nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	r.ParseForm()
	toPlayerIDStr := r.FormValue("playerId")
	roomInstanceIdStr := r.FormValue("roomInstanceId")
	toPlayerID, err := strconv.Atoi(toPlayerIDStr)
	if err != nil || toPlayerID == 0 {
		http.Error(w, "Bad Request", http.StatusBadRequest)
		return
	}

	if !utils.AccountActionAllowBurst("invite", currentUserID, 3*time.Second, 4) {
		writeRateLimited(w)
		return
	}
	if TargetIgnoresSender(currentUserID, uint(toPlayerID)) {
		w.WriteHeader(http.StatusOK)
		return
	}

	roomInstanceId, _ := strconv.ParseInt(roomInstanceIdStr, 10, 64)

	msg := models.Message{
		Id:           utils.NextLiveMessageID(),
		FromPlayerId: currentUserID,
		ToPlayerId:   uint(toPlayerID),
		Type:         int(models.MessageTypeGameInvite),
		Data:         roomInstanceIdStr,
		SentTime:     time.Now(),
	}

	if roomInstanceId > 0 {
		var instance models.RoomInstance
		if err := db.DB.First(&instance, roomInstanceId).Error; err == nil {
			roomId := uint(instance.RoomId)
			msg.RoomId = &roomId
			db.DB.Clauses(clause.OnConflict{DoNothing: true}).Create(&models.InstanceInvite{
				InstanceID: instance.Id,
				AccountID:  toPlayerID,
				InvitedBy:  int(currentUserID),
			})

			hub.HubSendToPlayer(toPlayerID, hub.NotifFrame("RoomInstanceUpdate", instance))

			var room models.Room
			if err := db.DB.Preload("SubRooms").Preload("Roles").Preload("Tags").
				First(&room, instance.RoomId).Error; err == nil {
				rooms := []models.Room{room}
				controllers.InitRoomSlices(rooms)
				hub.HubSendToPlayer(toPlayerID, hub.NotifFrame("RoomUpdate", rooms[0]))
			}
		}
	}

	hub.HubSendToPlayer(toPlayerID, hub.NotifFrame(models.MessageReceived, msg))

	w.WriteHeader(http.StatusOK)
}

func PlayerSubscriptions(w http.ResponseWriter, r *http.Request) {
	log.Printf("[SUBSCRIPTIONS] my")
	w.Header().Set("Content-Type", "application/json")
	fmt.Fprint(w, "[]")
}

func Thread(w http.ResponseWriter, r *http.Request) {
	log.Printf("[THREAD] %s", r.URL.RawQuery)
	w.Header().Set("Content-Type", "application/json")
	fmt.Fprint(w, "[]")
}

func AnnouncementsUnread(w http.ResponseWriter, r *http.Request) {
	log.Printf("[ANNOUNCEMENTS] %s", r.URL.Path)
	w.Header().Set("Content-Type", "application/json")
	fmt.Fprint(w, "[]")
}

func AnnouncementGet(w http.ResponseWriter, r *http.Request) {
	log.Printf("[ANNOUNCEMENT] get")
	var announcements []models.Announcement
	db.DB.Find(&announcements)
	if announcements == nil {
		announcements = []models.Announcement{}
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(announcements)
}

func SubscriptionMine(w http.ResponseWriter, r *http.Request) {
	log.Printf("[SUBSCRIPTION] mine %s", r.URL.Path)
	w.Header().Set("Content-Type", "application/json")
	fmt.Fprint(w, "[]")
}

func SubscriptionDetails(w http.ResponseWriter, r *http.Request) {
	log.Printf("[SUBSCRIPTION] details %s", r.URL.Path)

	parts := strings.Split(strings.TrimRight(r.URL.Path, "/"), "/")
	accountId, _ := strconv.Atoi(parts[len(parts)-1])

	var club models.Club
	if err := db.DB.Where("creator_account_id = ? AND club_type = ?", accountId, 1).First(&club).Error; err != nil {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"accountId":       accountId,
			"clubId":          0,
			"subscriberCount": 0,
		})
		return
	}

	subs := club.MemberCount - 1
	if subs < 0 {
		subs = 0
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"accountId":       accountId,
		"clubId":          club.ClubId,
		"subscriberCount": subs,
	})
}

func SubscriptionDispatch(w http.ResponseWriter, r *http.Request) {
	parts := strings.Split(strings.Trim(r.URL.Path, "/"), "/")
	if len(parts) >= 3 && parts[1] == "subscriberCount" {
		SubscriptionSubscriberCount(w, r)
		return
	}
	if len(parts) >= 2 {
		switch r.Method {
		case http.MethodPost:
			SubscriptionSubscribe(w, r)
			return
		case http.MethodDelete:
			SubscriptionUnsubscribe(w, r)
			return
		}
	}
	log.Printf("[SUBSCRIPTION] unhandled %s %s", r.Method, r.URL.Path)
	http.NotFound(w, r)
}

func SubscriptionUnsubscribe(w http.ResponseWriter, r *http.Request) {
	accountID, ok := controllers.AccountIDFromRequest(r)
	if !ok {
		writeJsonEnvelope(w, 0)
		return
	}

	parts := strings.Split(strings.TrimRight(r.URL.Path, "/"), "/")
	targetAccountId, _ := strconv.Atoi(parts[len(parts)-1])
	log.Printf("[SUBSCRIPTION] unsubscribe target=%d account=%d", targetAccountId, accountID)

	var club models.Club
	if err := db.DB.Where("creator_account_id = ? AND club_type = ?", targetAccountId, 1).First(&club).Error; err != nil {
		writeJsonEnvelope(w, 0)
		return
	}

	res := db.DB.Where("club_id = ? AND account_id = ?", club.ClubId, accountID).Delete(&models.ClubMember{})
	if res.RowsAffected > 0 {
		var count int64
		db.DB.Model(&models.ClubMember{}).Where("club_id = ?", club.ClubId).Count(&count)
		db.DB.Model(&club).Update("member_count", count)
	}

	writeJsonEnvelope(w, res.RowsAffected)
}

func SubscriptionSubscriberCount(w http.ResponseWriter, r *http.Request) {
	parts := strings.Split(strings.TrimRight(r.URL.Path, "/"), "/")
	accountId, _ := strconv.Atoi(parts[len(parts)-1])
	log.Printf("[SUBSCRIPTION] subscriberCount account=%d", accountId)

	var club models.Club
	var count int
	if err := db.DB.Where("creator_account_id = ? AND club_type = ?", accountId, 1).First(&club).Error; err == nil {
		count = club.MemberCount - 1
		if count < 0 {
			count = 0
		}
	}

	w.Header().Set("Content-Type", "application/json")
	fmt.Fprint(w, count)
}

func SubscriptionSubscribe(w http.ResponseWriter, r *http.Request) {
	accountID, ok := controllers.AccountIDFromRequest(r)
	if !ok {
		writeJsonEnvelope(w, nil)
		return
	}

	parts := strings.Split(strings.TrimRight(r.URL.Path, "/"), "/")
	targetAccountId, _ := strconv.Atoi(parts[len(parts)-1])
	_ = r.ParseForm()
	log.Printf("[SUBSCRIPTION] subscribe target=%d roomId=%s", targetAccountId, r.FormValue("roomId"))

	var club models.Club
	if err := db.DB.Where("creator_account_id = ? AND club_type = ?", targetAccountId, 1).First(&club).Error; err != nil {
		writeJsonEnvelope(w, 0)
		return
	}

	var existing models.ClubMember
	if db.DB.Where("club_id = ? AND account_id = ?", club.ClubId, accountID).First(&existing).Error != nil {
		db.DB.Create(&models.ClubMember{
			ClubId:         club.ClubId,
			AccountId:      int(accountID),
			MembershipType: int(models.ClubMembershipMember),
		})
		db.DB.Model(&club).Update("member_count", club.MemberCount+1)
	}

	writeJsonEnvelope(w, 0)
}

type SanitizeRequest struct {
	ReplacementChar int    `json:"ReplacementChar"`
	Value           string `json:"Value"`
}

func Sanitize(w http.ResponseWriter, r *http.Request) {
	var req SanitizeRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	value := req.Value
	if utils.IsTextFlagged(value) {
		replacement := "*"
		if req.ReplacementChar > 0 {
			replacement = string(rune(req.ReplacementChar))
		}
		value = strings.Repeat(replacement, len([]rune(req.Value)))
	}
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	json.NewEncoder(w).Encode(value)
}

func CommunityBoard(w http.ResponseWriter, r *http.Request) {
	log.Printf("[COMMUNITYBOARD] current")
	w.Header().Set("Content-Type", "application/json")
	http.ServeFile(w, r, "data/jsons/communityboard.json")
}
