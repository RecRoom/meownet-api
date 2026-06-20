package models

import "time"

type InventionPermission int

const (
	InventionPermissionUnassigned        InventionPermission = 0
	InventionPermissionLimitedOneUseOnly InventionPermission = 10
	InventionPermissionUseOnly           InventionPermission = 20
	InventionPermissionEditAndSave       InventionPermission = 40
	InventionPermissionPublish           InventionPermission = 60
	InventionPermissionCharge            InventionPermission = 80
	InventionPermissionUnlimited         InventionPermission = 100
)

type InventionTagType int

const (
	InventionTagCustom InventionTagType = 0
	InventionTagAuto   InventionTagType = 1
)

type Invention struct {
	InventionId              int64     `gorm:"primaryKey;column:invention_id;autoIncrement:false" json:"InventionId"`
	Name                     string    `gorm:"column:name" json:"Name"`
	Description              string    `gorm:"column:description" json:"Description"`
	ImageName                string    `gorm:"column:image_name" json:"ImageName"`
	CreatorPlayerId          int       `gorm:"column:creator_player_id;index" json:"CreatorPlayerId"`
	CreatorPermission        int       `gorm:"column:creator_permission;default:100" json:"CreatorPermission"`
	GeneralPermission        int       `gorm:"column:general_permission;default:100" json:"GeneralPermission"`
	AllowTrial               bool      `gorm:"column:allow_trial;default:true" json:"AllowTrial"`
	HideFromPlayer           bool      `gorm:"column:hide_from_player;default:false" json:"HideFromPlayer"`
	IsAGInvention            bool      `gorm:"column:is_ag_invention;default:false" json:"IsAGInvention"`
	IsCertifiedInvention     bool      `gorm:"column:is_certified_invention;default:false" json:"IsCertifiedInvention"`
	IsPublished              bool      `gorm:"column:is_published;default:false" json:"IsPublished"`
	Price                    int       `gorm:"column:price;default:0" json:"Price"`
	CheerCount               int       `gorm:"column:cheer_count;default:0" json:"CheerCount"`
	NumDownloads             int       `gorm:"column:num_downloads;default:0" json:"NumDownloads"`
	NumPlayersHaveUsedInRoom int       `gorm:"column:num_players_used_in_room;default:0" json:"NumPlayersHaveUsedInRoom"`
	CurrentVersionNumber     int       `gorm:"column:current_version_number;default:1" json:"CurrentVersionNumber"`
	ReplicationId            string    `gorm:"column:replication_id" json:"ReplicationId"`
	CreatedAt                time.Time `gorm:"column:created_at;autoCreateTime" json:"CreatedAt"`
	ModifiedAt               time.Time `gorm:"column:modified_at;autoUpdateTime" json:"ModifiedAt"`
}

func (Invention) TableName() string { return "inventions" }

type InventionVersion struct {
	Id                 uint   `gorm:"primaryKey;column:id;autoIncrement" json:"-"`
	InventionId        int64  `gorm:"column:invention_id;index:idx_inv_version,unique,priority:1" json:"InventionId"`
	VersionNumber      int    `gorm:"column:version_number;index:idx_inv_version,unique,priority:2" json:"VersionNumber"`
	BlobName           string `gorm:"column:blob_name" json:"BlobName"`
	ChipsCost          int    `gorm:"column:chips_cost;default:0" json:"ChipsCost"`
	CloudVariablesCost int    `gorm:"column:cloud_variables_cost;default:0" json:"CloudVariablesCost"`
	InstantiationCost  int    `gorm:"column:instantiation_cost;default:0" json:"InstantiationCost"`
	LightsCost         int    `gorm:"column:lights_cost;default:0" json:"LightsCost"`
	ReplicationId      string `gorm:"column:replication_id" json:"ReplicationId"`
}

func (InventionVersion) TableName() string { return "invention_versions" }

type InventionTag struct {
	Id          uint   `gorm:"primaryKey;column:id;autoIncrement" json:"-"`
	InventionId int64  `gorm:"column:invention_id;index" json:"-"`
	Tag         string `gorm:"column:tag" json:"Tag"`
	Type        int    `gorm:"column:type" json:"Type"`
}

func (InventionTag) TableName() string { return "invention_tags" }

type InventionOwnership struct {
	Id          uint      `gorm:"primaryKey;column:id;autoIncrement" json:"-"`
	InventionId int64     `gorm:"column:invention_id;uniqueIndex:idx_inv_owner" json:"InventionId"`
	AccountId   uint      `gorm:"column:account_id;uniqueIndex:idx_inv_owner;index" json:"AccountId"`
	AcquiredAt  time.Time `gorm:"column:acquired_at;autoCreateTime" json:"AcquiredAt"`
}

func (InventionOwnership) TableName() string { return "invention_ownerships" }

type InventionCheer struct {
	Id          uint      `gorm:"primaryKey;column:id;autoIncrement" json:"-"`
	InventionId int64     `gorm:"column:invention_id;uniqueIndex:idx_inv_cheer" json:"-"`
	AccountId   uint      `gorm:"column:account_id;uniqueIndex:idx_inv_cheer;index" json:"-"`
	CreatedAt   time.Time `gorm:"column:created_at;autoCreateTime" json:"-"`
}

func (InventionCheer) TableName() string { return "invention_cheers" }
