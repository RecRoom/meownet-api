package models

import "time"

type MessageType int

const (
	MessageTypeGameInvite                                   MessageType = 0
	MessageTypeGameInviteDeclined                           MessageType = 1
	MessageTypeGameJoinFailed                               MessageType = 2
	MessageTypePartyActivitySwitch                          MessageType = 3
	MessageTypeFriendInvite                                 MessageType = 4
	MessageTypeVoteToKick                                   MessageType = 5
	MessageTypeGameInviteV2                                 MessageType = 6
	MessageTypePartyActivitySwitchV2                        MessageType = 7
	MessageTypeRequestGameInvite                            MessageType = 10
	MessageTypeRequestGameInviteDeclined                    MessageType = 11
	MessageTypeFriendStatusOnline                           MessageType = 20
	MessageTypeTextMessage                                  MessageType = 30
	MessageTypeFriendRequestAccepted                        MessageType = 40
	MessageTypePlayerCheer                                  MessageType = 50
	MessageTypePlayerCheerAnonymous                         MessageType = 51
	MessageTypeRoomCoOwnerAdded                             MessageType = 60
	MessageTypeRoomCoOwnerRemoved                           MessageType = 61
	MessageTypeRoomCoOwnerInvited                           MessageType = 62
	MessageTypeCreatorPublishedNewRoom                      MessageType = 70
	MessageTypePlayerAttendingEvent                         MessageType = 80
	MessageTypePlayerEventInvitation                        MessageType = 81
	MessageTypeDeprecatedGroupInvitation                    MessageType = 90
	MessageTypeDeprecatedPlayerJoinedGroup                  MessageType = 91
	MessageTypeCoachMessage                                 MessageType = 100
	MessageTypeNewRoomComments                              MessageType = 110
	MessageTypePartyUpRequest                               MessageType = 120
	MessageTypeFriendIntroduction                           MessageType = 130
	MessageTypeClubMemberInvited                            MessageType = 200
	MessageTypeClubModeratorInvited                         MessageType = 201
	MessageTypeClubCoownerInvited                           MessageType = 202
	MessageTypeVirtualClubAnnouncementRoomPublished         MessageType = 100000
	MessageTypeVirtualClubAnnouncementInventionPublished    MessageType = 100001
	MessageTypeVirtualClubAnnouncementGeneric               MessageType = 100002
	MessageTypeVirtualClubAnnouncementPlayerEventPublished  MessageType = 100003
	MessageTypeVirtualClubAnnouncementClub                  MessageType = 100004
	MessageTypeVirtualClubAnnouncementPlayer                MessageType = 100005
	MessageTypeVirtualClubAnnouncementCode                  MessageType = 100006
	MessageTypeVirtualClubAnnouncementPhoto                 MessageType = 100007
	MessageTypeVirtualRoomNotification                      MessageType = 100008
)

type Message struct {
	Id            uint       `gorm:"primaryKey;autoIncrement" json:"Id"`
	FromPlayerId  uint       `gorm:"column:from_player_id;index" json:"FromPlayerId"`
	ToPlayerId    uint       `gorm:"column:to_player_id;index" json:"ToPlayerId"`
	SentTime      time.Time  `gorm:"column:sent_time;autoCreateTime" json:"SentTime"`
	Type          int        `gorm:"column:type" json:"Type"`
	RoomId        *uint      `gorm:"column:room_id" json:"RoomId"`
	InventionId   *uint      `gorm:"column:invention_id" json:"InventionId"`
	PlayerEventId *uint      `gorm:"column:player_event_id" json:"PlayerEventId"`
	Data          string     `gorm:"column:data" json:"Data"`
}

func (Message) TableName() string { return "messages" }
