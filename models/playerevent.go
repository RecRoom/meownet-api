package models

import "time"

type PlayerEventAccessibility int

const (
	PlayerEventAccessibilityPrivate  PlayerEventAccessibility = 0
	PlayerEventAccessibilityPublic   PlayerEventAccessibility = 1
	PlayerEventAccessibilityUnlisted PlayerEventAccessibility = 2
)

type PlayerEventResponseType int

const (
	PlayerEventResponseNone       PlayerEventResponseType = -1
	PlayerEventResponseYes        PlayerEventResponseType = 0
	PlayerEventResponseInterested PlayerEventResponseType = 1
	PlayerEventResponseNo         PlayerEventResponseType = 2
	PlayerEventResponsePending    PlayerEventResponseType = 3
)

type PlayerEvent struct {
	PlayerEventId   uint               `gorm:"primaryKey;column:player_event_id;autoIncrement" json:"PlayerEventId"`
	CreatorPlayerId uint               `gorm:"column:creator_player_id;index" json:"CreatorPlayerId"`
	RoomId          int64              `gorm:"column:room_id" json:"RoomId"`
	SubRoomId       *int64             `gorm:"column:sub_room_id" json:"SubRoomId"`
	ClubId          *int64             `gorm:"column:club_id" json:"ClubId"`
	Name            string             `gorm:"column:name" json:"Name"`
	Description     string             `gorm:"column:description" json:"Description"`
	ImageName       *string            `gorm:"column:image_name" json:"ImageName"`
	StartTime       time.Time          `gorm:"column:start_time" json:"StartTime"`
	EndTime         time.Time          `gorm:"column:end_time" json:"EndTime"`
	Accessibility   int                `gorm:"column:accessibility;default:1" json:"Accessibility"`
	State           int                `gorm:"column:state;default:0" json:"State"`
	AttendeeCount   int                `gorm:"column:attendee_count;default:0" json:"AttendeeCount"`
	Tags            []PlayerEventTag   `gorm:"foreignKey:PlayerEventId;references:PlayerEventId;constraint:OnDelete:CASCADE" json:"Tags"`
}

func (PlayerEvent) TableName() string { return "player_events" }

type PlayerEventTag struct {
	Id            uint   `gorm:"primaryKey;autoIncrement" json:"-"`
	PlayerEventId uint   `gorm:"column:player_event_id;index" json:"-"`
	Tag           string `gorm:"column:tag" json:"Tag"`
	Type          int    `gorm:"column:type;default:0" json:"Type"`
}

func (PlayerEventTag) TableName() string { return "player_event_tags" }

type PlayerEventResponse struct {
	PlayerEventResponseId uint      `gorm:"primaryKey;column:player_event_response_id;autoIncrement" json:"PlayerEventResponseId"`
	PlayerEventId         uint      `gorm:"column:player_event_id;index" json:"PlayerEventId"`
	PlayerId              uint      `gorm:"column:player_id;index" json:"PlayerId"`
	Type                  int       `gorm:"column:type" json:"Type"`
	CreatedAt             time.Time `gorm:"column:created_at;autoCreateTime" json:"CreatedAt"`
}

func (PlayerEventResponse) TableName() string { return "player_event_responses" }
