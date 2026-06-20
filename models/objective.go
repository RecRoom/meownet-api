package models

import "time"

type ObjectiveGroup struct {
	Id                       uint      `gorm:"primaryKey" json:"-"`
	AccountID                uint      `gorm:"column:account_id;index" json:"-"`
	Account                  Account   `gorm:"foreignKey:AccountID;references:AccountID;constraint:OnDelete:CASCADE" json:"-"`
	ClearedAt                time.Time `gorm:"column:cleared_at" json:"ClearedAt"`
	Group                    int       `gorm:"column:group_index" json:"Group"`
	IsCompleted              bool      `gorm:"column:is_completed" json:"IsCompleted"`
	RequiresCompleteOnServer bool      `gorm:"column:requires_complete_on_server" json:"RequiresCompleteOnServer"`
}

func (ObjectiveGroup) TableName() string { return "objective_groups" }

type Objective struct {
	Id               uint    `gorm:"primaryKey" json:"-"`
	AccountID        uint    `gorm:"column:account_id;index" json:"-"`
	Account          Account `gorm:"foreignKey:AccountID;references:AccountID;constraint:OnDelete:CASCADE" json:"-"`
	Group            int     `gorm:"column:group_index" json:"Group"`
	Type             int     `gorm:"column:type" json:"Type"`
	HasClaimedReward bool    `gorm:"column:has_claimed_reward" json:"HasClaimedReward"`
	Index            int     `gorm:"column:obj_index" json:"Index"`
	IsCompleted      bool    `gorm:"column:is_completed" json:"IsCompleted"`
	IsRewarded       bool    `gorm:"column:is_rewarded" json:"IsRewarded"`
	Progress         float64 `gorm:"column:progress" json:"Progress"`
	VisualProgress   float64 `gorm:"column:visual_progress" json:"VisualProgress"`
}

func (Objective) TableName() string { return "objectives" }
