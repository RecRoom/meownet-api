package models

import "time"

type RoomSceneLocations int

type CloudRegionCode int

const (
	CloudRegionCodeEU   CloudRegionCode = 0
	CloudRegionCodeUS   CloudRegionCode = 1
	CloudRegionCodeAsia CloudRegionCode = 2
	CloudRegionCodeJP   CloudRegionCode = 3
	CloudRegionCodeNone CloudRegionCode = 4
	CloudRegionCodeAU   CloudRegionCode = 5
	CloudRegionCodeUSW  CloudRegionCode = 6
	CloudRegionCodeSA   CloudRegionCode = 7
	CloudRegionCodeCAE  CloudRegionCode = 8
	CloudRegionCodeKR   CloudRegionCode = 9
	CloudRegionCodeIN   CloudRegionCode = 10
	CloudRegionCodeRU   CloudRegionCode = 11
	CloudRegionCodeRUE  CloudRegionCode = 12
)

type RoomRole int

const (
	RoomRoleNone             RoomRole = 0
	RoomRoleBanned           RoomRole = 1
	RoomRoleHost             RoomRole = 10
	RoomRoleModerator        RoomRole = 20
	RoomRoleCoOwner          RoomRole = 30
	RoomRoleTemporaryCoOwner RoomRole = 31
	RoomRoleCreator          RoomRole = 255
)

type RoomAccessibility int

const (
	RoomAccessibilityPrivate  RoomAccessibility = 0
	RoomAccessibilityPublic   RoomAccessibility = 1
	RoomAccessibilityUnlisted RoomAccessibility = 2
)

const (
	RoomWarningNone           = 0
	RoomWarningScary          = 1
	RoomWarningMature         = 2
	RoomWarningFlashingLights = 4
	RoomWarningIntenseMotion  = 8
	RoomWarningViolence       = 16
	RoomWarningCustom         = 32
	RoomWarningReports        = 64
)

type Room struct {
	RoomId                   uint            `gorm:"primaryKey;column:room_id" json:"RoomId"`
	Name                     string          `gorm:"column:name" json:"Name"`
	Description              string          `gorm:"column:description" json:"Description"`
	ImageName                string          `gorm:"column:image_name" json:"ImageName"`
	CreatorAccountId         int             `gorm:"column:creator_account_id" json:"CreatorAccountId"`
	State                    int             `gorm:"column:state" json:"State"`
	Accessibility            int             `gorm:"column:accessibility" json:"Accessibility"`
	AutoLocalizeRoom         bool            `gorm:"column:auto_localize_room" json:"AutoLocalizeRoom"`
	CloningAllowed           bool            `gorm:"column:cloning_allowed" json:"CloningAllowed"`
	CustomWarning            string          `gorm:"column:custom_warning" json:"CustomWarning"`
	DisableMicAutoMute       bool            `gorm:"column:disable_mic_auto_mute" json:"DisableMicAutoMute"`
	DisableRoomComments      bool            `gorm:"column:disable_room_comments" json:"DisableRoomComments"`
	EncryptVoiceChat         bool            `gorm:"column:encrypt_voice_chat" json:"EncryptVoiceChat"`
	IsDeveloperOwned         bool            `gorm:"column:is_developer_owned" json:"IsDeveloperOwned"`
	IsDorm                   bool            `gorm:"column:is_dorm" json:"IsDorm"`
	IsRRO                    bool            `gorm:"column:is_rro" json:"IsRRO"`
	LoadScreenLocked         bool            `gorm:"column:load_screen_locked" json:"LoadScreenLocked"`
	MaxPlayerCalculationMode int             `gorm:"column:max_player_calculation_mode" json:"MaxPlayerCalculationMode"`
	MaxPlayers               int             `gorm:"column:max_players" json:"MaxPlayers"`
	MinLevel                 int             `gorm:"column:min_level" json:"MinLevel"`
	PersistenceVersion       int             `gorm:"column:persistence_version" json:"PersistenceVersion"`
	RankedEntityId           string          `gorm:"column:ranked_entity_id" json:"RankedEntityId"`
	RankingContext           int             `gorm:"column:ranking_context" json:"RankingContext"`
	SupportsJuniors          bool            `gorm:"column:supports_juniors" json:"SupportsJuniors"`
	SupportsLevelVoting      bool            `gorm:"column:supports_level_voting" json:"SupportsLevelVoting"`
	SupportsMobile           bool            `gorm:"column:supports_mobile" json:"SupportsMobile"`
	SupportsQuest2           bool            `gorm:"column:supports_quest_2" json:"SupportsQuest2"`
	SupportsScreens          bool            `gorm:"column:supports_screens" json:"SupportsScreens"`
	SupportsTeleportVR       bool            `gorm:"column:supports_teleport_vr" json:"SupportsTeleportVR"`
	SupportsVRLow            bool            `gorm:"column:supports_vr_low" json:"SupportsVRLow"`
	SupportsWalkVR           bool            `gorm:"column:supports_walk_vr" json:"SupportsWalkVR"`
	ToxmodEnabled            bool            `gorm:"column:toxmod_enabled" json:"ToxmodEnabled"`
	UgcVersion               int             `gorm:"column:ugc_version" json:"UgcVersion"`
	WarningMask              int             `gorm:"column:warning_mask" json:"WarningMask"`
	DataBlob                 *string         `gorm:"column:data_blob" json:"DataBlob"`
	CreatedAt                time.Time       `gorm:"column:created_at" json:"CreatedAt"`
	Stats                    RoomStats       `gorm:"embedded" json:"Stats"`
	SubRooms                 []SubRoom       `gorm:"foreignKey:RoomId;references:RoomId;constraint:OnDelete:CASCADE" json:"SubRooms"`
	Roles                    []RoomRoleEntry `gorm:"foreignKey:RoomId;references:RoomId;constraint:OnDelete:CASCADE" json:"Roles"`
	Tags                     []RoomTag       `gorm:"foreignKey:RoomId;references:RoomId;constraint:OnDelete:CASCADE" json:"Tags"`
	LoadScreens              []interface{}   `gorm:"-" json:"LoadScreens"`
	PromoImages              []interface{}   `gorm:"-" json:"PromoImages"`
	PromoExternalContent     []interface{}   `gorm:"-" json:"PromoExternalContent"`
}

type RoomStats struct {
	CheerCount    int `gorm:"column:cheer_count;default:0" json:"CheerCount"`
	FavoriteCount int `gorm:"column:favorite_count;default:0" json:"FavoriteCount"`
	VisitorCount  int `gorm:"column:visitor_count;default:0" json:"VisitorCount"`
	VisitCount    int `gorm:"column:visit_count;default:0" json:"VisitCount"`
}

type RoomInteraction struct {
	AccountId uint `gorm:"primaryKey;autoIncrement:false"`
	RoomId    uint `gorm:"primaryKey;autoIncrement:false"`
	Cheered   bool
	Favorited bool
	Visited   bool
}

func (Room) TableName() string { return "rooms" }

type SubRoom struct {
	SubRoomId        uint   `gorm:"primaryKey;column:sub_room_id" json:"SubRoomId"`
	RoomId           uint   `gorm:"column:room_id;index" json:"RoomId"`
	Accessibility    int    `gorm:"column:accessibility" json:"Accessibility"`
	DataBlob         string `gorm:"column:data_blob" json:"DataBlob"`
	IsSandbox        bool   `gorm:"column:is_sandbox" json:"IsSandbox"`
	MaxPlayers       int    `gorm:"column:max_players" json:"MaxPlayers"`
	Name             string `gorm:"column:name" json:"Name"`
	SavedByAccountId int    `gorm:"column:saved_by_account_id" json:"SavedByAccountId"`
	UnitySceneId     string `gorm:"column:unity_scene_id" json:"UnitySceneId"`
}

func (SubRoom) TableName() string { return "sub_rooms" }

type SubRoomDataHistory struct {
	Id               uint      `gorm:"primaryKey;column:id" json:"-"`
	SubRoomId        int       `gorm:"column:sub_room_id;index" json:"SubRoomId"`
	DataBlob         string    `gorm:"column:data_blob" json:"DataBlob"`
	SavedByAccountId int       `gorm:"column:saved_by_account_id" json:"SavedByAccountId"`
	CreatedAt        time.Time `gorm:"column:created_at;autoCreateTime" json:"CreatedAt"`
}

func (SubRoomDataHistory) TableName() string { return "sub_room_data_history" }

type RoomRoleEntry struct {
	Id          uint `gorm:"primaryKey" json:"-"`
	RoomId      uint `gorm:"column:room_id;index" json:"-"`
	AccountId   int  `gorm:"column:account_id" json:"AccountId"`
	InvitedRole int  `gorm:"column:invited_role" json:"InvitedRole"`
	Role        int  `gorm:"column:role" json:"Role"`
}

func (RoomRoleEntry) TableName() string { return "room_roles" }

type RoomTag struct {
	Id     uint   `gorm:"primaryKey" json:"-"`
	RoomId uint   `gorm:"column:room_id;index" json:"-"`
	Tag    string `gorm:"column:tag" json:"Tag"`
	Type   int    `gorm:"column:type" json:"Type"`
}

func (RoomTag) TableName() string { return "room_tags" }

type RoomPlaylist struct {
	Name                string              `gorm:"column:name" json:"Name"`
	Description         string              `gorm:"column:description" json:"Description"`
	ImageName           string              `gorm:"column:image_name" json:"ImageName"`
	WarningMask         int                 `gorm:"column:warning_mask" json:"WarningMask"`
	CustomWarning       string              `gorm:"column:custom_warning" json:"CustomWarning"`
	CreatorAccountId    int                 `gorm:"column:creator_account_id" json:"CreatorAccountId"`
	State               int                 `gorm:"column:state" json:"State"`
	Accessibility       int                 `gorm:"column:accessibility" json:"Accessibility"`
	SupportsLevelVoting bool                `gorm:"column:supports_level_voting" json:"SupportsLevelVoting"`
	IsRRO               bool                `gorm:"column:is_rro" json:"IsRRO"`
	SupportsScreens     bool                `gorm:"column:supports_screens" json:"SupportsScreens"`
	SupportsWalkVR      bool                `gorm:"column:supports_walk_vr" json:"SupportsWalkVR"`
	SupportsTeleportVR  bool                `gorm:"column:supports_teleport_vr" json:"SupportsTeleportVR"`
	SupportsVRLow       bool                `gorm:"column:supports_vr_low" json:"SupportsVRLow"`
	SupportsQuest2      bool                `gorm:"column:supports_quest_2" json:"SupportsQuest2"`
	SupportsMobile      bool                `gorm:"column:supports_mobile" json:"SupportsMobile"`
	SupportsJuniors     bool                `gorm:"column:supports_juniors" json:"SupportsJuniors"`
	MinLevel            int                 `gorm:"column:min_level" json:"MinLevel"`
	CreatedAt           time.Time           `gorm:"column:created_at" json:"CreatedAt"`
	Stats               RoomStats           `gorm:"embedded" json:"Stats"`
	PlaylistId          uint                `gorm:"primaryKey;column:id" json:"PlaylistId"`
	SortOrder           int                 `gorm:"column:sort_order;default:0" json:"-"`
	Entries             []RoomPlaylistEntry `gorm:"foreignKey:PlaylistId;references:PlaylistId;constraint:OnDelete:CASCADE" json:"-"`
}

func (RoomPlaylist) TableName() string { return "room_playlists" }

type RoomPlaylistEntry struct {
	Id         uint `gorm:"primaryKey;column:id" json:"-"`
	PlaylistId uint `gorm:"column:playlist_id;index" json:"-"`
	RoomId     uint `gorm:"column:room_id" json:"RoomId"`
	SortOrder  int  `gorm:"column:sort_order;default:0" json:"-"`
}

func (RoomPlaylistEntry) TableName() string { return "room_playlist_entries" }

type FeaturedRoomGroup struct {
	Id        uint                `gorm:"primaryKey;column:id" json:"FeaturedRoomGroupId"`
	Name      string              `gorm:"column:name" json:"Name"`
	SortOrder int                 `gorm:"column:sort_order;default:0" json:"-"`
	Entries   []FeaturedRoomEntry `gorm:"foreignKey:GroupId;references:Id;constraint:OnDelete:CASCADE" json:"-"`
	Rooms     []FeaturedRoomItem  `gorm:"-" json:"Rooms"`
}

func (FeaturedRoomGroup) TableName() string { return "featured_room_groups" }

type FeaturedRoomEntry struct {
	Id        uint `gorm:"primaryKey;column:id" json:"-"`
	GroupId   uint `gorm:"column:group_id;index" json:"-"`
	RoomId    uint `gorm:"column:room_id" json:"-"`
	SortOrder int  `gorm:"column:sort_order;default:0" json:"-"`
}

func (FeaturedRoomEntry) TableName() string { return "featured_room_entries" }

type FeaturedRoomItem struct {
	RoomId    uint   `json:"RoomId"`
	RoomName  string `json:"RoomName"`
	ImageName string `json:"ImageName"`
}

type RoomInstance struct {
	Id                int64     `gorm:"primaryKey" json:"roomInstanceId"`
	OwnerAccountId    int       `gorm:"column:owner_account_id" json:"-"`
	RoomId            int64     `gorm:"column:room_id;index" json:"roomId"`
	SubRoomId         int64     `gorm:"column:sub_room_id" json:"subRoomId"`
	Location          string    `gorm:"column:location" json:"location"`
	DataBlob          string    `gorm:"column:data_blob" json:"-"`
	EventId           int64     `gorm:"column:event_id" json:"eventId"`
	PhotonRegionId    string    `gorm:"column:photon_region_id" json:"photonRegionId"`
	PhotonRoomId      string    `gorm:"column:photon_room_id" json:"photonRoomId"`
	Name              string    `gorm:"column:name" json:"name"`
	MaxCapacity       int       `gorm:"column:max_capacity" json:"maxCapacity"`
	IsFull            bool      `gorm:"column:is_full" json:"isFull"`
	IsPrivate         bool      `gorm:"column:is_private" json:"isPrivate"`
	IsInProgress      bool      `gorm:"column:is_in_progress" json:"isInProgress"`
	RoomCode          string    `gorm:"column:room_code" json:"roomCode"`
	RoomInstanceType  int       `gorm:"column:room_instance_type" json:"roomInstanceType"`
	ClubId            int64     `gorm:"column:club_id" json:"clubId"`
	EncryptVoiceChat  bool      `gorm:"column:encrypt_voice_chat" json:"EncryptVoiceChat"`
	MatchmakingPolicy int       `gorm:"column:matchmaking_policy" json:"matchmakingPolicy"`
	AllowNewUsers     bool      `gorm:"column:allow_new_users;default:true" json:"-"`
	JoinDisabled      bool      `gorm:"column:join_disabled;default:false" json:"-"`
	CreatedAt         time.Time `gorm:"column:created_at;autoCreateTime" json:"createdAt"`
}

func (RoomInstance) TableName() string { return "room_instances" }

type MatchmakingResponse struct {
	ErrorCode    int          `json:"ErrorCode"`
	RoomInstance RoomInstance `json:"RoomInstance"`
}

type MatchmakingErrorCode int

const (
	MMUnknownError                        MatchmakingErrorCode = -1
	MMSuccess                             MatchmakingErrorCode = 0
	MMNoSuchGame                          MatchmakingErrorCode = 1
	MMPlayerNotOnline                     MatchmakingErrorCode = 2
	MMInsufficientSpace                   MatchmakingErrorCode = 3
	MMEventNotStarted                     MatchmakingErrorCode = 4
	MMEventAlreadyFinished                MatchmakingErrorCode = 5
	MMBlockedFromRoom                     MatchmakingErrorCode = 7
	MMJuniorNotAllowed                    MatchmakingErrorCode = 11
	MMBanned                              MatchmakingErrorCode = 12
	MMAlreadyInBestInstance               MatchmakingErrorCode = 13
	MMInsufficientRelationship            MatchmakingErrorCode = 14
	MMUpdateRequired                      MatchmakingErrorCode = 16
	MMAlreadyInTargetInstance             MatchmakingErrorCode = 17
	MMUGCNotAllowed                       MatchmakingErrorCode = 19
	MMNoSuchRoom                          MatchmakingErrorCode = 20
	MMRoomIsNotActive                     MatchmakingErrorCode = 22
	MMRoomBlockedByCreator                MatchmakingErrorCode = 23
	MMRoomIsPrivate                       MatchmakingErrorCode = 25
	MMRoomInstanceIsPrivate               MatchmakingErrorCode = 26
	MMDeviceClassNotSupported             MatchmakingErrorCode = 30
	MMDeviceClassNotSupportedByRoomOwner  MatchmakingErrorCode = 31
	MMMovementModeNotSupportedByRoomOwner MatchmakingErrorCode = 32
	MMEventIsPrivate                      MatchmakingErrorCode = 35
	MMRoomInviteExpired                   MatchmakingErrorCode = 40
	MMNoAvailableRegion                   MatchmakingErrorCode = 45
	MMNotorietyTooPoor                    MatchmakingErrorCode = 50
	MMBannedFromRoom                      MatchmakingErrorCode = 55
	MMNoSuchRoomPlaylist                  MatchmakingErrorCode = 60
	MMRoomPlaylistIsNotActive             MatchmakingErrorCode = 61
	MMRoomPlaylistIsPrivate               MatchmakingErrorCode = 62
	MMNoSuchClub                          MatchmakingErrorCode = 70
	MMClubHasNoClubhouse                  MatchmakingErrorCode = 71
	MMClubIsNotActive                     MatchmakingErrorCode = 73
	MMNotAMemberOfClub                    MatchmakingErrorCode = 74
	MMBannedFromClub                      MatchmakingErrorCode = 75
	MMInstanceJoinNotPermitted            MatchmakingErrorCode = 76
	MMLevelTooLow                         MatchmakingErrorCode = 77
)
