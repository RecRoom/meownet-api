package inventions

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

func randomInventionId() int64 {
	var b [8]byte
	if _, err := rand.Read(b[:]); err != nil {
		return 0
	}
	v := int64(binary.BigEndian.Uint64(b[:]) >> 1)
	if v == 0 {
		v = 1
	}
	return v
}

func randomReplicationId() string {
	var b [8]byte
	if _, err := rand.Read(b[:]); err != nil {
		return "0"
	}
	v := binary.BigEndian.Uint64(b[:]) >> 1
	return strconv.FormatUint(v, 10)
}

func uniqueInventionId() int64 {
	for i := 0; i < 8; i++ {
		id := randomInventionId()
		var existing models.Invention
		if err := db.DB.Select("invention_id").Where("invention_id = ?", id).First(&existing).Error; err != nil {
			return id
		}
	}
	return randomInventionId()
}

func parseInt64Param(r *http.Request, key string) (int64, bool) {
	raw := strings.TrimSpace(r.URL.Query().Get(key))
	if raw == "" {
		return 0, false
	}
	v, err := strconv.ParseInt(raw, 10, 64)
	if err != nil {
		return 0, false
	}
	return v, true
}

func parseIntParam(r *http.Request, key string) (int, bool) {
	raw := strings.TrimSpace(r.URL.Query().Get(key))
	if raw == "" {
		return 0, false
	}
	v, err := strconv.Atoi(raw)
	if err != nil {
		return 0, false
	}
	return v, true
}

func parsePermissionLevel(s string) (int, bool) {
	switch strings.ToLower(strings.TrimSpace(s)) {
	case "unassigned", "0":
		return int(models.InventionPermissionUnassigned), true
	case "limitedoneuseonly", "limited_one_use_only", "10":
		return int(models.InventionPermissionLimitedOneUseOnly), true
	case "useonly", "use_only", "20":
		return int(models.InventionPermissionUseOnly), true
	case "editandsave", "edit_and_save", "40":
		return int(models.InventionPermissionEditAndSave), true
	case "publish", "60":
		return int(models.InventionPermissionPublish), true
	case "charge", "80":
		return int(models.InventionPermissionCharge), true
	case "unlimited", "100":
		return int(models.InventionPermissionUnlimited), true
	}
	if iv, err := strconv.Atoi(strings.TrimSpace(s)); err == nil {
		return iv, true
	}
	return 0, false
}

func loadInvention(id int64) (models.Invention, bool) {
	var inv models.Invention
	if err := db.DB.First(&inv, id).Error; err != nil {
		return models.Invention{}, false
	}
	return inv, true
}

func loadVersion(inventionId int64, versionNumber int) (models.InventionVersion, bool) {
	var v models.InventionVersion
	if err := db.DB.Where("invention_id = ? AND version_number = ?", inventionId, versionNumber).First(&v).Error; err != nil {
		return models.InventionVersion{}, false
	}
	return v, true
}

func loadCurrentVersion(inv models.Invention) (models.InventionVersion, bool) {
	return loadVersion(inv.InventionId, inv.CurrentVersionNumber)
}

func ownsInvention(accountId uint, inventionId int64) bool {
	if accountId == 0 {
		return false
	}
	var inv models.Invention
	if err := db.DB.Select("creator_player_id").Where("invention_id = ?", inventionId).First(&inv).Error; err == nil {
		if uint(inv.CreatorPlayerId) == accountId {
			return true
		}
	}
	var own models.InventionOwnership
	if err := db.DB.Where("invention_id = ? AND account_id = ?", inventionId, accountId).First(&own).Error; err == nil {
		return true
	}
	return false
}

func isCreator(accountId uint, inv models.Invention) bool {
	return accountId != 0 && uint(inv.CreatorPlayerId) == accountId
}

func writeJSON(w http.ResponseWriter, status int, value interface{}) {
	w.Header().Set("Content-Type", "application/json")
	if status != 0 && status != http.StatusOK {
		w.WriteHeader(status)
	}
	_ = json.NewEncoder(w).Encode(value)
}

func writeError(w http.ResponseWriter, status int, msg string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(map[string]string{"error": msg})
}
