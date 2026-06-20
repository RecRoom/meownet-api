using Microsoft.EntityFrameworkCore;
using MeowNet.API.Models;

namespace MeowNet.API.Data
{
    public class MeowNetDbContext : DbContext
    {
        public MeowNetDbContext(DbContextOptions<MeowNetDbContext> options)
            : base(options)
        {
        }

        public DbSet<Account> Accounts { get; set; }
        public DbSet<AccountBan> AccountBans { get; set; }
        public DbSet<Announcement> Announcements { get; set; }
        public DbSet<Avatar> Avatars { get; set; }
        public DbSet<AvatarItem> AvatarItems { get; set; }
        public DbSet<Balance> Balances { get; set; }
        public DbSet<Club> Clubs { get; set; }
        public DbSet<ClubAnnouncement> ClubAnnouncements { get; set; }
        public DbSet<ClubCustomTag> ClubCustomTags { get; set; }
        public DbSet<ClubMember> ClubMembers { get; set; }
        public DbSet<ClubPermission> ClubPermissions { get; set; }
        public DbSet<ClubReport> ClubReports { get; set; }
        public DbSet<DailyXpLedger> DailyXpLedgers { get; set; }
        public DbSet<DeviceBan> DeviceBans { get; set; }
        public DbSet<DeviceLogin> DeviceLogins { get; set; }
        public DbSet<FeaturedRoomEntry> FeaturedRoomEntrys { get; set; }
        public DbSet<FeaturedRoomGroup> FeaturedRoomGroups { get; set; }
        public DbSet<FeaturedRoomItem> FeaturedRoomItems { get; set; }
        public DbSet<Gift> Gifts { get; set; }
        public DbSet<GiftDrop> GiftDrops { get; set; }
        public DbSet<InstanceBan> InstanceBans { get; set; }
        public DbSet<InstanceInvite> InstanceInvites { get; set; }
        public DbSet<Invention> Inventions { get; set; }
        public DbSet<InventionCheer> InventionCheers { get; set; }
        public DbSet<InventionOwnership> InventionOwnerships { get; set; }
        public DbSet<InventionReport> InventionReports { get; set; }
        public DbSet<InventionTag> InventionTags { get; set; }
        public DbSet<InventionVersion> InventionVersions { get; set; }
        public DbSet<LeaderboardStat> LeaderboardStats { get; set; }
        public DbSet<MatchmakingResponse> MatchmakingResponses { get; set; }
        public DbSet<Message> Messages { get; set; }
        public DbSet<ModerationBlock> ModerationBlocks { get; set; }
        public DbSet<ModerationReport> ModerationReports { get; set; }
        public DbSet<Objective> Objectives { get; set; }
        public DbSet<ObjectiveGroup> ObjectiveGroups { get; set; }
        public DbSet<PlatformAccount> PlatformAccounts { get; set; }
        public DbSet<PlatformAccountLimit> PlatformAccountLimits { get; set; }
        public DbSet<PlayerBio> PlayerBios { get; set; }
        public DbSet<PlayerCheer> PlayerCheers { get; set; }
        public DbSet<PlayerEvent> PlayerEvents { get; set; }
        public DbSet<PlayerEventResponse> PlayerEventResponses { get; set; }
        public DbSet<PlayerEventTag> PlayerEventTags { get; set; }
        public DbSet<PlayerReport> PlayerReports { get; set; }
        public DbSet<PlayerSetting> PlayerSettings { get; set; }
        public DbSet<PlayerState> PlayerStates { get; set; }
        public DbSet<Progression> Progressions { get; set; }
        public DbSet<RefreshToken> RefreshTokens { get; set; }
        public DbSet<Relationship> Relationships { get; set; }
        public DbSet<RelationshipResponse> RelationshipResponses { get; set; }
        public DbSet<RewardDrop> RewardDrops { get; set; }
        public DbSet<RewardSelection> RewardSelections { get; set; }
        public DbSet<Room> Rooms { get; set; }
        public DbSet<RoomInstance> RoomInstances { get; set; }
        public DbSet<RoomInteraction> RoomInteractions { get; set; }
        public DbSet<RoomPlaylist> RoomPlaylists { get; set; }
        public DbSet<RoomPlaylistEntry> RoomPlaylistEntrys { get; set; }
        public DbSet<RoomReport> RoomReports { get; set; }
        public DbSet<RoomRoleEntry> RoomRoleEntrys { get; set; }
        public DbSet<RoomStats> RoomStatss { get; set; }
        public DbSet<RoomTag> RoomTags { get; set; }
        public DbSet<SavedOutfit> SavedOutfits { get; set; }
        public DbSet<ScreenShareReport> ScreenShareReports { get; set; }

        public DbSet<SubRoom> SubRooms { get; set; }
        public DbSet<SubRoomDataHistory> SubRoomDataHistorys { get; set; }
        public DbSet<UploadedPhoto> UploadedPhotos { get; set; }
        public DbSet<UploadedPhotoCheer> UploadedPhotoCheers { get; set; }
        public DbSet<UserAvatarItem> UserAvatarItems { get; set; }
        public DbSet<UserConsumable> UserConsumables { get; set; }
        public DbSet<UserEquipment> UserEquipments { get; set; }
        public DbSet<UserImage> UserImages { get; set; }
        public DbSet<WishlistItem> WishlistItems { get; set; }

        protected override void OnModelCreating(ModelBuilder modelBuilder)
        {
            base.OnModelCreating(modelBuilder);
            
            foreach (var entityType in modelBuilder.Model.GetEntityTypes())
            {
                foreach (var property in entityType.GetProperties())
                {
                    if (property.ClrType == typeof(bool) || property.ClrType == typeof(bool?))
                    {
                        property.SetColumnType("BOOLEAN");
                    }
                    else if (property.ClrType == typeof(DateTime) || property.ClrType == typeof(DateTime?))
                    {
                        property.SetColumnType("DATETIME");
                    }
                }
            }
        }
    }
}