package auth

import (
	"crypto/rand"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"math/big"
	"net/http"
	"strconv"
	"strings"
	"time"

	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"

	"meow.net/db"
	"meow.net/discord"
	"meow.net/models"
	"meow.net/utils"
)

func reportAuthError(r *http.Request, stage, reason string) {
	discord.SendAuthError(discord.AuthErrorInfo{
		Stage:      stage,
		Platform:   r.FormValue("platform"),
		PlatformID: r.FormValue("platform_id"),
		Username:   r.FormValue("username"),
		IP:         utils.ClientIP(r),
		Reason:     reason,
	})
}

func issueTokenPair(w http.ResponseWriter, r *http.Request, accountID uint, platformID, platformStr string, isJunior, isDeveloper, isModerator, noToken bool) {
	if noToken {
		log.Printf("[TOKEN] account %d has no_token set, withholding token", accountID)
		writeNoTokenResponse(w)
		return
	}

	recordDeviceLogin(r, accountID, platformID, platformStr)

	accessToken := utils.MakeJWT(fmt.Sprintf("%d", accountID), platformID, platformStr, isJunior, isDeveloper, isModerator)

	tokenBytes := make([]byte, 32)
	rand.Read(tokenBytes)
	refreshTokenStr := fmt.Sprintf("%x", tokenBytes)

	rt := models.RefreshToken{
		Token:      refreshTokenStr,
		AccountID:  accountID,
		PlatformID: platformID,
		Platform:   platformStr,
		ExpiresAt:  time.Now().Add(30 * 24 * time.Hour),
	}
	if err := db.DB.Create(&rt).Error; err != nil {
		log.Printf("[TOKEN] failed to store refresh token for account %d: %v", accountID, err)
	}

	json.NewEncoder(w).Encode(map[string]any{
		"access_token":      accessToken,
		"refresh_token":     refreshTokenStr,
		"error":             nil,
		"error_description": nil,
		"key":               "ZWZmNzk5ZGEtM2RmOC00NWQ5LTkwNjYtYTZmZWU1ZmIzMjI4",
	})
}

func recordDeviceLogin(r *http.Request, accountID uint, platformID, platformStr string) {
	deviceID := r.FormValue("device_id")
	if deviceID == "" {
		return
	}
	deviceClass, _ := strconv.Atoi(r.FormValue("device_class"))
	ip := utils.ClientIP(r)
	now := time.Now()

	var dl models.DeviceLogin
	err := db.DB.Where("account_id = ? AND device_id = ?", accountID, deviceID).First(&dl).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		dl = models.DeviceLogin{
			AccountID:   accountID,
			DeviceID:    deviceID,
			DeviceClass: deviceClass,
			PlatformID:  platformID,
			Platform:    platformStr,
			IP:          ip,
			LoginCount:  1,
			FirstSeen:   now,
			LastSeen:    now,
		}
		if err := db.DB.Create(&dl).Error; err != nil {
			log.Printf("[DEVICE] failed to record login account=%d device=%s: %v", accountID, deviceID, err)
		}
		return
	}
	if err != nil {
		log.Printf("[DEVICE] lookup failed account=%d device=%s: %v", accountID, deviceID, err)
		return
	}

	if err := db.DB.Model(&dl).Updates(map[string]any{
		"login_count":  dl.LoginCount + 1,
		"last_seen":    now,
		"ip":           ip,
		"device_class": deviceClass,
		"platform_id":  platformID,
		"platform":     platformStr,
	}).Error; err != nil {
		log.Printf("[DEVICE] failed to update login account=%d device=%s: %v", accountID, deviceID, err)
	}
}

func isJuniorAccount(a models.Account) bool {
	return a.IsJunior == nil || *a.IsJunior || a.TreatAsJunior
}

func findAccount(platform int, platformId string, accountId uint) (models.SelfAccount, error) {
	var pa models.PlatformAccount
	q := db.DB.Where("platform = ? AND platform_id = ?", platform, platformId)
	if accountId > 0 {
		q = q.Where("account_id = ?", accountId)
	}
	result := q.First(&pa)
	if result.Error != nil {
		return models.SelfAccount{}, fmt.Errorf("account not found")
	}
	var account models.SelfAccount
	if err := db.DB.First(&account, pa.AccountID).Error; err != nil {
		return models.SelfAccount{}, err
	}
	return account, nil
}

const (
	accountIDMin         = 1_000_000
	accountIDMax         = 999_999_999
	accountIDMaxAttempts = 10
)

func randomAccountID() (uint, error) {
	n, err := rand.Int(rand.Reader, big.NewInt(accountIDMax-accountIDMin+1))
	if err != nil {
		return 0, err
	}
	return uint(accountIDMin + n.Int64()), nil
}

func createWithRandomID(account *models.SelfAccount) error {
	for attempt := 0; attempt < accountIDMaxAttempts; attempt++ {
		id, err := randomAccountID()
		if err != nil {
			return fmt.Errorf("generate account id: %w", err)
		}
		account.AccountID = id
		err = db.DB.Create(account).Error
		if err == nil {
			return nil
		}
		if errors.Is(err, gorm.ErrDuplicatedKey) {
			continue
		}
		return err
	}
	return fmt.Errorf("could not allocate a unique account id")
}

func CreateAccount(platform int, platformId, deviceID string) (models.SelfAccount, error) {
	if err := checkDeviceBan(deviceID); err != nil {
		log.Printf("[AUTH] CreateAccount: rejected banned device=%s platform=%d platformId=%s", deviceID, platform, platformId)
		return models.SelfAccount{}, err
	}
	if !models.PlatformType(platform).IsKnown() {
		log.Printf("[AUTH] CreateAccount: rejected unknown platform=%d platformId=%s", platform, platformId)
		return models.SelfAccount{}, fmt.Errorf("invalid platform")
	}
	if models.PlatformType(platform) == models.PlatformSteam {
		if err := utils.ValidateSteamAccountStanding(platformId); err != nil {
			log.Printf("[AUTH] CreateAccount: steam standing rejected platformId=%s: %v", platformId, err)
			return models.SelfAccount{}, err
		}
	}

	var existingCount int64
	db.DB.Model(&models.PlatformAccount{}).Where("platform = ? AND platform_id = ?", platform, platformId).Count(&existingCount)

	var limit models.PlatformAccountLimit
	if err := db.DB.Where("platform = ? AND platform_id = ?", platform, platformId).Attrs(models.PlatformAccountLimit{
		Platform:    platform,
		PlatformID:  platformId,
		MaxAccounts: 1,
	}).FirstOrCreate(&limit).Error; err != nil {
		log.Printf("[AUTH] CreateAccount: error fetching limit for platform=%d platformId=%s: %v", platform, platformId, err)
		return models.SelfAccount{}, fmt.Errorf("internal server error")
	}

	if existingCount >= int64(limit.MaxAccounts) {
		log.Printf("[AUTH] CreateAccount: account limit reached for platform=%d platformId=%s (%d >= %d)", platform, platformId, existingCount, limit.MaxAccounts)
		return models.SelfAccount{}, fmt.Errorf("account creation limit reached for this platform ID")
	}

	suffix := platformId
	if len(suffix) > 6 {
		suffix = suffix[:6]
	}
	account := models.SelfAccount{
		Account: models.Account{
			RawUsername:  "Player_" + suffix,
			Username:     "player_" + suffix,
			DisplayName:  "Player",
			Platforms:    platform,
			CreatedAt:    time.Now(),
			ProfileImage: "DefaultImage.png",
		},
	}
	if err := createWithRandomID(&account); err != nil {
		return account, err
	}
	db.DB.Create(&models.PlatformAccount{
		AccountID:  account.AccountID,
		Platform:   platform,
		PlatformID: platformId,
	})
	setupNewAccountDefaults(&account.Account)
	return account, nil
}

func ForceCreateAccount(platform int, platformId string) (models.SelfAccount, error) {
	if !models.PlatformType(platform).IsKnown() {
		return models.SelfAccount{}, fmt.Errorf("invalid platform")
	}

	suffix := platformId
	if len(suffix) > 6 {
		suffix = suffix[:6]
	}
	account := models.SelfAccount{
		Account: models.Account{
			RawUsername:  "Player_" + suffix,
			Username:     "player_" + suffix,
			DisplayName:  "Player",
			Platforms:    platform,
			CreatedAt:    time.Now(),
			ProfileImage: "DefaultImage.png",
		},
	}
	if err := createWithRandomID(&account); err != nil {
		return account, err
	}
	db.DB.Create(&models.PlatformAccount{
		AccountID:  account.AccountID,
		Platform:   platform,
		PlatformID: platformId,
	})
	setupNewAccountDefaults(&account.Account)
	return account, nil
}

func FindOrCreateAccount(platform int, platformId, deviceID string) (models.SelfAccount, error) {
	var pa models.PlatformAccount
	result := db.DB.Where("platform = ? AND platform_id = ?", platform, platformId).Find(&pa)
	if result.RowsAffected == 0 {
		return CreateAccount(platform, platformId, deviceID)
	}

	var account models.SelfAccount
	if err := db.DB.First(&account, pa.AccountID).Error; err != nil {
		log.Printf("[AUTH] orphaned platform_account id=%d accountId=%d, cleaning up and recreating", pa.ID, pa.AccountID)
		db.DB.Delete(&pa)
		return CreateAccount(platform, platformId, deviceID)
	}
	return account, nil
}

func ConnectToken(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	if err := utils.CheckProxy(r); err != nil {
		log.Printf("[TOKEN] %v", err)
		reportAuthError(r, "connect token", fmt.Sprintf("VPN/proxy detected: %v", err))
		json.NewEncoder(w).Encode(map[string]any{
			"access_token":      nil,
			"refresh_token":     "",
			"error":             "VPN or Proxy detected",
			"error_description": "Please disable your VPN or proxy and try again.",
			"key":               "",
		})
		return
	}

	body, _ := io.ReadAll(r.Body)
	r.Body = io.NopCloser(strings.NewReader(string(body)))
	r.ParseForm()

	switch r.FormValue("grant_type") {
	case "password":
		if r.FormValue("username") != "" && r.FormValue("password") != "" {
			connectTokenPassword(w, r)
			return
		}
		fallthrough
	case "cached_login":
		platform, _ := strconv.Atoi(r.FormValue("platform"))
		if models.PlatformType(platform) == models.PlatformOculus {
			connectTokenOculus(w, r)
		} else {
			connectTokenSteam(w, r)
		}
	case "refresh_token":
		connectTokenRefresh(w, r)
	default:
		log.Printf("[TOKEN] rejected: unknown grant_type=%q", r.FormValue("grant_type"))
		reportAuthError(r, "connect token", fmt.Sprintf("unknown grant_type=%q", r.FormValue("grant_type")))
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
	}
}

var ErrDeviceBanned = errors.New("device is banned")

func checkDeviceBan(deviceID string) error {
	if deviceID == "" {
		return nil
	}
	var ban models.DeviceBan
	err := db.DB.Where("device_id = ? AND (expires_at IS NULL OR expires_at > ?)", deviceID, time.Now()).First(&ban).Error
	if err == nil {
		return fmt.Errorf("%w: %s", ErrDeviceBanned, deviceID)
	}
	return nil
}

func writeBannedTokenResponse(w http.ResponseWriter) {
	json.NewEncoder(w).Encode(map[string]any{
		"access_token":      nil,
		"refresh_token":     "",
		"error":             "Banned",
		"error_description": "Your device has been banned.",
		"key":               "",
	})
}

func writeNoTokenResponse(w http.ResponseWriter) {
	json.NewEncoder(w).Encode(map[string]any{
		"access_token":      nil,
		"refresh_token":     "",
		"error":             "No token",
		"error_description": "This account is not permitted to connect.",
		"key":               "",
	})
}

func validatePlatformAuth(r *http.Request) error {
	platform, _ := strconv.Atoi(r.FormValue("platform"))
	return ValidatePlatformOwnership(platform, r.FormValue("platform_id"), r.FormValue("platform_auth"))
}

func ValidatePlatformOwnership(platform int, platformId, platformAuthStr string) error {
	if platformId == "" || platformAuthStr == "" {
		return fmt.Errorf("missing platform_id or platform_auth")
	}

	if models.PlatformType(platform) == models.PlatformOculus {
		var pa struct {
			Nonce string `json:"Nonce"`
			AppId string `json:"AppId"`
		}
		if err := json.Unmarshal([]byte(platformAuthStr), &pa); err != nil || pa.Nonce == "" || pa.AppId == "" {
			return fmt.Errorf("invalid oculus platform_auth JSON: %v", err)
		}
		return utils.ValidateOculusNonce(pa.Nonce, platformId, pa.AppId)
	}

	var pa struct {
		Ticket string `json:"Ticket"`
		AppID  string `json:"AppId"`
	}
	if err := json.Unmarshal([]byte(platformAuthStr), &pa); err != nil || pa.Ticket == "" {
		return fmt.Errorf("invalid steam platform_auth JSON: %v", err)
	}
	steamID, err := utils.ValidateSteamTicket(pa.Ticket, pa.AppID)
	if err != nil {
		return err
	}
	if steamID != platformId {
		return fmt.Errorf("steam id mismatch got=%s want=%s", steamID, platformId)
	}
	return nil
}

func connectTokenPassword(w http.ResponseWriter, r *http.Request) {
	username := strings.ToLower(strings.TrimSpace(r.FormValue("username")))
	password := r.FormValue("password")
	platformId := r.FormValue("platform_id")
	platformStr := r.FormValue("platform")

	if err := validatePlatformAuth(r); err != nil {
		log.Printf("[TOKEN] password: platform auth check failed for username=%s: %v", username, err)
		reportAuthError(r, "password login", fmt.Sprintf("platform auth check failed: %v", err))
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	var account models.SelfAccount
	if err := db.DB.Where("LOWER(username) = ?", username).First(&account).Error; err != nil {
		log.Printf("[TOKEN] password: no account for username=%s", username)
		reportAuthError(r, "password login", fmt.Sprintf("no account for username=%s", username))
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	if account.PasswordHash == nil || *account.PasswordHash == "" {
		log.Printf("[TOKEN] password: account %d has no password set", account.AccountID)
		reportAuthError(r, "password login", fmt.Sprintf("account #%d has no password set", account.AccountID))
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	if err := bcrypt.CompareHashAndPassword([]byte(*account.PasswordHash), []byte(password)); err != nil {
		log.Printf("[TOKEN] password: bad password for account %d", account.AccountID)
		reportAuthError(r, "password login", fmt.Sprintf("bad password for account #%d", account.AccountID))
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	platform, _ := strconv.Atoi(platformStr)
	linkPlatformToAccount(account.AccountID, platform, platformId)

	issueTokenPair(w, r, account.AccountID, platformId, platformStr, isJuniorAccount(account.Account), account.IsDeveloper, account.IsModerator, account.NoToken)
}

func linkPlatformToAccount(accountID uint, platform int, platformId string) {
	if platformId == "" {
		return
	}
	var existing models.PlatformAccount
	err := db.DB.Where("account_id = ? AND platform = ? AND platform_id = ?", accountID, platform, platformId).First(&existing).Error
	if err == nil {
		return
	}
	if err := db.DB.Create(&models.PlatformAccount{
		AccountID:  accountID,
		Platform:   platform,
		PlatformID: platformId,
	}).Error; err != nil {
		log.Printf("[AUTH] linkPlatformToAccount: failed to link account=%d platform=%d platformId=%s: %v", accountID, platform, platformId, err)
		return
	}
	log.Printf("[AUTH] linked platform=%d platformId=%s to account=%d", platform, platformId, accountID)
}

func connectTokenOculus(w http.ResponseWriter, r *http.Request) {
	log.Printf("[TOKEN] oculus: incoming request from %s", utils.ClientIP(r))
	for k, v := range r.Form {
		log.Printf("[TOKEN] oculus:   %s = %v", k, v)
	}

	platformId := r.FormValue("platform_id")
	platformStr := r.FormValue("platform")

	if err := validatePlatformAuth(r); err != nil {
		log.Printf("[TOKEN] oculus: platform_auth check failed: %v", err)
		reportAuthError(r, "oculus token", fmt.Sprintf("platform auth check failed: %v", err))
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	platform, _ := strconv.Atoi(platformStr)
	accountIdVal, _ := strconv.ParseUint(r.FormValue("account_id"), 10, 64)

	var account models.SelfAccount
	var err error
	if accountIdVal > 0 {
		account, err = findAccount(platform, platformId, uint(accountIdVal))
	}
	if accountIdVal == 0 || err != nil {
		account, err = FindOrCreateAccount(platform, platformId, r.FormValue("device_id"))
	}
	if err != nil {
		log.Printf("[TOKEN] oculus: account lookup failed: %v", err)
		reportAuthError(r, "oculus token", fmt.Sprintf("account lookup/create failed: %v", err))
		if errors.Is(err, ErrDeviceBanned) {
			writeBannedTokenResponse(w)
			return
		}
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	issueTokenPair(w, r, account.AccountID, platformId, platformStr, isJuniorAccount(account.Account), account.IsDeveloper, account.IsModerator, account.NoToken)
}

func connectTokenSteam(w http.ResponseWriter, r *http.Request) {
	platformId := r.FormValue("platform_id")
	platformStr := r.FormValue("platform")

	if err := validatePlatformAuth(r); err != nil {
		log.Printf("[TOKEN] steam: platform_auth check failed: %v", err)
		reportAuthError(r, "steam token", fmt.Sprintf("platform auth check failed: %v", err))
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	platform, _ := strconv.Atoi(platformStr)
	accountIdVal, _ := strconv.ParseUint(r.FormValue("account_id"), 10, 64)

	var account models.SelfAccount
	var err error
	if accountIdVal > 0 {
		account, err = findAccount(platform, platformId, uint(accountIdVal))
	}
	if accountIdVal == 0 || err != nil {
		account, err = FindOrCreateAccount(platform, platformId, r.FormValue("device_id"))
	}
	if err != nil {
		log.Printf("[TOKEN] steam: account lookup failed: %v", err)
		reportAuthError(r, "steam token", fmt.Sprintf("account lookup/create failed: %v", err))
		if errors.Is(err, ErrDeviceBanned) {
			writeBannedTokenResponse(w)
			return
		}
		if errors.Is(err, utils.ErrSteamStandingFailed) {
			w.WriteHeader(http.StatusForbidden)
			json.NewEncoder(w).Encode(map[string]any{
				"access_token":      nil,
				"refresh_token":     "",
				"error":             "Steam account standing check failed",
				"error_description": err.Error(),
				"key":               "",
			})
			return
		}
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	issueTokenPair(w, r, account.AccountID, platformId, platformStr, isJuniorAccount(account.Account), account.IsDeveloper, account.IsModerator, account.NoToken)
}

func connectTokenRefresh(w http.ResponseWriter, r *http.Request) {
	tokenStr := r.FormValue("refresh_token")
	if tokenStr == "" {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	var rt models.RefreshToken
	if err := db.DB.Where("token = ?", tokenStr).First(&rt).Error; err != nil {
		log.Printf("[TOKEN] refresh: token not found")
		reportAuthError(r, "refresh token", "refresh token not found")
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}
	if time.Now().After(rt.ExpiresAt) {
		log.Printf("[TOKEN] refresh: token expired for account %d", rt.AccountID)
		discord.SendAuthError(discord.AuthErrorInfo{
			Stage:     "refresh token",
			Platform:  rt.Platform,
			AccountID: rt.AccountID,
			IP:        utils.ClientIP(r),
			Reason:    "refresh token expired",
		})
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}
	if rt.UsedAt != nil {
		log.Printf("[TOKEN] refresh: REUSE DETECTED account=%d, revoking all tokens", rt.AccountID)
		discord.SendAuthError(discord.AuthErrorInfo{
			Stage:     "refresh token",
			Platform:  rt.Platform,
			AccountID: rt.AccountID,
			IP:        utils.ClientIP(r),
			Reason:    "refresh token reuse detected",
		})
		db.DB.Where("account_id = ?", rt.AccountID).Delete(&models.RefreshToken{})
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	now := time.Now()
	db.DB.Model(&rt).Update("used_at", &now)

	var account models.Account
	if err := db.DB.First(&account, rt.AccountID).Error; err != nil {
		discord.SendAuthError(discord.AuthErrorInfo{
			Stage:     "refresh token",
			Platform:  rt.Platform,
			AccountID: rt.AccountID,
			IP:        utils.ClientIP(r),
			Reason:    fmt.Sprintf("account lookup failed: %v", err),
		})
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	issueTokenPair(w, r, rt.AccountID, rt.PlatformID, rt.Platform, isJuniorAccount(account), account.IsDeveloper, account.IsModerator, account.NoToken)
}

func PlatformLogin(w http.ResponseWriter, r *http.Request) {
	if err := utils.CheckProxy(r); err != nil {
		log.Printf("[AUTH] PlatformLogin %v", err)
		reportAuthError(r, "platform login", fmt.Sprintf("VPN/proxy detected: %v", err))
		http.Error(w, err.Error(), http.StatusForbidden)
		return
	}

	r.ParseForm()
	platformId := r.FormValue("platform_id")
	platformStr := r.FormValue("platform")
	platformAuthStr := r.FormValue("platform_auth")
	platform, _ := strconv.Atoi(platformStr)
	accountIdVal, _ := strconv.ParseUint(r.FormValue("account_id"), 10, 64)

	if platformId == "" {
		log.Printf("[AUTH] PlatformLogin: missing platform_id")
		reportAuthError(r, "platform login", "missing platform_id")
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	if platformAuthStr == "" {
		log.Printf("[AUTH] PlatformLogin: missing platform_auth")
		reportAuthError(r, "platform login", "missing platform auth")
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	if err := validatePlatformAuth(r); err != nil {
		log.Printf("[AUTH] PlatformLogin: platform_auth check failed for platformId=%s: %v", platformId, err)
		reportAuthError(r, "platform login", fmt.Sprintf("platform auth check failed: %v", err))
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	account, err := findAccount(platform, platformId, uint(accountIdVal))
	if err != nil {
		log.Printf("[AUTH] PlatformLogin: account not found for platform=%s platformId=%s", platformStr, platformId)
		reportAuthError(r, "platform login", fmt.Sprintf("account not found: %v", err))
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	issueTokenPair(w, r, account.AccountID, platformId, platformStr, isJuniorAccount(account.Account), account.IsDeveloper, account.IsModerator, account.NoToken)
}

func LoginToCachedAccount(w http.ResponseWriter, r *http.Request) {
	PlatformLogin(w, r)
}

func CachedLoginForPlatformId(w http.ResponseWriter, r *http.Request) {
	parts := strings.Split(strings.TrimPrefix(r.URL.Path, "/cachedlogin/forplatformid/"), "/")
	var platformId string
	platform := 0
	if len(parts) >= 2 {
		if p, err := strconv.Atoi(parts[0]); err == nil {
			platform = p
		}
		platformId = parts[1]
	}

	var platformAccounts []models.PlatformAccount
	db.DB.Where("platform = ? AND platform_id = ?", platform, platformId).Find(&platformAccounts)

	results := make([]map[string]any, 0, len(platformAccounts))
	for _, pa := range platformAccounts {
		var account models.SelfAccount
		if err := db.DB.First(&account, pa.AccountID).Error; err != nil {
			continue
		}
		results = append(results, map[string]any{
			"accountId":       account.AccountID,
			"lastLoginTime":   "2020-06-26T00:00:00Z",
			"platform":        platform,
			"platformId":      platformId,
			"requirePassword": false,
		})
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(results)
}

func CachedLoginForPlatformIds(w http.ResponseWriter, r *http.Request) {
	r.ParseForm()
	ids := r.Form["id"]

	var platformAccounts []models.PlatformAccount
	db.DB.Where("platform_id IN ?", ids).Find(&platformAccounts)

	results := make([]map[string]any, 0, len(platformAccounts))
	for _, pa := range platformAccounts {
		results = append(results, map[string]any{
			"platform":        pa.Platform,
			"platformId":      pa.PlatformID,
			"accountId":       pa.AccountID,
			"lastLoginTime":   "2020-06-26T00:00:00Z",
			"requirePassword": false,
		})
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(results)
}
