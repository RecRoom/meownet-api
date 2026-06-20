package importer

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"strings"

	"meow.net/controllers/admin"
	"meow.net/db"
	"meow.net/models"
)

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}

type tagInput struct {
	Tag  string `json:"tag"`
	Type int    `json:"type"`
}

type subroomInput struct {
	Name          string `json:"name"`
	UnitySceneId  string `json:"unity_scene_id"`
	IsSandbox     bool   `json:"is_sandbox"`
	MaxPlayers    int    `json:"max_players"`
	Accessibility int    `json:"accessibility"`
	BlobField     string `json:"blob_field"`
}

type metadataInput struct {
	Name                     string         `json:"name"`
	Description              string         `json:"description"`
	ImageFilename            string         `json:"image_filename"`
	WarningMask              int            `json:"warning_mask"`
	CustomWarning            string         `json:"custom_warning"`
	AutoLocalizeRoom         bool           `json:"auto_localize_room"`
	DisableMicAutoMute       bool           `json:"disable_mic_auto_mute"`
	DisableRoomComments      bool           `json:"disable_room_comments"`
	EncryptVoiceChat         bool           `json:"encrypt_voice_chat"`
	LoadScreenLocked         bool           `json:"load_screen_locked"`
	MaxPlayerCalculationMode int            `json:"max_player_calculation_mode"`
	MaxPlayers               int            `json:"max_players"`
	MinLevel                 int            `json:"min_level"`
	PersistenceVersion       int            `json:"persistence_version"`
	SupportsJuniors          bool           `json:"supports_juniors"`
	SupportsLevelVoting      bool           `json:"supports_level_voting"`
	SupportsMobile           bool           `json:"supports_mobile"`
	SupportsQuest2           bool           `json:"supports_quest_2"`
	SupportsScreens          bool           `json:"supports_screens"`
	SupportsTeleportVR       bool           `json:"supports_teleport_vr"`
	SupportsVRLow            bool           `json:"supports_vr_low"`
	SupportsWalkVR           bool           `json:"supports_walk_vr"`
	ToxmodEnabled            bool           `json:"toxmod_enabled"`
	UgcVersion               int            `json:"ugc_version"`
	IsDorm                   bool           `json:"is_dorm"`
	IsRRO                    bool           `json:"is_rro"`
	Tags                     []tagInput     `json:"tags"`
	Subrooms                 []subroomInput `json:"subrooms"`
}

func Upload(w http.ResponseWriter, r *http.Request) {
	if !admin.RequireAdmin(w, r) {
		return
	}

	if err := r.ParseMultipartForm(256 << 20); err != nil {
		log.Printf("[IMPORT] upload: bad multipart: %v", err)
		http.Error(w, "invalid multipart: "+err.Error(), http.StatusBadRequest)
		return
	}

	metaRaw := r.FormValue("metadata")
	if metaRaw == "" {
		http.Error(w, "missing metadata field", http.StatusBadRequest)
		return
	}
	var meta metadataInput
	if err := json.Unmarshal([]byte(metaRaw), &meta); err != nil {
		log.Printf("[IMPORT] upload: bad metadata json: %v", err)
		http.Error(w, "invalid metadata json: "+err.Error(), http.StatusBadRequest)
		return
	}
	if meta.Name == "" {
		http.Error(w, "metadata.name required", http.StatusBadRequest)
		return
	}
	if len(meta.Subrooms) == 0 {
		http.Error(w, "metadata.subrooms required (at least one)", http.StatusBadRequest)
		return
	}
	for i, sr := range meta.Subrooms {
		if sr.UnitySceneId == "" {
			http.Error(w, fmt.Sprintf("subrooms[%d].unity_scene_id required", i), http.StatusBadRequest)
			return
		}
		if sr.BlobField == "" {
			http.Error(w, fmt.Sprintf("subrooms[%d].blob_field required", i), http.StatusBadRequest)
			return
		}
	}
	log.Printf("[IMPORT] upload name=%q subrooms=%d has_image=%v",
		meta.Name, len(meta.Subrooms), meta.ImageFilename != "")

	imageName, err := maybeStoreImage(r, meta.ImageFilename)
	if err != nil {
		log.Printf("[IMPORT] upload image err: %v", err)
		http.Error(w, "image store failed: "+err.Error(), http.StatusInternalServerError)
		return
	}

	blobNames := make([]string, len(meta.Subrooms))
	for i, sr := range meta.Subrooms {
		name, err := storeSubroomBlob(r, sr.BlobField)
		if err != nil {
			log.Printf("[IMPORT] upload blob %s err: %v", sr.BlobField, err)
			http.Error(w, fmt.Sprintf("blob %s store failed: %v", sr.BlobField, err), http.StatusInternalServerError)
			return
		}
		blobNames[i] = name
	}

	if err := storeHolotars(r); err != nil {
		log.Printf("[IMPORT] upload holotar err: %v", err)
		http.Error(w, "holotar store failed: "+err.Error(), http.StatusInternalServerError)
		return
	}

	loaded, err := insertRoom(meta, imageName, blobNames)
	if err != nil {
		log.Printf("[IMPORT] upload db err: %v", err)
		writeJSON(w, http.StatusInternalServerError, map[string]any{"error": err.Error()})
		return
	}
	log.Printf("[IMPORT] upload success: new room_id=%d", loaded.RoomId)

	writeJSON(w, http.StatusOK, map[string]any{
		"error":   "",
		"success": true,
		"value":   loaded,
	})
}

func maybeStoreImage(r *http.Request, originalFilename string) (string, error) {
	file, header, err := r.FormFile("image")
	if err == http.ErrMissingFile {
		return "", nil
	}
	if err != nil {
		return "", err
	}
	defer file.Close()

	data, err := io.ReadAll(file)
	if err != nil {
		return "", err
	}

	ext := imageExtFromName(originalFilename)
	if ext == "" {
		ext = imageExtFromName(header.Filename)
	}
	newName := makeImageName(ext)
	contentType := header.Header.Get("Content-Type")
	if contentType == "" {
		contentType = imageContentTypeFromExt(ext)
	}
	if err := storeImage(newName, data, contentType); err != nil {
		return "", err
	}
	log.Printf("[IMPORT] upload stored image %s (%d bytes, ct=%s)", newName, len(data), contentType)
	return newName, nil
}

func storeSubroomBlob(r *http.Request, fieldName string) (string, error) {
	file, header, err := r.FormFile(fieldName)
	if err != nil {
		return "", fmt.Errorf("missing blob field %q: %w", fieldName, err)
	}
	defer file.Close()

	data, err := io.ReadAll(file)
	if err != nil {
		return "", err
	}
	newName := makeRoomBlobName()
	if err := storeRoomBlob(newName, data); err != nil {
		return "", err
	}
	log.Printf("[IMPORT] upload stored blob room/%s (%d bytes, src=%q)", newName, len(data), header.Filename)
	return newName, nil
}

func storeHolotars(r *http.Request) error {
	if r.MultipartForm == nil {
		return nil
	}
	for _, h := range r.MultipartForm.File["holotar"] {
		if h.Filename == "" {
			return fmt.Errorf("holotar missing filename")
		}
		file, err := h.Open()
		if err != nil {
			return fmt.Errorf("open holotar: %w", err)
		}
		data, err := io.ReadAll(file)
		file.Close()
		if err != nil {
			return fmt.Errorf("read holotar: %w", err)
		}
		if err := storeHolotar(h.Filename, data); err != nil {
			return fmt.Errorf("store holotar: %w", err)
		}
		log.Printf("[IMPORT] upload stored holotar data/%s (%d bytes)", h.Filename, len(data))
	}
	return nil
}

func insertRoom(meta metadataInput, imageName string, blobNames []string) (*models.Room, error) {
	metadataBlob := importerRoomMetadataBlob
	newRoom := models.Room{
		Name:                     meta.Name,
		Description:              meta.Description,
		ImageName:                imageName,
		CreatorAccountId:         importerOwnerAccountID,
		State:                    0,
		Accessibility:            0,
		AutoLocalizeRoom:         meta.AutoLocalizeRoom,
		CloningAllowed:           false,
		CustomWarning:            meta.CustomWarning,
		DisableMicAutoMute:       meta.DisableMicAutoMute,
		DisableRoomComments:      meta.DisableRoomComments,
		EncryptVoiceChat:         meta.EncryptVoiceChat,
		IsDeveloperOwned:         false,
		IsDorm:                   meta.IsDorm,
		IsRRO:                    meta.IsRRO,
		LoadScreenLocked:         meta.LoadScreenLocked,
		MaxPlayerCalculationMode: meta.MaxPlayerCalculationMode,
		MaxPlayers:               meta.MaxPlayers,
		MinLevel:                 meta.MinLevel,
		PersistenceVersion:       meta.PersistenceVersion,
		RankedEntityId:           "",
		RankingContext:           0,
		SupportsJuniors:          meta.SupportsJuniors,
		SupportsLevelVoting:      meta.SupportsLevelVoting,
		SupportsMobile:           meta.SupportsMobile,
		SupportsQuest2:           meta.SupportsQuest2,
		SupportsScreens:          meta.SupportsScreens,
		SupportsTeleportVR:       meta.SupportsTeleportVR,
		SupportsVRLow:            meta.SupportsVRLow,
		SupportsWalkVR:           meta.SupportsWalkVR,
		ToxmodEnabled:            meta.ToxmodEnabled,
		UgcVersion:               meta.UgcVersion,
		WarningMask:              meta.WarningMask,
		DataBlob:                 &metadataBlob,
		Roles: []models.RoomRoleEntry{
			{AccountId: importerOwnerAccountID, InvitedRole: 0, Role: 255},
		},
	}

	for i, sr := range meta.Subrooms {
		newRoom.SubRooms = append(newRoom.SubRooms, models.SubRoom{
			Accessibility:    sr.Accessibility,
			DataBlob:         blobNames[i],
			IsSandbox:        sr.IsSandbox,
			MaxPlayers:       sr.MaxPlayers,
			Name:             sr.Name,
			SavedByAccountId: importerOwnerAccountID,
			UnitySceneId:     sr.UnitySceneId,
		})
	}
	for _, t := range meta.Tags {
		newRoom.Tags = append(newRoom.Tags, models.RoomTag{Tag: t.Tag, Type: t.Type})
	}

	if err := db.DB.Create(&newRoom).Error; err != nil {
		return nil, fmt.Errorf("db insert: %w", err)
	}

	var loaded models.Room
	if err := db.DB.Preload("SubRooms").Preload("Roles").Preload("Tags").
		First(&loaded, newRoom.RoomId).Error; err != nil {
		return nil, fmt.Errorf("db reload: %w", err)
	}
	if loaded.SubRooms == nil {
		loaded.SubRooms = []models.SubRoom{}
	}
	if loaded.Roles == nil {
		loaded.Roles = []models.RoomRoleEntry{}
	}
	if loaded.Tags == nil {
		loaded.Tags = []models.RoomTag{}
	}
	loaded.LoadScreens = []interface{}{}
	loaded.PromoImages = []interface{}{}
	loaded.PromoExternalContent = []interface{}{}
	return &loaded, nil
}

func imageExtFromName(name string) string {
	if i := strings.LastIndex(name, "."); i >= 0 {
		return name[i:]
	}
	return ""
}

func imageContentTypeFromExt(ext string) string {
	switch strings.ToLower(ext) {
	case ".png":
		return "image/png"
	case ".gif":
		return "image/gif"
	case ".webp":
		return "image/webp"
	default:
		return "image/jpeg"
	}
}
