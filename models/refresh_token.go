package models

import "time"

type RefreshToken struct {
	ID         uint       `gorm:"primaryKey"`
	Token      string     `gorm:"uniqueIndex;not null"`
	AccountID  uint       `gorm:"index;not null"`
	PlatformID string     `gorm:"column:platform_id"`
	Platform   string     `gorm:"column:platform"`
	ExpiresAt  time.Time  `gorm:"not null"`
	UsedAt     *time.Time `gorm:"column:used_at"`
	CreatedAt  time.Time
}

func (RefreshToken) TableName() string { return "refresh_tokens" }
