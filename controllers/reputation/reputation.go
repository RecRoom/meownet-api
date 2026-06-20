package reputation

import (
	"time"

	"meow.net/db"
	"meow.net/models"
)

type Reputation struct {
	AccountId      uint    `json:"AccountId"`
	Noteriety      float64 `json:"Noteriety"`
	IsCheerful     bool    `json:"IsCheerful"`
	CheerGeneral   int     `json:"CheerGeneral"`
	CheerHelpful   int     `json:"CheerHelpful"`
	CheerGreatHost int     `json:"CheerGreatHost"`
	CheerSportsman int     `json:"CheerSportsman"`
	CheerCreative  int     `json:"CheerCreative"`
	CheerCredit    int     `json:"CheerCredit"`
	SelectedCheer  int     `json:"SelectedCheer"`
}

func Build(accountId uint) Reputation {
	rep := Reputation{
		AccountId:   accountId,
		IsCheerful:  true,
		CheerCredit: remainingCredit(accountId),
	}

	var selected int
	db.DB.Model(&models.Account{}).
		Select("selected_cheer").
		Where("account_id = ?", accountId).
		Scan(&selected)
	rep.SelectedCheer = selected

	type row struct {
		Category int
		Count    int
	}
	var rows []row
	db.DB.Model(&models.PlayerCheer{}).
		Select("category, COUNT(*) as count").
		Where("to_account_id = ?", accountId).
		Group("category").
		Scan(&rows)

	for _, r := range rows {
		switch models.CheerCategory(r.Category) {
		case models.CheerCategoryGeneral:
			rep.CheerGeneral = r.Count
		case models.CheerCategoryHelpful:
			rep.CheerHelpful = r.Count
		case models.CheerCategorySportmanship:
			rep.CheerSportsman = r.Count
		case models.CheerCategoryGreatHost:
			rep.CheerGreatHost = r.Count
		case models.CheerCategoryCreative:
			rep.CheerCreative = r.Count
		}
	}

	return rep
}

func remainingCredit(accountId uint) int {
	since := time.Now().UTC().Truncate(24 * time.Hour)
	var sent int64
	db.DB.Model(&models.PlayerCheer{}).
		Where("from_account_id = ? AND created_at >= ?", accountId, since).
		Count(&sent)
	credit := models.CheerDailyCredit - int(sent)*models.CheerCost
	if credit < 0 {
		credit = 0
	}
	return credit
}
