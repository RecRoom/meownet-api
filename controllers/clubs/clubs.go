package clubs

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"strings"

	"gorm.io/gorm"
	"meow.net/controllers"
	"meow.net/controllers/hub"
	"meow.net/db"
	"meow.net/models"
	"meow.net/utils"
)

var clubCategories = []string{
	"Social",
	"Creative",
	"Competitive",
	"Casual",
	"Entertainment",
}

func ClubDispatch(w http.ResponseWriter, r *http.Request) {
	path := strings.Trim(r.URL.Path, "/")
	parts := strings.Split(path, "/")

	if len(parts) >= 2 {
		if id, err := strconv.ParseInt(parts[1], 10, 64); err == nil {
			rest := strings.Join(parts[2:], "/")
			switch {
			case rest == "":
				if r.Method == http.MethodDelete {
					ClubDelete(w, r, id)
					return
				}
			case rest == "details":
				ClubDetails(w, r, id)
				return
			case rest == "members":
				if r.Method == http.MethodGet {
					ClubMembers(w, r, id)
					return
				}
			case rest == "members/requesttojoin":
				ClubRequestToJoin(w, r, id)
				return
			case rest == "members/invite":
				if r.Method == http.MethodPut {
					ClubMemberInvite(w, r, id)
					return
				}
			case rest == "modify":
				if r.Method == http.MethodPut {
					ClubModify(w, r, id)
					return
				}
			case rest == "modifydetails":
				ClubModifyDetails(w, r, id)
				return
			case rest == "mainimage":
				if r.Method == http.MethodPut {
					ClubMainImage(w, r, id)
					return
				}
			case rest == "clubhouse":
				if r.Method == http.MethodPut {
					ClubSetClubhouse(w, r, id)
					return
				}
			}
		}
	}

	log.Printf("[CLUB] unhandled %s %s", r.Method, r.URL.Path)
	http.NotFound(w, r)
}

func ClubCategoryTags(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(clubCategories)
}

func ClubCreate(w http.ResponseWriter, r *http.Request) {
	accountID, ok := controllers.AccountIDFromRequest(r)
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}
	if err := r.ParseForm(); err != nil {
		http.Error(w, "Bad Request", http.StatusBadRequest)
		return
	}
	name := strings.TrimSpace(r.FormValue("name"))
	description := r.FormValue("description")
	category := strings.TrimSpace(r.FormValue("category"))
	if name == "" {
		writeJsonError(w, http.StatusBadRequest, "name required")
		return
	}
	if !utils.IsValidName(name) {
		writeJsonError(w, http.StatusBadRequest, "club names can only use letters, numbers, and basic punctuation")
		return
	}
	if !utils.IsValidNameLength(name) {
		writeJsonError(w, http.StatusBadRequest, "club names can be at most 16 characters")
		return
	}
	if utils.IsAnyTextFlagged(name, description) {
		writeJsonError(w, http.StatusBadRequest, "club name or description violates the community guidelines")
		return
	}
	if category == "" {
		category = "Social"
	}
	clubId := randomClubId()
	for {
		var existing models.Club
		if err := db.DB.First(&existing, clubId).Error; err == gorm.ErrRecordNotFound {
			break
		}
		clubId = randomClubId()
	}

	club := models.Club{
		ClubId:           clubId,
		Name:             name,
		Description:      description,
		Category:         category,
		Visibility:       int(models.ClubVisibilityPublic),
		Joinability:      int(models.ClubJoinabilityOpen),
		AllowJuniors:     true,
		MainImageName:    "DefaultClub.png",
		ClubType:         0,
		CreatorAccountId: int(accountID),
		MemberCount:      1,
	}
	if err := db.DB.Create(&club).Error; err != nil {
		log.Printf("[CLUB] create error: %v", err)
		writeJsonError(w, http.StatusInternalServerError, "create failed")
		return
	}

	perms := newClubPermissions(clubId)
	db.DB.Create(&perms)

	owner := models.ClubMember{
		ClubId:         clubId,
		AccountId:      int(accountID),
		MembershipType: int(models.ClubMembershipCreator),
	}
	db.DB.Create(&owner)

	writeJsonEnvelope(w, clubDetailsResponse(club, int(accountID)))
}

func ClubDetails(w http.ResponseWriter, r *http.Request, clubId int64) {
	accountID, _ := controllers.AccountIDFromRequest(r)

	var club models.Club
	if err := db.DB.First(&club, clubId).Error; err != nil {
		http.Error(w, "Not Found", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(clubDetailsResponse(club, int(accountID)))
}

func ClubMembers(w http.ResponseWriter, r *http.Request, clubId int64) {
	log.Printf("[CLUB] members id=%d %s", clubId, r.URL.RawQuery)

	q := db.DB.Where("club_id = ?", clubId)
	if mtStr := r.URL.Query().Get("membershipType"); mtStr != "" {
		if mt, err := strconv.Atoi(mtStr); err == nil {
			q = q.Where("membership_type = ?", mt)
		}
	}

	sortBy := r.URL.Query().Get("sortBy")
	switch sortBy {
	case "1":
		q = q.Order("account_id asc")
	case "2":
		q = q.Order("created_at asc")
	default:
		q = q.Order("membership_type desc, created_at asc")
	}

	var members []models.ClubMember
	q.Find(&members)
	if members == nil {
		members = []models.ClubMember{}
	}

	writeJsonEnvelope(w, members)
}

func ClubMemberInvite(w http.ResponseWriter, r *http.Request, clubId int64) {
	accountID, ok := controllers.AccountIDFromRequest(r)
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}
	if err := r.ParseForm(); err != nil {
		http.Error(w, "Bad Request", http.StatusBadRequest)
		return
	}

	inviteeId, err := strconv.Atoi(r.FormValue("accountId"))
	if err != nil || inviteeId == 0 {
		http.Error(w, "Bad Request", http.StatusBadRequest)
		return
	}

	membershipType, err := strconv.Atoi(r.FormValue("membershipType"))
	if err != nil {
		membershipType = int(models.ClubMembershipMember)
	}

	var club models.Club
	if err := db.DB.First(&club, clubId).Error; err != nil {
		writeJsonError(w, http.StatusNotFound, "club not found")
		return
	}

	mt := myMembershipType(clubId, int(accountID))
	if mt < int(models.ClubMembershipCoowner) {
		writeJsonError(w, http.StatusForbidden, "insufficient permissions")
		return
	}

	var existing models.ClubMember
	if db.DB.Where("club_id = ? AND account_id = ?", clubId, inviteeId).First(&existing).Error != nil {
		db.DB.Create(&models.ClubMember{
			ClubId:         clubId,
			AccountId:      inviteeId,
			MembershipType: int(models.ClubMembershipPendingInvited),
		})
	}

	var msgType models.MessageType
	switch models.ClubMembershipType(membershipType) {
	case models.ClubMembershipModerator:
		msgType = models.MessageTypeClubModeratorInvited
	case models.ClubMembershipCoowner:
		msgType = models.MessageTypeClubCoownerInvited
	default:
		msgType = models.MessageTypeClubMemberInvited
	}

	msg := models.Message{
		FromPlayerId: accountID,
		ToPlayerId:   uint(inviteeId),
		Type:         int(msgType),
		Data:         strconv.FormatInt(clubId, 10),
	}
	db.DB.Create(&msg)
	hub.HubSendToPlayer(inviteeId, hub.NotifFrame(models.MessageReceived, msg))

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"error":   nil,
		"success": true,
		"value":   nil,
	})
}

func ClubRequestToJoin(w http.ResponseWriter, r *http.Request, clubId int64) {
	accountID, ok := controllers.AccountIDFromRequest(r)
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}
	log.Printf("[CLUB] requesttojoin id=%d account=%d", clubId, accountID)

	var club models.Club
	if err := db.DB.First(&club, clubId).Error; err != nil {
		writeJsonError(w, http.StatusNotFound, "club not found")
		return
	}

	var existing models.ClubMember
	if err := db.DB.Where("club_id = ? AND account_id = ?", clubId, accountID).First(&existing).Error; err == nil {
		writeJsonEnvelope(w, existing)
		return
	}

	member := models.ClubMember{
		ClubId:         clubId,
		AccountId:      int(accountID),
		MembershipType: int(models.ClubMembershipMember),
	}
	if err := db.DB.Create(&member).Error; err != nil {
		log.Printf("[CLUB] requesttojoin create error: %v", err)
		writeJsonError(w, http.StatusInternalServerError, "join failed")
		return
	}

	recountClubMembers(clubId)
	HubBroadcastClubMembershipUpdate(member)

	writeJsonEnvelope(w, member)
}

func ClubModifyDetails(w http.ResponseWriter, r *http.Request, clubId int64) {
	accountID, ok := controllers.AccountIDFromRequest(r)
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}
	if err := r.ParseForm(); err != nil {
		http.Error(w, "Bad Request", http.StatusBadRequest)
		return
	}

	var club models.Club
	if err := db.DB.First(&club, clubId).Error; err != nil {
		writeJsonError(w, http.StatusNotFound, "club not found")
		return
	}

	mt := myMembershipType(clubId, int(accountID))
	if mt < int(models.ClubMembershipCoowner) {
		writeJsonError(w, http.StatusForbidden, "insufficient permissions")
		return
	}

	updates := map[string]interface{}{}
	if v := r.FormValue("visibility"); v != "" {
		if iv, ok := parseVisibility(v); ok {
			updates["visibility"] = iv
		}
	}
	if v := r.FormValue("joinability"); v != "" {
		if iv, ok := parseJoinability(v); ok {
			updates["joinability"] = iv
		}
	}
	if v := r.FormValue("allowJuniors"); v != "" {
		if bv, ok := parseBoolForm(v); ok {
			updates["allow_juniors"] = bv
		}
	}
	if v := r.FormValue("name"); v != "" {
		if !utils.IsValidName(v) {
			writeJsonError(w, http.StatusBadRequest, "club names can only use letters, numbers, and basic punctuation")
			return
		}
		if !utils.IsValidNameLength(v) {
			writeJsonError(w, http.StatusBadRequest, "club names can be at most 16 characters")
			return
		}
		if utils.IsTextFlagged(v) {
			writeJsonError(w, http.StatusBadRequest, "club name violates the community guidelines")
			return
		}
		updates["name"] = v
	}
	if v := r.FormValue("description"); v != "" {
		if utils.IsTextFlagged(v) {
			writeJsonError(w, http.StatusBadRequest, "club description violates the community guidelines")
			return
		}
		updates["description"] = v
	}
	if v := r.FormValue("category"); v != "" {
		updates["category"] = v
	}
	if v := r.FormValue("mainImageName"); v != "" {
		updates["main_image_name"] = v
	}
	if v := r.FormValue("minLevel"); v != "" {
		if iv, err := strconv.Atoi(v); err == nil {
			updates["min_level"] = iv
		}
	}

	log.Printf("[CLUB] modifydetails id=%d updates=%v", clubId, updates)
	if len(updates) > 0 {
		db.DB.Model(&models.Club{}).Where("club_id = ?", clubId).Updates(updates)
		db.DB.First(&club, clubId)
	}

	if customTags, ok := r.Form["customTags"]; ok {
		db.DB.Where("club_id = ?", clubId).Delete(&models.ClubCustomTag{})
		seen := map[string]bool{}
		for _, tag := range customTags {
			tag = strings.TrimSpace(tag)
			if tag == "" {
				continue
			}
			key := strings.ToLower(tag)
			if seen[key] {
				continue
			}
			if utils.IsTextFlagged(tag) {
				writeJsonError(w, http.StatusBadRequest, "club tag violates the community guidelines")
				return
			}
			seen[key] = true
			db.DB.Create(&models.ClubCustomTag{ClubId: clubId, Tag: tag})
		}
	}

	writeJsonEnvelope(w, clubDetailsResponse(club, int(accountID)))
}

func ClubModify(w http.ResponseWriter, r *http.Request, clubId int64) {
	accountID, ok := controllers.AccountIDFromRequest(r)
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}
	if err := r.ParseForm(); err != nil {
		http.Error(w, "Bad Request", http.StatusBadRequest)
		return
	}

	var club models.Club
	if err := db.DB.First(&club, clubId).Error; err != nil {
		writeJsonError(w, http.StatusNotFound, "club not found")
		return
	}

	mt := myMembershipType(clubId, int(accountID))
	if mt < int(models.ClubMembershipCoowner) {
		writeJsonError(w, http.StatusForbidden, "insufficient permissions")
		return
	}

	updates := map[string]interface{}{}
	if v := r.FormValue("name"); v != "" {
		if !utils.IsValidName(v) {
			writeJsonError(w, http.StatusBadRequest, "club names can only use letters, numbers, and basic punctuation")
			return
		}
		if !utils.IsValidNameLength(v) {
			writeJsonError(w, http.StatusBadRequest, "club names can be at most 16 characters")
			return
		}
		if utils.IsTextFlagged(v) {
			writeJsonError(w, http.StatusBadRequest, "club name violates the community guidelines")
			return
		}
		updates["name"] = v
	}
	if r.Form.Has("description") {
		v := r.FormValue("description")
		if utils.IsTextFlagged(v) {
			writeJsonError(w, http.StatusBadRequest, "club description violates the community guidelines")
			return
		}
		updates["description"] = v
	}
	if v := r.FormValue("category"); v != "" {
		updates["category"] = v
	}

	log.Printf("[CLUB] modify id=%d updates=%v", clubId, updates)
	if len(updates) > 0 {
		db.DB.Model(&models.Club{}).Where("club_id = ?", clubId).Updates(updates)
		db.DB.First(&club, clubId)
	}

	writeJsonEnvelope(w, clubDetailsResponse(club, int(accountID)))
}

func ClubDelete(w http.ResponseWriter, r *http.Request, clubId int64) {
	accountID, ok := controllers.AccountIDFromRequest(r)
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	var club models.Club
	if err := db.DB.First(&club, clubId).Error; err != nil {
		writeJsonError(w, http.StatusNotFound, "club not found")
		return
	}

	if club.CreatorAccountId != int(accountID) {
		writeJsonError(w, http.StatusForbidden, "only the creator can delete the club")
		return
	}

	tx := db.DB.Begin()
	tx.Where("club_id = ?", clubId).Delete(&models.ClubMember{})
	tx.Where("club_id = ?", clubId).Delete(&models.ClubPermission{})
	tx.Where("club_id = ?", clubId).Delete(&models.ClubCustomTag{})
	tx.Where("club_id = ?", clubId).Delete(&models.ClubAnnouncement{})
	if err := tx.Delete(&models.Club{}, clubId).Error; err != nil {
		tx.Rollback()
		log.Printf("[CLUB] delete error: %v", err)
		writeJsonError(w, http.StatusInternalServerError, "delete failed")
		return
	}
	tx.Commit()

	writeJsonEnvelope(w, nil)
}

func ClubMainImage(w http.ResponseWriter, r *http.Request, clubId int64) {
	accountID, ok := controllers.AccountIDFromRequest(r)
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}
	if err := r.ParseForm(); err != nil {
		http.Error(w, "Bad Request", http.StatusBadRequest)
		return
	}

	var club models.Club
	if err := db.DB.First(&club, clubId).Error; err != nil {
		writeJsonError(w, http.StatusNotFound, "club not found")
		return
	}

	mt := myMembershipType(clubId, int(accountID))
	if mt < int(models.ClubMembershipCoowner) {
		writeJsonError(w, http.StatusForbidden, "insufficient permissions")
		return
	}

	imageName := r.FormValue("imageName")
	if imageName == "" {
		writeJsonError(w, http.StatusBadRequest, "imageName required")
		return
	}

	db.DB.Model(&models.Club{}).Where("club_id = ?", clubId).Update("main_image_name", imageName)
	club.MainImageName = imageName

	writeJsonEnvelope(w, clubDetailsResponse(club, int(accountID)))
}

func ClubSetClubhouse(w http.ResponseWriter, r *http.Request, clubId int64) {
	accountID, ok := controllers.AccountIDFromRequest(r)
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}
	if err := r.ParseForm(); err != nil {
		http.Error(w, "Bad Request", http.StatusBadRequest)
		return
	}

	var club models.Club
	if err := db.DB.First(&club, clubId).Error; err != nil {
		writeJsonError(w, http.StatusNotFound, "club not found")
		return
	}

	mt := myMembershipType(clubId, int(accountID))
	if mt < int(models.ClubMembershipCoowner) {
		writeJsonError(w, http.StatusForbidden, "insufficient permissions")
		return
	}

	var update map[string]interface{}
	if v := r.FormValue("roomId"); v != "" {
		roomId, err := strconv.ParseInt(v, 10, 64)
		if err != nil {
			writeJsonError(w, http.StatusBadRequest, "invalid roomId")
			return
		}
		update = map[string]interface{}{"clubhouse_room_id": roomId}
	} else {
		update = map[string]interface{}{"clubhouse_room_id": nil}
	}

	db.DB.Model(&models.Club{}).Where("club_id = ?", clubId).Updates(update)

	writeJsonEnvelope(w, nil)
}

func ClubHomeMe(w http.ResponseWriter, r *http.Request) {
	accountID, ok := controllers.AccountIDFromRequest(r)
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	if controllers.AccountIsBanned(accountID) {
		http.NotFound(w, r)
		return
	}

	switch r.Method {
	case http.MethodGet:
		clubHomeGet(w, accountID)
	case http.MethodPut:
		clubHomeSet(w, r, accountID)
	default:
		http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
	}
}

func clubHomeGet(w http.ResponseWriter, accountID uint) {
	var account models.Account
	if err := db.DB.First(&account, accountID).Error; err != nil || account.HomeClubId == nil {
		http.NotFound(w, nil)
		return
	}
	var club models.Club
	if err := db.DB.First(&club, *account.HomeClubId).Error; err != nil {
		http.NotFound(w, nil)
		return
	}
	if club.ClubhouseRoomId == nil {
		http.NotFound(w, nil)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(club)
}

func clubHomeSet(w http.ResponseWriter, r *http.Request, accountID uint) {
	if err := r.ParseForm(); err != nil {
		http.Error(w, "Bad Request", http.StatusBadRequest)
		return
	}
	clubId, err := strconv.ParseInt(r.FormValue("clubId"), 10, 64)
	if err != nil || clubId == 0 {
		writeJsonError(w, http.StatusBadRequest, "invalid clubId")
		return
	}

	var club models.Club
	if err := db.DB.First(&club, clubId).Error; err != nil {
		writeJsonError(w, http.StatusNotFound, "club not found")
		return
	}

	if myMembershipType(clubId, int(accountID)) < int(models.ClubMembershipMember) {
		writeJsonError(w, http.StatusForbidden, "not a member")
		return
	}

	db.DB.Model(&models.Account{}).Where("account_id = ?", accountID).Update("home_club_id", clubId)

	writeJsonEnvelope(w, club)
}

func ClubMineCreated(w http.ResponseWriter, r *http.Request) {
	accountID, ok := controllers.AccountIDFromRequest(r)
	if !ok {
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprint(w, "[]")
		return
	}
	log.Printf("[CLUB] mine/created account=%d", accountID)

	var clubs []models.Club
	db.DB.Where("creator_account_id = ? AND club_type != ?", accountID, 1).Order("created_at asc").Find(&clubs)
	if clubs == nil {
		clubs = []models.Club{}
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(clubs)
}

func ClubMineMember(w http.ResponseWriter, r *http.Request) {
	accountID, ok := controllers.AccountIDFromRequest(r)
	if !ok {
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprint(w, "[]")
		return
	}

	var clubs []models.Club
	db.DB.
		Joins("JOIN club_members ON club_members.club_id = clubs.club_id").
		Where("club_members.account_id = ? AND club_members.membership_type >= ?", accountID, int(models.ClubMembershipMember)).
		Where("clubs.club_type != ?", 1).
		Order("clubs.created_at asc").
		Find(&clubs)
	if clubs == nil {
		clubs = []models.Club{}
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(clubs)
}

func ClubSearch(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query()
	category := q.Get("category")
	query := q.Get("query")
	count, _ := strconv.Atoi(q.Get("count"))
	if count <= 0 || count > 100 {
		count = 30
	}
	log.Printf("[CLUB] search category=%q query=%q count=%d", category, query, count)

	dq := db.DB.Where("visibility = ?", int(models.ClubVisibilityPublic)).Where("club_type != ?", 1)
	if category != "" {
		dq = dq.Where("LOWER(category) = LOWER(?)", category)
	}
	if query != "" {
		like := "%" + utils.EscapeLike(strings.ToLower(query)) + "%"
		dq = dq.Where(`LOWER(name) LIKE ? ESCAPE '\' OR LOWER(description) LIKE ? ESCAPE '\'`, like, like)
	}

	switch q.Get("sort") {
	case "1":
		dq = dq.Order("created_at desc")
	case "2":
		dq = dq.Order("name asc")
	default:
		dq = dq.Order("member_count desc, created_at desc")
	}

	var total int64
	dq.Model(&models.Club{}).Count(&total)

	var clubs []models.Club
	dq.Limit(count).Find(&clubs)
	if clubs == nil {
		clubs = []models.Club{}
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"Clubs":             clubs,
		"ContinuationToken": nil,
		"TotalClubs":        total,
	})
}
