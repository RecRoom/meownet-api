package models

import "time"

type AccountBan struct {
	ID        uint       `gorm:"primaryKey" json:"id"`
	AccountID uint       `gorm:"index;not null" json:"account_id"`
	Reason    string     `json:"reason"`
	Message   string     `json:"message"`
	IsBan     bool       `gorm:"column:is_ban;default:true" json:"is_ban"`
	BannedBy  string     `json:"banned_by"`
	CreatedAt time.Time  `json:"created_at"`
	ExpiresAt *time.Time `gorm:"index" json:"expires_at"`
}

type DeviceBan struct {
	ID        uint       `gorm:"primaryKey" json:"id"`
	DeviceID  string     `gorm:"uniqueIndex;not null" json:"device_id"`
	AccountID uint       `gorm:"index" json:"account_id"`
	BanID     uint       `gorm:"index" json:"ban_id"`
	Reason    string     `json:"reason"`
	BannedBy  string     `json:"banned_by"`
	CreatedAt time.Time  `json:"created_at"`
	ExpiresAt *time.Time `gorm:"index" json:"expires_at"`
}
