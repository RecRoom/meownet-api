package social

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"gorm.io/gorm"

	"meow.net/controllers/hub"
	"meow.net/controllers/reputation"
	"meow.net/db"
	"meow.net/models"
	"meow.net/utils"
)

func relationshipResponse(rel models.Relationship, currentUserID uint) models.RelationshipResponse {
	if rel.RequesterID == currentUserID {
		return models.RelationshipResponse{
			Favorited:        rel.RequesterFavorited,
			Ignored:          rel.RequesterIgnored,
			Muted:            rel.RequesterMuted,
			PlayerID:         rel.TargetID,
			RelationshipType: rel.RelationshipType,
		}
	}
	relType := rel.RelationshipType
	if relType == models.RelationshipFriendRequestSent {
		relType = models.RelationshipFriendRequestReceived
	}
	return models.RelationshipResponse{
		Favorited:        rel.TargetFavorited,
		Ignored:          rel.TargetIgnored,
		Muted:            rel.TargetMuted,
		PlayerID:         rel.RequesterID,
		RelationshipType: relType,
	}
}

func pushPresence(a, b uint) {
	hub.HubSendToPlayer(int(a), hub.NotifFrame("PresenceUpdate", hub.BuildPresenceFor(int(a), int(b))))
	hub.HubSendToPlayer(int(b), hub.NotifFrame("PresenceUpdate", hub.BuildPresenceFor(int(b), int(a))))
}

func CurrentUserIDFromRequest(r *http.Request) (uint, error) {
	tokenStr := utils.GetBearerToken(r)
	if tokenStr == "" {
		return 0, fmt.Errorf("no token")
	}
	idStr, err := utils.ParseSubFromJWT(tokenStr)
	if err != nil || idStr == "" {
		return 0, fmt.Errorf("invalid token")
	}
	id, err := strconv.Atoi(idStr)
	if err != nil {
		return 0, fmt.Errorf("invalid id")
	}
	return uint(id), nil
}

func findRelationship(userA, userB uint) (models.Relationship, bool) {
	var rel models.Relationship
	err := db.DB.Where(
		"(requester_id = ? AND target_id = ?) OR (requester_id = ? AND target_id = ?)",
		userA, userB, userB, userA,
	).First(&rel).Error
	return rel, err == nil
}

func TargetIgnoresSender(senderID, targetID uint) bool {
	rel, ok := findRelationship(senderID, targetID)
	if !ok {
		return false
	}
	if rel.RequesterID == targetID {
		return rel.RequesterIgnored != 0
	}
	return rel.TargetIgnored != 0
}

func RelationshipsGet(w http.ResponseWriter, r *http.Request) {
	currentUserID, err := CurrentUserIDFromRequest(r)
	if err != nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	var relationships []models.Relationship
	db.DB.Where("requester_id = ? OR target_id = ?", currentUserID, currentUserID).Find(&relationships)

	resp := make([]models.RelationshipResponse, 0, len(relationships))
	for _, rel := range relationships {
		resp = append(resp, relationshipResponse(rel, currentUserID))
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

func SendCheer(w http.ResponseWriter, r *http.Request) {
	currentUserID, err := CurrentUserIDFromRequest(r)
	if err != nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	if err := r.ParseForm(); err != nil {
		http.Error(w, "Bad Request", http.StatusBadRequest)
		return
	}
	targetID, err := strconv.Atoi(r.FormValue("PlayerIdTo"))
	if err != nil || targetID == 0 || uint(targetID) == currentUserID {
		http.Error(w, "Bad Request", http.StatusBadRequest)
		return
	}
	category, _ := strconv.Atoi(r.FormValue("CheerCategory"))
	roomId, _ := strconv.ParseUint(r.FormValue("RoomId"), 10, 64)
	anonymous := strings.EqualFold(r.FormValue("Anonymous"), "true")

	cheer := models.PlayerCheer{
		FromAccountId: currentUserID,
		ToAccountId:   uint(targetID),
		Category:      category,
		Anonymous:     anonymous,
	}
	if roomId > 0 {
		rid := uint(roomId)
		cheer.RoomId = &rid
	}

	var hasCredit bool
	err = db.DB.Transaction(func(tx *gorm.DB) error {
		if err := db.AdvisoryLockCheer(tx, currentUserID); err != nil {
			return err
		}
		since := time.Now().UTC().Truncate(24 * time.Hour)
		var sent int64
		if err := tx.Model(&models.PlayerCheer{}).
			Where("from_account_id = ? AND created_at >= ?", currentUserID, since).
			Count(&sent).Error; err != nil {
			return err
		}
		if models.CheerDailyCredit-int(sent)*models.CheerCost < models.CheerCost {
			return nil
		}
		hasCredit = true
		return tx.Create(&cheer).Error
	})
	if err != nil {
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
	if !hasCredit {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"Message": "",
			"Success": false,
		})
		return
	}

	msgType := int(models.MessageTypePlayerCheer)
	fromId := currentUserID
	if anonymous {
		msgType = int(models.MessageTypePlayerCheerAnonymous)
		fromId = 0
	}
	msg := models.Message{
		FromPlayerId: fromId,
		ToPlayerId:   uint(targetID),
		Type:         msgType,
		Data:         strconv.Itoa(category),
	}
	db.DB.Create(&msg)
	hub.HubSendToPlayer(targetID, hub.NotifFrame(models.MessageReceived, msg))

	hub.HubBroadcastReputationUpdate(int(currentUserID))
	hub.HubBroadcastReputationUpdate(targetID)

	repJson, _ := json.Marshal(reputation.Build(currentUserID))
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"Message": string(repJson),
		"Success": true,
	})
}

func SetSelectedCheer(w http.ResponseWriter, r *http.Request) {
	currentUserID, err := CurrentUserIDFromRequest(r)
	if err != nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	if err := r.ParseForm(); err != nil {
		http.Error(w, "Bad Request", http.StatusBadRequest)
		return
	}
	category, _ := strconv.Atoi(r.FormValue("CheerCategory"))

	db.DB.Model(&models.Account{}).
		Where("account_id = ?", currentUserID).
		Update("selected_cheer", category)

	hub.HubBroadcastReputationUpdate(int(currentUserID))

	repJson, _ := json.Marshal(reputation.Build(currentUserID))
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"Message": string(repJson),
		"Success": true,
	})
}

func SendFriendRequest(w http.ResponseWriter, r *http.Request) {
	currentUserID, err := CurrentUserIDFromRequest(r)
	if err != nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	targetIDStr := r.URL.Query().Get("id")
	targetID, err := strconv.Atoi(targetIDStr)
	if err != nil || targetID == 0 || uint(targetID) == currentUserID {
		http.Error(w, "Bad Request", http.StatusBadRequest)
		return
	}

	if !utils.AccountActionAllowBurst("send_friend_request", currentUserID, 5*time.Second, 4) {
		writeRateLimited(w)
		return
	}

	const (
		outcomeNoChange = iota
		outcomeAutoAccept
		outcomeNewRequest
	)
	var (
		rel     models.Relationship
		outcome = outcomeNoChange
	)

	txErr := db.DB.Transaction(func(tx *gorm.DB) error {
		if err := db.AdvisoryLockRelationship(tx, currentUserID, uint(targetID)); err != nil {
			return err
		}

		var existing models.Relationship
		exists := tx.Where(
			"(requester_id = ? AND target_id = ?) OR (requester_id = ? AND target_id = ?)",
			currentUserID, uint(targetID), uint(targetID), currentUserID,
		).First(&existing).Error == nil

		ignored := false
		if exists {
			if existing.RequesterID == uint(targetID) {
				ignored = existing.RequesterIgnored != 0
			} else {
				ignored = existing.TargetIgnored != 0
			}
		}

		if ignored ||
			(exists && existing.RelationshipType == models.RelationshipFriend) ||
			(exists && existing.RelationshipType == models.RelationshipFriendRequestSent && existing.RequesterID == currentUserID) {
			rel = existing
			return nil
		}

		if exists && existing.RelationshipType == models.RelationshipFriendRequestSent && existing.TargetID == currentUserID {
			existing.RelationshipType = models.RelationshipFriend
			if err := tx.Save(&existing).Error; err != nil {
				return err
			}
			rel = existing
			outcome = outcomeAutoAccept
			return nil
		}

		if exists {
			existing.RequesterID = currentUserID
			existing.TargetID = uint(targetID)
			existing.RelationshipType = models.RelationshipFriendRequestSent
			if err := tx.Save(&existing).Error; err != nil {
				return err
			}
			rel = existing
		} else {
			rel = models.Relationship{
				RequesterID:      currentUserID,
				TargetID:         uint(targetID),
				RelationshipType: models.RelationshipFriendRequestSent,
			}
			if err := tx.Create(&rel).Error; err != nil {
				return err
			}
		}
		outcome = outcomeNewRequest
		return nil
	})
	if txErr != nil {
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	switch outcome {
	case outcomeAutoAccept:
		hub.HubSendToPlayer(int(rel.RequesterID), hub.NotifFrame(models.RelationshipChanged, relationshipResponse(rel, rel.RequesterID)))
		hub.HubSendToPlayer(int(currentUserID), hub.NotifFrame(models.RelationshipChanged, relationshipResponse(rel, currentUserID)))
		pushPresence(rel.RequesterID, currentUserID)
	case outcomeNewRequest:
		hub.HubSendToPlayer(int(currentUserID), hub.NotifFrame(models.RelationshipChanged, relationshipResponse(rel, currentUserID)))
		hub.HubSendToPlayer(targetID, hub.NotifFrame(models.RelationshipChanged, relationshipResponse(rel, uint(targetID))))

		reqMsg := models.Message{
			FromPlayerId: currentUserID,
			ToPlayerId:   uint(targetID),
			Type:         4,
			Data:         "",
		}
		db.DB.Create(&reqMsg)
		hub.HubSendToPlayer(targetID, hub.NotifFrame(models.MessageReceived, reqMsg))
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(relationshipResponse(rel, currentUserID))
}

func RemoveFriend(w http.ResponseWriter, r *http.Request) {
	currentUserID, err := CurrentUserIDFromRequest(r)
	if err != nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	targetIDStr := r.URL.Query().Get("id")
	targetID, err := strconv.Atoi(targetIDStr)
	if err != nil || targetID == 0 || uint(targetID) == currentUserID {
		http.Error(w, "Bad Request", http.StatusBadRequest)
		return
	}

	rel, exists := findRelationship(currentUserID, uint(targetID))
	if exists {
		rel.RelationshipType = models.RelationshipNone
		db.DB.Save(&rel)
	}

	resp := models.RelationshipResponse{
		PlayerID:         uint(targetID),
		RelationshipType: models.RelationshipNone,
	}

	hub.HubSendToPlayer(int(currentUserID), hub.NotifFrame(models.RelationshipChanged, resp))
	hub.HubSendToPlayer(targetID, hub.NotifFrame(models.RelationshipChanged, models.RelationshipResponse{
		PlayerID:         currentUserID,
		RelationshipType: models.RelationshipNone,
	}))

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

func AcceptFriendRequest(w http.ResponseWriter, r *http.Request) {
	currentUserID, err := CurrentUserIDFromRequest(r)
	if err != nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	requesterIDStr := r.URL.Query().Get("id")
	requesterID, err := strconv.Atoi(requesterIDStr)
	if err != nil || requesterID == 0 || uint(requesterID) == currentUserID {
		http.Error(w, "Bad Request", http.StatusBadRequest)
		return
	}

	rel, exists := findRelationship(currentUserID, uint(requesterID))
	if !exists {
		http.Error(w, "Not Found", http.StatusNotFound)
		return
	}

	if rel.RelationshipType != models.RelationshipFriendRequestSent || rel.RequesterID != uint(requesterID) || rel.TargetID != currentUserID {
		http.Error(w, "Forbidden", http.StatusForbidden)
		return
	}

	rel.RelationshipType = models.RelationshipFriend
	db.DB.Save(&rel)

	hub.HubSendToPlayer(int(rel.RequesterID), hub.NotifFrame(models.RelationshipChanged, relationshipResponse(rel, rel.RequesterID)))
	hub.HubSendToPlayer(int(currentUserID), hub.NotifFrame(models.RelationshipChanged, relationshipResponse(rel, currentUserID)))
	pushPresence(rel.RequesterID, currentUserID)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(relationshipResponse(rel, currentUserID))
}

func MutePlayer(w http.ResponseWriter, r *http.Request) {
	currentUserID, err := CurrentUserIDFromRequest(r)
	if err != nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	r.ParseForm()
	targetIDStr := r.FormValue("PlayerId")
	targetID, err := strconv.Atoi(targetIDStr)
	if err != nil || targetID == 0 || uint(targetID) == currentUserID {
		http.Error(w, "Bad Request", http.StatusBadRequest)
		return
	}

	rel, exists := findRelationship(currentUserID, uint(targetID))
	if !exists {
		rel = models.Relationship{
			RequesterID:      currentUserID,
			TargetID:         uint(targetID),
			RelationshipType: models.RelationshipNone,
		}
		db.DB.Create(&rel)
	}

	if rel.RequesterID == currentUserID {
		rel.RequesterMuted = 1
	} else {
		rel.TargetMuted = 1
	}
	db.DB.Save(&rel)

	hub.HubSendToPlayer(int(currentUserID), hub.NotifFrame(models.RelationshipChanged, relationshipResponse(rel, currentUserID)))

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(relationshipResponse(rel, currentUserID))
}

func IgnorePlayer(w http.ResponseWriter, r *http.Request) {
	currentUserID, err := CurrentUserIDFromRequest(r)
	if err != nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	r.ParseForm()
	targetIDStr := r.FormValue("PlayerId")
	targetID, err := strconv.Atoi(targetIDStr)
	if err != nil || targetID == 0 || uint(targetID) == currentUserID {
		http.Error(w, "Bad Request", http.StatusBadRequest)
		return
	}

	rel, exists := findRelationship(currentUserID, uint(targetID))
	if !exists {
		rel = models.Relationship{
			RequesterID:      currentUserID,
			TargetID:         uint(targetID),
			RelationshipType: models.RelationshipNone,
		}
		db.DB.Create(&rel)
	}

	if rel.RequesterID == currentUserID {
		rel.RequesterIgnored = 1
	} else {
		rel.TargetIgnored = 1
	}
	db.DB.Save(&rel)

	hub.HubSendToPlayer(int(currentUserID), hub.NotifFrame(models.RelationshipChanged, relationshipResponse(rel, currentUserID)))

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(relationshipResponse(rel, currentUserID))
}

func AddFriend(w http.ResponseWriter, r *http.Request) {
	currentUserID, err := CurrentUserIDFromRequest(r)
	if err != nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	targetIDStr := r.URL.Query().Get("id")
	targetID, err := strconv.Atoi(targetIDStr)
	if err != nil || targetID == 0 || uint(targetID) == currentUserID {
		http.Error(w, "Bad Request", http.StatusBadRequest)
		return
	}

	currentInstance, currentOk := hub.GetPlayerInstance(int(currentUserID))
	targetInstance, targetOk := hub.GetPlayerInstance(targetID)
	if !currentOk || !targetOk || currentInstance <= 0 || currentInstance != targetInstance {
		http.Error(w, "Forbidden", http.StatusForbidden)
		return
	}

	rel, exists := findRelationship(currentUserID, uint(targetID))
	if !exists {
		rel = models.Relationship{
			RequesterID:      currentUserID,
			TargetID:         uint(targetID),
			RelationshipType: models.RelationshipFriend,
		}
		db.DB.Create(&rel)
	} else {
		rel.RelationshipType = models.RelationshipFriend
		db.DB.Save(&rel)
	}

	hub.HubSendToPlayer(int(currentUserID), hub.NotifFrame(models.RelationshipChanged, relationshipResponse(rel, currentUserID)))
	hub.HubSendToPlayer(targetID, hub.NotifFrame(models.RelationshipChanged, relationshipResponse(rel, uint(targetID))))
	pushPresence(currentUserID, uint(targetID))

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(relationshipResponse(rel, currentUserID))
}

func UnignorePlayer(w http.ResponseWriter, r *http.Request) {
	currentUserID, err := CurrentUserIDFromRequest(r)
	if err != nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	r.ParseForm()
	targetIDStr := r.FormValue("PlayerId")
	targetID, err := strconv.Atoi(targetIDStr)
	if err != nil || targetID == 0 || uint(targetID) == currentUserID {
		http.Error(w, "Bad Request", http.StatusBadRequest)
		return
	}

	rel, exists := findRelationship(currentUserID, uint(targetID))
	if !exists {
		http.Error(w, "Not Found", http.StatusNotFound)
		return
	}

	if rel.RequesterID == currentUserID {
		rel.RequesterIgnored = 0
	} else {
		rel.TargetIgnored = 0
	}
	db.DB.Save(&rel)

	hub.HubSendToPlayer(int(currentUserID), hub.NotifFrame(models.RelationshipChanged, relationshipResponse(rel, currentUserID)))

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(relationshipResponse(rel, currentUserID))
}

func UnfavoritePlayer(w http.ResponseWriter, r *http.Request) {
	currentUserID, err := CurrentUserIDFromRequest(r)
	if err != nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	targetIDStr := r.URL.Query().Get("id")
	targetID, err := strconv.Atoi(targetIDStr)
	if err != nil || targetID == 0 || uint(targetID) == currentUserID {
		http.Error(w, "Bad Request", http.StatusBadRequest)
		return
	}

	rel, exists := findRelationship(currentUserID, uint(targetID))
	if !exists {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(models.RelationshipResponse{
			PlayerID:         uint(targetID),
			RelationshipType: models.RelationshipNone,
		})
		return
	}

	if rel.RequesterID == currentUserID {
		rel.RequesterFavorited = 0
	} else {
		rel.TargetFavorited = 0
	}
	db.DB.Save(&rel)

	hub.HubSendToPlayer(int(currentUserID), hub.NotifFrame(models.RelationshipChanged, relationshipResponse(rel, currentUserID)))

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(relationshipResponse(rel, currentUserID))
}

func FavoritePlayer(w http.ResponseWriter, r *http.Request) {
	currentUserID, err := CurrentUserIDFromRequest(r)
	if err != nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	targetIDStr := r.URL.Query().Get("id")
	targetID, err := strconv.Atoi(targetIDStr)
	if err != nil || targetID == 0 || uint(targetID) == currentUserID {
		http.Error(w, "Bad Request", http.StatusBadRequest)
		return
	}

	rel, exists := findRelationship(currentUserID, uint(targetID))
	if !exists {
		rel = models.Relationship{
			RequesterID:      currentUserID,
			TargetID:         uint(targetID),
			RelationshipType: models.RelationshipNone,
		}
		db.DB.Create(&rel)
	}

	if rel.RequesterID == currentUserID {
		rel.RequesterFavorited = 1
	} else {
		rel.TargetFavorited = 1
	}
	db.DB.Save(&rel)

	hub.HubSendToPlayer(int(currentUserID), hub.NotifFrame(models.RelationshipChanged, relationshipResponse(rel, currentUserID)))

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(relationshipResponse(rel, currentUserID))
}

func BulkIgnore(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	fmt.Fprint(w, "[]")
}

func UnmutePlayer(w http.ResponseWriter, r *http.Request) {
	currentUserID, err := CurrentUserIDFromRequest(r)
	if err != nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	r.ParseForm()
	targetIDStr := r.FormValue("PlayerId")
	targetID, err := strconv.Atoi(targetIDStr)
	if err != nil || targetID == 0 || uint(targetID) == currentUserID {
		http.Error(w, "Bad Request", http.StatusBadRequest)
		return
	}

	rel, exists := findRelationship(currentUserID, uint(targetID))
	if !exists {
		http.Error(w, "Not Found", http.StatusNotFound)
		return
	}

	if rel.RequesterID == currentUserID {
		rel.RequesterMuted = 0
	} else {
		rel.TargetMuted = 0
	}
	db.DB.Save(&rel)

	hub.HubSendToPlayer(int(currentUserID), hub.NotifFrame(models.RelationshipChanged, relationshipResponse(rel, currentUserID)))

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(relationshipResponse(rel, currentUserID))
}

func FavoriteFriendOnlineStatus(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
}
