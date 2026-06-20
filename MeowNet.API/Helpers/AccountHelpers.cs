using System;
using System.Linq;
using System.Threading.Tasks;
using Microsoft.EntityFrameworkCore;
using MeowNet.API.Data;
using MeowNet.API.Models;

namespace MeowNet.API.Helpers
{
    public static class AccountHelpers
    {
        public const string DormRoomUnitySceneId = "76d98498-60a1-430c-ab76-b54a29b7a163";
        public const int DormRoomMaxPlayers = 4;
        public const int OwnerRole = 255;

        private static long RandomClubId()
        {
            return new Random().NextInt64(100000000, 1000000000); // 9 digits
        }

        public static async Task SetupNewAccountDefaults(MeowNetDbContext context, Account account)
        {
            long accountId = account.AccountID;

            var existingDorm = await context.Rooms.FirstOrDefaultAsync(r => r.CreatorAccountId == accountId && r.IsDorm);
            if (existingDorm == null)
            {
                var dormRoom = new Room
                {
                    Name = "@" + account.Username + "'s Dorm",
                    Description = "Your personal room",
                    CreatorAccountId = (int)accountId,
                    ImageName = "DormRoom.png",
                    State = 0,
                    Accessibility = 0,
                    SupportsLevelVoting = false,
                    IsRRO = false,
                    IsDorm = true,
                    CloningAllowed = false,
                    SupportsVRLow = true,
                    SupportsMobile = true,
                    SupportsScreens = true,
                    SupportsWalkVR = true,
                    SupportsTeleportVR = true,
                    SupportsQuest2 = true,
                    SupportsJuniors = true,
                    MaxPlayers = DormRoomMaxPlayers,
                    PersistenceVersion = 1,
                    UgcVersion = 1,
                    WarningMask = 0,
                    DisableMicAutoMute = false,
                    CreatedAt = DateTime.UtcNow,
                    CustomWarning = "",
                    RankedEntityId = "",
                    DataBlob = ""
                };

                context.Rooms.Add(dormRoom);
                await context.SaveChangesAsync(); // generate RoomId

                var dormSubRoom = new SubRoom
                {
                    RoomId = dormRoom.RoomId,
                    Name = "Home",
                    Accessibility = 0,
                    MaxPlayers = DormRoomMaxPlayers,
                    SavedByAccountId = (int)accountId,
                    UnitySceneId = DormRoomUnitySceneId,
                    DataBlob = ""
                };

                context.SubRooms.Add(dormSubRoom);

                var roomRoleEntry = new RoomRoleEntry
                {
                    RoomId = dormRoom.RoomId,
                    AccountId = (int)accountId,
                    InvitedRole = 0,
                    Role = OwnerRole
                };

                context.RoomRoleEntrys.Add(roomRoleEntry);
            }

            long creatorClubId = RandomClubId();
            while (await context.Clubs.AnyAsync(c => c.ClubId == creatorClubId))
            {
                creatorClubId = RandomClubId();
            }

            var club = new Club
            {
                ClubId = creatorClubId,
                Name = Guid.NewGuid().ToString(),
                Description = "",
                Category = "CreatorClubs",
                Visibility = 1, // Public
                Joinability = 1, // Open
                AllowJuniors = true,
                MainImageName = "DefaultImage.png",
                ClubType = 1,
                CreatorAccountId = (int)accountId,
                MemberCount = 1,
                CreatedAt = DateTime.UtcNow
            };
            context.Clubs.Add(club);

            var clubMember = new ClubMember
            {
                ClubId = creatorClubId,
                AccountId = (int)accountId,
                MembershipType = 255 // Creator
            };
            context.ClubMembers.Add(clubMember);

            context.ClubPermissions.AddRange(
                new ClubPermission { ClubId = creatorClubId, Type = 255, ApproveMember = true, BanUnban = true, CreateEvent = true, EditDetails = true, EditPermissionSettings = true, PostAnnouncement = true },
                new ClubPermission { ClubId = creatorClubId, Type = 254, ApproveMember = true, BanUnban = true, CreateEvent = true, EditDetails = true, EditPermissionSettings = true, PostAnnouncement = true },
                new ClubPermission { ClubId = creatorClubId, Type = 1, ApproveMember = true, BanUnban = true, CreateEvent = false, EditDetails = false, EditPermissionSettings = false, PostAnnouncement = false },
                new ClubPermission { ClubId = creatorClubId, Type = 0, ApproveMember = false, BanUnban = false, CreateEvent = false, EditDetails = false, EditPermissionSettings = false, PostAnnouncement = false }
            );

            await context.SaveChangesAsync();
        }
    }
}
