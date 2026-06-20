package controllers

import (
	"crypto/rand"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"mime"
	"mime/multipart"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"gorm.io/gorm"
	"meow.net/controllers/hub"
	"meow.net/db"
	"meow.net/discord"
	"meow.net/models"
	"meow.net/utils"
)

func randomHex(n int) string {
	b := make([]byte, n)
	if _, err := rand.Read(b); err != nil {
		return fmt.Sprintf("%d", time.Now().UTC().UnixNano())
	}
	return fmt.Sprintf("%x", b)
}

func makeStoredImageName(accountID uint, originalName, contentType string) string {
	_ = originalName
	_ = contentType
	ts := time.Now().UTC().Format("20060102T150405.000")
	return fmt.Sprintf("img_%d_%s_%s.png", accountID, ts, randomHex(4))
}

func Images(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	fmt.Fprint(w, "[]")
}

func ImagesNamed(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	http.ServeFile(w, r, "data/jsons/namedimages.json")
}

func CurrentChallenge(w http.ResponseWriter, r *http.Request) {
	data, err := os.ReadFile("data/jsons/weeklychallenge.json")
	if err != nil {
		log.Printf("[CHALLENGE] failed to read file: %v", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	data = trimBOM(data)
	w.Header().Set("Content-Type", "application/json")
	w.Write(data)
}

func trimBOM(data []byte) []byte {
	bom := []byte{0xEF, 0xBB, 0xBF}
	if len(data) >= 3 && data[0] == bom[0] && data[1] == bom[1] && data[2] == bom[2] {
		return data[3:]
	}
	return data
}

type imgMetaPayload struct {
	PlayerIDs      []int `json:"playerIds"`
	SavedImageType int   `json:"savedImageType"`
	RoomID         int   `json:"roomId"`
	PlayerEventID  int   `json:"playerEventId"`
	Accessibility  int   `json:"accessibility"`
}

// POST /api/images/v4/uploadsaved
func ImagesUploadSaved(w http.ResponseWriter, r *http.Request) {
	log.Printf("[IMAGES] v4/uploadsaved ct=%q len=%d", r.Header.Get("Content-Type"), r.ContentLength)
	w.Header().Set("Content-Type", "application/json")

	accountID, ok := AccountIDFromRequest(r)
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	// must have an active session to upload
	if !hub.HubIsOnline(int(accountID)) {
		http.Error(w, "Upload unavailable", http.StatusForbidden)
		return
	}

	// to stop spam/junk uploads
	if !utils.AccountActionAllowBurst("image_upload", accountID, 3*time.Second, 10) {
		http.Error(w, "too many uploads", http.StatusTooManyRequests)
		return
	}

	var meta imgMetaPayload

	ct := r.Header.Get("Content-Type")
	if !strings.HasPrefix(ct, "multipart/form-data") {
		// Fallback
		body, _ := io.ReadAll(r.Body)
		var m struct {
			ImageName string `json:"ImageName"`
		}
		_ = json.Unmarshal(body, &m)
		if m.ImageName == "" {
			http.Error(w, "missing ImageName", http.StatusBadRequest)
			return
		}
		finalizeUploadSaved(w, accountID, m.ImageName, meta)
		return
	}

	_, params, err := mime.ParseMediaType(ct)
	if err != nil {
		http.Error(w, "invalid multipart", http.StatusBadRequest)
		return
	}
	mr := multipart.NewReader(r.Body, params["boundary"])

	var imageName string
	var storedName string
	var binaryContentType string

	for {
		part, err := mr.NextPart()
		if err == io.EOF {
			break
		}
		if err != nil {
			log.Printf("[IMAGES] multipart read error: %v", err)
			break
		}

		partName := part.FormName()
		partFile := part.FileName()
		partCT := part.Header.Get("Content-Type")
		log.Printf("[IMAGES] part name=%q filename=%q content-type=%q", partName, partFile, partCT)

		isBinary := partFile != "" || strings.HasPrefix(partCT, "image/") || partCT == "application/octet-stream"

		if isBinary {
			tmpName := makeStoredImageName(accountID, partFile, partCT)
			binaryContentType = partCT
			if err := saveFileBytes(tmpName, part, -1, partCT); err != nil {
				log.Printf("[IMAGES] save error: %v", err)
			} else {
				imageName = tmpName
				storedName = tmpName
			}
			part.Close()
			continue
		}

		buf, _ := io.ReadAll(io.LimitReader(part, 64*1024))
		part.Close()
		text := strings.TrimSpace(string(buf))
		log.Printf("[IMAGES]   text part value=%q", text)

		switch strings.ToLower(partName) {
		case "imagename", "name":
			if imageName == "" {
				imageName = text
			}
		case "metadata", "meta", "json":
			var m struct {
				ImageName string `json:"ImageName"`
			}
			if err := json.Unmarshal(buf, &m); err == nil && m.ImageName != "" && imageName == "" {
				imageName = m.ImageName
			}
		case "imgmeta":
			if err := json.Unmarshal(buf, &meta); err != nil {
				log.Printf("[IMAGES] imgmeta unmarshal error: %v", err)
			}
		}
	}

	if imageName == "" {
		http.Error(w, "missing ImageName", http.StatusBadRequest)
		return
	}
	_ = binaryContentType

	log.Printf("[IMAGES] meta: roomID=%d savedImageType=%d playerEventID=%d accessibility=%d playerIDs=%v",
		meta.RoomID, meta.SavedImageType, meta.PlayerEventID, meta.Accessibility, meta.PlayerIDs)

	if !validateUploadMeta(w, accountID, meta) {
		// drop the file we just stored so rejected uploads leave no junk behind
		if storedName != "" {
			if err := deleteStoredImage(storedName); err != nil {
				log.Printf("[IMAGES] failed to delete rejected upload %q: %v", storedName, err)
			}
		}
		return
	}

	finalizeUploadSaved(w, accountID, imageName, meta)
}

const maxTaggedPlayers = 20

func validateUploadMeta(w http.ResponseWriter, accountID uint, meta imgMetaPayload) bool {
	if meta.SavedImageType < int(models.SavedImageNone) || meta.SavedImageType > int(models.SavedImageRoomLoadScreen) {
		http.Error(w, "invalid savedImageType", http.StatusBadRequest)
		return false
	}

	if len(meta.PlayerIDs) > maxTaggedPlayers {
		http.Error(w, "Upload unavailable", http.StatusBadRequest)
		return false
	}

	if meta.SavedImageType == int(models.SavedImageShareCamera) && meta.RoomID <= 0 {
		http.Error(w, "Upload unavailable", http.StatusBadRequest)
		return false
	}

	if meta.RoomID > 0 {
		roomID, ok := hub.PlayerCurrentRoomID(int(accountID))
		if !ok || roomID != int64(meta.RoomID) {
			http.Error(w, "Upload unavailable", http.StatusForbidden)
			return false
		}
	}

	return true
}

func finalizeUploadSaved(w http.ResponseWriter, accountID uint, imageName string, meta imgMetaPayload) {
	img := models.UserImage{
		AccountID: accountID,
		ImageName: imageName,
		IsSaved:   true,
		CreatedAt: time.Now().UTC(),
	}
	db.DB.
		Where(models.UserImage{ImageName: imageName}).
		Assign(models.UserImage{AccountID: accountID, IsSaved: true}).
		FirstOrCreate(&img)

	photo := models.UploadedPhoto{
		AccountID:      accountID,
		ImageName:      imageName,
		PlayerIDs:      meta.PlayerIDs,
		SavedImageType: models.SavedImageType(meta.SavedImageType),
		RoomID:         meta.RoomID,
		PlayerEventID:  meta.PlayerEventID,
		Accessibility:  meta.Accessibility,
		CreatedAt:      time.Now().UTC(),
	}
	if err := db.DB.Create(&photo).Error; err != nil {
		log.Printf("[IMAGES] uploaded_photos insert error: %v", err)
	}

	notifyDiscordImageUploaded(accountID, imageName, meta)

	_ = json.NewEncoder(w).Encode(map[string]any{
		"ImageName": imageName,
		"Success":   true,
	})
}

func notifyDiscordImageUploaded(uploaderID uint, imageName string, meta imgMetaPayload) {
	if models.SavedImageType(meta.SavedImageType) != models.SavedImageShareCamera {
		return
	}
	info := discord.ImageUploadInfo{
		ImageURL: resolveImageURL(imageName),
	}

	var uploader models.Account
	if err := db.DB.Select("username", "display_name").
		Where("account_id = ?", uploaderID).First(&uploader).Error; err == nil {
		info.Uploader = firstNonEmpty(uploader.Username, uploader.DisplayName)
	}

	if len(meta.PlayerIDs) > 0 {
		var accts []models.Account
		if err := db.DB.Select("account_id", "username", "display_name").
			Where("account_id IN ?", meta.PlayerIDs).Find(&accts).Error; err == nil {
			for _, a := range accts {
				if a.AccountID == uploaderID {
					continue
				}
				info.PlayerNames = append(info.PlayerNames, firstNonEmpty(a.Username, a.DisplayName))
			}
		}
	}

	if meta.RoomID > 0 {
		var room models.Room
		if err := db.DB.Select("name", "creator_account_id").
			Where("room_id = ?", meta.RoomID).First(&room).Error; err == nil {
			info.RoomName = room.Name
			if room.CreatorAccountId > 0 {
				var owner models.Account
				if err := db.DB.Select("username", "display_name").
					Where("account_id = ?", room.CreatorAccountId).First(&owner).Error; err == nil {
					info.RoomOwner = firstNonEmpty(owner.Username, owner.DisplayName)
				}
			}
		}
	}

	discord.SendImageUploaded(info)
}

func resolveImageURL(imageName string) string {
	base := strings.TrimRight(os.Getenv("CDN_HOST"), "/")
	if base == "" {
		return ""
	}
	if !strings.HasPrefix(base, "http://") && !strings.HasPrefix(base, "https://") {
		base = "https://" + base
	}
	return base + "/" + imageName
}

func firstNonEmpty(values ...string) string {
	for _, v := range values {
		if v != "" {
			return v
		}
	}
	return ""
}

type roomImageResponse struct {
	Accessibility       int       `json:"Accessibility"`
	AccessibilityLocked bool      `json:"AccessibilityLocked"`
	CheerCount          int       `json:"CheerCount"`
	CommentCount        int       `json:"CommentCount"`
	CreatedAt           time.Time `json:"CreatedAt"`
	Id                  uint      `json:"Id"`
	ImageName           string    `json:"ImageName"`
	PlayerEventId       int       `json:"PlayerEventId"`
	PlayerId            uint      `json:"PlayerId"`
	RoomId              int       `json:"RoomId"`
	TaggedPlayerIds     []int     `json:"TaggedPlayerIds"`
	Type                int       `json:"Type"`
}

// GET /api/images/v4/room/{roomId}
func RoomImages(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	parts := strings.Split(strings.TrimPrefix(r.URL.Path, "/api/images/v4/room/"), "/")
	roomID, err := strconv.Atoi(parts[0])
	if err != nil || roomID <= 0 {
		http.Error(w, "invalid room id", http.StatusBadRequest)
		return
	}

	q := r.URL.Query()
	take, _ := strconv.Atoi(q.Get("take"))
	skip, _ := strconv.Atoi(q.Get("skip"))
	sort := q.Get("sort")

	if take <= 0 || take > 500 {
		take = 100
	}

	var photos []models.UploadedPhoto
	query := db.DB.Where("room_id = ?", roomID)
	if sort == "1" {
		query = query.Order("created_at DESC")
	} else {
		query = query.Order("created_at ASC")
	}
	query.Limit(take).Offset(skip).Find(&photos)

	log.Printf("[IMAGES] room %d: found %d photos (take=%d skip=%d)", roomID, len(photos), take, skip)

	out := make([]roomImageResponse, len(photos))
	for i, p := range photos {
		tagged := p.PlayerIDs
		if tagged == nil {
			tagged = []int{}
		}
		out[i] = roomImageResponse{
			Accessibility:       p.Accessibility,
			AccessibilityLocked: false,
			CheerCount:          p.CheerCount,
			CommentCount:        0,
			CreatedAt:           p.CreatedAt,
			Id:                  p.ID,
			ImageName:           p.ImageName,
			PlayerEventId:       p.PlayerEventID,
			PlayerId:            p.AccountID,
			RoomId:              p.RoomID,
			TaggedPlayerIds:     tagged,
			Type:                int(p.SavedImageType),
		}
	}
	json.NewEncoder(w).Encode(out)
}

type imageCheerRequest struct {
	Cheer        bool `json:"Cheer"`
	SavedImageId uint `json:"SavedImageId"`
}

// POST /api/images/v1/cheer
func ImageCheer(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	accountID, ok := AccountIDFromRequest(r)
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	var req imageCheerRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil || req.SavedImageId == 0 {
		json.NewEncoder(w).Encode(map[string]interface{}{
			"Message": "invalid request",
			"Success": false,
		})
		return
	}

	var photo models.UploadedPhoto
	if err := db.DB.First(&photo, req.SavedImageId).Error; err != nil {
		json.NewEncoder(w).Encode(map[string]interface{}{
			"Message": "image not found",
			"Success": false,
		})
		return
	}

	var existing models.UploadedPhotoCheer
	hasExisting := db.DB.Where("photo_id = ? AND account_id = ?", req.SavedImageId, accountID).First(&existing).Error == nil

	if req.Cheer && !hasExisting {
		db.DB.Create(&models.UploadedPhotoCheer{PhotoId: req.SavedImageId, AccountId: accountID})
		db.DB.Model(&models.UploadedPhoto{}).Where("id = ?", req.SavedImageId).
			UpdateColumn("cheer_count", gorm.Expr("cheer_count + ?", 1))
	} else if !req.Cheer && hasExisting {
		db.DB.Delete(&existing)
		db.DB.Model(&models.UploadedPhoto{}).Where("id = ? AND cheer_count > 0", req.SavedImageId).
			UpdateColumn("cheer_count", gorm.Expr("cheer_count - ?", 1))
	}

	json.NewEncoder(w).Encode(map[string]interface{}{
		"Message": "",
		"Success": true,
	})
}
