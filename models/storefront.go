package models

type GiftDrop struct {
	GiftDropId                int
	FriendlyName              string
	Tooltip                   string
	AvatarItemDesc            *string
	AvatarItemType            int
	ConsumableItemDesc        *string
	EquipmentPrefabName       *string
	EquipmentModificationGuid *string
	IsQuery                   bool
	Unique                    bool
	SubscribersOnly           bool
	Level                     int
	Rarity                    int
	CurrencyType              int
	Currency                  int
	Context                   int
	ItemSetId                 *int
	ItemSetFriendlyName       *string
}
