package models

import "time"

type SavedImageType int

const (
	SavedImageNone                 SavedImageType = 0
	SavedImageShareCamera          SavedImageType = 1
	SavedImageOutfitThumbnail      SavedImageType = 2
	SavedImageRoomThumbnail        SavedImageType = 3
	SavedImageProfileThumbnail     SavedImageType = 4
	SavedImageInventionThumbnail   SavedImageType = 5
	SavedImagePlayerEventThumbnail SavedImageType = 6
	SavedImageRoomLoadScreen       SavedImageType = 7
)

type UploadedPhoto struct {
	ID             uint           `gorm:"primaryKey;column:id"`
	AccountID      uint           `gorm:"column:account_id;index"`
	ImageName      string         `gorm:"column:image_name;index"`
	PlayerIDs      []int          `gorm:"column:player_ids;serializer:json"`
	SavedImageType SavedImageType `gorm:"column:saved_image_type"`
	RoomID         int            `gorm:"column:room_id;index"`
	PlayerEventID  int            `gorm:"column:player_event_id"`
	Accessibility  int            `gorm:"column:accessibility"`
	CheerCount     int            `gorm:"column:cheer_count;default:0"`
	CreatedAt      time.Time      `gorm:"column:created_at"`
}

func (UploadedPhoto) TableName() string { return "uploaded_photos" }

type UploadedPhotoCheer struct {
	Id        uint      `gorm:"primaryKey;column:id;autoIncrement"`
	PhotoId   uint      `gorm:"column:photo_id;uniqueIndex:idx_photo_cheer"`
	AccountId uint      `gorm:"column:account_id;uniqueIndex:idx_photo_cheer;index"`
	CreatedAt time.Time `gorm:"column:created_at;autoCreateTime"`
}

func (UploadedPhotoCheer) TableName() string { return "uploaded_photo_cheers" }
