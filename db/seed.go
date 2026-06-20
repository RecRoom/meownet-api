package db

import (
	"bytes"
	"encoding/csv"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"os"
	"strconv"
	"time"

	"gorm.io/gorm"

	"meow.net/models"
)

func Seed() {
	log.Println("Seeding default data...")

	seedCoachAccount()
	seedRRORooms()
	seedFeaturedRooms()
	seedCuratedPlaylists()
	seedAvatarItems()
	seedRewardDrops()
	resetSequences()

	log.Println("Seeding complete.")
}

func seedRewardDrops() {
	var existing int64
	DB.Model(&models.RewardDrop{}).Count(&existing)
	if existing > 0 {
		log.Printf("[SEED] reward_drops already seeded (%d rows), skipping", existing)
		return
	}

	rows := []models.RewardDrop{
		{GiftDropId: 100001, FriendlyName: "10 Tokens!", Tooltip: "Winner!", Currency: 10, CurrencyType: 2, Rarity: 0},
		{GiftDropId: 100002, FriendlyName: "20 Tokens!", Tooltip: "Winner!", Currency: 20, CurrencyType: 2, Rarity: 0},
		{GiftDropId: 100003, FriendlyName: "500 Tokens!", Tooltip: "Jackpot!", Currency: 500, CurrencyType: 2, Rarity: 20},
	}
	if err := DB.Create(&rows).Error; err != nil {
		log.Printf("[SEED] reward_drops insert error: %v", err)
		return
	}
	log.Printf("[SEED] reward_drops: inserted %d rows", len(rows))
}

func seedAvatarItems() {
	var existing int64
	DB.Model(&models.AvatarItem{}).Count(&existing)
	if existing > 0 {
		log.Printf("[SEED] avatar_items already seeded (%d rows), skipping", existing)
		return
	}

	f, err := os.Open("db/seeds/csv/2021_cleaned_avatar_items.csv")
	if err != nil {
		log.Printf("[SEED] avatar_items: failed to open CSV: %v", err)
		return
	}
	defer f.Close()

	rdr := csv.NewReader(f)
	rdr.FieldsPerRecord = -1

	seen := make(map[string]bool, 2000)
	rows := make([]models.AvatarItem, 0, 2000)
	for {
		rec, err := rdr.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			log.Printf("[SEED] avatar_items: csv parse error: %v", err)
			continue
		}
		if len(rec) < 5 {
			continue
		}
		desc := rec[0]
		if desc == "" || seen[desc] {
			continue
		}
		seen[desc] = true
		itemType, _ := strconv.Atoi(rec[1])
		rarity, _ := strconv.Atoi(rec[4])
		rows = append(rows, models.AvatarItem{
			AvatarItemDesc: desc,
			AvatarItemType: itemType,
			FriendlyName:   rec[2],
			ToolTip:        rec[3],
			Rarity:         rarity,
		})
	}

	if err := DB.CreateInBatches(&rows, 500).Error; err != nil {
		log.Printf("[SEED] avatar_items: insert error: %v", err)
		return
	}
	log.Printf("[SEED] avatar_items: inserted %d rows", len(rows))
}

func seedCoachAccount() {
	var existing models.Account
	if DB.First(&existing, 1).Error == nil {
		return
	}

	account := models.SelfAccount{
		Account: models.Account{
			AccountID:    1,
			Username:     "Coach",
			RawUsername:  "Coach",
			DisplayName:  "Coach",
			ProfileImage: "",
			CreatedAt:    time.Now(),
		},
	}
	if err := DB.Create(&account).Error; err != nil {
		log.Printf("Failed to seed coach account: %v", err)
		return
	}

	var maxRoomId uint = 0
	DB.Model(&models.Room{}).Select("COALESCE(MAX(room_id), 0)").Scan(&maxRoomId)
	dormRoomId := maxRoomId + 1

	dormRoom := models.Room{
		RoomId:             dormRoomId,
		Name:               "DormRoom",
		Description:        "Your personal room",
		CreatorAccountId:   1,
		IsDorm:             true,
		SupportsVRLow:      true,
		SupportsMobile:     true,
		SupportsScreens:    true,
		SupportsWalkVR:     true,
		SupportsTeleportVR: true,
		SupportsJuniors:    true,
		MaxPlayers:         4,
		PersistenceVersion: 1,
		UgcVersion:         1,
		CreatedAt:          time.Now(),
	}
	DB.Create(&dormRoom)

	var maxSubRoomId uint
	DB.Model(&models.SubRoom{}).Select("COALESCE(MAX(sub_room_id), 0)").Scan(&maxSubRoomId)

	DB.Create(&models.SubRoom{
		SubRoomId:        maxSubRoomId + 1,
		RoomId:           dormRoomId,
		Name:             "Home",
		MaxPlayers:       4,
		SavedByAccountId: 1,
		UnitySceneId:     "76d98498-60a1-430c-ab76-b54a29b7a163",
	})

	DB.Create(&models.RoomRoleEntry{
		RoomId:    dormRoomId,
		AccountId: 1,
		Role:      255,
	})

	log.Println("Seeded coach account (ID 1) and dorm")
}

func seedRRORooms() {
	data, err := os.ReadFile("db/seeds/defaultrooms.json")
	if err != nil {
		log.Printf("Failed to read defaultrooms.json: %v", err)
		return
	}

	var rooms []models.Room
	if err := json.NewDecoder(bytes.NewReader(data)).Decode(&rooms); err != nil {
		log.Printf("Failed to parse defaultrooms.json: %v", err)
		return
	}

	var existingIDs []uint
	DB.Model(&models.Room{}).Pluck("room_id", &existingIDs)
	existingSet := make(map[uint]bool, len(existingIDs))
	for _, id := range existingIDs {
		existingSet[id] = true
	}

	var newRooms []models.Room
	var allSubRooms []models.SubRoom
	var allRoles []models.RoomRoleEntry
	var allTags []models.RoomTag
	for _, room := range rooms {
		if existingSet[room.RoomId] {
			continue
		}
		for i := range room.Roles {
			room.Roles[i].RoomId = room.RoomId
		}
		for i := range room.Tags {
			room.Tags[i].RoomId = room.RoomId
		}
		allSubRooms = append(allSubRooms, room.SubRooms...)
		allRoles = append(allRoles, room.Roles...)
		allTags = append(allTags, room.Tags...)
		room.SubRooms = nil
		room.Roles = nil
		room.Tags = nil
		newRooms = append(newRooms, room)
	}

	if len(newRooms) == 0 {
		return
	}

	DB.Create(&newRooms)
	if len(allSubRooms) > 0 {
		DB.Create(&allSubRooms)
	}
	if len(allRoles) > 0 {
		DB.Create(&allRoles)
	}
	if len(allTags) > 0 {
		DB.Create(&allTags)
	}

	log.Printf("Seeded %d rooms", len(newRooms))
}

func seedFeaturedRooms() {
	var existing int64
	DB.Model(&models.FeaturedRoomGroup{}).Count(&existing)
	if existing > 0 {
		log.Printf("[SEED] featured_room_groups already seeded (%d rows), skipping", existing)
		return
	}

	var roomIds []uint
	DB.Model(&models.Room{}).
		Where("is_dorm = ?", false).
		Where("accessibility = ?", int(models.RoomAccessibilityPublic)).
		Order("room_id ASC").
		Limit(10).
		Pluck("room_id", &roomIds)

	if len(roomIds) == 0 {
		log.Printf("[SEED] featured rooms: no eligible rooms, skipping")
		return
	}

	group := models.FeaturedRoomGroup{Name: "Featured Rooms", SortOrder: 0}
	err := DB.Transaction(func(tx *gorm.DB) error {
		if err := tx.Create(&group).Error; err != nil {
			return err
		}
		entries := make([]models.FeaturedRoomEntry, 0, len(roomIds))
		for i, id := range roomIds {
			entries = append(entries, models.FeaturedRoomEntry{
				GroupId:   group.Id,
				RoomId:    id,
				SortOrder: i,
			})
		}
		return tx.Create(&entries).Error
	})
	if err != nil {
		log.Printf("[SEED] featured rooms insert error: %v", err)
		return
	}
	log.Printf("[SEED] featured rooms: created group %d with %d rooms", group.Id, len(roomIds))
}

func seedCuratedPlaylists() {
	var existing int64
	DB.Model(&models.RoomPlaylist{}).Count(&existing)
	if existing > 0 {
		log.Printf("[SEED] room_playlists already seeded (%d rows), skipping", existing)
		return
	}

	var rooms []models.Room
	DB.Where("is_dorm = ?", false).
		Where("accessibility = ?", int(models.RoomAccessibilityPublic)).
		Order("room_id ASC").
		Limit(10).
		Find(&rooms)

	if len(rooms) == 0 {
		log.Printf("[SEED] curated playlists: no eligible rooms, skipping")
		return
	}

	playlists := make([]models.RoomPlaylist, 0, len(rooms))
	for i, rm := range rooms {
		playlists = append(playlists, models.RoomPlaylist{
			Name:                rm.Name,
			Description:         rm.Description,
			ImageName:           rm.ImageName,
			WarningMask:         rm.WarningMask,
			CustomWarning:       rm.CustomWarning,
			CreatorAccountId:    rm.CreatorAccountId,
			State:               rm.State,
			Accessibility:       rm.Accessibility,
			SupportsLevelVoting: rm.SupportsLevelVoting,
			IsRRO:               rm.IsRRO,
			SupportsScreens:     rm.SupportsScreens,
			SupportsWalkVR:      rm.SupportsWalkVR,
			SupportsTeleportVR:  rm.SupportsTeleportVR,
			SupportsVRLow:       rm.SupportsVRLow,
			SupportsQuest2:      rm.SupportsQuest2,
			SupportsMobile:      rm.SupportsMobile,
			SupportsJuniors:     rm.SupportsJuniors,
			MinLevel:            rm.MinLevel,
			CreatedAt:           rm.CreatedAt,
			Stats:               rm.Stats,
			SortOrder:           i,
		})
	}
	if err := DB.Create(&playlists).Error; err != nil {
		log.Printf("[SEED] room_playlists insert error: %v", err)
		return
	}
	log.Printf("[SEED] curated playlists: created %d playlists", len(playlists))
}

func resetSequences() {
	tables := []struct {
		table  string
		column string
	}{
		{"accounts", "account_id"},
		{"rooms", "room_id"},
		{"sub_rooms", "sub_room_id"},
	}
	for _, t := range tables {
		query := fmt.Sprintf(
			"SELECT setval(pg_get_serial_sequence('%s', '%s'), COALESCE((SELECT MAX(%s) FROM %s), 0) + 1, false)",
			t.table, t.column, t.column, t.table,
		)
		DB.Exec(query)
	}
}
