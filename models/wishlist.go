package models

import "time"

type WishlistItem struct {
	WishlistItemId    string    `gorm:"primaryKey;column:wishlist_item_id" json:"WishlistItemId"`
	AccountId         int       `gorm:"column:account_id;index" json:"AccountId"`
	PurchasableItemId int       `gorm:"column:purchasable_item_id" json:"PurchasableItemId"`
	CreatedAt         time.Time `gorm:"column:created_at;autoCreateTime" json:"CreatedAt"`
}

func (WishlistItem) TableName() string { return "wishlist_items" }
