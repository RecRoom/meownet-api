package inventions

import (
	"encoding/json"
	"log"
	"net/http"
	"strconv"
	"strings"
	"time"

	"meow.net/controllers"
	"meow.net/db"
	"meow.net/models"
	"meow.net/utils"
)

type inventionEnvelope struct {
	Invention        *models.Invention        `json:"Invention"`
	InventionVersion *models.InventionVersion `json:"InventionVersion"`
	Status           int                      `json:"Status"`
}

func Details(w http.ResponseWriter, r *http.Request) {
	id, ok := parseInt64Param(r, "inventionId")
	if !ok {
		writeError(w, http.StatusBadRequest, "inventionId required")
		return
	}
	if _, ok := loadInvention(id); !ok {
		writeError(w, http.StatusNotFound, "invention not found")
		return
	}
	var tags []models.InventionTag
	db.DB.Where("invention_id = ?", id).Find(&tags)
	if tags == nil {
		tags = []models.InventionTag{}
	}
	writeJSON(w, http.StatusOK, map[string]interface{}{
		"Tags": tags,
	})
}

func Version(w http.ResponseWriter, r *http.Request) {
	id, ok := parseInt64Param(r, "inventionId")
	if !ok {
		writeError(w, http.StatusBadRequest, "inventionId required")
		return
	}
	version, ok := parseIntParam(r, "version")
	if !ok {
		writeError(w, http.StatusBadRequest, "version required")
		return
	}
	v, ok := loadVersion(id, version)
	if !ok {
		writeError(w, http.StatusNotFound, "version not found")
		return
	}
	writeJSON(w, http.StatusOK, v)
}

func PersonalDetails(w http.ResponseWriter, r *http.Request) {
	parts := strings.Split(strings.TrimRight(r.URL.Path, "/"), "/")
	idStr := parts[len(parts)-1]
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid invention id")
		return
	}

	accountID, _ := controllers.AccountIDFromRequest(r)
	cheering := false
	if accountID != 0 {
		var cheer models.InventionCheer
		if err := db.DB.Where("invention_id = ? AND account_id = ?", id, accountID).First(&cheer).Error; err == nil {
			cheering = true
		}
	}
	writeJSON(w, http.StatusOK, map[string]interface{}{
		"IsCheering": cheering,
	})
}

func Search(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query()
	value := strings.TrimSpace(q.Get("value"))
	take := 100
	if t, err := strconv.Atoi(q.Get("take")); err == nil && t > 0 {
		take = t
	}
	if take > 1000 {
		take = 1000
	}
	skip := 0
	if s, err := strconv.Atoi(q.Get("skip")); err == nil && s > 0 {
		skip = s
	}

	dq := db.DB.Model(&models.Invention{}).
		Where("is_published = ? AND hide_from_player = ?", true, false)

	if value != "" {
		if strings.HasPrefix(value, "#") {
			tag := strings.ToLower(strings.TrimPrefix(value, "#"))
			dq = dq.Where("invention_id IN (?)",
				db.DB.Model(&models.InventionTag{}).
					Select("invention_id").
					Where("LOWER(tag) = ?", tag))
		} else {
			like := "%" + utils.EscapeLike(strings.ToLower(value)) + "%"
			dq = dq.Where(`LOWER(name) LIKE ? ESCAPE '\' OR LOWER(description) LIKE ? ESCAPE '\'`, like, like)
		}
	}

	var list []models.Invention
	if err := dq.Order("modified_at desc").Offset(skip).Limit(take).Find(&list).Error; err != nil {
		log.Printf("[INVENTIONS] search error: %v", err)
		writeJSON(w, http.StatusOK, []models.Invention{})
		return
	}
	if list == nil {
		list = []models.Invention{}
	}
	writeJSON(w, http.StatusOK, list)
}

func TopToday(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query()
	take := 100
	if t, err := strconv.Atoi(q.Get("take")); err == nil && t > 0 {
		take = t
	}
	if take > 1000 {
		take = 1000
	}
	skip := 0
	if s, err := strconv.Atoi(q.Get("skip")); err == nil && s > 0 {
		skip = s
	}

	var list []models.Invention
	if err := db.DB.
		Where("is_published = ? AND hide_from_player = ?", true, false).
		Order("cheer_count desc, num_downloads desc, modified_at desc").
		Offset(skip).Limit(take).
		Find(&list).Error; err != nil {
		log.Printf("[INVENTIONS] toptoday error: %v", err)
		writeJSON(w, http.StatusOK, []models.Invention{})
		return
	}
	if list == nil {
		list = []models.Invention{}
	}
	writeJSON(w, http.StatusOK, list)
}

func Mine(w http.ResponseWriter, r *http.Request) {
	accountID, ok := controllers.AccountIDFromRequest(r)
	if !ok {
		writeError(w, http.StatusUnauthorized, "unauthorized")
		return
	}
	var list []models.Invention
	db.DB.Where("creator_player_id = ? OR invention_id IN (?)",
		accountID,
		db.DB.Model(&models.InventionOwnership{}).
			Select("invention_id").
			Where("account_id = ?", accountID)).
		Order("modified_at desc").
		Find(&list)
	if list == nil {
		list = []models.Invention{}
	}
	writeJSON(w, http.StatusOK, list)
}

func FromCreators(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query()
	creatorID, err := strconv.Atoi(q.Get("id"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "id required")
		return
	}
	take := 100
	if t, err := strconv.Atoi(q.Get("take")); err == nil && t > 0 {
		take = t
	}
	if take > 1000 {
		take = 1000
	}
	skip := 0
	if s, err := strconv.Atoi(q.Get("skip")); err == nil && s > 0 {
		skip = s
	}

	accountID, _ := controllers.AccountIDFromRequest(r)

	dq := db.DB.Where("creator_player_id = ?", creatorID)
	if uint(creatorID) != accountID {
		dq = dq.Where("is_published = ? AND hide_from_player = ?", true, false)
	}

	var list []models.Invention
	dq.Order("modified_at desc").Offset(skip).Limit(take).Find(&list)
	if list == nil {
		list = []models.Invention{}
	}
	writeJSON(w, http.StatusOK, list)
}

func Batch(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query()
	rawIds := q["id"]
	if len(rawIds) == 0 {
		writeJSON(w, http.StatusOK, []models.Invention{})
		return
	}
	ids := make([]int64, 0, len(rawIds))
	for _, raw := range rawIds {
		for _, part := range strings.Split(raw, ",") {
			if v, err := strconv.ParseInt(strings.TrimSpace(part), 10, 64); err == nil {
				ids = append(ids, v)
			}
		}
	}
	if len(ids) == 0 {
		writeJSON(w, http.StatusOK, []models.Invention{})
		return
	}
	accountID, _ := controllers.AccountIDFromRequest(r)

	var list []models.Invention
	db.DB.Where("invention_id IN ?", ids).Find(&list)
	if list == nil {
		list = []models.Invention{}
	}
	filtered := make([]models.Invention, 0, len(list))
	for _, inv := range list {
		if inv.IsPublished || isCreator(accountID, inv) {
			filtered = append(filtered, inv)
		}
	}
	writeJSON(w, http.StatusOK, filtered)
}

type saveRequest struct {
	AICost                int     `json:"aiCost"`
	ChipsCost             int     `json:"chipsCost"`
	CloudVariablesCost    int     `json:"cloudVariablesCost"`
	CreationRoomId        int64   `json:"creationRoomId"`
	CreatorAccountRole    int     `json:"creatorAccountRole"`
	Description           string  `json:"description"`
	ImageName             string  `json:"imageName"`
	InstantiationCost     int     `json:"instantiationCost"`
	InventionDataFilename string  `json:"inventionDataFilename"`
	LightsCost            int     `json:"lightsCost"`
	Name                  string  `json:"name"`
	ReferencedInventions  []int64 `json:"referencedInventions"`
}

func Save(w http.ResponseWriter, r *http.Request) {
	accountID, ok := controllers.AccountIDFromRequest(r)
	if !ok {
		writeError(w, http.StatusUnauthorized, "unauthorized")
		return
	}
	var req saveRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid body")
		return
	}
	if strings.TrimSpace(req.InventionDataFilename) == "" {
		writeError(w, http.StatusBadRequest, "inventionDataFilename required")
		return
	}
	if strings.TrimSpace(req.Name) == "" {
		req.Name = "Untitled"
	}
	if strings.TrimSpace(req.Description) == "" {
		req.Description = "No description yet"
	}

	now := time.Now().UTC()
	inv := models.Invention{
		InventionId:          uniqueInventionId(),
		Name:                 req.Name,
		Description:          req.Description,
		ImageName:            req.ImageName,
		CreatorPlayerId:      int(accountID),
		CreatorPermission:    int(models.InventionPermissionUnlimited),
		GeneralPermission:    int(models.InventionPermissionUnlimited),
		AllowTrial:           true,
		IsPublished:          false,
		CurrentVersionNumber: 1,
		ReplicationId:        randomReplicationId(),
		CreatedAt:            now,
		ModifiedAt:           now,
	}
	if err := db.DB.Create(&inv).Error; err != nil {
		log.Printf("[INVENTIONS] save create error: %v", err)
		writeError(w, http.StatusInternalServerError, "create failed")
		return
	}

	version := models.InventionVersion{
		InventionId:        inv.InventionId,
		VersionNumber:      1,
		BlobName:           req.InventionDataFilename,
		ChipsCost:          req.ChipsCost,
		CloudVariablesCost: req.CloudVariablesCost,
		InstantiationCost:  req.InstantiationCost,
		LightsCost:         req.LightsCost,
		ReplicationId:      randomReplicationId(),
	}
	if err := db.DB.Create(&version).Error; err != nil {
		log.Printf("[INVENTIONS] save version error: %v", err)
		db.DB.Delete(&inv)
		writeError(w, http.StatusInternalServerError, "create failed")
		return
	}

	writeJSON(w, http.StatusOK, inventionEnvelope{
		Invention:        &inv,
		InventionVersion: &version,
		Status:           0,
	})
}

type setTagsRequest struct {
	AutoTags    []string `json:"AutoTags"`
	CustomTags  []string `json:"CustomTags"`
	InventionId int64    `json:"InventionId"`
}

func SetTags(w http.ResponseWriter, r *http.Request) {
	accountID, ok := controllers.AccountIDFromRequest(r)
	if !ok {
		writeError(w, http.StatusUnauthorized, "unauthorized")
		return
	}
	var req setTagsRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid body")
		return
	}
	inv, ok := loadInvention(req.InventionId)
	if !ok {
		writeError(w, http.StatusNotFound, "invention not found")
		return
	}
	if !isCreator(accountID, inv) {
		writeError(w, http.StatusForbidden, "forbidden")
		return
	}

	tx := db.DB.Begin()
	if err := tx.Where("invention_id = ?", req.InventionId).Delete(&models.InventionTag{}).Error; err != nil {
		tx.Rollback()
		writeError(w, http.StatusInternalServerError, "tag update failed")
		return
	}
	combined := make([]string, 0, len(req.AutoTags)+len(req.CustomTags))
	for _, t := range req.AutoTags {
		t = strings.TrimSpace(t)
		if t == "" {
			continue
		}
		if err := tx.Create(&models.InventionTag{InventionId: req.InventionId, Tag: t, Type: int(models.InventionTagAuto)}).Error; err != nil {
			tx.Rollback()
			writeError(w, http.StatusInternalServerError, "tag update failed")
			return
		}
		combined = append(combined, t)
	}
	for _, t := range req.CustomTags {
		t = strings.TrimSpace(t)
		if t == "" {
			continue
		}
		if err := tx.Create(&models.InventionTag{InventionId: req.InventionId, Tag: t, Type: int(models.InventionTagCustom)}).Error; err != nil {
			tx.Rollback()
			writeError(w, http.StatusInternalServerError, "tag update failed")
			return
		}
		combined = append(combined, t)
	}
	if err := tx.Commit().Error; err != nil {
		writeError(w, http.StatusInternalServerError, "tag update failed")
		return
	}

	writeJSON(w, http.StatusOK, map[string]interface{}{
		"Result": 0,
		"Tags":   combined,
	})
}

func Update(w http.ResponseWriter, r *http.Request) {
	accountID, ok := controllers.AccountIDFromRequest(r)
	if !ok {
		writeError(w, http.StatusUnauthorized, "unauthorized")
		return
	}
	q := r.URL.Query()
	id, err := strconv.ParseInt(q.Get("inventionId"), 10, 64)
	if err != nil {
		writeError(w, http.StatusBadRequest, "inventionId required")
		return
	}
	inv, ok := loadInvention(id)
	if !ok {
		writeError(w, http.StatusNotFound, "invention not found")
		return
	}
	if !isCreator(accountID, inv) {
		writeError(w, http.StatusForbidden, "forbidden")
		return
	}

	updates := map[string]interface{}{}
	if v := q.Get("name"); v != "" {
		name := strings.TrimSpace(v)
		if name != "" {
			updates["name"] = name
		}
	}
	if q.Has("description") {
		updates["description"] = q.Get("description")
	}
	if v := q.Get("permission"); v != "" {
		if iv, ok := parsePermissionLevel(v); ok {
			updates["general_permission"] = iv
		}
	}
	if v := q.Get("allowTrial"); v != "" {
		updates["allow_trial"] = strings.EqualFold(v, "true") || v == "1"
	}
	if v := q.Get("imageName"); v != "" {
		updates["image_name"] = v
	}

	if len(updates) > 0 {
		updates["modified_at"] = time.Now().UTC()
		if err := db.DB.Model(&models.Invention{}).Where("invention_id = ?", id).Updates(updates).Error; err != nil {
			log.Printf("[INVENTIONS] update error: %v", err)
			writeError(w, http.StatusInternalServerError, "update failed")
			return
		}
		db.DB.First(&inv, id)
	}

	version, _ := loadCurrentVersion(inv)
	writeJSON(w, http.StatusOK, inventionEnvelope{
		Invention:        &inv,
		InventionVersion: &version,
		Status:           0,
	})
}

func FullLineageOwner(w http.ResponseWriter, r *http.Request) {
	accountID, ok := controllers.AccountIDFromRequest(r)
	if !ok {
		writeJSON(w, http.StatusOK, false)
		return
	}
	id, err := strconv.ParseInt(r.URL.Query().Get("id"), 10, 64)
	if err != nil {
		writeError(w, http.StatusBadRequest, "id required")
		return
	}
	writeJSON(w, http.StatusOK, ownsInvention(accountID, id))
}

type addVersionRequest struct {
	AICost                int     `json:"aiCost"`
	ChipsCost             int     `json:"chipsCost"`
	CloudVariablesCost    int     `json:"cloudVariablesCost"`
	CreationRoomId        int64   `json:"creationRoomId"`
	InstantiationCost     int     `json:"instantiationCost"`
	InventionDataFilename string  `json:"inventionDataFilename"`
	InventionId           int64   `json:"inventionId"`
	LightsCost            int     `json:"lightsCost"`
	ReferencedInventions  []int64 `json:"referencedInventions"`
}

func AddVersion(w http.ResponseWriter, r *http.Request) {
	accountID, ok := controllers.AccountIDFromRequest(r)
	if !ok {
		writeError(w, http.StatusUnauthorized, "unauthorized")
		return
	}
	var req addVersionRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid body")
		return
	}
	if strings.TrimSpace(req.InventionDataFilename) == "" {
		writeError(w, http.StatusBadRequest, "inventionDataFilename required")
		return
	}
	inv, ok := loadInvention(req.InventionId)
	if !ok {
		writeError(w, http.StatusNotFound, "invention not found")
		return
	}
	if !isCreator(accountID, inv) {
		writeError(w, http.StatusForbidden, "forbidden")
		return
	}

	newVersionNumber := inv.CurrentVersionNumber + 1
	version := models.InventionVersion{
		InventionId:        inv.InventionId,
		VersionNumber:      newVersionNumber,
		BlobName:           req.InventionDataFilename,
		ChipsCost:          req.ChipsCost,
		CloudVariablesCost: req.CloudVariablesCost,
		InstantiationCost:  req.InstantiationCost,
		LightsCost:         req.LightsCost,
		ReplicationId:      randomReplicationId(),
	}
	if err := db.DB.Create(&version).Error; err != nil {
		log.Printf("[INVENTIONS] addversion create error: %v", err)
		writeError(w, http.StatusInternalServerError, "create failed")
		return
	}
	now := time.Now().UTC()
	if err := db.DB.Model(&models.Invention{}).Where("invention_id = ?", inv.InventionId).Updates(map[string]interface{}{
		"current_version_number": newVersionNumber,
		"modified_at":            now,
	}).Error; err != nil {
		log.Printf("[INVENTIONS] addversion update error: %v", err)
	}
	inv.CurrentVersionNumber = newVersionNumber
	inv.ModifiedAt = now

	writeJSON(w, http.StatusOK, inventionEnvelope{
		Invention:        &inv,
		InventionVersion: &version,
		Status:           0,
	})
}

func Delete(w http.ResponseWriter, r *http.Request) {
	accountID, ok := controllers.AccountIDFromRequest(r)
	if !ok {
		writeError(w, http.StatusUnauthorized, "unauthorized")
		return
	}
	id, err := strconv.ParseInt(r.URL.Query().Get("inventionId"), 10, 64)
	if err != nil {
		writeError(w, http.StatusBadRequest, "inventionId required")
		return
	}
	inv, ok := loadInvention(id)
	if !ok {
		writeError(w, http.StatusNotFound, "invention not found")
		return
	}
	if !isCreator(accountID, inv) {
		writeError(w, http.StatusForbidden, "forbidden")
		return
	}

	db.DB.Where("invention_id = ?", id).Delete(&models.InventionVersion{})
	db.DB.Where("invention_id = ?", id).Delete(&models.InventionTag{})
	db.DB.Where("invention_id = ?", id).Delete(&models.InventionOwnership{})
	db.DB.Where("invention_id = ?", id).Delete(&models.InventionCheer{})
	db.DB.Delete(&models.Invention{}, id)

	writeJSON(w, http.StatusOK, inventionEnvelope{
		Invention:        nil,
		InventionVersion: nil,
		Status:           0,
	})
}

type updatePriceRequest struct {
	InventionId int64 `json:"InventionId"`
	Price       int   `json:"Price"`
}

func UpdatePrice(w http.ResponseWriter, r *http.Request) {
	accountID, ok := controllers.AccountIDFromRequest(r)
	if !ok {
		writeError(w, http.StatusUnauthorized, "unauthorized")
		return
	}
	var req updatePriceRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid body")
		return
	}
	if req.Price < 0 {
		writeError(w, http.StatusBadRequest, "price must be >= 0")
		return
	}
	inv, ok := loadInvention(req.InventionId)
	if !ok {
		writeError(w, http.StatusNotFound, "invention not found")
		return
	}
	if !isCreator(accountID, inv) {
		writeError(w, http.StatusForbidden, "forbidden")
		return
	}

	now := time.Now().UTC()
	if err := db.DB.Model(&models.Invention{}).Where("invention_id = ?", req.InventionId).Updates(map[string]interface{}{
		"price":       req.Price,
		"modified_at": now,
	}).Error; err != nil {
		log.Printf("[INVENTIONS] updateprice error: %v", err)
		writeError(w, http.StatusInternalServerError, "update failed")
		return
	}
	db.DB.First(&inv, req.InventionId)
	version, _ := loadCurrentVersion(inv)

	writeJSON(w, http.StatusOK, inventionEnvelope{
		Invention:        &inv,
		InventionVersion: &version,
		Status:           0,
	})
}

func Publish(w http.ResponseWriter, r *http.Request) {
	accountID, ok := controllers.AccountIDFromRequest(r)
	if !ok {
		writeError(w, http.StatusUnauthorized, "unauthorized")
		return
	}
	q := r.URL.Query()
	id, err := strconv.ParseInt(q.Get("inventionId"), 10, 64)
	if err != nil {
		writeError(w, http.StatusBadRequest, "inventionId required")
		return
	}
	inv, ok := loadInvention(id)
	if !ok {
		writeError(w, http.StatusNotFound, "invention not found")
		return
	}
	if !isCreator(accountID, inv) {
		writeError(w, http.StatusForbidden, "forbidden")
		return
	}

	permission := int(models.InventionPermissionUseOnly)
	if v := q.Get("permissionLevel"); v != "" {
		if iv, ok := parsePermissionLevel(v); ok {
			permission = iv
		}
	}
	price := 0
	if v := q.Get("price"); v != "" {
		if iv, err := strconv.Atoi(v); err == nil && iv >= 0 {
			price = iv
		}
	}

	now := time.Now().UTC()
	updates := map[string]interface{}{
		"is_published":       true,
		"general_permission": permission,
		"price":              price,
		"modified_at":        now,
	}
	if err := db.DB.Model(&models.Invention{}).Where("invention_id = ?", id).Updates(updates).Error; err != nil {
		log.Printf("[INVENTIONS] publish update error: %v", err)
		writeError(w, http.StatusInternalServerError, "publish failed")
		return
	}
	db.DB.First(&inv, id)
	version, _ := loadCurrentVersion(inv)

	writeJSON(w, http.StatusOK, inventionEnvelope{
		Invention:        &inv,
		InventionVersion: &version,
		Status:           0,
	})
}
