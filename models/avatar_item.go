package models

type AvatarItem struct {
	AvatarItemDesc string `gorm:"primaryKey;column:avatar_item_desc"`
	AvatarItemType int    `gorm:"column:avatar_item_type;index"`
	FriendlyName   string `gorm:"column:friendly_name"`
	ToolTip        string `gorm:"column:tool_tip"`
	Rarity         int    `gorm:"column:rarity"`
}

func (AvatarItem) TableName() string { return "avatar_items" }
