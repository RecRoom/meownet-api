package rooms

import (
	"encoding/json"
	"net/http"
	"strconv"
	"strings"

	"gorm.io/gorm"
	"meow.net/controllers"
	"meow.net/db"
	"meow.net/models"
)

func getRoomInteraction(roomId uint, accountId uint) models.RoomInteraction {
	var interaction models.RoomInteraction
	db.DB.Where("room_id = ? AND account_id = ?", roomId, accountId).FirstOrCreate(&interaction, models.RoomInteraction{
		RoomId:    roomId,
		AccountId: accountId,
	})
	return interaction
}

func RoomInteractionGet(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	parts := strings.Split(strings.Trim(r.URL.Path, "/"), "/")
	if len(parts) < 2 {
		http.Error(w, "Bad Request", http.StatusBadRequest)
		return
	}
	roomId, err := strconv.Atoi(parts[1])
	if err != nil {
		http.Error(w, "Bad Request", http.StatusBadRequest)
		return
	}

	userId, err := controllers.CurrentUserIDFromRequest(r)
	if err != nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	interaction := getRoomInteraction(uint(roomId), userId)
	json.NewEncoder(w).Encode(map[string]bool{
		"Cheered":   interaction.Cheered,
		"Favorited": interaction.Favorited,
	})
}

func RoomInteractionCheer(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	parts := strings.Split(strings.Trim(r.URL.Path, "/"), "/")
	if len(parts) < 2 {
		http.Error(w, "Bad Request", http.StatusBadRequest)
		return
	}
	roomId, err := strconv.Atoi(parts[1])
	if err != nil {
		http.Error(w, "Bad Request", http.StatusBadRequest)
		return
	}

	userId, err := controllers.CurrentUserIDFromRequest(r)
	if err != nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	interaction := getRoomInteraction(uint(roomId), userId)
	if !interaction.Cheered {
		interaction.Cheered = true
		db.DB.Save(&interaction)
		db.DB.Model(&models.Room{}).Where("room_id = ?", roomId).UpdateColumn("cheer_count", gorm.Expr("cheer_count + ?", 1))
	}

	json.NewEncoder(w).Encode(map[string]bool{
		"Cheered":   interaction.Cheered,
		"Favorited": interaction.Favorited,
	})
}

func RoomInteractionFavorite(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	parts := strings.Split(strings.Trim(r.URL.Path, "/"), "/")
	if len(parts) < 2 {
		http.Error(w, "Bad Request", http.StatusBadRequest)
		return
	}
	roomId, err := strconv.Atoi(parts[1])
	if err != nil {
		http.Error(w, "Bad Request", http.StatusBadRequest)
		return
	}

	userId, err := controllers.CurrentUserIDFromRequest(r)
	if err != nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	interaction := getRoomInteraction(uint(roomId), userId)
	if !interaction.Favorited {
		interaction.Favorited = true
		db.DB.Save(&interaction)
		db.DB.Model(&models.Room{}).Where("room_id = ?", roomId).UpdateColumn("favorite_count", gorm.Expr("favorite_count + ?", 1))
	}

	json.NewEncoder(w).Encode(map[string]bool{
		"Cheered":   interaction.Cheered,
		"Favorited": interaction.Favorited,
	})
}

func RoomInteractionUncheer(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	parts := strings.Split(strings.Trim(r.URL.Path, "/"), "/")
	if len(parts) < 2 {
		http.Error(w, "Bad Request", http.StatusBadRequest)
		return
	}
	roomId, err := strconv.Atoi(parts[1])
	if err != nil {
		http.Error(w, "Bad Request", http.StatusBadRequest)
		return
	}

	userId, err := controllers.CurrentUserIDFromRequest(r)
	if err != nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	interaction := getRoomInteraction(uint(roomId), userId)
	if interaction.Cheered {
		interaction.Cheered = false
		db.DB.Save(&interaction)
		db.DB.Model(&models.Room{}).Where("room_id = ?", roomId).UpdateColumn("cheer_count", gorm.Expr("cheer_count - ?", 1))
	}

	json.NewEncoder(w).Encode(map[string]bool{
		"Cheered":   interaction.Cheered,
		"Favorited": interaction.Favorited,
	})
}

func RoomInteractionUnfavorite(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	parts := strings.Split(strings.Trim(r.URL.Path, "/"), "/")
	if len(parts) < 2 {
		http.Error(w, "Bad Request", http.StatusBadRequest)
		return
	}
	roomId, err := strconv.Atoi(parts[1])
	if err != nil {
		http.Error(w, "Bad Request", http.StatusBadRequest)
		return
	}

	userId, err := controllers.CurrentUserIDFromRequest(r)
	if err != nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	interaction := getRoomInteraction(uint(roomId), userId)
	if interaction.Favorited {
		interaction.Favorited = false
		db.DB.Save(&interaction)
		db.DB.Model(&models.Room{}).Where("room_id = ?", roomId).UpdateColumn("favorite_count", gorm.Expr("favorite_count - ?", 1))
	}

	json.NewEncoder(w).Encode(map[string]bool{
		"Cheered":   interaction.Cheered,
		"Favorited": interaction.Favorited,
	})
}
