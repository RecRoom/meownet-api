package models

type NotificationType int

const (
	RelationshipChanged                NotificationType = 1
	MessageReceived                    NotificationType = 2
	MessageDeleted                     NotificationType = 3
	PresenceHeartbeatResponse          NotificationType = 4
	RefreshLogin                       NotificationType = 5
	Logout                             NotificationType = 6
	SubscriptionUpdateProfile          NotificationType = 11
	SubscriptionUpdatePresence         NotificationType = 12
	SubscriptionUpdateGameSession      NotificationType = 13
	SubscriptionUpdateRoom             NotificationType = 15
	SubscriptionUpdateRoomPlaylist     NotificationType = 16
	ModerationQuitGame                 NotificationType = 20
	ModerationUpdateRequired           NotificationType = 21
	ModerationKick                     NotificationType = 22
	ModerationKickAttemptFailed        NotificationType = 23
	ModerationRoomBan                  NotificationType = 24
	ServerMaintenance                  NotificationType = 25
	GiftPackageReceived                NotificationType = 30
	GiftPackageReceivedImmediate       NotificationType = 31
	GiftPackageRewardSelectionReceived NotificationType = 32
	ProfileJuniorStatusUpdate          NotificationType = 40
	RelationshipsInvalid               NotificationType = 50
	StorefrontBalanceAdd               NotificationType = 60
	StorefrontBalanceUpdate            NotificationType = 61
	StorefrontBalancePurchase          NotificationType = 62
	ConsumableMappingAdded             NotificationType = 70
	ConsumableMappingRemoved           NotificationType = 71
	PlayerEventCreated                 NotificationType = 80
	PlayerEventUpdated                 NotificationType = 81
	PlayerEventDeleted                 NotificationType = 82
	PlayerEventResponseChanged         NotificationType = 83
	PlayerEventResponseDeleted         NotificationType = 84
	PlayerEventStateChanged            NotificationType = 85
	ChatMessageReceived                NotificationType = 90
	CommunityBoardUpdate               NotificationType = 95
	CommunityBoardAnnouncementUpdate   NotificationType = 96
	InventionModerationStateChanged    NotificationType = 100
	FreeGiftButtonItemsAdded           NotificationType = 110
	LocalRoomKeyCreated                NotificationType = 120
	LocalRoomKeyDeleted                NotificationType = 121
)
