package clubs

import (
	"crypto/rand"
	"encoding/binary"
	"encoding/json"
	"net/http"
	"strconv"
	"strings"

	"meow.net/db"
	"meow.net/models"
)

func randomClubId() int64 {
	var b [4]byte
	rand.Read(b[:])
	return int64(binary.BigEndian.Uint32(b[:]) >> 1)
}

func newClubPermissions(clubId int64) []models.ClubPermission {
	return []models.ClubPermission{
		{
			ClubId:                 clubId,
			Type:                   int(models.ClubMembershipCoowner),
			ApproveMember:          true,
			BanUnban:               true,
			CreateEvent:            true,
			EditDetails:            true,
			EditPermissionSettings: true,
			PostAnnouncement:       true,
		},
		{
			ClubId:        clubId,
			Type:          int(models.ClubMembershipModerator),
			ApproveMember: true,
			BanUnban:      true,
		},
		{
			ClubId: clubId,
			Type:   int(models.ClubMembershipMember),
		},
	}
}

func loadClubPermissions(clubId int64) (coowner, moderator, member models.ClubPermission) {
	var perms []models.ClubPermission
	db.DB.Where("club_id = ?", clubId).Find(&perms)
	for _, p := range perms {
		switch p.Type {
		case int(models.ClubMembershipCoowner):
			coowner = p
		case int(models.ClubMembershipModerator):
			moderator = p
		case int(models.ClubMembershipMember):
			member = p
		}
	}
	return
}

func loadClubCustomTags(clubId int64) []string {
	var rows []models.ClubCustomTag
	db.DB.Where("club_id = ?", clubId).Find(&rows)
	tags := make([]string, 0, len(rows))
	for _, r := range rows {
		tags = append(tags, r.Tag)
	}
	return tags
}

func myMembershipType(clubId int64, accountId int) int {
	if accountId == 0 {
		return 0
	}
	var m models.ClubMember
	if err := db.DB.Where("club_id = ? AND account_id = ?", clubId, accountId).First(&m).Error; err != nil {
		return 0
	}
	return m.MembershipType
}

func recountClubMembers(clubId int64) int {
	var count int64
	db.DB.Model(&models.ClubMember{}).Where("club_id = ?", clubId).Count(&count)
	db.DB.Model(&models.Club{}).Where("club_id = ?", clubId).Update("member_count", count)
	return int(count)
}

func clubDetailsResponse(club models.Club, accountId int) map[string]interface{} {
	coowner, moderator, member := loadClubPermissions(club.ClubId)
	return map[string]interface{}{
		"AdditionalImages":     []interface{}{},
		"Club":                 club,
		"ClubId":               club.ClubId,
		"CoownerPermissions":   coowner,
		"CustomTags":           loadClubCustomTags(club.ClubId),
		"MemberPermissions":    member,
		"ModeratorPermissions": moderator,
		"MyMembershipType":     myMembershipType(club.ClubId, accountId),
	}
}

func parseBoolForm(s string) (bool, bool) {
	switch strings.ToLower(strings.TrimSpace(s)) {
	case "true", "1", "yes":
		return true, true
	case "false", "0", "no":
		return false, true
	}
	return false, false
}

func parseVisibility(s string) (int, bool) {
	switch strings.ToLower(strings.TrimSpace(s)) {
	case "private", "0":
		return int(models.ClubVisibilityPrivate), true
	case "public", "1":
		return int(models.ClubVisibilityPublic), true
	}
	return 0, false
}

func parseJoinability(s string) (int, bool) {
	switch strings.ToLower(strings.TrimSpace(s)) {
	case "open", "0":
		return int(models.ClubJoinabilityOpen), true
	case "inviteonly", "invite_only", "1":
		return int(models.ClubJoinabilityInviteOnly), true
	case "requesttojoin", "request_to_join", "2":
		return int(models.ClubJoinabilityAskToJoin), true
	}
	return 0, false
}

func parseClubIdFromPath(path string) (int64, bool) {
	parts := strings.Split(strings.Trim(path, "/"), "/")
	for i, seg := range parts {
		if seg == "club" && i+1 < len(parts) {
			id, err := strconv.ParseInt(parts[i+1], 10, 64)
			if err != nil {
				return 0, false
			}
			return id, true
		}
	}
	return 0, false
}

func writeJsonEnvelope(w http.ResponseWriter, value interface{}) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"error":   "",
		"success": true,
		"value":   value,
	})
}

func writeJsonError(w http.ResponseWriter, status int, msg string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"error":   msg,
		"success": false,
		"value":   nil,
	})
}
