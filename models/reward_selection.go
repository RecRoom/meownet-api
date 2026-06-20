package models

import "time"

type RewardSelection struct {
	ID           uint      `gorm:"primaryKey;column:id" json:"RewardSelectionId"`
	AccountID    uint      `gorm:"column:account_id;index" json:"-"`
	Message      string    `gorm:"column:message" json:"Message"`
	GiftContext  int       `gorm:"column:gift_context" json:"GiftContext"`
	RewardType   int       `gorm:"column:reward_type" json:"RewardType"`
	GiftDrop1Id  int       `gorm:"column:gift_drop_1_id" json:"-"`
	GiftDrop2Id  int       `gorm:"column:gift_drop_2_id" json:"-"`
	GiftDrop3Id  int       `gorm:"column:gift_drop_3_id" json:"-"`
	Consumed     bool      `gorm:"column:consumed;default:false" json:"-"`
	CreatedAt    time.Time `gorm:"column:created_at;autoCreateTime" json:"CreatedAt"`
}

func (RewardSelection) TableName() string { return "reward_selections" }
