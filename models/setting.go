package models

type PlayerSetting struct {
	ID        uint    `gorm:"primaryKey" json:"-"`
	AccountID uint    `gorm:"column:account_id;index" json:"-"`
	Account   Account `gorm:"foreignKey:AccountID;references:AccountID;constraint:OnDelete:CASCADE" json:"-"`
	Key       string  `json:"Key"`
	Value     string  `json:"Value"`
}

func (PlayerSetting) TableName() string {
	return "player_settings"
}
