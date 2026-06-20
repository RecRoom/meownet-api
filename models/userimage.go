package models

import "time"

type UserImage struct {
	ID        uint      `gorm:"primaryKey;column:id"`
	AccountID uint      `gorm:"column:account_id;index"`
	ImageName string    `gorm:"column:image_name;uniqueIndex"`
	IsSaved   bool      `gorm:"column:is_saved;default:false"`
	CreatedAt time.Time `gorm:"column:created_at"`
}

func (UserImage) TableName() string { return "user_images" }
