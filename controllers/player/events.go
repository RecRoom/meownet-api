package player

import (
	"encoding/json"
	"io"
	"net/http"
	"strconv"
	"strings"
	"time"

	"meow.net/controllers/hub"
	"meow.net/db"
	"meow.net/models"
)

func parseAccessibilityString(s string) int {
	switch strings.ToLower(s) {
	case "private":
		return int(models.PlayerEventAccessibilityPrivate)
	case "unlisted":
		return int(models.PlayerEventAccessibilityUnlisted)
	default:
		return int(models.PlayerEventAccessibilityPublic)
	}
}

func parseResponseTypeString(s string) int {
	switch strings.ToLower(s) {
	case "yes":
		return int(models.PlayerEventResponseYes)
	case "interested":
		return int(models.PlayerEventResponseInterested)
	case "no":
		return int(models.PlayerEventResponseNo)
	case "pending":
		return int(models.PlayerEventResponsePending)
	default:
		return int(models.PlayerEventResponseNone)
	}
}

// GET /api/playerevents/v1/search
func PlayerEventsSearch(w http.ResponseWriter, r *http.Request) {
	scheduleFilter := r.URL.Query().Get("scheduleFilter")
	tagFilter := r.URL.Query().Get("tag")
	sortBy := r.URL.Query().Get("sort")
	now := time.Now()

	query := db.DB.Model(&models.PlayerEvent{}).Preload("Tags")

	switch scheduleFilter {
	case "Past":
		query = query.Where("end_time < ?", now)
	case "All":
		// no time filter
	default: // Upcoming
		query = query.Where("end_time >= ?", now)
	}

	if tagFilter != "" {
		query = query.
			Joins("JOIN player_event_tags ON player_event_tags.player_event_id = player_events.player_event_id").
			Where("player_event_tags.tag = ?", tagFilter)
	}

	switch sortBy {
	case "Attendance":
		query = query.Order("attendee_count DESC, start_time ASC")
	default:
		query = query.Order("start_time ASC")
	}

	var events []models.PlayerEvent
	query.Find(&events)
	if events == nil {
		events = []models.PlayerEvent{}
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(events)
}

// GET /api/playerevents/v1/tagfilters
func PlayerEventTagFilters(w http.ResponseWriter, r *http.Request) {
	var rows []struct {
		Tag   string
		Count int
	}
	db.DB.Model(&models.PlayerEventTag{}).
		Select("tag, count(*) as count").
		Where("tag != ''").
		Group("tag").
		Order("count DESC").
		Limit(20).
		Scan(&rows)

	tags := make([]string, 0, len(rows))
	for _, row := range rows {
		tags = append(tags, row.Tag)
	}

	pinnedCount := len(tags)
	if pinnedCount > 5 {
		pinnedCount = 5
	}

	var pinned []string
	if pinnedCount > 0 {
		pinned = tags[:pinnedCount]
	} else {
		pinned = []string{}
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"PinnedFilters":   pinned,
		"PopularFilters":  tags,
		"TrendingFilters": nil,
	})
}

type liveEventView struct {
	models.PlayerEvent
	PlayerCount int  `json:"PlayerCount"`
	IsFull      bool `json:"IsFull"`
}

// GET /api/playerevents/v1/searchlive
func PlayerEventsSearchLive(w http.ResponseWriter, r *http.Request) {
	now := time.Now()
	var events []models.PlayerEvent
	db.DB.Preload("Tags").
		Where("start_time <= ? AND end_time >= ?", now, now).
		Order("start_time ASC").
		Find(&events)

	results := make([]liveEventView, 0, len(events))
	for _, e := range events {
		var playerCount int64
		db.DB.Model(&models.RoomInstance{}).
			Where("room_id = ?", e.RoomId).
			Count(&playerCount)
		results = append(results, liveEventView{
			PlayerEvent: e,
			PlayerCount: int(playerCount),
			IsFull:      false,
		})
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(results)
}

// GET /api/playerevents/v1/{id}/responses
func PlayerEventResponses(w http.ResponseWriter, r *http.Request, eventId uint) {
	var responses []models.PlayerEventResponse
	db.DB.Where("player_event_id = ?", eventId).Order("created_at ASC").Find(&responses)
	if responses == nil {
		responses = []models.PlayerEventResponse{}
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(responses)
}

// GET /api/playerevents/v1/{id}
func PlayerEventGet(w http.ResponseWriter, r *http.Request, eventId uint) {
	var event models.PlayerEvent
	if err := db.DB.Preload("Tags").First(&event, eventId).Error; err != nil {
		http.Error(w, "Not Found", http.StatusNotFound)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(event)
}

// handles /api/playerevents/v1/ prefix
func PlayerEventsV1Dispatch(w http.ResponseWriter, r *http.Request) {
	path := strings.Trim(r.URL.Path, "/")
	parts := strings.Split(path, "/")
	if len(parts) < 4 {
		http.NotFound(w, r)
		return
	}

	sub := parts[3]

	switch sub {
	case "search":
		PlayerEventsSearch(w, r)
	case "tagfilters":
		PlayerEventTagFilters(w, r)
	case "searchlive":
		PlayerEventsSearchLive(w, r)
	case "all":
		PlayerEventsAll(w, r)
	case "respond":
		PlayerEventRespond(w, r)
	case "room":
		if len(parts) < 5 {
			http.NotFound(w, r)
			return
		}
		roomId, err := strconv.ParseInt(parts[4], 10, 64)
		if err != nil {
			http.Error(w, "Bad Request", http.StatusBadRequest)
			return
		}
		RoomPlayerEventsById(w, r, roomId)
	default:
		eventId, err := strconv.ParseUint(sub, 10, 64)
		if err != nil {
			http.NotFound(w, r)
			return
		}
		if len(parts) >= 5 && parts[4] == "responses" {
			PlayerEventResponses(w, r, uint(eventId))
		} else {
			PlayerEventGet(w, r, uint(eventId))
		}
	}
}

// GET /api/playerevents/v1/all
func PlayerEventsAll(w http.ResponseWriter, r *http.Request) {
	accountId, ok := accountIDFromRequest(r)
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	var created []models.PlayerEvent
	db.DB.Preload("Tags").Where("creator_player_id = ?", accountId).Find(&created)
	if created == nil {
		created = []models.PlayerEvent{}
	}

	var responses []models.PlayerEventResponse
	db.DB.Where("player_id = ?", accountId).Find(&responses)
	if responses == nil {
		responses = []models.PlayerEventResponse{}
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"Created":   created,
		"Responses": responses,
	})
}

// GET /api/playerevents/v1/room/
func RoomPlayerEventsById(w http.ResponseWriter, r *http.Request, roomId int64) {
	var events []models.PlayerEvent
	db.DB.Preload("Tags").Where("room_id = ?", roomId).Order("start_time ASC").Find(&events)
	if events == nil {
		events = []models.PlayerEvent{}
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(events)
}

// POST /api/playerevents/v1/respond
func PlayerEventRespond(w http.ResponseWriter, r *http.Request) {
	accountId, ok := accountIDFromRequest(r)
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	body, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "Bad Request", http.StatusBadRequest)
		return
	}

	var req struct {
		PlayerEventId uint   `json:"PlayerEventId"`
		Type          string `json:"Type"`
	}
	if err := json.Unmarshal(body, &req); err != nil {
		http.Error(w, "Bad Request", http.StatusBadRequest)
		return
	}

	var event models.PlayerEvent
	if err := db.DB.Preload("Tags").First(&event, req.PlayerEventId).Error; err != nil {
		http.Error(w, "Not Found", http.StatusNotFound)
		return
	}

	responseType := parseResponseTypeString(req.Type)

	var existing models.PlayerEventResponse
	result := db.DB.Where("player_event_id = ? AND player_id = ?", req.PlayerEventId, accountId).First(&existing)
	if result.Error == nil {
		db.DB.Model(&existing).Update("type", responseType)
	} else {
		existing = models.PlayerEventResponse{
			PlayerEventId: req.PlayerEventId,
			PlayerId:      accountId,
			Type:          responseType,
		}
		db.DB.Create(&existing)
	}

	var yesCount int64
	db.DB.Model(&models.PlayerEventResponse{}).
		Where("player_event_id = ? AND type = ?", req.PlayerEventId, models.PlayerEventResponseYes).
		Count(&yesCount)
	db.DB.Model(&event).Update("attendee_count", int(yesCount))

	hub.HubSendToPlayer(int(accountId), hub.NotifFrame(int(models.PlayerEventResponseChanged), map[string]interface{}{
		"PlayerEventId":         req.PlayerEventId,
		"PlayerId":              accountId,
		"PlayerEventResponseId": existing.PlayerEventResponseId,
		"Type":                  responseType,
	}))

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"PlayerEvent": event,
		"Result":      0,
	})
}

type playerEventInput struct {
	Accessibility string   `json:"Accessibility"`
	ClubId        *int64   `json:"ClubId"`
	Description   string   `json:"Description"`
	EndTime       string   `json:"EndTime"`
	ImageName     *string  `json:"ImageName"`
	Name          string   `json:"Name"`
	RoomId        int64    `json:"RoomId"`
	StartTime     string   `json:"StartTime"`
	SubRoomId     *int64   `json:"SubRoomId"`
	Tags          []string `json:"Tags"`
}

// POST /api/playerevents/v2
func PlayerEventCreate(w http.ResponseWriter, r *http.Request) {
	accountId, ok := accountIDFromRequest(r)
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	body, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "Bad Request", http.StatusBadRequest)
		return
	}

	var req playerEventInput
	if err := json.Unmarshal(body, &req); err != nil {
		http.Error(w, "Bad Request", http.StatusBadRequest)
		return
	}

	startTime, err := time.Parse(time.RFC3339, req.StartTime)
	if err != nil {
		http.Error(w, "Bad Request: invalid StartTime", http.StatusBadRequest)
		return
	}
	endTime, err := time.Parse(time.RFC3339, req.EndTime)
	if err != nil {
		http.Error(w, "Bad Request: invalid EndTime", http.StatusBadRequest)
		return
	}

	tags := make([]models.PlayerEventTag, 0, len(req.Tags))
	for _, t := range req.Tags {
		tags = append(tags, models.PlayerEventTag{Tag: t, Type: 0})
	}

	event := models.PlayerEvent{
		CreatorPlayerId: accountId,
		RoomId:          req.RoomId,
		SubRoomId:       req.SubRoomId,
		ClubId:          req.ClubId,
		Name:            req.Name,
		Description:     req.Description,
		ImageName:       req.ImageName,
		StartTime:       startTime,
		EndTime:         endTime,
		Accessibility:   parseAccessibilityString(req.Accessibility),
		Tags:            tags,
	}

	if err := db.DB.Create(&event).Error; err != nil {
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	db.DB.Preload("Tags").First(&event, event.PlayerEventId)

	hub.HubSendToPlayer(int(accountId), hub.NotifFrame(int(models.PlayerEventCreated), event))

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"PlayerEvent": event,
		"Result":      0,
	})
}

// POST /api/playerevents/v2/{id}
func PlayerEventUpdate(w http.ResponseWriter, r *http.Request, eventId uint) {
	accountId, ok := accountIDFromRequest(r)
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	var event models.PlayerEvent
	if err := db.DB.Preload("Tags").First(&event, eventId).Error; err != nil {
		http.Error(w, "Not Found", http.StatusNotFound)
		return
	}
	if event.CreatorPlayerId != accountId {
		http.Error(w, "Forbidden", http.StatusForbidden)
		return
	}

	body, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "Bad Request", http.StatusBadRequest)
		return
	}

	var req playerEventInput
	if err := json.Unmarshal(body, &req); err != nil {
		http.Error(w, "Bad Request", http.StatusBadRequest)
		return
	}

	startTime, err := time.Parse(time.RFC3339, req.StartTime)
	if err != nil {
		http.Error(w, "Bad Request: invalid StartTime", http.StatusBadRequest)
		return
	}
	endTime, err := time.Parse(time.RFC3339, req.EndTime)
	if err != nil {
		http.Error(w, "Bad Request: invalid EndTime", http.StatusBadRequest)
		return
	}

	event.Name = req.Name
	event.Description = req.Description
	event.ImageName = req.ImageName
	event.RoomId = req.RoomId
	event.SubRoomId = req.SubRoomId
	event.ClubId = req.ClubId
	event.StartTime = startTime
	event.EndTime = endTime
	event.Accessibility = parseAccessibilityString(req.Accessibility)

	db.DB.Save(&event)

	if req.Tags != nil {
		db.DB.Where("player_event_id = ?", eventId).Delete(&models.PlayerEventTag{})
		newTags := make([]models.PlayerEventTag, 0, len(req.Tags))
		for _, t := range req.Tags {
			newTags = append(newTags, models.PlayerEventTag{PlayerEventId: eventId, Tag: t, Type: 0})
		}
		if len(newTags) > 0 {
			db.DB.Create(&newTags)
		}
	}

	db.DB.Preload("Tags").First(&event, eventId)

	hub.HubSendToPlayer(int(accountId), hub.NotifFrame(int(models.PlayerEventUpdated), event))

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"PlayerEvent": event,
		"Result":      0,
	})
}

// POST /api/playerevents/v2/delete/{id}
func PlayerEventDelete(w http.ResponseWriter, r *http.Request, eventId uint) {
	accountId, ok := accountIDFromRequest(r)
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	var event models.PlayerEvent
	if err := db.DB.First(&event, eventId).Error; err != nil {
		http.Error(w, "Not Found", http.StatusNotFound)
		return
	}
	if event.CreatorPlayerId != accountId {
		http.Error(w, "Forbidden", http.StatusForbidden)
		return
	}

	db.DB.Delete(&event)

	hub.HubSendToPlayer(int(accountId), hub.NotifFrame(int(models.PlayerEventDeleted), map[string]interface{}{
		"PlayerEventId": eventId,
	}))

	w.WriteHeader(http.StatusOK)
}

// handles /api/playerevents/v2/ prefix
func PlayerEventsV2Dispatch(w http.ResponseWriter, r *http.Request) {
	path := strings.Trim(r.URL.Path, "/")
	parts := strings.Split(path, "/")

	if len(parts) < 4 {
		// POST /api/playerevents/v2
		if r.Method == http.MethodPost {
			PlayerEventCreate(w, r)
			return
		}
		http.NotFound(w, r)
		return
	}

	sub := parts[3]

	if sub == "delete" && len(parts) >= 5 {
		eventId, err := strconv.ParseUint(parts[4], 10, 64)
		if err != nil {
			http.Error(w, "Bad Request", http.StatusBadRequest)
			return
		}
		PlayerEventDelete(w, r, uint(eventId))
		return
	}

	eventId, err := strconv.ParseUint(sub, 10, 64)
	if err != nil {
		http.NotFound(w, r)
		return
	}
	if r.Method == http.MethodPost {
		PlayerEventUpdate(w, r, uint(eventId))
		return
	}

	http.NotFound(w, r)
}
