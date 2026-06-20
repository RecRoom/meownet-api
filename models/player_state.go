package models

type PlayerState struct {
	AccountID        uint    `gorm:"primaryKey;column:account_id" json:"-"`
	StatusVisibility int     `gorm:"column:status_visibility;default:0" json:"-"`
	VrMovementMode   int     `gorm:"column:vr_movement_mode;default:1" json:"-"`
	AvoidJuniors     bool    `gorm:"column:avoid_juniors;default:false" json:"-"`
	LoginLockToken   *string `gorm:"column:login_lock_token" json:"-"`
}

func (PlayerState) TableName() string { return "player_states" }
