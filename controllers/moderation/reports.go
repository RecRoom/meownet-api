package moderation

import (
	"encoding/json"
	"net/http"
	"strconv"
	"time"

	"meow.net/controllers"
	"meow.net/controllers/hub"
	"meow.net/db"
	"meow.net/discord"
	"meow.net/models"
)

const DevReportBanDuration = time.Hour

type thornBlockRequest struct {
	PlayerId         uint   `json:"PlayerId"`
	TargetId         uint   `json:"TargetId"`
	ReportedPlayerId uint   `json:"ReportedPlayerId"`
	GameSessionId    int64  `json:"GameSessionId"`
	ReportCategory   int    `json:"ReportCategory"`
	Message          string `json:"Message"`
}

func ThornBlock(w http.ResponseWriter, r *http.Request) {
	currentUserID, _ := controllers.CurrentUserIDFromRequest(r)

	var req thornBlockRequest
	if r.Body != nil {
		_ = json.NewDecoder(r.Body).Decode(&req)
	}

	target := req.TargetId
	if target == 0 {
		target = req.ReportedPlayerId
	}
	if target == 0 {
		target = req.PlayerId
	}

	if currentUserID != 0 && target != 0 {
		report := models.ModerationReport{
			ReporterID:     currentUserID,
			TargetID:       target,
			ReportCategory: req.ReportCategory,
			Message:        req.Message,
			GameSessionID:  req.GameSessionId,
			Source:         "thorn",
		}
		db.DB.Create(&report)
	}

	w.WriteHeader(http.StatusOK)
}

func Thorn(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
}

func Hile(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.Write([]byte("false"))
}

func ScreenShareReport(w http.ResponseWriter, r *http.Request) {
	currentUserID, _ := controllers.CurrentUserIDFromRequest(r)

	r.ParseForm()
	reportedPlayerID, _ := strconv.Atoi(r.FormValue("ReportedPlayerId"))
	roomID, _ := strconv.ParseInt(r.FormValue("RoomId"), 10, 64)
	roomInstanceID, _ := strconv.ParseInt(r.FormValue("RoomInstanceId"), 10, 64)
	roomInstanceType, _ := strconv.Atoi(r.FormValue("RoomInstanceType"))

	if currentUserID != 0 && reportedPlayerID != 0 {
		report := models.ScreenShareReport{
			ReporterID:       currentUserID,
			ReportedPlayerID: uint(reportedPlayerID),
			RoomID:           roomID,
			RoomInstanceID:   roomInstanceID,
			RoomInstanceType: roomInstanceType,
			ImageName:        r.FormValue("ImageName"),
			Details:          r.FormValue("Details"),
		}
		db.DB.Create(&report)
	}

	w.WriteHeader(http.StatusOK)
}

func PlayerReportCreate(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	currentUserID, _ := controllers.CurrentUserIDFromRequest(r)

	r.ParseForm()
	reportedPlayerID, _ := strconv.Atoi(r.FormValue("PlayerIdReported"))
	reportCategory, _ := strconv.Atoi(r.FormValue("ReportCategory"))
	roomID, _ := strconv.ParseInt(r.FormValue("RoomId"), 10, 64)
	roomInstanceType, _ := strconv.Atoi(r.FormValue("RoomInstanceType"))
	heightReporter, _ := strconv.ParseFloat(r.FormValue("HeightReporter"), 64)
	heightReported, _ := strconv.ParseFloat(r.FormValue("HeightReported"), 64)

	if currentUserID != 0 && reportedPlayerID != 0 && (isDeveloper(currentUserID) || isModerator(currentUserID)) {
		if models.ReportCategory(reportCategory) == models.ReportCategoryUnderage {
			markPlayerJuniorFromDevReport(uint(reportedPlayerID))
		} else {
			banPlayerFromDevReport(uint(reportedPlayerID), currentUserID, r.FormValue("Details"))
		}
		json.NewEncoder(w).Encode(map[string]interface{}{
			"Message": "",
			"Success": true,
		})
		return
	}

	if currentUserID != 0 && reportedPlayerID != 0 {
		report := models.PlayerReport{
			ReporterID:       currentUserID,
			ReportedPlayerID: uint(reportedPlayerID),
			ReportCategory:   reportCategory,
			Details:          r.FormValue("Details"),
			HeightReporter:   heightReporter,
			HeightReported:   heightReported,
			RoomID:           roomID,
			RoomInstanceType: roomInstanceType,
		}
		db.DB.Create(&report)

		discord.SendPlayerReport(discord.PlayerReportInfo{
			ReporterID:       currentUserID,
			ReporterName:     lookupUsername(currentUserID),
			ReportedID:       uint(reportedPlayerID),
			ReportedName:     lookupUsername(uint(reportedPlayerID)),
			CategoryID:       reportCategory,
			CategoryName:     reportCategoryName(reportCategory),
			Details:          r.FormValue("Details"),
			HeightReporter:   heightReporter,
			HeightReported:   heightReported,
			RoomID:           roomID,
			RoomInstanceType: roomInstanceType,
		})
	}

	json.NewEncoder(w).Encode(map[string]interface{}{
		"Message": "",
		"Success": true,
	})
}

func isDeveloper(accountID uint) bool {
	if accountID == 0 {
		return false
	}
	var acc models.Account
	if err := db.DB.Select("is_developer", "is_moderator").First(&acc, accountID).Error; err != nil {
		return false
	}
	return acc.IsDeveloper || acc.IsModerator
}

func isModerator(accountID uint) bool {
	if accountID == 0 {
		return false
	}
	var acc models.Account
	if err := db.DB.Select("is_moderator").First(&acc, accountID).Error; err != nil {
		return false
	}
	return acc.IsModerator
}

func banPlayerFromDevReport(reportedID, reporterID uint, reason string) {
	expiresAt := time.Now().Add(DevReportBanDuration)
	ban := models.AccountBan{
		AccountID: reportedID,
		Reason:    reason,
		Message:   reason,
		IsBan:     true,
		BannedBy:  lookupUsername(reporterID),
		ExpiresAt: &expiresAt,
	}
	if err := db.DB.Save(&ban).Error; err != nil {
		return
	}
	hub.HubKickPlayer(int(reportedID))
}

func markPlayerJuniorFromDevReport(reportedID uint) {
	if err := db.DB.Model(&models.Account{}).
		Where("account_id = ?", reportedID).
		Update("treat_as_junior", true).Error; err != nil {
		return
	}
	hub.HubKickPlayer(int(reportedID))
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

func reportCategoryName(c int) string {
	switch models.ReportCategory(c) {
	case models.ReportCategoryModerator:
		return "Moderator"
	case models.ReportCategoryUnknown:
		return "Unknown"
	case models.ReportCategoryHarassment:
		return "Harassment"
	case models.ReportCategoryCheating:
		return "Cheating"
	case models.ReportCategoryAFK:
		return "AFK"
	case models.ReportCategoryMisc:
		return "Misc"
	case models.ReportCategoryUnderage:
		return "Underage"
	case models.ReportCategoryVoteKick:
		return "VoteKick"
	case models.ReportCategoryMisleadingPurchases:
		return "MisleadingPurchases"
	case models.ReportCategoryCoC_Underage:
		return "CoC: Underage"
	case models.ReportCategoryCoC_Sexual:
		return "CoC: Sexual"
	case models.ReportCategoryCoC_Discrimination:
		return "CoC: Discrimination"
	case models.ReportCategoryCoC_Trolling:
		return "CoC: Trolling"
	case models.ReportCategoryCoC_NameOrProfile:
		return "CoC: Name or Profile"
	case models.ReportCategoryIssuingInaccurateReports:
		return "Issuing Inaccurate Reports"
	}
	return ""
}
