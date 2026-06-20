package clubs

import (
	"encoding/json"
	"net/http"
	"strconv"

	"meow.net/controllers"
	"meow.net/db"
	"meow.net/discord"
	"meow.net/models"
)

func ClubReportCreate(w http.ResponseWriter, r *http.Request) {
	accountID, _ := controllers.AccountIDFromRequest(r)

	r.ParseForm()
	clubId, _ := strconv.ParseInt(r.FormValue("clubId"), 10, 64)
	reportCategory, _ := strconv.Atoi(r.FormValue("reportCategory"))
	details := r.FormValue("details")

	if clubId == 0 {
		writeReportResult(w, false, "invalid club")
		return
	}

	var club models.Club
	if err := db.DB.First(&club, clubId).Error; err != nil {
		writeReportResult(w, false, "club not found")
		return
	}

	report := models.ClubReport{
		ReporterID:     accountID,
		ClubID:         clubId,
		ReportCategory: reportCategory,
		Details:        details,
	}
	db.DB.Create(&report)

	discord.SendClubReport(discord.ClubReportInfo{
		ReporterID:   accountID,
		ReporterName: lookupUsername(accountID),
		ClubID:       club.ClubId,
		ClubName:     club.Name,
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
