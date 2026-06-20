package controllers

import (
	"net/http"
	"strconv"
	"time"

	"meow.net/db"
	"meow.net/models"
	"meow.net/utils"
)

func AccountIDFromRequest(r *http.Request) (uint, bool) {
	token := utils.GetBearerToken(r)
	if token == "" {
		return 0, false
	}
	sub, err := utils.ParseSubFromJWT(token)
	if err != nil || sub == "" {
		return 0, false
	}
	id, err := strconv.ParseUint(sub, 10, 64)
	if err != nil {
		return 0, false
	}
	return uint(id), true
}

func CurrentUserIDFromRequest(r *http.Request) (uint, error) {
	id, ok := AccountIDFromRequest(r)
	if !ok {
		return 0, http.ErrNoCookie
	}
	return id, nil
}

func AccountIsBanned(accountID uint) bool {
	if accountID == 0 {
		return false
	}
	var count int64
	db.DB.Model(&models.AccountBan{}).
		Where("account_id = ? AND (expires_at IS NULL OR expires_at > ?)", accountID, time.Now()).
		Count(&count)
	return count > 0
}
