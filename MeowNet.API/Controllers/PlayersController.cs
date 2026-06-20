using MeowNet.API.Data;
using MeowNet.API.Models;
using Microsoft.AspNetCore.Mvc;
using Microsoft.EntityFrameworkCore;
using System.Security.Claims;
using System.Text.Json;

namespace MeowNet.API.Controllers
{
    [ApiController]
    [Route("player")]
    public class PlayersController : ControllerBase
    {
        private readonly MeowNetDbContext _context;

        public PlayersController(MeowNetDbContext context)
        {
            _context = context;
        }

        private long? GetAccountId()
        {
            var authHeader = Request.Headers.Authorization.ToString();
            if (string.IsNullOrEmpty(authHeader) || !authHeader.StartsWith("Bearer ", StringComparison.OrdinalIgnoreCase))
                return null;

            var idClaim = User.FindFirstValue(ClaimTypes.NameIdentifier) 
                          ?? User.FindFirstValue("sub") 
                          ?? User.FindFirstValue("nameid");
                          
            if (idClaim != null && long.TryParse(idClaim, out var id)) return id;
            return null;
        }

        [HttpGet("")]
        public async Task<IActionResult> PlayerGet([FromQuery] List<long> id, [FromServices] MeowNet.API.Services.HubService hubService)
        {
            var viewerId = GetAccountId();
            if (viewerId == null) return Unauthorized();

            var results = new List<object>();

            // For fetching offline last_online time
            var lastOnlineData = await _context.Accounts
                .Where(a => id.Contains(a.AccountID))
                .ToDictionaryAsync(a => a.AccountID, a => a.LastOnline);

            // Fetch player states
            var playerStates = await _context.PlayerStates
                .Where(p => id.Contains(p.AccountID))
                .ToDictionaryAsync(p => p.AccountID);

            foreach (var i in id)
            {
                bool isOnline = hubService.IsOnline(i);
                string lastOnline = isOnline 
                    ? DateTime.UtcNow.ToString("yyyy-MM-ddTHH:mm:ss")
                    : (lastOnlineData.ContainsKey(i) && lastOnlineData[i].HasValue 
                        ? lastOnlineData[i].Value.ToString("yyyy-MM-ddTHH:mm:ss")
                        : "0001-01-01T00:00:00");

                RoomInstance instance = null;
                if (isOnline)
                {
                    var instanceId = hubService.GetPlayerInstance(i);
                    if (instanceId.HasValue && instanceId.Value > 0)
                    {
                        instance = await _context.RoomInstances.FirstOrDefaultAsync(r => r.Id == instanceId.Value);
                    }
                }

                playerStates.TryGetValue(i, out var st);

                results.Add(new { 
                    playerId = i, 
                    isOnline = isOnline, 
                    roomInstance = instance,
                    lastOnline = lastOnline,
                    appVersion = "20210827",
                    statusVisibility = st?.StatusVisibility ?? 0,
                    vrMovementMode = st?.VrMovementMode ?? 1,
                    deviceClass = 0,
                    errorCode = 0,
                    clientJoinData = ""
                });
            }
            return Ok(results);
        }

        [HttpPost("login")]
        public async Task<IActionResult> PlayerLogin()
        {
            using var reader = new StreamReader(Request.Body);
            var token = await reader.ReadToEndAsync();
            var accountId = GetAccountId();
            if (accountId == null) return Content("0");

            token = token?.Trim('"')?.Trim();
            if (string.IsNullOrEmpty(token)) return Content("0");

            var st = await _context.PlayerStates.FirstOrDefaultAsync(p => p.AccountID == accountId);

            if (st == null)
            {
                st = new PlayerState { AccountID = accountId.Value, LoginLockToken = token };
                _context.PlayerStates.Add(st);
            }
            else
            {
                st.LoginLockToken = token;
                _context.PlayerStates.Update(st);
            }

            await _context.SaveChangesAsync();
            return Content("0");
        }

        [HttpPost("logout")]
        public async Task<IActionResult> PlayerLogout()
        {
            var accountId = GetAccountId();
            if (accountId != null)
            {
                var st = await _context.PlayerStates.FirstOrDefaultAsync(p => p.AccountID == accountId);
                if (st != null)
                {
                    st.LoginLockToken = null;
                    _context.PlayerStates.Update(st);
                    await _context.SaveChangesAsync();
                }
            }
            return Ok();
        }

        [HttpGet("heartbeat")]
        [HttpPost("heartbeat")]
        [HttpPut("heartbeat")]
        public IActionResult PlayerHeartbeat([FromServices] MeowNet.API.Services.HubService hubService)
        {
            var accountId = GetAccountId();
            if (accountId == null) return Unauthorized();

            var instanceId = hubService.GetPlayerInstance(accountId.Value);
            RoomInstance instance = null;
            if (instanceId.HasValue && instanceId.Value > 0)
            {
                instance = _context.RoomInstances.FirstOrDefault(i => i.Id == instanceId.Value);
            }

            var st = _context.PlayerStates.FirstOrDefault(p => p.AccountID == accountId.Value);

            return Ok(new
            {
                playerId = accountId,
                statusVisibility = st?.StatusVisibility ?? 0,
                deviceClass = 0,
                vrMovementMode = st?.VrMovementMode ?? 1,
                roomInstance = instance,
                isOnline = true,
                appVersion = "20210827",
                errorCode = 0,
                lastOnline = DateTime.UtcNow.ToString("yyyy-MM-ddTHH:mm:ss"),
                clientJoinData = ""
            });
        }

        [HttpGet("avoidjuniors")]
        public async Task<IActionResult> PlayerAvoidJuniors()
        {
            var accountId = GetAccountId();
            if (accountId == null) return Content("false");

            var st = await _context.PlayerStates.FirstOrDefaultAsync(p => p.AccountID == accountId);
            if (st == null) return Content("false");

            return Content(st.AvoidJuniors ? "true" : "false");
        }



        [HttpGet("/api/players/v1/progression")]
        [HttpGet("/api/players/v2/progression")]
        public async Task<IActionResult> GetProgression()
        {
            var accountId = GetAccountId();
            if (accountId == null) return Unauthorized();

            var prog = await _context.Progressions.FindAsync(accountId.Value);
            if (prog == null)
            {
                prog = new Progression { AccountID = accountId.Value, Level = 1, XP = 0 };
                _context.Progressions.Add(prog);
                await _context.SaveChangesAsync();
            }
            return Ok(prog);
        }

        [HttpGet("/api/players/v2/progression/bulk")]
        public async Task<IActionResult> ProgressionBulk([FromQuery] List<long> id)
        {
            if (id == null || id.Count == 0) return Ok(new object[0]);

            var progs = await _context.Progressions.Where(p => id.Contains(p.AccountID)).ToListAsync();
            var results = new List<Progression>();

            foreach (var i in id)
            {
                var p = progs.FirstOrDefault(x => x.AccountID == i);
                if (p != null) results.Add(p);
                else results.Add(new Progression { AccountID = i, Level = 1, XP = 0 });
            }
            return Ok(results);
        }

        [HttpGet("/api/playerReputation/v2/bulk")]
        public IActionResult ReputationBulk([FromQuery] List<long> id)
        {
            var results = id.Select(i => new { AccountId = i, Noteriety = 0, IsCheerful = true, CheerGeneral = 0, CheerHelpful = 0, CheerGreatHost = 0, CheerSportsman = 0, CheerCreative = 0, CheerCredit = 0, SelectedCheer = 0 }).ToList();
            return Ok(results);
        }

        [HttpGet("/api/avatar/v4/items")]
        public async Task<IActionResult> AvatarItems([FromServices] IWebHostEnvironment env)
        {
            var filePath = Path.Combine(env.ContentRootPath, "..", "data", "jsons", "defaultUnlocked.json");
            if (System.IO.File.Exists(filePath))
            {
                var json = await System.IO.File.ReadAllTextAsync(filePath);
                return Content(json, "application/json");
            }
            return Ok(new object[] { });
        }

        public class AvatarDto
        {
            public string FaceFeatures { get; set; } = "";
            public string HairColor { get; set; } = "";
            public string OutfitSelections { get; set; } = "";
            public string SkinColor { get; set; } = "";
        }

        [HttpGet("/api/avatar/v2")]
        public async Task<IActionResult> AvatarV2()
        {
            var accountId = GetAccountId();
            if (accountId == null) return Unauthorized();

            var avatar = await _context.Avatars.FirstOrDefaultAsync(a => a.AccountID == accountId.Value);
            if (avatar == null)
            {
                return Ok(new { FaceFeatures = "", HairColor = "", OutfitSelections = "", SkinColor = "" });
            }

            return Ok(new
            {
                FaceFeatures = avatar.FaceFeatures ?? "",
                HairColor = avatar.HairColor ?? "",
                OutfitSelections = avatar.OutfitSelections ?? "",
                SkinColor = avatar.SkinColor ?? ""
            });
        }

        [HttpPost("/api/avatar/v2/set")]
        public async Task<IActionResult> AvatarSet([FromBody] AvatarDto dto)
        {
            var accountId = GetAccountId();
            if (accountId == null) return Unauthorized();

            var avatar = await _context.Avatars.FirstOrDefaultAsync(a => a.AccountID == accountId.Value);
            if (avatar == null)
            {
                avatar = new Avatar { AccountID = accountId.Value };
                _context.Avatars.Add(avatar);
            }

            avatar.FaceFeatures = dto.FaceFeatures;
            avatar.HairColor = dto.HairColor;
            avatar.OutfitSelections = dto.OutfitSelections;
            avatar.SkinColor = dto.SkinColor;

            await _context.SaveChangesAsync();

            return Ok(new
            {
                FaceFeatures = avatar.FaceFeatures ?? "",
                HairColor = avatar.HairColor ?? "",
                OutfitSelections = avatar.OutfitSelections ?? "",
                SkinColor = avatar.SkinColor ?? ""
            });
        }

        [HttpGet("/api/avatar/v2/saved")]
        public IActionResult AvatarSavedV2()
        {
            var accountId = GetAccountId();
            if (accountId == null) return Ok(Array.Empty<object>());

            var av = _context.Avatars.FirstOrDefault(a => a.AccountID == accountId.Value);
            if (av == null) return Ok(Array.Empty<object>());

            return Ok(new[]
            {
                new {
                    FaceFeatures = av.FaceFeatures ?? "",
                    HairColor = av.HairColor ?? "",
                    OutfitSelections = av.OutfitSelections ?? "",
                    SkinColor = av.SkinColor ?? ""
                }
            });
        }

        [HttpGet("/api/avatar/v3/saved")]
        public IActionResult AvatarSaved()
        {
            return Ok(Array.Empty<object>());
        }

        [HttpGet("/api/avatar/v1/defaultunlocked")]
        public async Task<IActionResult> DefaultUnlocked([FromServices] IWebHostEnvironment env)
        {
            var filePath = Path.Combine(env.ContentRootPath, "..", "data", "jsons", "defaultUnlocked.json");
            if (System.IO.File.Exists(filePath))
            {
                var json = await System.IO.File.ReadAllTextAsync(filePath);
                return Content(json, "application/json; charset=utf-8");
            }
            return Ok(new object[] { });
        }

        [HttpGet("/api/PlayerReporting/v1/moderationBlockDetails")]
        public IActionResult ModerationBlockDetails()
        {
            return Ok(new
            {
                ReportCategory = 0,
                Duration = 0,
                GameSessionId = 0,
                Message = "You are not blocked."
            });
        }



        [HttpPost("/api/relationships/v1/bulkignoreplatformusers")]
        public IActionResult BulkIgnorePlatformUsers()
        {
            return Ok(new object[0]);
        }

        [HttpPost("/api/PlayerReporting/v1/{actionName}")]
        [HttpPut("/api/PlayerReporting/v1/{actionName}")]
        public IActionResult PlayerReportingAction(string actionName)
        {
            return Ok(new object[0]);
        }

        [HttpGet("/api/challenge/v2/getCurrent")]
        public IActionResult GetCurrentChallenge()
        {
            return Ok(new
            {
                ChallengeMapId = 1,
                CompletedRequired = false,
                StartAt = "2026-03-25T21:00:00",
                EndAt = "2027-04-01T21:00:00",
                ServerTime = "2026-03-31T14:42:54.2754728Z",
                Challenges = Array.Empty<object>(),
                Gift = (object)null,
                FallbackGiftName = "",
                ChallengeThemeString = "\"no weekly's at this time.\""
            });
        }

        [HttpGet("/api/settings/v2")]
        public async Task<IActionResult> Settings()
        {
            var accountId = GetAccountId();
            if (accountId == null) return Unauthorized();

            var settings = await _context.PlayerSettings.Where(s => s.AccountID == accountId.Value).ToListAsync();
            return Ok(settings);
        }

        [HttpPost("/api/settings/v2/set")]
        public async Task<IActionResult> SettingsSet([FromBody] JsonElement body)
        {
            var accountId = GetAccountId();
            if (accountId == null) return Unauthorized();

            var reqSettings = new List<PlayerSetting>();
            if (body.ValueKind == JsonValueKind.Array)
            {
                reqSettings = JsonSerializer.Deserialize<List<PlayerSetting>>(body.GetRawText());
            }
            else if (body.ValueKind == JsonValueKind.Object)
            {
                var single = JsonSerializer.Deserialize<PlayerSetting>(body.GetRawText());
                if (single != null) reqSettings.Add(single);
            }

            if (reqSettings != null)
            {
                foreach (var setting in reqSettings)
                {
                    var existing = await _context.PlayerSettings.FirstOrDefaultAsync(s => s.AccountID == accountId.Value && s.Key == setting.Key);
                    if (existing != null)
                    {
                        existing.Value = setting.Value;
                    }
                    else
                    {
                        setting.AccountID = accountId.Value;
                        _context.PlayerSettings.Add(setting);
                    }
                }
                await _context.SaveChangesAsync();
            }

            var allSettings = await _context.PlayerSettings.Where(s => s.AccountID == accountId.Value).ToListAsync();
            return Ok(allSettings);
        }
    }

    [ApiController]
    [Route("econ/balance")]
    public class BalanceController : ControllerBase
    {
        private readonly MeowNetDbContext _context;

        public BalanceController(MeowNetDbContext context)
        {
            _context = context;
        }

        private long? GetAccountId()
        {
            var authHeader = Request.Headers.Authorization.ToString();
            if (string.IsNullOrEmpty(authHeader) || !authHeader.StartsWith("Bearer ", StringComparison.OrdinalIgnoreCase))
                return null;

            var idClaim = User.FindFirstValue(ClaimTypes.NameIdentifier) 
                          ?? User.FindFirstValue("sub") 
                          ?? User.FindFirstValue("nameid");
            if (idClaim != null && long.TryParse(idClaim, out var id)) return id;
            return null;
        }

        [HttpGet("{currencyType:int}")]
        public async Task<IActionResult> BalanceGet(int currencyType)
        {
            var accountId = GetAccountId();
            if (accountId == null)
            {
                return Ok(new[]
                {
                    new { Balance = 0, BalanceType = -2, CurrencyType = currencyType }
                });
            }

            var bal = await _context.Balances.FirstOrDefaultAsync(b => b.AccountID == accountId.Value && b.CurrencyType == currencyType);
            if (bal == null)
            {
                bal = new Balance
                {
                    AccountID = accountId.Value,
                    CurrencyType = currencyType,
                    Amount = 0,
                    BalanceType = -2
                };
                _context.Balances.Add(bal);
                await _context.SaveChangesAsync();
            }

            return Ok(new[]
            {
                new { Balance = bal.Amount, BalanceType = bal.BalanceType, CurrencyType = currencyType }
            });
        }

        [HttpPost("/api/PlayerReporting/v1/hile")]
        public IActionResult Hile()
        {
            return Ok(new { error = "", success = true, value = new object() });
        }

        [HttpGet("/api/playerevents/v1/all")]
        public IActionResult PlayerEventsAll()
        {
            return Ok(new { Created = Array.Empty<object>(), Attending = Array.Empty<object>(), Hosting = Array.Empty<object>(), Responses = Array.Empty<object>() });
        }

        [HttpGet("/api/playerevents/v1/room/{id}")]
        public IActionResult RoomEventsById(long id)
        {
            return Ok(Array.Empty<object>());
        }

        [HttpGet("/api/objectives/v1/myprogress")]
        public IActionResult ObjectivesMyProgress()
        {
            return Ok(new { Objectives = Array.Empty<object>(), ObjectiveGroups = Array.Empty<object>() });
        }

        [HttpPost("/player/photonregionpings")]
        [HttpGet("/player/photonregionpings")]
        [HttpPut("/player/photonregionpings")]
        public IActionResult PhotonRegionPings()
        {
            return Ok(new { error = "", success = true, value = new object() });
        }

        [HttpPost("/api/CampusCard/v1/UpdateAndGetSubscription")]
        [HttpGet("/api/CampusCard/v1/UpdateAndGetSubscription")]
        public IActionResult UpdateAndGetSubscription()
        {
            return Ok(new { error = "", success = true, value = new { IsSubscribed = false } });
        }

        [HttpGet("/api/avatar/v2/gifts")]
        public IActionResult AvatarGifts()
        {
            return Ok(Array.Empty<object>());
        }

        [HttpGet("/api/gamerewards/v1/pending")]
        public IActionResult GameRewardsPending()
        {
            return Ok(Array.Empty<object>());
        }


        [HttpGet("/api/roomkeys/v1/mine")]
        public IActionResult RoomKeysMine()
        {
            return Ok(Array.Empty<object>());
        }

        [HttpPut("/player/statusvisibility")]
        public IActionResult StatusVisibility()
        {
            return Ok();
        }

        [HttpGet("/api/PlayerReporting/v1/voteToKickReasons")]
        public IActionResult VoteToKickReasons()
        {
            return Ok(Array.Empty<object>());
        }
    }
}
