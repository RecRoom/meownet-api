package models

import "time"

type LeaderboardStat struct {
	ID           uint      `gorm:"primaryKey;column:id"`
	AccountID    uint      `gorm:"column:account_id;uniqueIndex:idx_leaderboard_stat"`
	RoomID       int       `gorm:"column:room_id;uniqueIndex:idx_leaderboard_stat"`
	StatChannel  int       `gorm:"column:stat_channel;uniqueIndex:idx_leaderboard_stat"`
	Score        int       `gorm:"column:score;default:0"`
	UpdatedAt    time.Time `gorm:"column:updated_at;autoUpdateTime"`
}

func (LeaderboardStat) TableName() string { return "leaderboard_stats" }
