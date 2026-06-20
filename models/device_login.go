package models

import "time"

type DeviceLogin struct {
	ID          uint      `gorm:"primaryKey" json:"id"`
	AccountID   uint      `gorm:"not null;uniqueIndex:idx_device_account" json:"account_id"`
	DeviceID    string    `gorm:"index;uniqueIndex:idx_device_account" json:"device_id"`
	DeviceClass int       `json:"device_class"`
	PlatformID  string    `gorm:"column:platform_id;index" json:"platform_id"`
	Platform    string    `gorm:"column:platform" json:"platform"`
	IP          string    `gorm:"column:ip" json:"ip"`
	LoginCount  int       `json:"login_count"`
	FirstSeen   time.Time `json:"first_seen"`
	LastSeen    time.Time `json:"last_seen"`
}

func (DeviceLogin) TableName() string { return "device_logins" }
