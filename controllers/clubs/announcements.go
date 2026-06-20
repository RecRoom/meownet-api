package clubs

import (
	"encoding/json"
	"log"
	"net/http"
	"strconv"
	"strings"

	"meow.net/controllers"
	"meow.net/controllers/hub"
	"meow.net/db"
	"meow.net/models"
)

func ClubAnnouncementsGet(w http.ResponseWriter, r *http.Request) {
	clubId, ok := parseClubIdFromPath(r.URL.Path)
	if !ok {
		http.Error(w, "Bad Request", http.StatusBadRequest)
		return
	}
	var announcements []models.ClubAnnouncement
	db.DB.Where("club_id = ?", clubId).Order("created_at desc").Find(&announcements)
	if announcements == nil {
		announcements = []models.ClubAnnouncement{}
	}

	var lastId interface{} = nil
	if len(announcements) > 0 {
		lastId = announcements[0].AnnouncementId
	}

	writeJsonEnvelope(w, map[string]interface{}{
		"Announcements":          announcements,
		"ClubId":                 clubId,
		"LastAnnouncementId":     lastId,
		"LastReadAnnouncementId": 0,
	})
}

func ClubAnnouncementCreate(w http.ResponseWriter, r *http.Request) {
	accountID, ok := controllers.AccountIDFromRequest(r)
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}
	clubId, ok := parseClubIdFromPath(r.URL.Path)
	if !ok {
		http.Error(w, "Bad Request", http.StatusBadRequest)
		return
	}
	if err := r.ParseForm(); err != nil {
		http.Error(w, "Bad Request", http.StatusBadRequest)
		return
	}

	mt := myMembershipType(clubId, int(accountID))
	if mt < int(models.ClubMembershipCoowner) {
		writeJsonError(w, http.StatusForbidden, "insufficient permissions")
		return
	}

	ann := models.ClubAnnouncement{
		ClubId:    clubId,
		AccountId: int(accountID),
		Title:     r.FormValue("title"),
		Body:      r.FormValue("body"),
		ImageName: r.FormValue("imageName"),
		Meta:      r.FormValue("meta"),
	}
	if err := db.DB.Create(&ann).Error; err != nil {
		log.Printf("[CLUB] announcement create error: %v", err)
		writeJsonError(w, http.StatusInternalServerError, "create failed")
		return
	}
	writeJsonEnvelope(w, ann.AnnouncementId)
}

func ClubAnnouncementsRouter(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		ClubAnnouncementsGet(w, r)
	case http.MethodPost:
		ClubAnnouncementCreate(w, r)
	default:
		http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
	}
}

func ClubPlayerEvents(w http.ResponseWriter, r *http.Request) {
	path := r.URL.Path
	parts := strings.Split(strings.Trim(path, "/"), "/")
	var clubId int64
	for i, p := range parts {
		if p == "club" && i+1 < len(parts) {
			id, err := strconv.ParseInt(parts[i+1], 10, 64)
			if err == nil {
				clubId = id
			}
			break
		}
	}

	var events []models.PlayerEvent
	db.DB.Preload("Tags").Where("club_id = ?", clubId).Order("start_time ASC").Find(&events)
	if events == nil {
		events = []models.PlayerEvent{}
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"ContinuationToken": "",
		"Events":            events,
	})
}

func HubBroadcastClubMembershipUpdate(m models.ClubMember) {
	frame := hub.NotifFrame("ClubMembershipUpdate", map[string]interface{}{
		"ClubMemberId":          m.ClubMemberId,
		"AccountId":             m.AccountId,
		"ClubId":                m.ClubId,
		"MembershipType":        m.MembershipType,
		"CreatedAt":             m.CreatedAt,
		"InvitedMembershipType": 0,
	})
	hub.HubSendToPlayer(m.AccountId, frame)
}
