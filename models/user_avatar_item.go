package models

import "time"

type UserAvatarItem struct {
	AccountID      uint      `gorm:"primaryKey;column:account_id"`
	AvatarItemDesc string    `gorm:"primaryKey;column:avatar_item_desc"`
	CreatedAt      time.Time `gorm:"column:created_at"`
}

func (UserAvatarItem) TableName() string { return "user_avatar_items" }
