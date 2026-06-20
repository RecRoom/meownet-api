package models

type SavedOutfit struct {
	AccountID        uint    `gorm:"primaryKey;column:account_id" json:"-"`
	Slot             string  `gorm:"primaryKey;column:slot" json:"Slot"`
	Account          Account `gorm:"foreignKey:AccountID;references:AccountID;constraint:OnDelete:CASCADE" json:"-"`
	PreviewImageName string  `gorm:"column:preview_image_name" json:"PreviewImageName"`
	OutfitSelections string  `gorm:"column:outfit_selections;type:text" json:"OutfitSelections"`
	FaceFeatures     string  `gorm:"column:face_features;type:text" json:"FaceFeatures"`
	SkinColor        string  `gorm:"column:skin_color" json:"SkinColor"`
	HairColor        string  `gorm:"column:hair_color" json:"HairColor"`
}

func (SavedOutfit) TableName() string { return "saved_outfits" }
