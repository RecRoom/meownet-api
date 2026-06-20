package store

import (
	"encoding/json"
	"net/http"
	"strconv"
	"strings"

	"github.com/google/uuid"
	"meow.net/controllers"
	"meow.net/db"
	"meow.net/models"
)

func WishlistDispatch(w http.ResponseWriter, r *http.Request) {
	parts := strings.Split(strings.Trim(r.URL.Path, "/"), "/")
	if len(parts) >= 5 && parts[4] == "me" {
		if r.Method == http.MethodGet {
			wishlistMe(w, r)
		} else if r.Method == http.MethodPut && len(parts) >= 6 {
			wishlistAdd(w, r, parts[5])
		} else if r.Method == http.MethodDelete && len(parts) >= 6 {
			wishlistRemove(w, r, parts[5])
		} else {
			http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
		}
		return
	}
	if len(parts) >= 5 && r.Method == http.MethodGet {
		wishlistByAccount(w, r, parts[4])
		return
	}
	http.Error(w, "Not Found", http.StatusNotFound)
}

func wishlistMe(w http.ResponseWriter, r *http.Request) {
	accountID, ok := controllers.AccountIDFromRequest(r)
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	var items []models.WishlistItem
	db.DB.Where("account_id = ?", accountID).Find(&items)
	if items == nil {
		items = []models.WishlistItem{}
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(items)
}

func wishlistAdd(w http.ResponseWriter, r *http.Request, itemIdStr string) {
	accountID, ok := controllers.AccountIDFromRequest(r)
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	purchasableItemId, err := strconv.Atoi(itemIdStr)
	if err != nil {
		http.Error(w, "Bad Request", http.StatusBadRequest)
		return
	}

	var existing models.WishlistItem
	if db.DB.Where("account_id = ? AND purchasable_item_id = ?", accountID, purchasableItemId).First(&existing).Error == nil {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"error":   "",
			"success": true,
			"value":   existing,
		})
		return
	}

	item := models.WishlistItem{
		WishlistItemId:    uuid.New().String(),
		AccountId:         int(accountID),
		PurchasableItemId: purchasableItemId,
	}
	db.DB.Create(&item)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"error":   "",
		"success": true,
		"value":   item,
	})
}

func wishlistRemove(w http.ResponseWriter, r *http.Request, itemIdStr string) {
	accountID, ok := controllers.AccountIDFromRequest(r)
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	purchasableItemId, err := strconv.Atoi(itemIdStr)
	if err != nil {
		http.Error(w, "Bad Request", http.StatusBadRequest)
		return
	}

	db.DB.Where("account_id = ? AND purchasable_item_id = ?", accountID, purchasableItemId).Delete(&models.WishlistItem{})
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"error":   "",
		"success": true,
		"value":   nil,
	})
}

func wishlistByAccount(w http.ResponseWriter, r *http.Request, accountIdStr string) {
	accountId, err := strconv.Atoi(accountIdStr)
	if err != nil {
		http.Error(w, "Bad Request", http.StatusBadRequest)
		return
	}

	var items []models.WishlistItem
	db.DB.Where("account_id = ?", accountId).Find(&items)
	if items == nil {
		items = []models.WishlistItem{}
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(items)
}

func AdCarouselItems(w http.ResponseWriter, r *http.Request) {
	http.ServeFile(w, r, "data/jsons/adcarouselitems.json")
}
