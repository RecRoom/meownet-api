package models

type Progression struct {
	AccountID uint    `gorm:"primaryKey;column:account_id" json:"-"`
	Account   Account `gorm:"foreignKey:AccountID;references:AccountID;constraint:OnDelete:CASCADE" json:"-"`
	Level     int     `gorm:"column:level;default:1" json:"Level"`
	XP        int     `gorm:"column:xp;default:0" json:"XP"`
}

func (Progression) TableName() string { return "progressions" }
