package models

import "time"

type DailyXpLedger struct {
	ID        uint      `gorm:"primaryKey;column:id"`
	AccountID uint      `gorm:"column:account_id;uniqueIndex:idx_daily_xp_account_day,priority:1"`
	Day       time.Time `gorm:"column:day;type:date;uniqueIndex:idx_daily_xp_account_day,priority:2"`
	Xp        int       `gorm:"column:xp;default:0"`
	UpdatedAt time.Time `gorm:"column:updated_at;autoUpdateTime"`
}

func (DailyXpLedger) TableName() string { return "daily_xp_ledgers" }
