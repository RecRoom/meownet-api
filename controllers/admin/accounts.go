package admin

import (
	"encoding/json"
	"net/http"
	"strconv"

	"meow.net/controllers/auth"
	"meow.net/db"
	"meow.net/models"
)

func ForceCreateAccount(w http.ResponseWriter, r *http.Request) {
	if !requireAdmin(w, r) {
		return
	}

	var body struct {
		Platform   int    `json:"platform"`
		PlatformID string `json:"platform_id"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}

	if body.PlatformID == "" {
		http.Error(w, "missing platform_id", http.StatusBadRequest)
		return
	}

	account, err := auth.ForceCreateAccount(body.Platform, body.PlatformID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	writeJSON(w, http.StatusOK, account)
}

func AdjustBalance(w http.ResponseWriter, r *http.Request) {
	if !requireAdmin(w, r) {
		return
	}
	accountID, err := strconv.Atoi(r.PathValue("id"))
	if err != nil {
		http.Error(w, "bad account id", http.StatusBadRequest)
		return
	}

	var body struct {
		CurrencyType int `json:"CurrencyType"`
		Delta        int `json:"Delta"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		http.Error(w, "invalid body", http.StatusBadRequest)
		return
	}

	var bal models.Balance
	db.DB.
		Where("account_id = ? AND currency_type = ?", accountID, body.CurrencyType).
		Attrs(models.Balance{AccountID: uint(accountID), CurrencyType: body.CurrencyType, BalanceType: -2}).
		FirstOrCreate(&bal)

	bal.Amount += body.Delta
	if err := db.DB.Save(&bal).Error; err != nil {
		http.Error(w, "db error", http.StatusInternalServerError)
		return
	}
	writeJSON(w, http.StatusOK, bal)
}

func RevokeAvatarItem(w http.ResponseWriter, r *http.Request) {
	if !requireAdmin(w, r) {
		return
	}
	accountID, err := strconv.Atoi(r.PathValue("id"))
	if err != nil {
		http.Error(w, "bad account id", http.StatusBadRequest)
		return
	}
	desc := r.PathValue("desc")
	if desc == "" {
		http.Error(w, "bad desc", http.StatusBadRequest)
		return
	}
	if err := db.DB.
		Where("account_id = ? AND avatar_item_desc = ?", accountID, desc).
		Delete(&models.UserAvatarItem{}).Error; err != nil {
		http.Error(w, "db error", http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}
