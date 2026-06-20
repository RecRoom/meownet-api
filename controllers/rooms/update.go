package rooms

import (
	"net/http"
	"strconv"
	"strings"

	"meow.net/controllers"
	"meow.net/controllers/hub"
	"meow.net/db"
	"meow.net/models"
	"meow.net/utils"
)

func RoomUpdateName(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	parts := strings.Split(strings.Trim(r.URL.Path, "/"), "/")
	if len(parts) < 2 {
		http.Error(w, "Bad Request", http.StatusBadRequest)
		return
	}
	roomId, err := strconv.Atoi(parts[1])
	if err != nil {
		http.Error(w, "Bad Request", http.StatusBadRequest)
		return
	}

	accountId, err := controllers.CurrentUserIDFromRequest(r)
	if err != nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	room, ok := loadRoom(roomId)
	if !ok {
		http.Error(w, "Not Found", http.StatusNotFound)
		return
	}

	if !isRoomOwner(room, int(accountId)) {
		http.Error(w, "Forbidden", http.StatusForbidden)
		return
	}

	name := strings.TrimSpace(r.FormValue("name"))
	if msg, ok := validateRoomName(name, roomId); !ok {
		writeRoomModerationRejection(w, msg)
		return
	}

	db.DB.Model(&models.Room{}).Where("room_id = ?", roomId).Update("name", name)
	room.Name = name

	updated, ok := loadRoom(roomId)
	if !ok {
		updated = room
	}
	frame := hub.NotifFrame("RoomUpdate", updated)
	notified := map[int]bool{}
	for _, role := range updated.Roles {
		if role.AccountId != 0 && !notified[role.AccountId] {
			hub.HubSendToPlayer(role.AccountId, frame)
			notified[role.AccountId] = true
		}
	}
	if updated.CreatorAccountId != 0 && !notified[updated.CreatorAccountId] {
		hub.HubSendToPlayer(updated.CreatorAccountId, frame)
	}

	var instances []models.RoomInstance
	db.DB.Where("room_id = ?", roomId).Find(&instances)
	for _, inst := range instances {
		hub.HubBroadcastRoomInstanceUpdate(inst.Id)
	}

	roomSuccessResponse(w, updated)
}

func RoomUpdateImage(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	parts := strings.Split(strings.Trim(r.URL.Path, "/"), "/")
	if len(parts) < 2 {
		http.Error(w, "Bad Request", http.StatusBadRequest)
		return
	}
	roomId, err := strconv.Atoi(parts[1])
	if err != nil {
		http.Error(w, "Bad Request", http.StatusBadRequest)
		return
	}

	accountId, err := controllers.CurrentUserIDFromRequest(r)
	if err != nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	room, ok := loadRoom(roomId)
	if !ok {
		http.Error(w, "Not Found", http.StatusNotFound)
		return
	}

	if !isRoomOwner(room, int(accountId)) {
		http.Error(w, "Forbidden", http.StatusForbidden)
		return
	}

	imageName := r.FormValue("imageName")
	if imageName == "" {
		http.Error(w, "Bad Request", http.StatusBadRequest)
		return
	}

	db.DB.Model(&models.Room{}).Where("room_id = ?", roomId).Update("image_name", imageName)
	room.ImageName = imageName
	roomSuccessResponse(w, room)
}

func RoomUpdateTags(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	parts := strings.Split(strings.Trim(r.URL.Path, "/"), "/")
	if len(parts) < 2 {
		http.Error(w, "Bad Request", http.StatusBadRequest)
		return
	}
	roomId, err := strconv.Atoi(parts[1])
	if err != nil {
		http.Error(w, "Bad Request", http.StatusBadRequest)
		return
	}

	accountId, err := controllers.CurrentUserIDFromRequest(r)
	if err != nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	room, ok := loadRoom(roomId)
	if !ok {
		http.Error(w, "Not Found", http.StatusNotFound)
		return
	}

	if !isRoomOwner(room, int(accountId)) {
		http.Error(w, "Forbidden", http.StatusForbidden)
		return
	}

	r.ParseForm()

	type tagInput struct {
		value string
		typ   int
	}
	var inputs []tagInput
	for _, v := range r.Form["autoTag"] {
		if v != "" {
			inputs = append(inputs, tagInput{value: v, typ: 1})
		}
	}
	for _, v := range r.Form["tag"] {
		if v != "" {
			inputs = append(inputs, tagInput{value: v, typ: 0})
		}
	}

	db.DB.Where("room_id = ?", roomId).Delete(&models.RoomTag{})
	room.Tags = nil

	for _, in := range inputs {
		newTag := models.RoomTag{RoomId: room.RoomId, Tag: in.value, Type: in.typ}
		db.DB.Create(&newTag)
		room.Tags = append(room.Tags, newTag)
	}

	roomSuccessResponse(w, room)
}

func RoomUpdateDescription(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	parts := strings.Split(strings.Trim(r.URL.Path, "/"), "/")
	if len(parts) < 2 {
		http.Error(w, "Bad Request", http.StatusBadRequest)
		return
	}
	roomId, err := strconv.Atoi(parts[1])
	if err != nil {
		http.Error(w, "Bad Request", http.StatusBadRequest)
		return
	}

	accountId, err := controllers.CurrentUserIDFromRequest(r)
	if err != nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	room, ok := loadRoom(roomId)
	if !ok {
		http.Error(w, "Not Found", http.StatusNotFound)
		return
	}

	if !isRoomOwner(room, int(accountId)) {
		http.Error(w, "Forbidden", http.StatusForbidden)
		return
	}

	description := r.FormValue("description")
	if utils.IsTextFlagged(description) {
		writeRoomModerationRejection(w, "Room description violates the community guidelines.")
		return
	}
	db.DB.Model(&models.Room{}).Where("room_id = ?", roomId).Update("description", description)
	room.Description = description
	roomSuccessResponse(w, room)
}

func RoomUpdateWarning(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	parts := strings.Split(strings.Trim(r.URL.Path, "/"), "/")
	if len(parts) < 2 {
		http.Error(w, "Bad Request", http.StatusBadRequest)
		return
	}
	roomId, err := strconv.Atoi(parts[1])
	if err != nil {
		http.Error(w, "Bad Request", http.StatusBadRequest)
		return
	}

	accountId, err := controllers.CurrentUserIDFromRequest(r)
	if err != nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	room, ok := loadRoom(roomId)
	if !ok {
		http.Error(w, "Not Found", http.StatusNotFound)
		return
	}

	if !isRoomOwner(room, int(accountId)) {
		http.Error(w, "Forbidden", http.StatusForbidden)
		return
	}

	warningMask := 0
	if maskStr := r.FormValue("warningMask"); maskStr != "" {
		warningMask, err = strconv.Atoi(maskStr)
		if err != nil {
			http.Error(w, "Bad Request", http.StatusBadRequest)
			return
		}
	}
	customWarning := r.FormValue("customWarning")
	if utils.IsTextFlagged(customWarning) {
		writeRoomModerationRejection(w, "Custom warning violates the community guidelines.")
		return
	}

	db.DB.Model(&models.Room{}).Where("room_id = ?", roomId).Updates(map[string]interface{}{
		"warning_mask":   warningMask,
		"custom_warning": customWarning,
	})
	room.WarningMask = warningMask
	room.CustomWarning = customWarning
	roomSuccessResponse(w, room)
}

func RoomUpdateAccessibility(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	parts := strings.Split(strings.Trim(r.URL.Path, "/"), "/")
	if len(parts) < 2 {
		http.Error(w, "Bad Request", http.StatusBadRequest)
		return
	}
	roomId, err := strconv.Atoi(parts[1])
	if err != nil {
		http.Error(w, "Bad Request", http.StatusBadRequest)
		return
	}

	accountId, err := controllers.CurrentUserIDFromRequest(r)
	if err != nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	room, ok := loadRoom(roomId)
	if !ok {
		http.Error(w, "Not Found", http.StatusNotFound)
		return
	}

	if !isRoomOwner(room, int(accountId)) {
		http.Error(w, "Forbidden", http.StatusForbidden)
		return
	}

	accessibilityStr := r.FormValue("accessibility")
	accessibility, err := strconv.Atoi(accessibilityStr)
	if err != nil {
		http.Error(w, "Bad Request", http.StatusBadRequest)
		return
	}

	db.DB.Model(&models.Room{}).Where("room_id = ?", roomId).Update("accessibility", accessibility)
	room.Accessibility = accessibility
	roomSuccessResponse(w, room)
}

func RoomUpdateRestrictions(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	parts := strings.Split(strings.Trim(r.URL.Path, "/"), "/")
	if len(parts) < 2 {
		http.Error(w, "Bad Request", http.StatusBadRequest)
		return
	}
	roomId, err := strconv.Atoi(parts[1])
	if err != nil {
		http.Error(w, "Bad Request", http.StatusBadRequest)
		return
	}

	accountId, err := controllers.CurrentUserIDFromRequest(r)
	if err != nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	room, ok := loadRoom(roomId)
	if !ok {
		http.Error(w, "Not Found", http.StatusNotFound)
		return
	}

	if !isRoomOwner(room, int(accountId)) {
		http.Error(w, "Forbidden", http.StatusForbidden)
		return
	}

	r.ParseForm()

	updates := map[string]interface{}{}
	fields := []struct {
		formKey string
		column  string
		target  *bool
	}{
		{"supportsScreens", "supports_screens", &room.SupportsScreens},
		{"supportsWalkVR", "supports_walk_vr", &room.SupportsWalkVR},
		{"supportsTeleportVR", "supports_teleport_vr", &room.SupportsTeleportVR},
		{"supportsJuniors", "supports_juniors", &room.SupportsJuniors},
		{"supportsMobile", "supports_mobile", &room.SupportsMobile},
		{"supportsQuest2", "supports_quest_2", &room.SupportsQuest2},
		{"supportsVRLow", "supports_vr_low", &room.SupportsVRLow},
		{"supportsLevelVoting", "supports_level_voting", &room.SupportsLevelVoting},
	}
	for _, f := range fields {
		if _, present := r.Form[f.formKey]; present {
			v := parseRoomBool(r, f.formKey)
			updates[f.column] = v
			*f.target = v
		}
	}

	if len(updates) > 0 {
		db.DB.Model(&models.Room{}).Where("room_id = ?", roomId).Updates(updates)
	}

	roomSuccessResponse(w, room)
}

func RoomUpdateAutoMute(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	parts := strings.Split(strings.Trim(r.URL.Path, "/"), "/")
	if len(parts) < 2 {
		http.Error(w, "Bad Request", http.StatusBadRequest)
		return
	}
	roomId, err := strconv.Atoi(parts[1])
	if err != nil {
		http.Error(w, "Bad Request", http.StatusBadRequest)
		return
	}

	accountId, err := controllers.CurrentUserIDFromRequest(r)
	if err != nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	room, ok := loadRoom(roomId)
	if !ok {
		http.Error(w, "Not Found", http.StatusNotFound)
		return
	}

	if !isRoomOwner(room, int(accountId)) {
		http.Error(w, "Forbidden", http.StatusForbidden)
		return
	}

	disable := parseRoomBool(r, "disable")
	db.DB.Model(&models.Room{}).Where("room_id = ?", roomId).Update("disable_mic_auto_mute", disable)
	room.DisableMicAutoMute = disable
	roomSuccessResponse(w, room)
}

func RoomUpdateComments(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	parts := strings.Split(strings.Trim(r.URL.Path, "/"), "/")
	if len(parts) < 2 {
		http.Error(w, "Bad Request", http.StatusBadRequest)
		return
	}
	roomId, err := strconv.Atoi(parts[1])
	if err != nil {
		http.Error(w, "Bad Request", http.StatusBadRequest)
		return
	}

	accountId, err := controllers.CurrentUserIDFromRequest(r)
	if err != nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	room, ok := loadRoom(roomId)
	if !ok {
		http.Error(w, "Not Found", http.StatusNotFound)
		return
	}

	if !isRoomOwner(room, int(accountId)) {
		http.Error(w, "Forbidden", http.StatusForbidden)
		return
	}

	disable := parseRoomBool(r, "disable")
	db.DB.Model(&models.Room{}).Where("room_id = ?", roomId).Update("disable_room_comments", disable)
	room.DisableRoomComments = disable
	roomSuccessResponse(w, room)
}

func RoomUpdateVoiceChatEncryption(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	parts := strings.Split(strings.Trim(r.URL.Path, "/"), "/")
	if len(parts) < 2 {
		http.Error(w, "Bad Request", http.StatusBadRequest)
		return
	}
	roomId, err := strconv.Atoi(parts[1])
	if err != nil {
		http.Error(w, "Bad Request", http.StatusBadRequest)
		return
	}

	accountId, err := controllers.CurrentUserIDFromRequest(r)
	if err != nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	room, ok := loadRoom(roomId)
	if !ok {
		http.Error(w, "Not Found", http.StatusNotFound)
		return
	}

	if !isRoomOwner(room, int(accountId)) {
		http.Error(w, "Forbidden", http.StatusForbidden)
		return
	}

	encrypt := parseRoomBool(r, "encryptVoiceChat")
	db.DB.Model(&models.Room{}).Where("room_id = ?", roomId).Update("encrypt_voice_chat", encrypt)
	room.EncryptVoiceChat = encrypt
	roomSuccessResponse(w, room)
}

func RoomUpdateCloning(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	parts := strings.Split(strings.Trim(r.URL.Path, "/"), "/")
	if len(parts) < 2 {
		http.Error(w, "Bad Request", http.StatusBadRequest)
		return
	}
	roomId, err := strconv.Atoi(parts[1])
	if err != nil {
		http.Error(w, "Bad Request", http.StatusBadRequest)
		return
	}

	accountId, err := controllers.CurrentUserIDFromRequest(r)
	if err != nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	room, ok := loadRoom(roomId)
	if !ok {
		http.Error(w, "Not Found", http.StatusNotFound)
		return
	}

	if !isRoomOwner(room, int(accountId)) {
		http.Error(w, "Forbidden", http.StatusForbidden)
		return
	}

	cloningStr := r.FormValue("cloningAllowed")
	cloningAllowed := strings.EqualFold(cloningStr, "true") || cloningStr == "1"

	db.DB.Model(&models.Room{}).Where("room_id = ?", roomId).Update("cloning_allowed", cloningAllowed)
	room.CloningAllowed = cloningAllowed
	roomSuccessResponse(w, room)
}
