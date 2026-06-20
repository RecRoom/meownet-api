package models

import "time"

type UserConsumable struct {
	ID                    uint      `gorm:"primaryKey;column:id"`
	AccountID             uint      `gorm:"column:account_id;index"`
	ConsumableItemDesc    string    `gorm:"column:consumable_item_desc;index"`
	ActiveDurationMinutes int       `gorm:"column:active_duration_minutes"`
	InitialCount          int       `gorm:"column:initial_count;default:0"`
	IsActive              bool      `gorm:"column:is_active;default:false"`
	IsTransferable        bool      `gorm:"column:is_transferable;default:false"`
	CreatedAt             time.Time `gorm:"column:created_at"`
}

func (UserConsumable) TableName() string { return "user_consumables" }
