package player

import (
	"bytes"
	"encoding/json"
	"log"
	"net/http"
	"os"
	"strconv"
	"sync"

	"meow.net/db"
	"meow.net/models"
	"meow.net/utils"
)

func accountIDFromRequest(r *http.Request) (uint, bool) {
	token := utils.GetBearerToken(r)
	if token == "" {
		return 0, false
	}
	sub, err := utils.ParseSubFromJWT(token)
	if err != nil || sub == "" {
		return 0, false
	}
	id, err := strconv.ParseUint(sub, 10, 64)
	if err != nil {
		return 0, false
	}
	return uint(id), true
}

func writeAvatar(w http.ResponseWriter, a models.Avatar) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"FaceFeatures":     a.FaceFeatures,
		"HairColor":        a.HairColor,
		"OutfitSelections": a.OutfitSelections,
		"SkinColor":        a.SkinColor,
	})
}

func Avatar(w http.ResponseWriter, r *http.Request) {
	log.Printf("[AVATAR] v2 get")

	accountID, ok := accountIDFromRequest(r)
	if !ok {
		writeAvatar(w, models.Avatar{})
		return
	}

	var avatar models.Avatar
	if err := db.DB.Where("account_id = ?", accountID).First(&avatar).Error; err != nil {
		writeAvatar(w, models.Avatar{})
		return
	}
	writeAvatar(w, avatar)
}

var (
	avatarItemsCache     []byte
	avatarItemsCacheOnce sync.Once
	avatarItemsHasItems  bool
	defaultUnlockedDescs map[string]struct{}
)

type avatarItemOut struct {
	AvatarItemDesc string `json:"AvatarItemDesc"`
	AvatarItemType int    `json:"AvatarItemType"`
	FriendlyName   string `json:"FriendlyName"`
	Rarity         int    `json:"Rarity"`
	ToolTip        string `json:"ToolTip"`
}

func BuildAvatarItemsCache() {
	avatarItemsCacheOnce.Do(func() {
		raw, err := os.ReadFile("data/jsons/defaultUnlocked.json")
		if err != nil {
			log.Printf("[AVATAR] items cache: read error: %v", err)
			avatarItemsCache = []byte("[]")
			defaultUnlockedDescs = map[string]struct{}{}
			return
		}

		var out []avatarItemOut
		if err := json.Unmarshal(raw, &out); err != nil {
			log.Printf("[AVATAR] items cache: JSON decode error: %v", err)
			avatarItemsCache = []byte("[]")
			defaultUnlockedDescs = map[string]struct{}{}
			return
		}

		defaultUnlockedDescs = make(map[string]struct{}, len(out))
		for _, it := range out {
			defaultUnlockedDescs[it.AvatarItemDesc] = struct{}{}
		}

		avatarItemsCache = raw
		avatarItemsHasItems = len(out) > 0
		log.Printf("[AVATAR] items cache: %d items (%d bytes)", len(out), len(avatarItemsCache))
	})
}

func AvatarItems(w http.ResponseWriter, r *http.Request) {
	log.Printf("[AVATAR] items")
	BuildAvatarItemsCache()
	w.Header().Set("Content-Type", "application/json")

	accountID, ok := accountIDFromRequest(r)
	if !ok {
		w.Write(avatarItemsCache)
		return
	}

	extras, err := queryUserAvatarExtras(accountID)
	if err != nil || len(extras) == 0 {
		w.Write(avatarItemsCache)
		return
	}

	trimmed := bytes.TrimRight(avatarItemsCache, " \n\r\t")
	if len(trimmed) == 0 || trimmed[len(trimmed)-1] != ']' {
		_ = json.NewEncoder(w).Encode(extras)
		return
	}
	trimmed = trimmed[:len(trimmed)-1]

	w.Write(trimmed)
	needsComma := avatarItemsHasItems
	for _, e := range extras {
		raw, err := json.Marshal(e)
		if err != nil {
			continue
		}
		if needsComma {
			w.Write([]byte{','})
		}
		w.Write(raw)
		needsComma = true
	}
	w.Write([]byte{']'})
}

func queryUserAvatarExtras(accountID uint) ([]avatarItemOut, error) {
	var rows []models.AvatarItem
	err := db.DB.
		Table("user_avatar_items AS uai").
		Select("ai.avatar_item_desc, ai.avatar_item_type, ai.friendly_name, ai.tool_tip, ai.rarity").
		Joins("JOIN avatar_items AS ai ON ai.avatar_item_desc = uai.avatar_item_desc").
		Where("uai.account_id = ?", accountID).
		Scan(&rows).Error
	if err != nil {
		log.Printf("[AVATAR] items: extras query error: %v", err)
		return nil, err
	}
	out := make([]avatarItemOut, 0, len(rows))
	for _, it := range rows {
		if _, dup := defaultUnlockedDescs[it.AvatarItemDesc]; dup {
			continue
		}
		out = append(out, avatarItemOut{
			AvatarItemDesc: it.AvatarItemDesc,
			AvatarItemType: it.AvatarItemType,
			FriendlyName:   it.FriendlyName,
			Rarity:         it.Rarity,
			ToolTip:        it.ToolTip,
		})
	}
	return out, nil
}

func DefaultUnlocked(w http.ResponseWriter, r *http.Request) {
	log.Printf("[AVATAR] defaultunlocked")
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	http.ServeFile(w, r, "data/jsons/defaultUnlocked.json")
}

func AvatarSet(w http.ResponseWriter, r *http.Request) {
	log.Printf("[AVATAR] set")

	accountID, ok := accountIDFromRequest(r)
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	var body struct {
		FaceFeatures     string `json:"FaceFeatures"`
		HairColor        string `json:"HairColor"`
		OutfitSelections string `json:"OutfitSelections"`
		SkinColor        string `json:"SkinColor"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		http.Error(w, "invalid body", http.StatusBadRequest)
		return
	}

	avatar := models.Avatar{
		AccountID:        accountID,
		FaceFeatures:     body.FaceFeatures,
		HairColor:        body.HairColor,
		OutfitSelections: body.OutfitSelections,
		SkinColor:        body.SkinColor,
	}
	if err := db.DB.Save(&avatar).Error; err != nil {
		log.Printf("[AVATAR] save error: %v", err)
		http.Error(w, "save failed", http.StatusInternalServerError)
		return
	}

	writeAvatar(w, avatar)
}

func AvatarSaved(w http.ResponseWriter, r *http.Request) {
	log.Printf("[AVATAR] saved get")

	accountID, ok := accountIDFromRequest(r)
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	var outfits []models.SavedOutfit
	if err := db.DB.Where("account_id = ?", accountID).Find(&outfits).Error; err != nil {
		log.Printf("[AVATAR] saved get error: %v", err)
		http.Error(w, "query failed", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	if outfits == nil {
		outfits = []models.SavedOutfit{}
	}
	json.NewEncoder(w).Encode(outfits)
}

func AvatarSavedSet(w http.ResponseWriter, r *http.Request) {
	log.Printf("[AVATAR] saved set")

	accountID, ok := accountIDFromRequest(r)
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	var body struct {
		Slot             string `json:"Slot"`
		PreviewImageName string `json:"PreviewImageName"`
		OutfitSelections string `json:"OutfitSelections"`
		FaceFeatures     string `json:"FaceFeatures"`
		SkinColor        string `json:"SkinColor"`
		HairColor        string `json:"HairColor"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		http.Error(w, "invalid body", http.StatusBadRequest)
		return
	}
	if body.Slot == "" {
		http.Error(w, "Slot is required", http.StatusBadRequest)
		return
	}

	outfit := models.SavedOutfit{
		AccountID:        accountID,
		Slot:             body.Slot,
		PreviewImageName: body.PreviewImageName,
		OutfitSelections: body.OutfitSelections,
		FaceFeatures:     body.FaceFeatures,
		SkinColor:        body.SkinColor,
		HairColor:        body.HairColor,
	}
	if err := db.DB.Save(&outfit).Error; err != nil {
		log.Printf("[AVATAR] saved set error: %v", err)
		http.Error(w, "save failed", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(outfit)
}
