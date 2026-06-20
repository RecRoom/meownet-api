package rooms

import (
	"encoding/json"
	"net/http"
	"strconv"
	"strings"

	"meow.net/controllers"
	"meow.net/db"
	"meow.net/models"
	"meow.net/utils"
)

func RoomsDispatch(w http.ResponseWriter, r *http.Request) {
	parts := strings.Split(strings.Trim(r.URL.Path, "/"), "/")
	if len(parts) >= 4 && parts[2] == "playerdata" && parts[3] == "me" {
		RoomPlayerDataMe(w, r)
		return
	}
	if len(parts) >= 5 && parts[2] == "subrooms" && parts[4] == "data" {
		RoomSaveSubRoomData(w, r)
		return
	}
	if len(parts) >= 5 && parts[2] == "subrooms" && parts[4] == "datahistory" && r.Method == http.MethodGet {
		RoomSubRoomDataHistory(w, r)
		return
	}
	if len(parts) >= 5 && parts[2] == "subrooms" && parts[4] == "restoredata" && r.Method == http.MethodPost {
		RoomRestoreSubRoomData(w, r)
		return
	}
	if len(parts) == 5 && parts[2] == "subrooms" && r.Method == http.MethodPut {
		switch parts[4] {
		case "modify":
			RoomSubRoomModify(w, r)
			return
		case "accessibility":
			RoomSubRoomAccessibility(w, r)
			return
		}
	}
	if len(parts) >= 5 && parts[2] == "roles" {
		switch parts[4] {
		case "invite":
			RoomRoleInvite(w, r)
			return
		case "acceptinvite":
			RoomRoleAcceptInvite(w, r)
			return
		}
	}
	if len(parts) == 4 && parts[2] == "roles" && r.Method == http.MethodPut {
		RoomRoleSet(w, r)
		return
	}
	if len(parts) >= 4 && parts[2] == "interactionby" && parts[3] == "me" {
		if r.Method == http.MethodGet {
			RoomInteractionGet(w, r)
			return
		} else if len(parts) >= 5 {
			if r.Method == http.MethodPut {
				if parts[4] == "cheer" {
					RoomInteractionCheer(w, r)
					return
				} else if parts[4] == "favorite" {
					RoomInteractionFavorite(w, r)
					return
				}
			} else if r.Method == http.MethodDelete {
				if parts[4] == "cheer" {
					RoomInteractionUncheer(w, r)
					return
				} else if parts[4] == "favorite" {
					RoomInteractionUnfavorite(w, r)
					return
				}
			}
		}
	}

	if len(parts) >= 3 {
		switch parts[2] {
		case "clone":
			if r.Method == http.MethodPost {
				RoomClone(w, r)
				return
			}
		case "image":
			if r.Method == http.MethodPut {
				RoomUpdateImage(w, r)
				return
			}
		case "name":
			if r.Method == http.MethodPut {
				RoomUpdateName(w, r)
				return
			}
		case "tags":
			if r.Method == http.MethodPut {
				RoomUpdateTags(w, r)
				return
			}
		case "description":
			if r.Method == http.MethodPut {
				RoomUpdateDescription(w, r)
				return
			}
		case "warning":
			if r.Method == http.MethodPut {
				RoomUpdateWarning(w, r)
				return
			}
		case "cloning":
			if r.Method == http.MethodPut {
				RoomUpdateCloning(w, r)
				return
			}
		case "accessibility":
			if r.Method == http.MethodPut {
				RoomUpdateAccessibility(w, r)
				return
			}
		case "restrictions":
			if r.Method == http.MethodPut {
				RoomUpdateRestrictions(w, r)
				return
			}
		case "automute":
			if r.Method == http.MethodPut {
				RoomUpdateAutoMute(w, r)
				return
			}
		case "comments":
			if r.Method == http.MethodPut {
				RoomUpdateComments(w, r)
				return
			}
		case "voice_chat_encryption":
			if r.Method == http.MethodPut {
				RoomUpdateVoiceChatEncryption(w, r)
				return
			}
		case "allow_new_users":
			if r.Method == http.MethodPut {
				RoomAllowNewUsers(w, r)
				return
			}
		case "subrooms":
			if r.Method == http.MethodPost && len(parts) == 3 {
				RoomCreateSubRoom(w, r)
				return
			}
		}
	}

	if len(parts) == 2 && r.Method == http.MethodDelete {
		RoomDelete(w, r)
		return
	}

	RoomsGetById(w, r)
}

func RoomsGetById(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	parts := strings.Split(strings.TrimRight(r.URL.Path, "/"), "/")
	idStr := parts[len(parts)-1]
	id, err := strconv.Atoi(idStr)
	if err != nil {
		http.Error(w, "Bad Request", http.StatusBadRequest)
		return
	}

	var room models.Room
	if err := db.DB.Preload("SubRooms").Preload("Roles").Preload("Tags").First(&room, id).Error; err != nil {
		http.Error(w, "Not Found", http.StatusNotFound)
		return
	}
	serializeSingleRoom(w, room)
}

func RoomsGet(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	name := r.URL.Query().Get("name")
	if name == "" {
		http.Error(w, "Bad Request", http.StatusBadRequest)
		return
	}

	room, err := roomByName(name)
	if err != nil {
		http.Error(w, "Not Found", http.StatusNotFound)
		return
	}
	serializeSingleRoom(w, room)
}

func serializeRoomsByIds(w http.ResponseWriter, ids []int) {
	roomList := []models.Room{}
	if len(ids) > 0 {
		db.DB.Preload("SubRooms").Preload("Roles").Preload("Tags").
			Where("room_id IN ?", ids).
			Find(&roomList)
	}
	controllers.InitRoomSlices(roomList)
	json.NewEncoder(w).Encode(roomList)
}

func RoomsBulk(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	if r.Method == http.MethodPost {
		var ids []int
		if err := json.NewDecoder(r.Body).Decode(&ids); err == nil && len(ids) > 0 {
			serializeRoomsByIds(w, ids)
			return
		}
	}

	if idParams := r.URL.Query()["id"]; len(idParams) > 0 {
		var ids []int
		for _, idStr := range idParams {
			if id, err := strconv.Atoi(idStr); err == nil {
				ids = append(ids, id)
			}
		}
		serializeRoomsByIds(w, ids)
		return
	}

	name := r.URL.Query().Get("name")
	if name == "" {
		http.Error(w, "Bad Request", http.StatusBadRequest)
		return
	}

	room, err := roomByName(name)
	if err != nil {
		http.Error(w, "Not Found", http.StatusNotFound)
		return
	}
	roomList := []models.Room{room}
	controllers.InitRoomSlices(roomList)
	json.NewEncoder(w).Encode(roomList)
}

func RoomsHot(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	tag := strings.ToLower(r.URL.Query().Get("tag"))
	query := db.DB.Preload("SubRooms").Preload("Roles").Preload("Tags").
		Where("is_dorm = ?", false).
		Where("accessibility = ?", int(models.RoomAccessibilityPublic))

	switch tag {
	case "":
		query = query.Order("visit_count DESC")
	case "new":
		query = query.Order("created_at DESC")
	default:
		query = query.Where("room_id IN (?)",
			db.DB.Table("room_tags").Select("room_id").Where("LOWER(tag) = ?", tag)).
			Order("visit_count DESC")
	}

	query = query.Limit(parseLimit(r, 50, 200))

	var roomList []models.Room
	query.Find(&roomList)
	controllers.InitRoomSlices(roomList)

	json.NewEncoder(w).Encode(map[string]interface{}{
		"TotalResults": len(roomList),
		"Results":      roomList,
	})
}

func RoomsFeaturedCurrent(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	var group models.FeaturedRoomGroup
	if err := db.DB.Order("sort_order ASC, id ASC").First(&group).Error; err != nil {
		group.Rooms = []models.FeaturedRoomItem{}
		json.NewEncoder(w).Encode(group)
		return
	}

	var entries []models.FeaturedRoomEntry
	db.DB.Where("group_id = ?", group.Id).Order("sort_order ASC, id ASC").Find(&entries)

	roomIds := make([]uint, 0, len(entries))
	for _, e := range entries {
		roomIds = append(roomIds, e.RoomId)
	}

	group.Rooms = []models.FeaturedRoomItem{}
	if len(roomIds) > 0 {
		var roomList []models.Room
		db.DB.Where("room_id IN ?", roomIds).Find(&roomList)
		byId := make(map[uint]models.Room, len(roomList))
		for _, rm := range roomList {
			byId[rm.RoomId] = rm
		}
		for _, e := range entries {
			if rm, ok := byId[e.RoomId]; ok {
				group.Rooms = append(group.Rooms, models.FeaturedRoomItem{
					RoomId:    rm.RoomId,
					RoomName:  rm.Name,
					ImageName: rm.ImageName,
				})
			}
		}
	}

	json.NewEncoder(w).Encode(group)
}

func RoomsCuratedPlaylists(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	playlistIds := []int64{}
	db.DB.Model(&models.RoomPlaylist{}).
		Order("sort_order ASC, id ASC").
		Pluck("id", &playlistIds)

	json.NewEncoder(w).Encode(playlistIds)
}

func PlaylistGet(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	id, err := strconv.ParseUint(r.PathValue("id"), 10, 64)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	var playlist models.RoomPlaylist
	if err := db.DB.First(&playlist, id).Error; err != nil {
		w.WriteHeader(http.StatusNotFound)
		return
	}

	json.NewEncoder(w).Encode(playlist)
}

func RoomsTopCreators(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	creatorLimit := parseLimit(r, 20, 100)
	perCreator := parseIntQuery(r, "perCreator", 5, 1, 25)

	type creatorAgg struct {
		CreatorAccountId int
		TotalVisits      int
	}
	var creators []creatorAgg
	db.DB.Table("rooms").
		Select("creator_account_id, SUM(visit_count) as total_visits").
		Where("is_dorm = ?", false).
		Where("accessibility = ?", int(models.RoomAccessibilityPublic)).
		Group("creator_account_id").
		Order("total_visits DESC").
		Limit(creatorLimit).
		Scan(&creators)

	out := make([]models.Room, 0, creatorLimit*perCreator)
	for _, c := range creators {
		var rooms []models.Room
		db.DB.Preload("SubRooms").Preload("Roles").Preload("Tags").
			Where("creator_account_id = ?", c.CreatorAccountId).
			Where("is_dorm = ?", false).
			Where("accessibility = ?", int(models.RoomAccessibilityPublic)).
			Order("created_at DESC").
			Limit(perCreator).
			Find(&rooms)
		controllers.InitRoomSlices(rooms)
		out = append(out, rooms...)
	}

	json.NewEncoder(w).Encode(out)
}

func parseIntQuery(r *http.Request, key string, def, min, max int) int {
	raw := r.URL.Query().Get(key)
	if raw == "" {
		return def
	}
	n, err := strconv.Atoi(raw)
	if err != nil {
		return def
	}
	if n < min {
		return min
	}
	if n > max {
		return max
	}
	return n
}

func parseLimit(r *http.Request, def, max int) int {
	raw := r.URL.Query().Get("limit")
	if raw == "" {
		return def
	}
	n, err := strconv.Atoi(raw)
	if err != nil || n <= 0 {
		return def
	}
	if n > max {
		return max
	}
	return n
}

func RoomsSearch(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	raw := r.URL.Query().Get("query")
	query := db.DB.Preload("SubRooms").Preload("Roles").Preload("Tags").
		Where("is_dorm = ?", false).
		Where("accessibility = ?", int(models.RoomAccessibilityPublic))

	var tags []string
	var nameTerms []string

	for _, term := range strings.Fields(raw) {
		if strings.HasPrefix(term, "#") {
			tag := strings.ToLower(strings.TrimPrefix(term, "#"))
			tags = append(tags, tag)
		} else {
			nameTerms = append(nameTerms, term)
		}
	}

	if len(tags) > 0 {
		query = query.Where("room_id IN (?)",
			db.DB.Table("room_tags").Select("room_id").
				Where("LOWER(tag) IN ?", tags))
	}

	for _, term := range nameTerms {
		query = query.Where(`LOWER(name) LIKE ? ESCAPE '\'`, "%"+utils.EscapeLike(strings.ToLower(term))+"%")
	}

	var roomList []models.Room
	query.Find(&roomList)
	controllers.InitRoomSlices(roomList)

	json.NewEncoder(w).Encode(map[string]interface{}{
		"TotalResults": len(roomList),
		"Results":      roomList,
	})
}

func RoomRooms(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	name := r.URL.Query().Get("name")
	if name != "" {
		room, err := roomByName(name)
		if err != nil {
			http.Error(w, "Not Found", http.StatusNotFound)
			return
		}
		serializeSingleRoom(w, room)
		return
	}

	includeIds := r.URL.Query()["include"]
	var roomList []models.Room
	query := db.DB.Preload("SubRooms").Preload("Roles").Preload("Tags").
		Where("is_dorm = ?", false).
		Where("accessibility = ?", int(models.RoomAccessibilityPublic))

	if len(includeIds) > 0 {
		var ids []int
		for _, idStr := range includeIds {
			if id, err := strconv.Atoi(idStr); err == nil {
				ids = append(ids, id)
			}
		}
		query = query.Where("room_id IN ?", ids)
	}

	query.Find(&roomList)
	controllers.InitRoomSlices(roomList)
	json.NewEncoder(w).Encode(roomList)
}

func RoomCreatedByMe(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	tokenStr := utils.GetBearerToken(r)
	if tokenStr == "" {
		json.NewEncoder(w).Encode([]interface{}{})
		return
	}
	accountIdStr, err := utils.ParseSubFromJWT(tokenStr)
	if err != nil {
		json.NewEncoder(w).Encode([]interface{}{})
		return
	}
	accountId, _ := strconv.Atoi(accountIdStr)

	var roomList []models.Room
	db.DB.Preload("SubRooms").Preload("Roles").Preload("Tags").Where("creator_account_id = ?", accountId).Find(&roomList)
	controllers.InitRoomSlices(roomList)

	json.NewEncoder(w).Encode(roomList)
}

func RoomVisitedByMe(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	tokenStr := utils.GetBearerToken(r)
	if tokenStr == "" {
		json.NewEncoder(w).Encode([]interface{}{})
		return
	}
	accountIdStr, err := utils.ParseSubFromJWT(tokenStr)
	if err != nil {
		json.NewEncoder(w).Encode([]interface{}{})
		return
	}
	accountId, _ := strconv.Atoi(accountIdStr)

	var interactions []models.RoomInteraction
	db.DB.Where("account_id = ? AND visited = ?", accountId, true).Find(&interactions)

	roomIds := make([]uint, 0, len(interactions))
	for _, i := range interactions {
		roomIds = append(roomIds, i.RoomId)
	}

	var roomList []models.Room
	if len(roomIds) > 0 {
		db.DB.Preload("SubRooms").Preload("Roles").Preload("Tags").Where("room_id IN ?", roomIds).Find(&roomList)
	}
	controllers.InitRoomSlices(roomList)
	json.NewEncoder(w).Encode(roomList)
}

func RoomFavoritedByMe(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	tokenStr := utils.GetBearerToken(r)
	if tokenStr == "" {
		json.NewEncoder(w).Encode([]interface{}{})
		return
	}
	accountIdStr, err := utils.ParseSubFromJWT(tokenStr)
	if err != nil {
		json.NewEncoder(w).Encode([]interface{}{})
		return
	}
	accountId, _ := strconv.Atoi(accountIdStr)

	var interactions []models.RoomInteraction
	db.DB.Where("account_id = ? AND favorited = ?", accountId, true).Find(&interactions)

	roomIds := make([]uint, 0, len(interactions))
	for _, i := range interactions {
		roomIds = append(roomIds, i.RoomId)
	}

	var roomList []models.Room
	if len(roomIds) > 0 {
		db.DB.Preload("SubRooms").Preload("Roles").Preload("Tags").Where("room_id IN ?", roomIds).Find(&roomList)
	}
	controllers.InitRoomSlices(roomList)
	json.NewEncoder(w).Encode(roomList)
}

func RoomModeratedByMe(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	tokenStr := utils.GetBearerToken(r)
	if tokenStr == "" {
		json.NewEncoder(w).Encode([]interface{}{})
		return
	}
	accountIdStr, err := utils.ParseSubFromJWT(tokenStr)
	if err != nil {
		json.NewEncoder(w).Encode([]interface{}{})
		return
	}
	accountId, _ := strconv.Atoi(accountIdStr)

	var roomList []models.Room
	db.DB.Preload("SubRooms").Preload("Roles").Preload("Tags").
		Where("room_id IN (?)",
			db.DB.Table("room_roles").Select("room_id").
				Where("account_id = ? AND (role > 0 OR invited_role > 0)", accountId)).
		Find(&roomList)
	controllers.InitRoomSlices(roomList)

	json.NewEncoder(w).Encode(roomList)
}

func RoomFilters(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"PinnedFilters":   []string{"rro", "community", "featured", "quest", "pvp", "hangout", "game", "art", "horror"},
		"PopularFilters":  []string{"pvp", "quest", "game", "hangout", "art"},
		"TrendingFilters": []string{"featured", "game", "horror", "quest"},
	})
}
