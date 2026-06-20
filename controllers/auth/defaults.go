package auth

import (
	"log"
	"math/rand"
	"time"

	"github.com/google/uuid"
	"meow.net/db"
	"meow.net/models"
)

func randomClubId() int64 {
	return rand.Int63n(900000000) + 100000000
}

const (
	dormRoomUnitySceneId = "76d98498-60a1-430c-ab76-b54a29b7a163"
	dormRoomMaxPlayers   = 4
	ownerRole            = 255
)

func setupNewAccountDefaults(account *models.Account) {
	accountId := int(account.AccountID)
	var existingDorm models.Room
	if err := db.DB.Where("creator_account_id = ? AND is_dorm = ?", accountId, true).First(&existingDorm).Error; err == nil {
		log.Printf("[AUTH] dorm already exists for accountId=%d (roomId=%d), skipping dorm setup", accountId, existingDorm.RoomId)
	} else {
		dormRoom := models.Room{
			Name:                "@" + account.Username + "'s Dorm",
			Description:         "Your personal room",
			CreatorAccountId:    accountId,
			ImageName:           "DormRoom.png",
			State:               0,
			Accessibility:       0,
			SupportsLevelVoting: false,
			IsRRO:               false,
			IsDorm:              true,
			CloningAllowed:      false,
			SupportsVRLow:       true,
			SupportsMobile:      true,
			SupportsScreens:     true,
			SupportsWalkVR:      true,
			SupportsTeleportVR:  true,
			SupportsQuest2:      true,
			SupportsJuniors:     true,
			MaxPlayers:          dormRoomMaxPlayers,
			PersistenceVersion:  1,
			UgcVersion:          1,
			WarningMask:         0,
			DisableMicAutoMute:  false,
			CreatedAt:           time.Now(),
		}
		if err := db.DB.Create(&dormRoom).Error; err != nil {
			log.Printf("[AUTH] failed to create dorm room for accountId=%d: %v", accountId, err)
			return
		}

		dormSubRoom := models.SubRoom{
			RoomId:           dormRoom.RoomId,
			Name:             "Home",
			Accessibility:    0,
			MaxPlayers:       dormRoomMaxPlayers,
			SavedByAccountId: accountId,
			UnitySceneId:     dormRoomUnitySceneId,
		}
		if err := db.DB.Create(&dormSubRoom).Error; err != nil {
			log.Printf("[AUTH] failed to create dorm sub room for accountId=%d roomId=%d: %v", accountId, dormRoom.RoomId, err)
		}

		if err := db.DB.Create(&models.RoomRoleEntry{
			RoomId:      dormRoom.RoomId,
			AccountId:   accountId,
			InvitedRole: 0,
			Role:        ownerRole,
		}).Error; err != nil {
			log.Printf("[AUTH] failed to create owner role for accountId=%d roomId=%d: %v", accountId, dormRoom.RoomId, err)
		}
	}

	creatorClubId := randomClubId()
	var existingClub models.Club
	for db.DB.Where("club_id = ?", creatorClubId).First(&existingClub).Error == nil {
		creatorClubId = randomClubId()
	}

	db.DB.Create(&models.Club{
		ClubId:           creatorClubId,
		Name:             uuid.New().String(),
		Description:      "",
		Category:         "CreatorClubs",
		Visibility:       int(models.ClubVisibilityPublic),
		Joinability:      int(models.ClubJoinabilityOpen),
		AllowJuniors:     true,
		MainImageName:    "DefaultImage.png",
		ClubType:         1,
		CreatorAccountId: accountId,
		MemberCount:      1,
	})

	db.DB.Create(&models.ClubMember{
		ClubId:         creatorClubId,
		AccountId:      accountId,
		MembershipType: int(models.ClubMembershipCreator),
	})

	db.DB.Create(&models.ClubPermission{
		ClubId:                 creatorClubId,
		Type:                   int(models.ClubMembershipCreator),
		ApproveMember:          true,
		BanUnban:               true,
		CreateEvent:            true,
		EditDetails:            true,
		EditPermissionSettings: true,
		PostAnnouncement:       true,
	})
	db.DB.Create(&models.ClubPermission{
		ClubId:                 creatorClubId,
		Type:                   int(models.ClubMembershipCoowner),
		ApproveMember:          true,
		BanUnban:               true,
		CreateEvent:            true,
		EditDetails:            true,
		EditPermissionSettings: true,
		PostAnnouncement:       true,
	})
	db.DB.Create(&models.ClubPermission{
		ClubId:                 creatorClubId,
		Type:                   int(models.ClubMembershipModerator),
		ApproveMember:          true,
		BanUnban:               true,
		CreateEvent:            false,
		EditDetails:            false,
		EditPermissionSettings: false,
		PostAnnouncement:       false,
	})
	db.DB.Create(&models.ClubPermission{
		ClubId:                 creatorClubId,
		Type:                   int(models.ClubMembershipMember),
		ApproveMember:          false,
		BanUnban:               false,
		CreateEvent:            false,
		EditDetails:            false,
		EditPermissionSettings: false,
		PostAnnouncement:       false,
	})
}
