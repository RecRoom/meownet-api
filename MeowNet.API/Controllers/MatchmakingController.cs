using MeowNet.API.Data;
using MeowNet.API.Models;
using MeowNet.API.Services;
using Microsoft.AspNetCore.Mvc;
using Microsoft.EntityFrameworkCore;
using System.Security.Claims;

namespace MeowNet.API.Controllers
{
    [ApiController]
    public class MatchmakingController : ControllerBase
    {
        private readonly MeowNetDbContext _context;
        private readonly HubService _hubService;

        public MatchmakingController(MeowNetDbContext context, HubService hubService)
        {
            _context = context;
            _hubService = hubService;
        }

        private long? GetAccountId()
        {
            var idClaim = User.FindFirstValue(ClaimTypes.NameIdentifier) ?? User.FindFirstValue("sub") ?? User.FindFirstValue("nameid");
            if (long.TryParse(idClaim, out var id)) return id;
            return null;
        }

        private bool IsAccountBanned(long accountId)
        {
            return _context.AccountBans.Any(b => b.AccountID == accountId && (b.ExpiresAt == null || b.ExpiresAt > DateTime.UtcNow));
        }

        private object BuildSelfStatus(long accountId, RoomInstance instance, int errorCode = 0)
        {
            var st = _context.PlayerStates.FirstOrDefault(p => p.AccountID == accountId);
            return new
            {
                playerId = accountId,
                statusVisibility = st?.StatusVisibility ?? 0,
                deviceClass = 0,
                vrMovementMode = st?.VrMovementMode ?? 1,
                roomInstance = instance,
                isOnline = true,
                appVersion = "20210827",
                errorCode = errorCode,
                lastOnline = DateTime.UtcNow.ToString("yyyy-MM-ddTHH:mm:ss"),
                clientJoinData = ""
            };
        }

        [HttpGet("goto/none")]
        [HttpPost("goto/none")]
        public IActionResult GotoNone()
        {
            var accountId = GetAccountId() ?? 0;
            Console.WriteLine($"[GOTO] goto/none called for accountId={accountId}");
            _hubService.ClearPlayerInstance(accountId);
            return Ok(BuildSelfStatus(accountId, null, 0));
        }

        [HttpGet("api/quickPlay/v1/getandclear")]
        public IActionResult QuickPlay()
        {
            return Ok();
        }

        [HttpPost("goto/room/{roomParam}/{subRoomParam?}")]
        [HttpGet("goto/room/{roomParam}/{subRoomParam?}")]
        public async Task<IActionResult> GotoRoom(string roomParam, string subRoomParam = "")
        {
            var accountId = GetAccountId();
            Console.WriteLine($"[GOTO] goto/room called: roomParam={roomParam}, subRoomParam={subRoomParam}, accountId={accountId}");
            if (accountId == null) return Unauthorized();

            var roomLower = roomParam.ToLower();
            if (roomLower != "dormroom" && IsAccountBanned(accountId.Value)) return StatusCode(403, "Account banned");

            Room roomData = null;

            if (roomLower == "dormroom")
            {
                roomData = await _context.Rooms.FirstOrDefaultAsync(r => r.CreatorAccountId == accountId.Value && (r.IsDorm || r.Name.ToLower() == "dormroom"));
                if (roomData == null)
                {
                    var acc = await _context.Accounts.FindAsync(accountId.Value);
                    if (acc != null)
                    {
                        await Helpers.AccountHelpers.SetupNewAccountDefaults(_context, acc);
                        roomData = await _context.Rooms.FirstOrDefaultAsync(r => r.CreatorAccountId == accountId.Value && (r.IsDorm || r.Name.ToLower() == "dormroom"));
                    }
                }
            }
            else if (long.TryParse(roomParam, out var roomId))
            {
                roomData = await _context.Rooms.FirstOrDefaultAsync(r => r.RoomId == roomId);
            }
            else
            {
                roomData = await _context.Rooms.FirstOrDefaultAsync(r => r.Name.ToLower() == roomLower);
            }

            if (roomData == null) return NotFound("Room not found");

            return await EnterRoom(accountId.Value, roomData, subRoomParam);
        }

        private async Task<IActionResult> EnterRoom(long accountId, Room roomData, string subRoomParam)
        {
            var query = _context.SubRooms.Where(s => s.RoomId == roomData.RoomId);
            if (!string.IsNullOrEmpty(subRoomParam))
                query = query.Where(s => s.Name.ToLower() == subRoomParam.ToLower());

            var subRoom = await query.FirstOrDefaultAsync();
            if (subRoom == null) return NotFound("Sub-room not found");

            int maxCapacity = subRoom.MaxPlayers > 0 ? subRoom.MaxPlayers : 4;
            
            bool wantPrivate = roomData.IsDorm || Request.Query["JoinMode"] == "2" || Request.Form["JoinMode"] == "2";
            long currentInstanceId = _hubService.GetPlayerInstance(accountId) ?? 0;

            RoomInstance instance = null;

            if (!wantPrivate)
            {
                var candidates = await _context.RoomInstances
                    .Where(i => i.RoomId == roomData.RoomId && !i.IsPrivate && !i.IsInProgress && !i.JoinDisabled && i.Id != currentInstanceId)
                    .OrderBy(i => i.CreatedAt)
                    .ToListAsync();

                foreach (var c in candidates)
                {
                    if (_hubService.LivePlayerCountInInstance(c.Id, accountId) >= maxCapacity) continue;

                    instance = c;
                    if (string.IsNullOrEmpty(instance.PhotonRoomId))
                    {
                        instance.PhotonRoomId = Guid.NewGuid().ToString();
                        _context.RoomInstances.Update(instance);
                        await _context.SaveChangesAsync();
                    }
                    break;
                }
            }

            if (instance == null)
            {
                instance = new RoomInstance
                {
                    OwnerAccountId = (int)accountId,
                    RoomId = roomData.RoomId,
                    SubRoomId = subRoom.SubRoomId,
                    Location = subRoom.UnitySceneId.Trim(),
                    PhotonRegionId = "us",
                    PhotonRoomId = Guid.NewGuid().ToString(),
                    Name = roomData.Name,
                    MaxCapacity = maxCapacity,
                    IsPrivate = wantPrivate,
                    CreatedAt = DateTime.UtcNow,
                    DataBlob = "",
                    RoomCode = ""
                };
                _context.RoomInstances.Add(instance);
                await _context.SaveChangesAsync();
            }

            _hubService.SetPlayerInstance(accountId, instance.Id);

            var visit = await _context.RoomInteractions.FirstOrDefaultAsync(v => v.RoomId == roomData.RoomId && v.AccountId == accountId);
            
            if (roomData.Stats == null) roomData.Stats = new RoomStats();
            
            if (visit == null || !visit.Visited) roomData.Stats.VisitorCount++;
            roomData.Stats.VisitCount++;

            if (visit == null)
            {
                visit = new RoomInteraction { RoomId = roomData.RoomId, AccountId = (uint)accountId, Visited = true };
                _context.RoomInteractions.Add(visit);
            }
            else if (!visit.Visited)
            {
                visit.Visited = true;
                _context.RoomInteractions.Update(visit);
            }
            
            _context.Rooms.Update(roomData);
            await _context.SaveChangesAsync();

            return Ok(BuildSelfStatus(accountId, instance, 0));
        }

        [HttpPost("goto/player/{targetPlayerId}")]
        [HttpGet("goto/player/{targetPlayerId}")]
        public async Task<IActionResult> GotoPlayer(long targetPlayerId)
        {
            var accountId = GetAccountId();
            if (accountId == null) return Unauthorized();

            if (IsAccountBanned(accountId.Value)) return Ok(BuildSelfStatus(accountId.Value, null, 1)); // MMBanned

            var instanceId = _hubService.GetPlayerInstance(targetPlayerId);
            if (instanceId == null || instanceId == 0) return Ok(BuildSelfStatus(accountId.Value, null, 7)); // MMPlayerNotOnline

            var instance = await _context.RoomInstances.FirstOrDefaultAsync(i => i.Id == instanceId.Value);
            if (instance == null) return Ok(BuildSelfStatus(accountId.Value, null, 2)); // MMNoSuchGame

            if (instance.JoinDisabled) return Ok(BuildSelfStatus(accountId.Value, null, 4));

            if (instance.IsPrivate && accountId.Value != instance.OwnerAccountId) return Ok(BuildSelfStatus(accountId.Value, null, 12));

            if (_hubService.LivePlayerCountInInstance(instance.Id, accountId.Value) >= (instance.MaxCapacity > 0 ? instance.MaxCapacity : 4))
                return Ok(BuildSelfStatus(accountId.Value, null, 5)); // MMInsufficientSpace

            _hubService.SetPlayerInstance(accountId.Value, instance.Id);

            return Ok(BuildSelfStatus(accountId.Value, instance, 0));
        }
    }
}
