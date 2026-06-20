package moderation

import (
	"net/http"

	"meow.net/db"
	"meow.net/models"
)

func DeviceIdUpdate(w http.ResponseWriter, r *http.Request) {
	r.ParseForm()
	oldDeviceID := r.FormValue("oldDeviceId")
	newDeviceID := r.FormValue("newDeviceId")

	if oldDeviceID != "" && newDeviceID != "" && oldDeviceID != newDeviceID {
		migrateDeviceID(oldDeviceID, newDeviceID)
	}

	w.WriteHeader(http.StatusOK)
}

func migrateDeviceID(oldDeviceID, newDeviceID string) {
	var existingBan models.DeviceBan
	if err := db.DB.Where("device_id = ?", newDeviceID).First(&existingBan).Error; err != nil {
		db.DB.Model(&models.DeviceBan{}).Where("device_id = ?", oldDeviceID).
			Update("device_id", newDeviceID)
	} else {
		db.DB.Where("device_id = ?", oldDeviceID).Delete(&models.DeviceBan{})
	}

	var logins []models.DeviceLogin
	db.DB.Where("device_id = ?", oldDeviceID).Find(&logins)
	for _, login := range logins {
		var count int64
		db.DB.Model(&models.DeviceLogin{}).
			Where("account_id = ? AND device_id = ?", login.AccountID, newDeviceID).
			Count(&count)
		if count > 0 {
			db.DB.Delete(&models.DeviceLogin{}, login.ID)
			continue
		}
		db.DB.Model(&models.DeviceLogin{}).Where("id = ?", login.ID).
			Update("device_id", newDeviceID)
	}
}
