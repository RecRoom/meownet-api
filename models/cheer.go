package models

import "time"

type CheerCategory int

const (
	CheerCategoryGeneral          CheerCategory = 0
	CheerCategoryHelpful          CheerCategory = 10
	CheerCategorySportmanship     CheerCategory = 20
	CheerCategoryGreatHost        CheerCategory = 30
	CheerCategoryCreative         CheerCategory = 40
	CheerCategoryRecRoomDeveloper CheerCategory = 9000
)

const (
	CheerDailyCredit = 20
	CheerCost        = 1
)

type PlayerCheer struct {
	Id            uint      `gorm:"primaryKey;column:id;autoIncrement"`
	FromAccountId uint      `gorm:"column:from_account_id;index"`
	ToAccountId   uint      `gorm:"column:to_account_id;index"`
	Category      int       `gorm:"column:category"`
	RoomId        *uint     `gorm:"column:room_id"`
	Anonymous     bool      `gorm:"column:anonymous"`
	CreatedAt     time.Time `gorm:"column:created_at;autoCreateTime;index"`
}

func (PlayerCheer) TableName() string { return "player_cheers" }
