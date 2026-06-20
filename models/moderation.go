package models

import "time"

type ReportCategory int

const (
	ReportCategoryModerator                   ReportCategory = -1
	ReportCategoryUnknown                     ReportCategory = 0
	ReportCategoryDEPRECATED_MicrophoneAbuse  ReportCategory = 1
	ReportCategoryHarassment                  ReportCategory = 2
	ReportCategoryCheating                    ReportCategory = 3
	ReportCategoryDEPRECATED_ImmatureBehavior ReportCategory = 4
	ReportCategoryAFK                         ReportCategory = 5
	ReportCategoryMisc                        ReportCategory = 6
	ReportCategoryUnderage                    ReportCategory = 7
	ReportCategoryVoteKick                    ReportCategory = 10
	ReportCategoryMisleadingPurchases         ReportCategory = 11
	ReportCategoryCoC_Underage                ReportCategory = 100
	ReportCategoryCoC_Sexual                  ReportCategory = 101
	ReportCategoryCoC_Discrimination          ReportCategory = 102
	ReportCategoryCoC_Trolling                ReportCategory = 103
	ReportCategoryCoC_NameOrProfile           ReportCategory = 104
	ReportCategoryIssuingInaccurateReports    ReportCategory = 1000
)

type ModerationBlock struct {
	ID             uint       `gorm:"primaryKey;autoIncrement" json:"-"`
	AccountID      uint       `gorm:"column:account_id;index" json:"-"`
	ReporterID     *uint      `gorm:"column:reporter_id" json:"PlayerIdReporter"`
	GameSessionID  int64      `gorm:"column:game_session_id" json:"GameSessionId"`
	IsBan          bool       `gorm:"column:is_ban;default:false" json:"IsBan"`
	IsHostKick     bool       `gorm:"column:is_host_kick;default:false" json:"IsHostKick"`
	Message        *string    `gorm:"column:message" json:"Message"`
	ReportCategory int        `gorm:"column:report_category;default:0" json:"ReportCategory"`
	Duration       int        `gorm:"column:duration;default:0" json:"Duration"`
	CreatedAt      time.Time  `gorm:"column:created_at;autoCreateTime" json:"-"`
	ExpiresAt      *time.Time `gorm:"column:expires_at;index" json:"-"`
}

func (ModerationBlock) TableName() string { return "moderation_blocks" }

type ModerationReport struct {
	ID             uint      `gorm:"primaryKey;autoIncrement"`
	ReporterID     uint      `gorm:"column:reporter_id;index"`
	TargetID       uint      `gorm:"column:target_id;index"`
	ReportCategory int       `gorm:"column:report_category"`
	Message        string    `gorm:"column:message"`
	GameSessionID  int64     `gorm:"column:game_session_id"`
	Source         string    `gorm:"column:source"`
	CreatedAt      time.Time `gorm:"column:created_at;autoCreateTime"`
	Resolved       bool      `gorm:"column:resolved;default:false"`
}

func (ModerationReport) TableName() string { return "moderation_reports" }

type ScreenShareReport struct {
	ID               uint      `gorm:"primaryKey;autoIncrement"`
	ReporterID       uint      `gorm:"column:reporter_id;index"`
	ReportedPlayerID uint      `gorm:"column:reported_player_id;index"`
	RoomID           int64     `gorm:"column:room_id"`
	RoomInstanceID   int64     `gorm:"column:room_instance_id"`
	RoomInstanceType int       `gorm:"column:room_instance_type"`
	ImageName        string    `gorm:"column:image_name"`
	Details          string    `gorm:"column:details"`
	CreatedAt        time.Time `gorm:"column:created_at;autoCreateTime"`
	Resolved         bool      `gorm:"column:resolved;default:false"`
}

func (ScreenShareReport) TableName() string { return "screen_share_reports" }

type PlayerReport struct {
	ID               uint      `gorm:"primaryKey;autoIncrement"`
	ReporterID       uint      `gorm:"column:reporter_id;index"`
	ReportedPlayerID uint      `gorm:"column:reported_player_id;index"`
	ReportCategory   int       `gorm:"column:report_category"`
	Details          string    `gorm:"column:details"`
	HeightReporter   float64   `gorm:"column:height_reporter"`
	HeightReported   float64   `gorm:"column:height_reported"`
	RoomID           int64     `gorm:"column:room_id"`
	RoomInstanceType int       `gorm:"column:room_instance_type"`
	CreatedAt        time.Time `gorm:"column:created_at;autoCreateTime"`
	Resolved         bool      `gorm:"column:resolved;default:false"`
}

func (PlayerReport) TableName() string { return "player_reports" }

type InventionReport struct {
	ID             uint      `gorm:"primaryKey;autoIncrement"`
	ReporterID     uint      `gorm:"column:reporter_id;index"`
	InventionID    int64     `gorm:"column:invention_id;index"`
	ReportCategory int       `gorm:"column:report_category"`
	Details        string    `gorm:"column:details"`
	CreatedAt      time.Time `gorm:"column:created_at;autoCreateTime"`
	Resolved       bool      `gorm:"column:resolved;default:false"`
}

func (InventionReport) TableName() string { return "invention_reports" }

type ClubReport struct {
	ID             uint      `gorm:"primaryKey;autoIncrement"`
	ReporterID     uint      `gorm:"column:reporter_id;index"`
	ClubID         int64     `gorm:"column:club_id;index"`
	ReportCategory int       `gorm:"column:report_category"`
	Details        string    `gorm:"column:details"`
	CreatedAt      time.Time `gorm:"column:created_at;autoCreateTime"`
	Resolved       bool      `gorm:"column:resolved;default:false"`
}

func (ClubReport) TableName() string { return "club_reports" }

type RoomReport struct {
	ID             uint      `gorm:"primaryKey;autoIncrement"`
	ReporterID     uint      `gorm:"column:reporter_id;index"`
	RoomID         int64     `gorm:"column:room_id;index"`
	ReportCategory int       `gorm:"column:report_category"`
	Details        string    `gorm:"column:details"`
	CreatedAt      time.Time `gorm:"column:created_at;autoCreateTime"`
	Resolved       bool      `gorm:"column:resolved;default:false"`
}

func (RoomReport) TableName() string { return "room_reports" }

type InstanceBan struct {
	ID         uint      `gorm:"primaryKey;autoIncrement"`
	InstanceID int64     `gorm:"column:instance_id;index:idx_instance_account,priority:1"`
	AccountID  uint      `gorm:"column:account_id;index:idx_instance_account,priority:2"`
	IssuedBy   uint      `gorm:"column:issued_by"`
	ExpiresAt  time.Time `gorm:"column:expires_at;index"`
	CreatedAt  time.Time `gorm:"column:created_at;autoCreateTime"`
}

func (InstanceBan) TableName() string { return "instance_bans" }

type InstanceInvite struct {
	InstanceID int64     `gorm:"column:instance_id;primaryKey;autoIncrement:false"`
	AccountID  int       `gorm:"column:account_id;primaryKey;autoIncrement:false"`
	InvitedBy  int       `gorm:"column:invited_by"`
	CreatedAt  time.Time `gorm:"column:created_at;autoCreateTime"`
}

func (InstanceInvite) TableName() string { return "instance_invites" }
