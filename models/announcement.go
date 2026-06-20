package models

import "time"

type Announcement struct {
	AnnouncementId   uint      `gorm:"primaryKey;column:announcement_id" json:"AnnouncementId"`
	AnnouncementType int       `gorm:"column:announcement_type" json:"AnnouncementType"`
	Body             string    `gorm:"column:body" json:"Body"`
	CreatedAt        time.Time `gorm:"column:created_at;autoCreateTime" json:"CreatedAt"`
	ImageName        string    `gorm:"column:image_name" json:"ImageName"`
	LinkName         string    `gorm:"column:link_name" json:"LinkName"`
	LinkType         int       `gorm:"column:link_type" json:"LinkType"`
	LinkUri          string    `gorm:"column:link_uri" json:"LinkUri"`
	Platform         int       `gorm:"column:platform;default:-1" json:"Platform"`
	Title            string    `gorm:"column:title" json:"Title"`
}

func (Announcement) TableName() string { return "announcements" }
