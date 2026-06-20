package routes

import (
	"encoding/json"
	"net/http"
	"strings"

	"meow.net/controllers"
	"meow.net/controllers/account"
	"meow.net/controllers/auth"
	"meow.net/controllers/clubs"
	"meow.net/controllers/hub"
	"meow.net/controllers/inventions"
	"meow.net/controllers/moderation"
	"meow.net/controllers/player"
	"meow.net/controllers/rooms"
	"meow.net/controllers/social"
	"meow.net/controllers/store"
	"meow.net/utils"
)

// trying to start http method validation, i forgot if some are post or put :(

func routeByPathSuffix(suffix string, matched, fallback http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if strings.HasSuffix(r.URL.Path, suffix) {
			matched(w, r)
			return
		}
		fallback(w, r)
	}
}

func RegisterRoutes() {
	// Account API
	http.HandleFunc("GET /account/search", account.AccountSearch)
	http.HandleFunc("/account/create", account.AccountCreate)
	http.HandleFunc("GET /account/bulk", account.AccountBulk)
	http.HandleFunc("/account/me/displayname", account.AccountUpdateDisplayName)
	http.HandleFunc("/account/me/username", account.AccountUpdateUsername)
	http.HandleFunc("/account/me/bio", account.AccountUpdateBio)
	http.HandleFunc("/account/me/birthday", account.AccountUpdateBirthday)
	http.HandleFunc("GET /account/me/haspassword", account.AccountHasPassword)
	http.HandleFunc("/account/me/changepassword", account.AccountChangePassword)
	http.HandleFunc("/account/me/profileimage", account.AccountProfileImage)
	http.HandleFunc("GET /account/me", account.AccountMe)
	http.HandleFunc("/account/", routeByPathSuffix("/bio", account.AccountGetBio, account.AccountGet))
	http.HandleFunc("GET /api/players/v1/me", account.AccountMe)
	http.HandleFunc("GET /api/account/me", account.AccountMe)
	http.HandleFunc("GET /api/accounts/account/me", account.AccountMe)
	http.HandleFunc("GET /parentalcontrol/me", account.ParentalControlMe)

	// Clubs
	http.HandleFunc("GET /club/home/me", clubs.ClubHomeMe)
	http.HandleFunc("GET /club/categoryTags", clubs.ClubCategoryTags)
	http.HandleFunc("/club/create", clubs.ClubCreate)
	http.HandleFunc("GET /club/search", clubs.ClubSearch)
	http.HandleFunc("GET /club/mine/created", clubs.ClubMineCreated)
	http.HandleFunc("GET /club/mine/member", clubs.ClubMineMember)
	http.HandleFunc("/club/", clubs.ClubDispatch)
	http.HandleFunc("/api/clubreporting/v1/report", clubs.ClubReportCreate)

	// Player and Progression API
	http.HandleFunc("/player/login", player.PlayerLogin)
	http.HandleFunc("/player/logout", player.PlayerLogout)
	http.HandleFunc("/player/heartbeat", player.PlayerHeartbeat)
	http.HandleFunc("/player/photonregionpings", player.PlayerPhotonRegionPings)
	http.HandleFunc("/player/statusvisibility", player.PlayerStatusVisibility)
	http.HandleFunc("/player/notifydisconnect", player.NotifyDisconnect)
	http.HandleFunc("GET /player/avoidjuniors", player.PlayerAvoidJuniors)
	http.HandleFunc("GET /player", player.PlayerGet)
	http.HandleFunc("GET /api/players/v1/progression/", player.PlayerProgression)
	http.HandleFunc("GET /api/players/v2/progression/", player.PlayerProgression)
	http.HandleFunc("GET /api/players/v2/progression/bulk", player.PlayerProgressionBulk)
	http.HandleFunc("GET /api/playerReputation/v1/", player.PlayerReputation)
	http.HandleFunc("GET /api/playerReputation/v2/", player.PlayerReputation)
	http.HandleFunc("GET /api/playerReputation/v2/bulk", player.PlayerReputationBulk)
	http.HandleFunc("/api/playersubscriptions/v1/my", social.PlayerSubscriptions)

	// Social and Relationships
	http.HandleFunc("GET /api/relationships/v2/get", social.RelationshipsGet)
	http.HandleFunc("/api/relationships/v2/sendfriendrequest", social.SendFriendRequest)
	http.HandleFunc("/api/relationships/v2/acceptfriendrequest", social.AcceptFriendRequest)
	http.HandleFunc("/api/relationships/v2/removefriend", social.RemoveFriend)
	http.HandleFunc("/api/relationships/v2/addfriend", social.AddFriend)
	http.HandleFunc("/api/relationships/v1/favorite", social.FavoritePlayer)
	http.HandleFunc("/api/relationships/v1/unfavorite", social.UnfavoritePlayer)
	http.HandleFunc("/api/relationships/v1/mute", social.MutePlayer)
	http.HandleFunc("/api/relationships/v1/unmute", social.UnmutePlayer)
	http.HandleFunc("/api/relationships/v1/ignore", social.IgnorePlayer)
	http.HandleFunc("/api/relationships/v1/unignore", social.UnignorePlayer)
	http.HandleFunc("/api/relationships/v1/bulkignoreplatformusers", social.BulkIgnore)
	http.HandleFunc("GET /api/messages/v2/get", social.Messages)
	http.HandleFunc("/api/messages/v2/send", social.SendMessage)
	http.HandleFunc("/api/messages/v3/delete", social.DeleteMessages)
	http.HandleFunc("/invite", social.Invite)
	http.HandleFunc("GET /api/messages/v1/favoriteFriendOnlineStatus", social.FavoriteFriendOnlineStatus)
	http.HandleFunc("/api/sanitize/v1", social.Sanitize)
	http.HandleFunc("GET /thread", social.Thread)
	http.HandleFunc("GET /chat/thread", social.Thread)
	http.HandleFunc("/api/PlayerCheer/v1/create", social.SendCheer)
	http.HandleFunc("/api/PlayerCheer/v1/SetSelectedCheer", social.SetSelectedCheer)

	// Announcements
	http.HandleFunc("/announcements/club/", clubs.ClubAnnouncementsRouter)
	http.HandleFunc("/announcements/", social.AnnouncementsUnread)
	http.HandleFunc("GET /api/announcement/v1/get", social.AnnouncementGet)

	// Subscriptions
	http.HandleFunc("GET /subscription/mine/", social.SubscriptionMine)
	http.HandleFunc("GET /subscription/details/", social.SubscriptionDetails)
	http.HandleFunc("/subscription/", social.SubscriptionDispatch)

	// Auth
	http.HandleFunc("/api/platformlogin/", auth.PlatformLogin)
	http.HandleFunc("/api/platformlogin/logintocachedaccount", auth.LoginToCachedAccount)
	http.HandleFunc("/connect/token", auth.ConnectToken)
	http.HandleFunc("GET /cachedlogin/forplatformid/", auth.CachedLoginForPlatformId)
	http.HandleFunc("/cachedlogin/forplatformids", auth.CachedLoginForPlatformIds)

	// WebSocket Hub
	http.HandleFunc("/hub/v1/negotiate", hub.HubNegotiate)
	http.HandleFunc("GET /hub/v1", hub.HubWebSocket)

	// Moderation
	http.HandleFunc("GET /api/PlayerReporting/v1/moderationBlockDetails", moderation.BlockDetails)
	http.HandleFunc("GET /api/PlayerReporting/v1/voteToKickReasons", moderation.VoteToKickReasons)
	http.HandleFunc("/api/PlayerReporting/v1/instantKick", moderation.InstantKick)
	http.HandleFunc("/api/PlayerReporting/v1/hile", moderation.Hile)
	http.HandleFunc("/api/PlayerReporting/v3/create", moderation.PlayerReportCreate)
	http.HandleFunc("POST /api/PlayerReporting/v1/deviceId", moderation.DeviceIdUpdate)
	http.HandleFunc("/api/thorn/v1/moderation/block", moderation.ThornBlock)
	http.HandleFunc("/api/thorn/", moderation.Thorn)
	http.HandleFunc("/api/screensharereports/v1/report", moderation.ScreenShareReport)
	http.HandleFunc("/api/sanitize/v1/isPure", moderation.SanitizeIsPure)

	// Avatar
	http.HandleFunc("GET /api/avatar/v1/defaultunlocked", player.DefaultUnlocked)
	http.HandleFunc("/api/avatar/v2/gifts/consume/", controllers.GiftsConsume)
	http.HandleFunc("GET /api/avatar/v2/gifts", controllers.GiftsList)
	http.HandleFunc("/api/avatar/v3/gifts/generate", controllers.GiftsGenerate)
	http.HandleFunc("/api/avatar/v2/set", player.AvatarSet)
	http.HandleFunc("/api/avatar/v3/saved/set", player.AvatarSavedSet)
	http.HandleFunc("GET /api/avatar/v3/saved", player.AvatarSaved)
	http.HandleFunc("GET /api/avatar/v2", player.Avatar)
	http.HandleFunc("GET /api/avatar/v4/items", player.AvatarItems)

	// Store and Economy
	http.HandleFunc("GET /api/storefronts/v4/balance/", player.BalanceGet)
	http.HandleFunc("/api/storefronts/v2/buyItem", store.BuyItem)
	http.HandleFunc("/api/storefronts/v2/buyInvention", inventions.BuyInvention)
	http.HandleFunc("GET /api/storefronts/v1/adcarouselitems", store.AdCarouselItems)
	http.HandleFunc("/api/storefronts/", store.StorefrontByType)

	// Inventions
	http.HandleFunc("GET /api/inventions/v1/details", inventions.Details)
	http.HandleFunc("GET /api/inventions/v1/version", inventions.Version)
	http.HandleFunc("GET /api/inventions/v1/personaldetails/", inventions.PersonalDetails)
	http.HandleFunc("GET /api/inventions/v2/search", inventions.Search)
	http.HandleFunc("GET /api/inventions/v2/mine", inventions.Mine)
	http.HandleFunc("GET /api/inventions/v1/toptoday", inventions.TopToday)
	http.HandleFunc("/api/inventions/v1/fromcreators", inventions.FromCreators)
	http.HandleFunc("GET /api/inventions/v2/batch", inventions.Batch)
	http.HandleFunc("/api/inventions/v6/save", inventions.Save)
	http.HandleFunc("/api/inventions/v1/settags", inventions.SetTags)
	http.HandleFunc("/api/inventions/v1/update", inventions.Update)
	http.HandleFunc("/api/inventions/v1/fulllineageowner", inventions.FullLineageOwner)
	http.HandleFunc("/api/inventions/v4/addversion", inventions.AddVersion)
	http.HandleFunc("/api/inventions/v1/delete", inventions.Delete)
	http.HandleFunc("/api/inventions/v1/updateprice", inventions.UpdatePrice)
	http.HandleFunc("/api/inventions/v3/publish", inventions.Publish)
	http.HandleFunc("/api/inventions/v1/report", inventions.Report)

	// Item Wishlists
	http.HandleFunc("/api/itemWishlists/v1/wishlist/", store.WishlistDispatch)
	http.HandleFunc("GET /api/challenge/v2/getCurrent", controllers.CurrentChallenge)

	// Matchmaking and Play
	http.HandleFunc("GET /api/quickPlay/v1/getandclear", player.QuickPlay)
	http.HandleFunc("GET /api/rooms/v1/filters", rooms.RoomFilters)
	http.HandleFunc("/api/rooms/v1/verifyRole", rooms.RoomVerifyRole)
	http.HandleFunc("/api/rooms/v2/report", rooms.RoomReportCreate)
	http.HandleFunc("/api/rooms/", rooms.Rooms)
	http.HandleFunc("/goto/event/", player.GotoEvent)
	http.HandleFunc("/goto/player/", player.GotoPlayer)
	http.HandleFunc("/goto/room/", player.GotoRoom)
	http.HandleFunc("/goto/club/", player.GotoClub)
	http.HandleFunc("/goto/none", player.GotoNone)

	// Rooms
	http.HandleFunc("GET /rooms/bulk", rooms.RoomsBulk)
	http.HandleFunc("GET /rooms/hot", rooms.RoomsHot)
	http.HandleFunc("GET /rooms/search", rooms.RoomsSearch)
	http.HandleFunc("GET /rooms/curated_playlists", rooms.RoomsCuratedPlaylists)
	http.HandleFunc("GET /playlists/{id}", rooms.PlaylistGet)
	http.HandleFunc("GET /rooms/topcreators", rooms.RoomsTopCreators)
	http.HandleFunc("GET /featuredrooms/current", rooms.RoomsFeaturedCurrent)
	http.HandleFunc("GET /rooms/base", rooms.RoomsBase)
	http.HandleFunc("GET /rooms/createdby/me", rooms.RoomCreatedByMe)
	http.HandleFunc("GET /rooms/visitedby/me", rooms.RoomVisitedByMe)
	http.HandleFunc("GET /rooms/favoritedby/me", rooms.RoomFavoritedByMe)
	http.HandleFunc("GET /rooms/moderatedby/me", rooms.RoomModeratedByMe)
	http.HandleFunc("GET /rooms/ownedby/", rooms.RoomsOwnedBy)
	http.HandleFunc("/rooms/", rooms.RoomsDispatch)
	http.HandleFunc("/rooms", rooms.RoomsGet)

	// Room data
	http.HandleFunc("/room/", rooms.RoomData)

	// Room Keys
	http.HandleFunc("/api/roomkeys/", rooms.RoomKeys)

	// Images
	http.HandleFunc("GET /api/images/v2/named", controllers.ImagesNamed)
	http.HandleFunc("/api/images/v4/uploadsaved", controllers.ImagesUploadSaved)
	http.HandleFunc("/api/images/v1/cheer", controllers.ImageCheer)
	http.HandleFunc("GET /api/images/v4/room/", controllers.RoomImages)
	http.HandleFunc("/api/images/", controllers.Images)
	http.HandleFunc("/upload", controllers.Upload)

	// Objectives, Events, Equipment, Consumables
	http.HandleFunc("GET /api/objectives/v1/myprogress", player.Objectives)
	http.HandleFunc("/api/objectives/v1/updateobjective", player.UpdateObjective)
	http.HandleFunc("/api/players/v2/objectives", player.ObjectivesV2)
	http.HandleFunc("/api/playerevents/v1/", player.PlayerEventsV1Dispatch)
	http.HandleFunc("/api/playerevents/v2/", player.PlayerEventsV2Dispatch)
	http.HandleFunc("POST /api/playerevents/v2", player.PlayerEventCreate)
	http.HandleFunc("GET /api/playerevents/v1/club/", clubs.ClubPlayerEvents)
	http.HandleFunc("GET /api/equipment/v2/getUnlocked", store.Equipment)
	http.HandleFunc("/api/equipment/v1/update", store.EquipmentUpdate)
	http.HandleFunc("GET /api/consumables/v2/getUnlocked", player.Consumables)
	http.HandleFunc("/api/consumables/v1/consume", player.ConsumableConsume)

	// Activities
	http.HandleFunc("GET /api/activities/charades/v1/words/Icebreakers", controllers.CharadesIcebreakers)
	http.HandleFunc("GET /api/activities/charades/v1/words/Charades", controllers.CharadesWords)

	// Game Rewards and Community
	http.HandleFunc("GET /api/gamerewards/v1/pending", store.GameRewards)
	http.HandleFunc("/api/gamerewards/v1/request", controllers.GameRewardsRequest)
	http.HandleFunc("POST /api/gamerewards/v1/select", controllers.GameRewardsSelect)
	http.HandleFunc("GET /api/communityboard/v2/current", social.CommunityBoard)

	// Leaderboard
	http.HandleFunc("/leaderboard/GetRanks", controllers.LeaderboardGetRanks)
	http.HandleFunc("/leaderboard/GetPlayerRank", controllers.LeaderboardGetPlayerRank)
	http.HandleFunc("/leaderboard/GetNearbyScores", controllers.LeaderboardGetNearbyScores)
	http.HandleFunc("/leaderboard/CheckAndSetStat", controllers.LeaderboardCheckAndSetStat)

	// Room instance state
	http.HandleFunc("/roominstance/", rooms.RoomInstanceDispatch)

	// Room Currencies
	http.HandleFunc("/api/roomcurrencies/", rooms.RoomCurrencies)

	// Configuration and Settings
	http.HandleFunc("GET /api/gameconfigs/v1/all", controllers.GameConfigs)
	http.HandleFunc("/api/settings/v2/", player.Settings)
	http.HandleFunc("/api/settings/v2/set", player.SettingsSet)
	http.HandleFunc("GET /api/config/v2", controllers.ConfigV2)
	http.HandleFunc("GET /config/LoadingScreenTipData", controllers.LoadingScreenTips)

	// Roles
	http.HandleFunc("GET /role/", account.RoleCheck)

	// Name generation
	http.HandleFunc("GET /namegen/options", account.NamegenOptions)

	// Pageview / analytics
	http.HandleFunc("/pageview/consume", player.PageviewConsume)

	// Campus card
	http.HandleFunc("/api/CampusCard/", store.CampusCard)

	// System and Analytics
	http.HandleFunc("GET /eac/challenge", controllers.EACChallenge)
	http.HandleFunc("POST /anticheat/callback", controllers.AnticheatCallback)
	http.HandleFunc("GET /anticheat/hashes", controllers.AnticheatHashes)
	http.HandleFunc("/anticheat/verifyowner", controllers.AnticheatVerifyOwner)
	versionHandler := utils.JsonHandler(map[string]interface{}{
		"VersionStatus": 0,
	})
	http.HandleFunc("GET /api/versioncheck/", versionHandler)
	http.HandleFunc("GET /api/versioncheck", versionHandler)

	// Admin
	RegisterAdminRoutes()

	// Patcher
	RegisterPatcherRoutes()

	http.HandleFunc("GET /api/config/v1/amplitude", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/plain; charset=utf-8")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"AmplitudeKey":   "",
			"UseRudderStack": false,
			"RudderStackKey": "",
			"UseStatSig":     false,
			"StatSigKey":     "",
		})
	})
}
