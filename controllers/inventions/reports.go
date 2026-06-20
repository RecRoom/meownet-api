package inventions

import (
	"encoding/json"
	"net/http"

	"meow.net/controllers"
	"meow.net/db"
	"meow.net/discord"
	"meow.net/models"
)

type reportRequest struct {
	InventionId    int64  `json:"InventionId"`
	Details        string `json:"Details"`
	ReportCategory int    `json:"ReportCategory"`
}

func Report(w http.ResponseWriter, r *http.Request) {
	accountID, _ := controllers.AccountIDFromRequest(r)

	var req reportRequest
	if r.Body != nil {
		_ = json.NewDecoder(r.Body).Decode(&req)
	}
	if req.InventionId == 0 {
		writeReportResult(w, false, "invalid invention")
		return
	}

	inv, ok := loadInvention(req.InventionId)
	if !ok {
		writeReportResult(w, false, "invention not found")
		return
	}

	report := models.InventionReport{
		ReporterID:     accountID,
		InventionID:    req.InventionId,
		ReportCategory: req.ReportCategory,
		Details:        req.Details,
	}
	db.DB.Create(&report)

	discord.SendInventionReport(discord.InventionReportInfo{
		ReporterID:    accountID,
		ReporterName:  lookupUsername(accountID),
		InventionID:   inv.InventionId,
		InventionName: inv.Name,
		CreatorID:     uint(inv.CreatorPlayerId),
		CreatorName:   lookupUsername(uint(inv.CreatorPlayerId)),
		CategoryID:    req.ReportCategory,
		Details:       req.Details,
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
