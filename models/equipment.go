package models

type UserEquipment struct {
	ID               uint   `gorm:"primaryKey;autoIncrement;column:id"`
	AccountID        uint   `gorm:"column:account_id;index"`
	ModificationGuid string `gorm:"column:modification_guid"`
	PrefabName       string `gorm:"column:prefab_name"`
	FriendlyName     string `gorm:"column:friendly_name"`
	Tooltip          string `gorm:"column:tooltip"`
	Rarity           int    `gorm:"column:rarity;default:-1"`
	Favorited        bool   `gorm:"column:favorited;default:false"`
}

func (UserEquipment) TableName() string { return "user_equipment" }
