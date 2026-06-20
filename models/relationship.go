package models

type RelationshipType int

const (
	RelationshipNone                  RelationshipType = 0
	RelationshipFriendRequestSent     RelationshipType = 1
	RelationshipFriendRequestReceived RelationshipType = 2
	RelationshipFriend                RelationshipType = 3
)

type Relationship struct {
	ID               uint             `gorm:"primaryKey"`
	RequesterID      uint             `gorm:"column:requester_id;uniqueIndex:idx_relationship"`
	TargetID         uint             `gorm:"column:target_id;uniqueIndex:idx_relationship"`
	RelationshipType RelationshipType `gorm:"column:relationship_type;default:0"`

	RequesterFavorited int `gorm:"column:requester_favorited;default:0"`
	RequesterIgnored   int `gorm:"column:requester_ignored;default:0"`
	RequesterMuted     int `gorm:"column:requester_muted;default:0"`

	TargetFavorited int `gorm:"column:target_favorited;default:0"`
	TargetIgnored   int `gorm:"column:target_ignored;default:0"`
	TargetMuted     int `gorm:"column:target_muted;default:0"`
}

func (Relationship) TableName() string { return "relationships" }

type RelationshipResponse struct {
	Favorited        int              `json:"Favorited"`
	Ignored          int              `json:"Ignored"`
	Muted            int              `json:"Muted"`
	PlayerID         uint             `json:"PlayerID"`
	RelationshipType RelationshipType `json:"RelationshipType"`
}
