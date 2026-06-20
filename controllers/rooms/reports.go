package rooms

import (
	"encoding/json"
	"net/http"
	"strconv"

	"meow.net/controllers"
	"meow.net/db"
	"meow.net/discord"
	"meow.net/models"
)

func RoomReportCreate(w http.ResponseWriter, r *http.Request) {
	accountID, _ := controllers.AccountIDFromRequest(r)

	r.ParseForm()
	roomID, _ := strconv.ParseInt(r.FormValue("RoomId"), 10, 64)
	reportCategory, _ := strconv.Atoi(r.FormValue("ReportCategory"))
	details := r.FormValue("Details")

	if roomID == 0 {
		writeReportResult(w, false, "invalid room")
		return
	}

	var room models.Room
	if err := db.DB.First(&room, roomID).Error; err != nil {
		writeReportResult(w, false, "room not found")
		return
	}

	report := models.RoomReport{
		ReporterID:     accountID,
		RoomID:         roomID,
		ReportCategory: reportCategory,
		Details:        details,
	}
	db.DB.Create(&report)

	discord.SendRoomReport(discord.RoomReportInfo{
		ReporterID:   accountID,
		ReporterName: lookupUsername(accountID),
		RoomID:       int64(room.RoomId),
		RoomName:     room.Name,
		CreatorID:    uint(room.CreatorAccountId),
		CreatorName:  lookupUsername(uint(room.CreatorAccountId)),
		CategoryID:   reportCategory,
		Details:      details,
	})

	writeReportResult(w, true, "")
}

func writeReportResult(w http.ResponseWriter, success bool, message string) {
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(map[string]interface{}{
		"Success": success,
		"Message": message,
	})
}

func lookupUsername(accountID uint) string {
	if accountID == 0 {
		return ""
	}
	var acc models.Account
	if err := db.DB.Select("username").First(&acc, accountID).Error; err != nil {
		return ""
	}
	return acc.Username
}
