package controllers

import (
	"encoding/json"
	"errors"
	"io"
	"log"
	"net/http"
	"strconv"

	"gorm.io/gorm"
	"meow.net/db"
	"meow.net/discord"
	"meow.net/models"
	"meow.net/utils"
)

const hashesURL = "https://cdn.cookedasset.com/build/hashes.json"

type anticheatCallbackBody struct {
	Type     string `json:"type"`
	DeviceID string `json:"device_id"`
	Details  string `json:"details"`
}

func AnticheatCallback(w http.ResponseWriter, r *http.Request) {
	var body anticheatCallbackBody
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"error": "invalid payload"})
		return
	}

	detection, ok := models.ParseDetectionType(body.Type)
	if !ok {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"error": "unknown detection type"})
		return
	}
	if body.DeviceID == "" {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"error": "device_id required"})
		return
	}

	accountID, accountName := accountForDevice(body.DeviceID)

	log.Printf("[ANTICHEAT] detection type=%s device=%s account=%d details=%q", body.Type, body.DeviceID, accountID, body.Details)

	discord.SendAnticheatDetection(discord.AnticheatInfo{
		DetectionType: body.Type,
		DeviceID:      body.DeviceID,
		Details:       body.Details,
		IP:            utils.ClientIP(r),
		AccountID:     accountID,
		AccountName:   accountName,
	})

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]bool{"ok": true, "close": detection.ShouldClose()})
}

func AnticheatHashes(w http.ResponseWriter, r *http.Request) {
	resp, err := http.Get(hashesURL)
	if err != nil {
		log.Printf("[ANTICHEAT] failed to fetch hashes: %v", err)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadGateway)
		json.NewEncoder(w).Encode(map[string]string{"error": "failed to fetch hashes"})
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		log.Printf("[ANTICHEAT] hashes fetch returned status %d", resp.StatusCode)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadGateway)
		json.NewEncoder(w).Encode(map[string]string{"error": "failed to fetch hashes"})
		return
	}

	w.Header().Set("Content-Type", "application/json")
	if _, err := io.Copy(w, resp.Body); err != nil {
		log.Printf("[ANTICHEAT] failed to write hashes response: %v", err)
	}
}

// reports whether a given platform id owns a given user
func AnticheatVerifyOwner(w http.ResponseWriter, r *http.Request) {
	r.ParseForm()

	platform, platformErr := strconv.Atoi(r.FormValue("platform"))
	platformID := r.FormValue("platform_id")

	userIDStr := r.FormValue("user_id")
	if userIDStr == "" {
		userIDStr = r.FormValue("account_id")
	}
	userID, userErr := strconv.ParseUint(userIDStr, 10, 64)

	if platformErr != nil || platformID == "" || userErr != nil || userID == 0 {
		writeOwnership(w, http.StatusBadRequest, false)
		return
	}

	var count int64
	if err := db.DB.Model(&models.PlatformAccount{}).
		Where("platform = ? AND platform_id = ? AND account_id = ?", platform, platformID, uint(userID)).
		Count(&count).Error; err != nil {
		log.Printf("[ANTICHEAT] ownership lookup failed platform=%d platformId=%s userId=%d: %v", platform, platformID, userID, err)
		writeOwnership(w, http.StatusInternalServerError, false)
		return
	}

	writeOwnership(w, http.StatusOK, count > 0)
}

func writeOwnership(w http.ResponseWriter, status int, owns bool) {
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	w.WriteHeader(status)
	if owns {
		w.Write([]byte("true"))
	} else {
		w.Write([]byte("false"))
	}
}

func accountForDevice(deviceID string) (uint, string) {
	var dl models.DeviceLogin
	err := db.DB.Where("device_id = ?", deviceID).Order("last_seen DESC").First(&dl).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return 0, ""
	}
	if err != nil {
		log.Printf("[ANTICHEAT] device lookup failed device=%s: %v", deviceID, err)
		return 0, ""
	}

	var account models.Account
	if err := db.DB.First(&account, dl.AccountID).Error; err != nil {
		return dl.AccountID, ""
	}
	name := account.Username
	if name == "" {
		name = account.DisplayName
	}
	return dl.AccountID, name
}
