package account

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"strings"

	"golang.org/x/crypto/bcrypt"
	"meow.net/controllers/auth"
	"meow.net/db"
	"meow.net/discord"
	"meow.net/models"
	"meow.net/utils"
)

const (
	defaultProfileImage    = "DefaultImage.png"
	defaultUsernameChanges = 3
	juniorAgeThreshold     = 13
)

func AccountMe(w http.ResponseWriter, r *http.Request) {
	tokenStr := utils.GetBearerToken(r)
	if tokenStr == "" {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}
	accountIdStr, err := utils.ParseSubFromJWT(tokenStr)
	if err != nil || accountIdStr == "" {
		log.Printf("Error parsing JWT: %v", err)
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	accountId, err := strconv.Atoi(accountIdStr)
	if err != nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	var acc models.SelfAccount
	result := db.DB.First(&acc, accountId)
	if result.Error != nil {
		log.Printf("Error fetching account: %v", result.Error)
		http.Error(w, "Account not found", http.StatusNotFound)
		return
	}

	if accountId == 0 {
		http.Error(w, "Account not found", http.StatusNotFound)
		return
	}

	if acc.ProfileImage == "" {
		acc.ProfileImage = defaultProfileImage
	}
	acc.AvailableUsernameChanges = defaultUsernameChanges

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(acc)
}

func ParentalControlMe(w http.ResponseWriter, r *http.Request) {
	tokenStr := utils.GetBearerToken(r)
	if tokenStr == "" {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}
	accountId, err := utils.ParseSubFromJWT(tokenStr)
	if err != nil {
		log.Printf("Error parsing JWT: %v", err)
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	fmt.Fprintf(w, `{"accountId":%s,"disallowInAppPurchases":false}`, accountId)
}

func AccountCreate(w http.ResponseWriter, r *http.Request) {
	r.ParseForm()
	platformStr := r.FormValue("platform")
	platformId := r.FormValue("platformId")
	platform, _ := strconv.Atoi(platformStr)

	w.Header().Set("Content-Type", "application/json")

	reject := func(reason string) {
		discord.SendAuthError(discord.AuthErrorInfo{
			Stage:      "create account",
			Platform:   platformStr,
			PlatformID: platformId,
			IP:         utils.ClientIP(r),
			Reason:     reason,
		})

		w.WriteHeader(http.StatusForbidden)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"error":   "Account creation is not available",
			"success": false,
			"value":   nil,
		})
	}

	switch models.PlatformType(platform) {
	case models.PlatformOculus:
		if err := utils.ValidateOculusUserExists(platformId); err != nil {
			reject(fmt.Sprintf("oculus user validation failed: %v", err))
			return
		}
	case models.PlatformSteam:
		ticket := r.Header.Get("x-steam-ticket")
		if ticket == "" {
			reject("missing x-steam-ticket header")
			return
		}
		steamID, err := utils.ValidateSteamTicket(ticket, "")
		if err != nil {
			reject(fmt.Sprintf("steam ticket validation failed: %v", err))
			return
		}
		if steamID != platformId {
			reject(fmt.Sprintf("steam id mismatch got=%s want=%s", steamID, platformId))
			return
		}
	default:
		reject("account/create disabled")
		return
	}

	account, err := auth.CreateAccount(platform, platformId, r.FormValue("device_id"))
	if err != nil {
		log.Printf("[ACCOUNT] create failed platform=%d platformId=%s: %v", platform, platformId, err)
		discord.SendAuthError(discord.AuthErrorInfo{
			Stage:      "create account",
			Platform:   platformStr,
			PlatformID: platformId,
			IP:         utils.ClientIP(r),
			Reason:     fmt.Sprintf("account create failed: %v", err),
		})

		w.WriteHeader(http.StatusForbidden)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"error":   err.Error(),
			"success": false,
			"value":   nil,
		})
		return
	}

	json.NewEncoder(w).Encode(map[string]interface{}{
		"error":   nil,
		"success": true,
		"value":   account,
	})
}

const (
	minSearchNameLen   = 2
	accountSearchLimit = 50
)

func AccountSearch(w http.ResponseWriter, r *http.Request) {
	name := strings.TrimSpace(r.URL.Query().Get("name"))
	var accounts []models.Account
	if len([]rune(name)) >= minSearchNameLen {
		pattern := "%" + utils.EscapeLike(strings.ToLower(name)) + "%"
		db.DB.Where(`LOWER(username) LIKE ? ESCAPE '\'`, pattern).
			Limit(accountSearchLimit).
			Find(&accounts)
	}
	if accounts == nil {
		accounts = []models.Account{}
	}

	for i := range accounts {
		if accounts[i].ProfileImage == "" {
			accounts[i].ProfileImage = defaultProfileImage
		}
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(accounts)
}

func AccountBulk(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	ids := r.URL.Query()["id"]
	var accountIds []int
	for _, idStr := range ids {
		if id, err := strconv.Atoi(idStr); err == nil {
			accountIds = append(accountIds, id)
		}
	}

	accounts := make([]models.Account, 0)
	if len(accountIds) > 0 {
		db.DB.Where("account_id IN ?", accountIds).Find(&accounts)
	}

	for i := range accounts {
		if accounts[i].ProfileImage == "" {
			accounts[i].ProfileImage = defaultProfileImage
		}
	}

	json.NewEncoder(w).Encode(accounts)
}

func AccountGet(w http.ResponseWriter, r *http.Request) {
	parts := strings.Split(r.URL.Path, "/")
	if len(parts) < 3 {
		http.Error(w, "Bad Request", http.StatusBadRequest)
		return
	}
	idStr := parts[2]
	accountId, err := strconv.Atoi(idStr)
	if err != nil {
		http.Error(w, "Bad Request", http.StatusBadRequest)
		return
	}

	var acc models.Account
	result := db.DB.First(&acc, accountId)
	if result.Error != nil {
		http.Error(w, "Not Found", http.StatusNotFound)
		return
	}

	if acc.ProfileImage == "" {
		acc.ProfileImage = defaultProfileImage
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(acc)
}

func RoleCheck(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	parts := strings.Split(strings.Trim(r.URL.Path, "/"), "/")
	if len(parts) >= 3 && strings.EqualFold(parts[1], "developer") {
		accountId, err := strconv.Atoi(parts[2])
		if err == nil {
			var acc models.Account
			if err := db.DB.First(&acc, accountId).Error; err == nil && (acc.IsDeveloper || acc.IsModerator) {
				w.Write([]byte("true"))
				return
			}
		}
	}

	w.Write([]byte("false"))
}

func AccountHasPassword(w http.ResponseWriter, r *http.Request) {
	tokenStr := utils.GetBearerToken(r)
	if tokenStr == "" {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}
	accountIdStr, err := utils.ParseSubFromJWT(tokenStr)
	if err != nil || accountIdStr == "" {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	accountId, _ := strconv.Atoi(accountIdStr)
	var acc models.Account
	if err := db.DB.First(&acc, accountId).Error; err != nil {
		http.Error(w, "Not Found", http.StatusNotFound)
		return
	}

	if acc.PasswordHash != nil && *acc.PasswordHash != "" {
		fmt.Fprint(w, "true")
	} else {
		fmt.Fprint(w, "false")
	}
}

func AccountChangePassword(w http.ResponseWriter, r *http.Request) {
	tokenStr := utils.GetBearerToken(r)
	if tokenStr == "" {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}
	accountIdStr, err := utils.ParseSubFromJWT(tokenStr)
	if err != nil || accountIdStr == "" {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	r.ParseForm()
	newPassword := r.FormValue("newPassword")
	if newPassword == "" {
		http.Error(w, "Bad Request", http.StatusBadRequest)
		return
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(newPassword), bcrypt.DefaultCost)
	if err != nil {
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	accountId, _ := strconv.Atoi(accountIdStr)
	hashStr := string(hash)
	if err := db.DB.Model(&models.Account{}).Where("account_id = ?", accountId).Update("password_hash", hashStr).Error; err != nil {
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	var acc models.Account
	if err := db.DB.First(&acc, accountId).Error; err != nil {
		http.Error(w, "Not Found", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"error":   nil,
		"success": true,
		"value":   acc,
	})
}

func NamegenOptions(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"Adjectives": []string{
			"Purrfect", "Fluffy", "Whiskered", "Sleek", "Playful", "Cuddly", "Mysterious",
		},
		"Nouns": []string{
			"Meow", "Cat", "Kitty", "Purr", "Whisker", "Claw", "Tail", "Fur",
		},
	})
}
