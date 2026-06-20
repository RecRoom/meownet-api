package models

import "time"

type Account struct {
	AccountID        uint       `gorm:"primaryKey;column:account_id" json:"accountId"`
	CreatedAt        time.Time  `gorm:"column:created_at" json:"createdAt"`
	DisplayName      string     `gorm:"column:display_name" json:"displayName"`
	IsJunior         *bool      `gorm:"column:is_junior" json:"isJunior"`
	Platforms        int        `gorm:"column:platforms" json:"platforms"`
	ProfileImage     string     `gorm:"column:profile_image" json:"profileImage"`
	Username         string     `gorm:"column:username" json:"username"`
	RawUsername      string     `gorm:"column:raw_username" json:"-"`
	TreatAsJunior    bool       `gorm:"column:treat_as_junior" json:"-"`
	HasBirthday      bool       `gorm:"column:has_birthday" json:"-"`
	IsDeveloper      bool       `gorm:"column:is_developer;default:false" json:"-"`
	IsModerator      bool       `gorm:"column:is_moderator;default:false" json:"-"`
	PasswordHash     *string    `gorm:"column:password_hash" json:"-"`
	HomeClubId       *int64     `gorm:"column:home_club_id" json:"-"`
	SelectedCheer int        `gorm:"column:selected_cheer;default:0" json:"-"`
	LastOnline    *time.Time `gorm:"column:last_online" json:"-"`
	NoToken       bool       `gorm:"column:no_token;default:false" json:"-"`
}

func (Account) TableName() string {
	return "accounts"
}

type PlayerBio struct {
	AccountID uint   `gorm:"primaryKey;column:account_id" json:"accountId"`
	Bio       string `gorm:"column:bio" json:"bio"`
}

func (PlayerBio) TableName() string { return "player_bios" }

func (SelfAccount) TableName() string {
	return "accounts"
}

type SelfAccount struct {
	Account
	AvailableUsernameChanges int     `gorm:"-" json:"availableUsernameChanges"`
	Email                    *string `gorm:"column:email" json:"email"`
	Phone                    *string `gorm:"column:phone" json:"phone"`
	Birthday                 *string `gorm:"column:birthday" json:"birthday"`
	JuniorState              int     `gorm:"column:junior_state" json:"juniorState"`
	ParentAccountID          *uint   `gorm:"column:parent_account_id" json:"parentAccountId"`
}

type PlatformAccount struct {
	ID         uint    `gorm:"primaryKey"`
	AccountID  uint    `gorm:"column:account_id;index"`
	Account    Account `gorm:"foreignKey:AccountID;references:AccountID;constraint:OnDelete:CASCADE"`
	Platform   int     `gorm:"column:platform"`
	PlatformID string  `gorm:"column:platform_id;index"`
}

func (PlatformAccount) TableName() string { return "platform_accounts" }

type PlatformAccountLimit struct {
	Platform    int    `gorm:"primaryKey;column:platform"`
	PlatformID  string `gorm:"primaryKey;column:platform_id"`
	MaxAccounts int    `gorm:"column:max_accounts;default:1"`
}

func (PlatformAccountLimit) TableName() string { return "platform_account_limits" }
