package models

import "time"

type ClubMembershipType int

const (
	ClubMembershipBanned           ClubMembershipType = -1
	ClubMembershipNone             ClubMembershipType = 0
	ClubMembershipPendingRequested ClubMembershipType = 1
	ClubMembershipPendingInvited   ClubMembershipType = 2
	ClubMembershipPendingDenied    ClubMembershipType = 3
	ClubMembershipMember           ClubMembershipType = 10
	ClubMembershipModerator        ClubMembershipType = 20
	ClubMembershipCoowner          ClubMembershipType = 30
	ClubMembershipCreator          ClubMembershipType = 100
)

type ClubVisibility int

const (
	ClubVisibilityPrivate ClubVisibility = 0
	ClubVisibilityPublic  ClubVisibility = 1
)

type ClubJoinability int

const (
	ClubJoinabilityOpen       ClubJoinability = 0
	ClubJoinabilityInviteOnly ClubJoinability = 1
	ClubJoinabilityAskToJoin ClubJoinability = 2
)

type Club struct {
	ClubId           int64     `gorm:"primaryKey;column:club_id;autoIncrement:false" json:"ClubId"`
	Name             string    `gorm:"column:name" json:"Name"`
	Description      string    `gorm:"column:description" json:"Description"`
	Category         string    `gorm:"column:category" json:"Category"`
	Visibility       int       `gorm:"column:visibility;default:1" json:"Visibility"`
	Joinability      int       `gorm:"column:joinability;default:0" json:"Joinability"`
	AllowJuniors     bool      `gorm:"column:allow_juniors;default:true" json:"AllowJuniors"`
	MainImageName    string    `gorm:"column:main_image_name;default:DefaultImgPurple" json:"MainImageName"`
	ClubType         int       `gorm:"column:club_type;default:0" json:"ClubType"`
	ClubhouseRoomId  *int64    `gorm:"column:clubhouse_room_id" json:"ClubhouseRoomId"`
	CreatorAccountId int       `gorm:"column:creator_account_id;index" json:"CreatorAccountId"`
	IsRRO            bool      `gorm:"column:is_rro;default:false" json:"IsRRO"`
	MinLevel         int       `gorm:"column:min_level;default:0" json:"MinLevel"`
	State            int       `gorm:"column:state;default:0" json:"State"`
	MemberCount      int       `gorm:"column:member_count;default:0" json:"MemberCount"`
	CreatedAt        time.Time `gorm:"column:created_at;autoCreateTime" json:"-"`
}

func (Club) TableName() string { return "clubs" }

type ClubMember struct {
	ClubMemberId   int64     `gorm:"primaryKey;column:club_member_id;autoIncrement" json:"ClubMemberId"`
	ClubId         int64     `gorm:"column:club_id;index" json:"ClubId"`
	AccountId      int       `gorm:"column:account_id;index" json:"AccountId"`
	MembershipType int       `gorm:"column:membership_type" json:"MembershipType"`
	CreatedAt      time.Time `gorm:"column:created_at;autoCreateTime" json:"CreatedAt"`
}

func (ClubMember) TableName() string { return "club_members" }

type ClubPermission struct {
	ClubPermissionsId      int64 `gorm:"primaryKey;column:club_permissions_id;autoIncrement" json:"ClubPermissionsId"`
	ClubId                 int64 `gorm:"column:club_id;index" json:"ClubId"`
	Type                   int   `gorm:"column:type" json:"Type"`
	ApproveMember          bool  `gorm:"column:approve_member" json:"ApproveMember"`
	BanUnban               bool  `gorm:"column:ban_unban" json:"BanUnban"`
	CreateEvent            bool  `gorm:"column:create_event" json:"CreateEvent"`
	EditDetails            bool  `gorm:"column:edit_details" json:"EditDetails"`
	EditPermissionSettings bool  `gorm:"column:edit_permission_settings" json:"EditPermissionSettings"`
	PostAnnouncement       bool  `gorm:"column:post_announcement" json:"PostAnnouncement"`
}

func (ClubPermission) TableName() string { return "club_permissions" }

type ClubCustomTag struct {
	Id     int64  `gorm:"primaryKey;column:id;autoIncrement" json:"-"`
	ClubId int64  `gorm:"column:club_id;index" json:"-"`
	Tag    string `gorm:"column:tag" json:"-"`
}

func (ClubCustomTag) TableName() string { return "club_custom_tags" }

type ClubAnnouncement struct {
	AnnouncementId int64     `gorm:"primaryKey;column:announcement_id;autoIncrement" json:"AnnouncementId"`
	ClubId         int64     `gorm:"column:club_id;index" json:"ClubId"`
	AccountId      int       `gorm:"column:account_id" json:"AccountId"`
	Title          string    `gorm:"column:title" json:"Title"`
	Body           string    `gorm:"column:body" json:"Body"`
	ImageName      string    `gorm:"column:image_name" json:"ImageName"`
	Meta           string    `gorm:"column:meta" json:"Meta"`
	CreatedAt      time.Time `gorm:"column:created_at;autoCreateTime" json:"CreatedAt"`
}

func (ClubAnnouncement) TableName() string { return "club_announcements" }
