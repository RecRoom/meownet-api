package models

type RewardDrop struct {
	GiftDropId                int     `gorm:"primaryKey;column:gift_drop_id"`
	FriendlyName              string  `gorm:"column:friendly_name"`
	Tooltip                   string  `gorm:"column:tooltip"`
	AvatarItemDesc            *string `gorm:"column:avatar_item_desc"`
	AvatarItemType            int     `gorm:"column:avatar_item_type"`
	ConsumableItemDesc        *string `gorm:"column:consumable_item_desc"`
	EquipmentPrefabName       *string `gorm:"column:equipment_prefab_name"`
	EquipmentModificationGuid *string `gorm:"column:equipment_modification_guid"`
	IsQuery                   bool    `gorm:"column:is_query"`
	Unique                    bool    `gorm:"column:unique"`
	SubscribersOnly           bool    `gorm:"column:subscribers_only"`
	Level                     int     `gorm:"column:level"`
	Rarity                    int     `gorm:"column:rarity"`
	CurrencyType              int     `gorm:"column:currency_type"`
	Currency                  int     `gorm:"column:currency"`
	Context                   int     `gorm:"column:context"`
	ItemSetId                 *int    `gorm:"column:item_set_id"`
	ItemSetFriendlyName       *string `gorm:"column:item_set_friendly_name"`
}

func (RewardDrop) TableName() string { return "reward_drops" }

func (r RewardDrop) ToGiftDrop() GiftDrop {
	return GiftDrop{
		GiftDropId:                r.GiftDropId,
		FriendlyName:              r.FriendlyName,
		Tooltip:                   r.Tooltip,
		AvatarItemDesc:            r.AvatarItemDesc,
		AvatarItemType:            r.AvatarItemType,
		ConsumableItemDesc:        r.ConsumableItemDesc,
		EquipmentPrefabName:       r.EquipmentPrefabName,
		EquipmentModificationGuid: r.EquipmentModificationGuid,
		IsQuery:                   r.IsQuery,
		Unique:                    r.Unique,
		SubscribersOnly:           r.SubscribersOnly,
		Level:                     r.Level,
		Rarity:                    r.Rarity,
		CurrencyType:              r.CurrencyType,
		Currency:                  r.Currency,
		Context:                   r.Context,
		ItemSetId:                 r.ItemSetId,
		ItemSetFriendlyName:       r.ItemSetFriendlyName,
	}
}
