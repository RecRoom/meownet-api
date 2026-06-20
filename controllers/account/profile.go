package account

import (
	"encoding/json"
	"io"
	"log"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	"meow.net/controllers"
	"meow.net/controllers/hub"
	"meow.net/db"
	"meow.net/models"
	"meow.net/utils"
)

func AccountUpdateDisplayName(w http.ResponseWriter, r *http.Request) {
	tokenStr := utils.GetBearerToken(r)
	if tokenStr == "" {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}
	accountId, err := utils.ParseSubFromJWT(tokenStr)
	if err != nil || accountId == "" {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	r.ParseForm()
	body, _ := io.ReadAll(r.Body)
	r.Body = io.NopCloser(strings.NewReader(string(body)))

	var displayName string
	if r.FormValue("displayName") != "" {
		displayName = r.FormValue("displayName")
	} else {
		for _, pair := range strings.Split(string(body), "&") {
			kv := strings.Split(pair, "=")
			if len(kv) == 2 && kv[0] == "displayName" {
				displayName = kv[1]
			}
		}
	}

	if !utils.IsValidName(displayName) {
		writeModerationRejection(w, "Display names can only use letters, numbers, and basic punctuation.")
		return
	}

	if !utils.IsValidNameLength(displayName) {
		writeModerationRejection(w, "Display names can be at most 16 characters.")
		return
	}

	if utils.IsTextFlagged(displayName) {
		writeModerationRejection(w, "Display name violates the community guidelines.")
		return
	}

	var acc models.Account
	if err := db.DB.First(&acc, accountId).Error; err != nil {
		http.Error(w, "Not Found", http.StatusNotFound)
		return
	}

	acc.DisplayName = displayName
	db.DB.Save(&acc)

	var selfAcc models.SelfAccount
	if db.DB.First(&selfAcc, acc.AccountID).Error == nil {
		hub.HubSendToPlayer(int(acc.AccountID), hub.NotifFrame("SelfAccountUpdate", selfAcc))
		hub.HubSendToPlayer(int(acc.AccountID), hub.NotifFrame("AccountUpdate", selfAcc.Account))
		hub.HubBroadcastAccountUpdate(int(acc.AccountID))
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"error":   "",
		"success": true,
		"value":   acc,
	})
}

func AccountUpdateUsername(w http.ResponseWriter, r *http.Request) {
	tokenStr := utils.GetBearerToken(r)
	if tokenStr == "" {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}
	accountId, err := utils.ParseSubFromJWT(tokenStr)
	if err != nil || accountId == "" {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	r.ParseForm()
	body, _ := io.ReadAll(r.Body)
	r.Body = io.NopCloser(strings.NewReader(string(body)))

	var username string
	if r.FormValue("username") != "" {
		username = r.FormValue("username")
	} else {
		for _, pair := range strings.Split(string(body), "&") {
			kv := strings.Split(pair, "=")
			if len(kv) == 2 && kv[0] == "username" {
				username = kv[1]
			}
		}
	}

	accountIdInt, _ := strconv.Atoi(accountId)

	sendSelfAccountToPlayer := func() {
		var selfAcc models.SelfAccount
		if db.DB.First(&selfAcc, accountIdInt).Error == nil {
			hub.HubSendToPlayer(accountIdInt, hub.NotifFrame("SelfAccountUpdate", selfAcc))
		}
	}

	if !utils.IsValidUsername(username) || utils.IsTextFlagged(username) {
		sendSelfAccountToPlayer()
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"error":   "Usernames can only contain letters, numbers, and underscores",
			"success": false,
			"value":   "",
		})
		return
	}

	if !utils.IsValidUsernameLength(username) {
		sendSelfAccountToPlayer()
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"error":   "Usernames must be between 3 and 16 characters",
			"success": false,
			"value":   "",
		})
		return
	}

	lowerUsername := strings.ToLower(username)

	var existing models.Account
	if err := db.DB.Where("LOWER(username) = ?", lowerUsername).First(&existing).Error; err == nil && int(existing.AccountID) != accountIdInt {
		sendSelfAccountToPlayer()
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"error":   "This username is already taken",
			"success": false,
			"value":   "",
		})
		return
	}

	var acc models.Account
	if err := db.DB.First(&acc, accountId).Error; err != nil {
		http.Error(w, "Not Found", http.StatusNotFound)
		return
	}

	acc.Username = lowerUsername
	acc.RawUsername = username
	acc.DisplayName = username
	db.DB.Save(&acc)

	db.DB.Model(&models.Room{}).Where("creator_account_id = ? AND is_dorm = ?", acc.AccountID, true).Update("name", "@"+acc.Username+"'s Dorm")

	var selfAcc models.SelfAccount
	if db.DB.First(&selfAcc, acc.AccountID).Error == nil {
		selfAcc.AvailableUsernameChanges = defaultUsernameChanges
		hub.HubSendToPlayer(int(acc.AccountID), hub.NotifFrame("SelfAccountUpdate", selfAcc))
		hub.HubSendToPlayer(int(acc.AccountID), hub.NotifFrame("AccountUpdate", selfAcc.Account))
		hub.HubBroadcastAccountUpdate(int(acc.AccountID))
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"error":   "",
		"success": true,
		"value":   acc,
	})
}

func AccountGetBio(w http.ResponseWriter, r *http.Request) {
	parts := strings.Split(r.URL.Path, "/")
	if len(parts) < 4 {
		http.Error(w, "Bad Request", http.StatusBadRequest)
		return
	}
	idStr := parts[2]
	accountId, err := strconv.Atoi(idStr)
	if err != nil {
		http.Error(w, "Bad Request", http.StatusBadRequest)
		return
	}

	var bio models.PlayerBio
	if err := db.DB.First(&bio, accountId).Error; err != nil {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{"accountId": accountId, "bio": nil})
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{"accountId": bio.AccountID, "bio": bio.Bio})
}

func AccountUpdateBio(w http.ResponseWriter, r *http.Request) {
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

	r.ParseForm()
	body, _ := io.ReadAll(r.Body)
	r.Body = io.NopCloser(strings.NewReader(string(body)))

	var newBio string
	if r.FormValue("bio") != "" {
		newBio = r.FormValue("bio")
	} else {
		for _, pair := range strings.Split(string(body), "&") {
			kv := strings.Split(pair, "=")
			if len(kv) == 2 && kv[0] == "bio" {
				newBio, _ = url.QueryUnescape(kv[1])
			}
		}
	}

	if utils.IsTextFlagged(newBio) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"Message": "Bio violates the community guidelines.",
			"Success": false,
		})
		return
	}

	var bio models.PlayerBio
	if err := db.DB.First(&bio, accountId).Error; err != nil {
		bio = models.PlayerBio{
			AccountID: uint(accountId),
			Bio:       newBio,
		}
		db.DB.Create(&bio)
	} else {
		bio.Bio = newBio
		db.DB.Save(&bio)
	}

	var acc models.Account
	db.DB.First(&acc, accountId)
	if acc.ProfileImage == "" {
		acc.ProfileImage = defaultProfileImage
	}
	accountJSON, _ := json.Marshal(acc)

	hub.HubBroadcastAccountUpdate(accountId)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"Message": string(accountJSON),
		"Success": true,
	})
}

func AccountUpdateBirthday(w http.ResponseWriter, r *http.Request) {
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

	body, _ := io.ReadAll(r.Body)
	r.Body = io.NopCloser(strings.NewReader(string(body)))

	var birthday string
	if r.FormValue("birthday") != "" {
		birthday = r.FormValue("birthday")
	} else {
		var j map[string]interface{}
		if err := json.Unmarshal(body, &j); err == nil {
			if v, ok := j["birthday"].(string); ok {
				birthday = v
			}
		}
		if birthday == "" {
			for _, pair := range strings.Split(string(body), "&") {
				kv := strings.Split(pair, "=")
				if len(kv) == 2 && kv[0] == "birthday" {
					birthday, _ = url.QueryUnescape(kv[1])
				}
			}
		}
	}

	parsed, err := time.Parse("2006-01-02", birthday)
	if err != nil {
		parsed, err = time.Parse(time.RFC3339, birthday)
	}
	if err != nil {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"error":   "Invalid date format",
			"success": false,
			"value":   nil,
		})
		return
	}

	age := time.Now().Year() - parsed.Year()
	if time.Now().YearDay() < parsed.YearDay() {
		age--
	}

	if age < juniorAgeThreshold {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"error":   "You must be at least 13 years old to create a non-junior account",
			"success": false,
			"value":   nil,
		})
		return
	}

	var acc models.SelfAccount
	if err := db.DB.First(&acc, accountId).Error; err != nil {
		http.Error(w, "Not Found", http.StatusNotFound)
		return
	}

	acc.Birthday = &birthday
	notJunior := false
	acc.IsJunior = &notJunior
	db.DB.Save(&acc)
	acc.AvailableUsernameChanges = defaultUsernameChanges

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"error":   nil,
		"success": true,
		"value":   acc,
	})
}

func AccountProfileImage(w http.ResponseWriter, r *http.Request) {
	accountID, ok := controllers.AccountIDFromRequest(r)
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	if !utils.AccountActionAllowBurst("profile_image", accountID, 10*time.Second, 5) {
		http.Error(w, "too many requests", http.StatusTooManyRequests)
		return
	}

	body, _ := io.ReadAll(r.Body)
	r.Body = io.NopCloser(strings.NewReader(string(body)))

	imageName := r.FormValue("imageName")
	if imageName == "" {
		var j map[string]interface{}
		if err := json.Unmarshal(body, &j); err == nil {
			if v, ok := j["imageName"].(string); ok {
				imageName = v
			}
		}
	}
	if imageName == "" {
		for _, pair := range strings.Split(string(body), "&") {
			kv := strings.SplitN(pair, "=", 2)
			if len(kv) == 2 && kv[0] == "imageName" {
				imageName, _ = url.QueryUnescape(kv[1])
			}
		}
	}

	if imageName == "" {
		http.Error(w, "missing imageName", http.StatusBadRequest)
		return
	}

	raw, err := controllers.LoadStoredImage(imageName)
	if err != nil {
		log.Printf("[ACCOUNT] profileimage load %q error: %v", imageName, err)
		http.Error(w, "image not found", http.StatusBadGateway)
		return
	}

	square, err := squareProfilePNG(raw, profileImageSize)
	if err != nil {
		log.Printf("[ACCOUNT] profileimage decode %q error: %v", imageName, err)
		http.Error(w, "invalid image", http.StatusBadRequest)
		return
	}
	profileName := makeProfileImageName(accountID)
	if err := controllers.SaveStoredImage(profileName, square, "image/png"); err != nil {
		log.Printf("[ACCOUNT] profileimage save %q error: %v", profileName, err)
		http.Error(w, "save failed", http.StatusInternalServerError)
		return
	}

	if err := db.DB.Model(&models.Account{}).
		Where("account_id = ?", accountID).
		Update("profile_image", profileName).Error; err != nil {
		log.Printf("[ACCOUNT] profileimage update error: %v", err)
		if delErr := controllers.DeleteStoredImage(profileName); delErr != nil {
			log.Printf("[ACCOUNT] profileimage cleanup %q error: %v", profileName, delErr)
		}
		http.Error(w, "update failed", http.StatusInternalServerError)
		return
	}

	if imageName != profileName {
		if err := controllers.DeleteStoredImage(imageName); err != nil {
			log.Printf("[ACCOUNT] profileimage delete original %q error: %v", imageName, err)
		}
	}

	var selfAcc models.SelfAccount
	if db.DB.First(&selfAcc, accountID).Error == nil {
		hub.HubSendToPlayer(int(accountID), hub.NotifFrame("SelfAccountUpdate", selfAcc))
		hub.HubSendToPlayer(int(accountID), hub.NotifFrame("AccountUpdate", selfAcc.Account))
		hub.HubBroadcastAccountUpdate(int(accountID))
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success":   true,
		"imageName": profileName,
	})
}
