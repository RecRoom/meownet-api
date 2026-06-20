package player

import (
	"encoding/json"
	"io"
	"net/http"
	"strconv"
	"strings"
	"time"

	"gorm.io/gorm/clause"
	"meow.net/controllers/hub"
	"meow.net/db"
	"meow.net/models"
	"meow.net/utils"
)

const (
	settingsRateEvery = 90 * time.Second
	settingsRateBurst = 20
)

func PlayerGet(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	viewerID, ok := accountIDFromRequest(r)
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	idStrs := r.URL.Query()["id"]
	ids := make([]int, 0, len(idStrs))
	for _, idStr := range idStrs {
		if id, err := strconv.Atoi(idStr); err == nil {
			ids = append(ids, id)
		}
	}
	results := hub.BuildPresenceForBatch(int(viewerID), ids)
	json.NewEncoder(w).Encode(results)
}

func PlayerPhotonRegionPings(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
}

// for icy :3
func PlayerStatusVisibility(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
}

func PlayerLogin(w http.ResponseWriter, r *http.Request) {
	writeGranted := func() {
		w.Header().Set("Content-Type", "application/json; charset=utf-8")
		w.Write([]byte("0"))
	}

	accountId, ok := accountIDFromRequest(r)
	if !ok {
		writeGranted()
		return
	}

	body, _ := io.ReadAll(r.Body)
	token := strings.TrimSpace(strings.Trim(string(body), `"`))
	if token == "" {
		writeGranted()
		return
	}

	var st models.PlayerState
	err := db.DB.Select("login_lock_token").First(&st, accountId).Error
	held := err == nil && st.LoginLockToken != nil && *st.LoginLockToken != ""
	if held && *st.LoginLockToken != token && hub.HubIsOnline(int(accountId)) {
		http.Error(w, "Already logged in", http.StatusConflict)
		return
	}

	upsertPlayerState(accountId, models.PlayerState{AccountID: accountId, LoginLockToken: &token}, "login_lock_token")

	writeGranted()
}

func PlayerLogout(w http.ResponseWriter, r *http.Request) {
	if accountId, ok := accountIDFromRequest(r); ok {
		hub.ClearLoginLock(int(accountId))
	}
	w.WriteHeader(http.StatusOK)
}

func PlayerHeartbeat(w http.ResponseWriter, r *http.Request) {
	authHeader := r.Header.Get("Authorization")
	tokenStr := strings.TrimPrefix(authHeader, "Bearer ")
	if tokenStr == "" {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}
	accountIdStr, err := utils.ParseSubFromJWT(tokenStr)
	if err != nil || accountIdStr == "" {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}
	accountId, err := strconv.Atoi(accountIdStr)
	if err != nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	var roomInstance interface{} = nil
	if instanceId, ok := hub.GetPlayerInstance(accountId); ok && instanceId > 0 {
		var instance models.RoomInstance
		if err := db.DB.First(&instance, instanceId).Error; err == nil {
			roomInstance = instance
		}
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(hub.BuildSelfStatus(accountId, roomInstance, 0))
}

func PlayerAvoidJuniors(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	accountId, ok := accountIDFromRequest(r)
	if !ok {
		w.Write([]byte("false"))
		return
	}

	var st models.PlayerState
	if err := db.DB.Select("avoid_juniors").First(&st, accountId).Error; err != nil {
		w.Write([]byte("false"))
		return
	}
	writeBool(w, st.AvoidJuniors)
}

func writeBool(w http.ResponseWriter, v bool) {
	if v {
		w.Write([]byte("true"))
	} else {
		w.Write([]byte("false"))
	}
}

func upsertPlayerState(accountId uint, st models.PlayerState, column string) {
	db.DB.Select("account_id", column).
		Clauses(clause.OnConflict{
			Columns:   []clause.Column{{Name: "account_id"}},
			DoUpdates: clause.AssignmentColumns([]string{column}),
		}).
		Create(&st)
}

func Settings(w http.ResponseWriter, r *http.Request) {
	tokenStr := utils.GetBearerToken(r)
	if tokenStr == "" {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}
	accountId, err := utils.ParseSubFromJWT(tokenStr)
	if err != nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	accId, _ := strconv.ParseUint(accountId, 10, 32)
	uintAccountId := uint(accId)

	if !utils.AccountActionAllowBurst("settings_v2", uintAccountId, settingsRateEvery, settingsRateBurst) {
		http.Error(w, "too many requests", http.StatusTooManyRequests)
		return
	}

	var settings []models.PlayerSetting
	db.DB.Where("account_id = ?", uintAccountId).Find(&settings)

	if settings == nil {
		settings = []models.PlayerSetting{}
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(settings)
}

func SettingsSet(w http.ResponseWriter, r *http.Request) {
	tokenStr := utils.GetBearerToken(r)
	if tokenStr == "" {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}
	accountId, err := utils.ParseSubFromJWT(tokenStr)
	if err != nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	accId, _ := strconv.ParseUint(accountId, 10, 32)
	uintAccountId := uint(accId)

	if !utils.AccountActionAllowBurst("settings_v2", uintAccountId, settingsRateEvery, settingsRateBurst) {
		http.Error(w, "too many requests", http.StatusTooManyRequests)
		return
	}

	bodyBytes, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "Bad Request", http.StatusBadRequest)
		return
	}

	var reqSettings []models.PlayerSetting
	if err := json.Unmarshal(bodyBytes, &reqSettings); err != nil {
		var singleSetting models.PlayerSetting
		if err := json.Unmarshal(bodyBytes, &singleSetting); err != nil {
			http.Error(w, "Bad Request", http.StatusBadRequest)
			return
		}
		reqSettings = append(reqSettings, singleSetting)
	}

	for _, setting := range reqSettings {
		var existing models.PlayerSetting
		result := db.DB.Where("account_id = ? AND key = ?", uintAccountId, setting.Key).First(&existing)
		if result.Error == nil {
			db.DB.Model(&existing).Update("value", setting.Value)
		} else {
			setting.AccountID = uintAccountId
			db.DB.Omit("Account").Create(&setting)
		}
	}

	var allSettings []models.PlayerSetting
	db.DB.Where("account_id = ?", uintAccountId).Find(&allSettings)
	if allSettings == nil {
		allSettings = []models.PlayerSetting{}
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(allSettings)
}

func PageviewConsume(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
}
