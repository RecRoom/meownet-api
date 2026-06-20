package models

type Avatar struct {
	AccountID        uint    `gorm:"primaryKey;column:account_id" json:"-"`
	Account          Account `gorm:"foreignKey:AccountID;references:AccountID;constraint:OnDelete:CASCADE" json:"-"`
	FaceFeatures     string  `gorm:"column:face_features;type:text" json:"FaceFeatures"`
	HairColor        string  `gorm:"column:hair_color" json:"HairColor"`
	OutfitSelections string  `gorm:"column:outfit_selections;type:text" json:"OutfitSelections"`
	SkinColor        string  `gorm:"column:skin_color" json:"SkinColor"`
}

func (Avatar) TableName() string { return "avatars" }
