package db

import (
	"log"

	"meow.net/models"
)

func Migrate() {
	log.Println("Running database migrations...")

	dropDeprecatedStorefrontTables()

	err := DB.AutoMigrate(
		&models.SelfAccount{},
		&models.PlatformAccount{},
		&models.PlatformAccountLimit{},
		&models.PlayerState{},
		&models.PlayerSetting{},
		&models.Room{},
		&models.SubRoom{},
		&models.SubRoomDataHistory{},
		&models.RoomRoleEntry{},
		&models.RoomTag{},
		&models.RoomPlaylist{},
		&models.RoomPlaylistEntry{},
		&models.FeaturedRoomGroup{},
		&models.FeaturedRoomEntry{},
		&models.RoomInteraction{},
		&models.RoomInstance{},
		&models.ObjectiveGroup{},
		&models.Objective{},
		&models.Avatar{},
		&models.Balance{},
		&models.UserConsumable{},
		&models.UserImage{},
		&models.UploadedPhoto{},
		&models.UploadedPhotoCheer{},
		&models.AvatarItem{},
		&models.UserAvatarItem{},
		&models.PlayerBio{},
		&models.Announcement{},
		&models.Relationship{},
		&models.Message{},
		&models.Progression{},
		&models.DailyXpLedger{},
		&models.ModerationBlock{},
		&models.ModerationReport{},
		&models.ScreenShareReport{},
		&models.PlayerReport{},
		&models.InventionReport{},
		&models.ClubReport{},
		&models.RoomReport{},
		&models.InstanceBan{},
		&models.InstanceInvite{},
		&models.SavedOutfit{},
		&models.LeaderboardStat{},
		&models.RewardSelection{},
		&models.Gift{},
		&models.RewardDrop{},
		&models.Club{},
		&models.ClubMember{},
		&models.ClubPermission{},
		&models.ClubCustomTag{},
		&models.ClubAnnouncement{},
		&models.Invention{},
		&models.InventionVersion{},
		&models.InventionTag{},
		&models.InventionOwnership{},
		&models.InventionCheer{},
		&models.PlayerCheer{},
		&models.UserEquipment{},
		&models.WishlistItem{},
		&models.RefreshToken{},
		&models.AccountBan{},
		&models.DeviceBan{},
		&models.DeviceLogin{},
		&models.PlayerEvent{},
		&models.PlayerEventTag{},
		&models.PlayerEventResponse{},
	)
	if err != nil {
		log.Fatalf("Migration failed: %v", err)
	}

	fixPlatformAccountLimitPK()

	migrateBanIdentity()

	clearStaleRoomInstances()

	log.Println("Database migration completed successfully.")
}

func clearStaleRoomInstances() {
	tables := []string{"instance_invites", "instance_bans", "room_instances"}
	for _, t := range tables {
		if err := DB.Exec("DELETE FROM " + t).Error; err != nil {
			log.Printf("failed to clear stale %s on startup: %v", t, err)
			return
		}
	}
	log.Println("Cleared stale room instances on startup.")
}

func dropDeprecatedStorefrontTables() {
	tables := []string{"item_prices", "gift_drops", "storefront_items", "storefronts"}
	for _, t := range tables {
		if !DB.Migrator().HasTable(t) {
			continue
		}
		if err := DB.Migrator().DropTable(t); err != nil {
			log.Printf("failed to drop deprecated table %s: %v", t, err)
			continue
		}
		log.Printf("dropped deprecated table %s (storefronts now served from db/seeds/storefronts/*.json)", t)
	}
}

func migrateBanIdentity() {
	var pkCols []string
	DB.Raw(`
		SELECT a.attname
		FROM pg_index i
		JOIN pg_attribute a ON a.attrelid = i.indrelid AND a.attnum = ANY(i.indkey)
		WHERE i.indrelid = 'account_bans'::regclass AND i.indisprimary
	`).Scan(&pkCols)

	if len(pkCols) != 1 || pkCols[0] != "id" {
		log.Println("Migrating account_bans PK from account_id to id")
		if err := DB.Exec(`ALTER TABLE account_bans DROP CONSTRAINT IF EXISTS account_bans_pkey`).Error; err != nil {
			log.Fatalf("failed to drop old PK on account_bans: %v", err)
		}
		if err := DB.Exec(`ALTER TABLE account_bans ADD PRIMARY KEY (id)`).Error; err != nil {
			log.Fatalf("failed to add id PK on account_bans: %v", err)
		}
	}

	if err := DB.Exec(`
		UPDATE device_bans b
		SET account_id = dl.account_id
		FROM device_logins dl
		WHERE b.account_id = 0 AND b.device_id = dl.device_id
	`).Error; err != nil {
		log.Printf("failed to backfill device_bans.account_id: %v", err)
	}
}

func fixPlatformAccountLimitPK() {
	var pkCols []string
	DB.Raw(`
		SELECT a.attname
		FROM pg_index i
		JOIN pg_attribute a ON a.attrelid = i.indrelid AND a.attnum = ANY(i.indkey)
		WHERE i.indrelid = 'platform_account_limits'::regclass AND i.indisprimary
	`).Scan(&pkCols)

	if len(pkCols) == 2 {
		return
	}

	log.Println("Migrating platform_account_limits PK to (platform, platform_id)")
	if err := DB.Exec(`UPDATE platform_account_limits SET platform = 0 WHERE platform IS NULL`).Error; err != nil {
		log.Fatalf("failed to backfill platform on platform_account_limits: %v", err)
	}
	if err := DB.Exec(`ALTER TABLE platform_account_limits ALTER COLUMN platform SET NOT NULL`).Error; err != nil {
		log.Fatalf("failed to set platform NOT NULL on platform_account_limits: %v", err)
	}
	if err := DB.Exec(`ALTER TABLE platform_account_limits DROP CONSTRAINT IF EXISTS platform_account_limits_pkey`).Error; err != nil {
		log.Fatalf("failed to drop old PK on platform_account_limits: %v", err)
	}
	if err := DB.Exec(`ALTER TABLE platform_account_limits ADD PRIMARY KEY (platform, platform_id)`).Error; err != nil {
		log.Fatalf("failed to add composite PK on platform_account_limits: %v", err)
	}
}
