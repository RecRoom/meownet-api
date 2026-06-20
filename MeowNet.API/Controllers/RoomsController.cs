using MeowNet.API.Data;
using MeowNet.API.Models;
using Microsoft.AspNetCore.Mvc;
using Microsoft.EntityFrameworkCore;
using System.Security.Claims;
using System.Text.Json;
namespace MeowNet.API.Controllers
{
    [ApiController]
    [Route("rooms")]
    public class RoomsController : ControllerBase
    {
        private readonly MeowNetDbContext _context;
        private readonly IWebHostEnvironment _env;
        public RoomsController(MeowNetDbContext context, IWebHostEnvironment env)
        {
            _context = context;
            _env = env;
        }
        private long? GetAccountId()
        {
            var authHeader = Request.Headers.Authorization.ToString();
            if (string.IsNullOrEmpty(authHeader) || !authHeader.StartsWith("Bearer "))
                return null;
            var idClaim = User.FindFirstValue(ClaimTypes.NameIdentifier) ?? User.FindFirstValue("sub") ?? User.FindFirstValue("nameid");
            if (long.TryParse(idClaim, out var id)) return id;
            return null;
        }
        private async Task InitRoomSlices(List<Room> rooms)
        {
            foreach (var room in rooms)
            {
                if (room.SubRooms == null || !room.SubRooms.Any()) 
                {
                    room.SubRooms = await _context.SubRooms.Where(s => s.RoomId == room.RoomId).ToListAsync();
                }
                if (room.Roles == null || room.Roles.Count == 0)
                {
                    room.Roles = await _context.RoomRoleEntrys.Where(r => r.RoomId == room.RoomId).ToListAsync();
                    if (room.Roles.Count == 0)
                    {
                        room.Roles = new List<RoomRoleEntry>
                        {
                            new RoomRoleEntry { AccountId = (int)room.CreatorAccountId, Role = 255, InvitedRole = 255 }
                        };
                    }
                }
                if (room.Tags == null || room.Tags.Count == 0)
                {
                    room.Tags = await _context.RoomTags.Where(t => t.RoomId == room.RoomId).ToListAsync();
                }
                if (room.LoadScreens == null) room.LoadScreens = new List<object>();
                if (room.PromoImages == null) room.PromoImages = new List<object>();
                if (room.PromoExternalContent == null) room.PromoExternalContent = new List<object>();
                if (room.Stats == null) room.Stats = new RoomStats { CheerCount = 0, FavoriteCount = 0, VisitorCount = 0, VisitCount = 0 };
                if (room.DataBlob == null) room.DataBlob = "";
            }
        }
        [HttpGet("{id:long}")]
        public async Task<IActionResult> RoomsGetById(long id)
        {
            Console.WriteLine($"[ROOMS] RoomsGetById called for id={id}");
            var room = await _context.Rooms
                .FirstOrDefaultAsync(r => r.RoomId == id);
            if (room == null) return NotFound("Not Found");
            var list = new List<Room> { room };
            await InitRoomSlices(list);
            return Ok(room);
        }
        [HttpGet("")]
        public async Task<IActionResult> RoomsGet([FromQuery] string name)
        {
            if (string.IsNullOrEmpty(name)) return BadRequest("Bad Request");
            var pattern = name.ToLower();
            var room = await _context.Rooms
                .FirstOrDefaultAsync(r => EF.Functions.Like(r.Name, pattern));
            if (room == null) return NotFound("Not Found");
            var list = new List<Room> { room };
            await InitRoomSlices(list);
            return Ok(room);
        }
        [HttpGet("bulk")]
        [HttpPost("bulk")]
        public async Task<IActionResult> RoomsBulk([FromQuery] string? name, [FromQuery] List<long>? id)
        {
            if (Request.Method == "POST")
            {
                using var reader = new StreamReader(Request.Body);
                var body = await reader.ReadToEndAsync();
                try {
                    var postIds = JsonSerializer.Deserialize<List<long>>(body);
                    if (postIds != null && postIds.Count > 0)
                    {
                        id = postIds;
                    }
                } catch { }
            }

            if (id != null && id.Count > 0)
            {
                var rooms = await _context.Rooms.Where(r => id.Contains(r.RoomId)).ToListAsync();
                await InitRoomSlices(rooms);
                return Ok(rooms);
            }

            if (!string.IsNullOrEmpty(name))
            {
                var pattern = name.ToLower();
                var room = await _context.Rooms.FirstOrDefaultAsync(r => r.Name.ToLower() == pattern);
                if (room == null) return Ok(new List<Room>()); // Return empty list instead of 404 to mimic Go
                var list = new List<Room> { room };
                await InitRoomSlices(list);
                return Ok(list);
            }

            return BadRequest("Bad Request");
        }

        [HttpGet("hot")]
        public async Task<IActionResult> RoomsHot([FromQuery] string? tag, [FromQuery] int limit = 50)
        {
            if (limit <= 0) limit = 50;
            if (limit > 200) limit = 200;
            var query = _context.Rooms
                .Where(r => !r.IsDorm && r.Accessibility == 1);
            tag = tag?.ToLower();
            if (string.IsNullOrEmpty(tag))
            {
                query = query.OrderByDescending(r => r.CreatedAt);
            }
            else if (tag == "new")
            {
                query = query.OrderByDescending(r => r.CreatedAt);
            }
            else if (tag == "rro")
            {
                query = query.Where(r => r.IsRRO).OrderByDescending(r => r.CreatedAt);
            }
            else
            {
                var roomIdsWithTag = await _context.RoomTags.Where(t => t.Tag.ToLower() == tag).Select(t => t.RoomId).ToListAsync();
                query = query.Where(r => roomIdsWithTag.Contains(r.RoomId))
                             .OrderByDescending(r => r.CreatedAt);
            }
            var rooms = await query.Take(limit).ToListAsync();
            await InitRoomSlices(rooms);
            return Ok(new
            {
                TotalResults = rooms.Count,
                Results = rooms
            });
        }
        [HttpGet("search")]
        public async Task<IActionResult> RoomsSearch([FromQuery] string query)
        {
            if (string.IsNullOrWhiteSpace(query)) return Ok(new { TotalResults = 0, Results = new List<Room>() });
            var q = _context.Rooms
                .Where(r => !r.IsDorm && r.Accessibility == 1);
            var terms = query.Split(new[] { ' ' }, StringSplitOptions.RemoveEmptyEntries);
            var tags = terms.Where(t => t.StartsWith("#")).Select(t => t.Substring(1).ToLower()).ToList();
            var nameTerms = terms.Where(t => !t.StartsWith("#")).ToList();
            if (tags.Any())
            {
                // Tags is NotMapped, so we must filter in memory
                var roomsList = await q.ToListAsync();
                var filtered = roomsList.Where(r => r.Tags != null && r.Tags.Any(t => tags.Contains(t.Tag.ToLower()))).Take(50).ToList();
                await InitRoomSlices(filtered);
                return Ok(new
                {
                    TotalResults = filtered.Count,
                    Results = filtered
                });
            }
            
            foreach (var term in nameTerms)
            {
                var p = $"%{term.ToLower()}%";
                q = q.Where(r => EF.Functions.Like(r.Name, p));
            }

            var rooms = await q.Take(50).ToListAsync();
            await InitRoomSlices(rooms);
            return Ok(new
            {
                TotalResults = rooms.Count,
                Results = rooms
            });
        }
        [HttpGet("createdby/me")]
        [HttpGet("createdbyme")]
        public async Task<IActionResult> RoomCreatedByMe()
        {
            var accountId = GetAccountId();
            if (accountId == null) return Ok(new List<object>());
            var rooms = await _context.Rooms
                .Where(r => r.CreatorAccountId == accountId.Value)
                .ToListAsync();
            await InitRoomSlices(rooms);
            return Ok(rooms);
        }

        [HttpGet("visitedby/me")]
        public IActionResult RoomVisitedByMe()
        {
            return Ok(new List<object>());
        }

        [HttpGet("favoritedby/me")]
        public IActionResult RoomFavoritedByMe()
        {
            return Ok(new List<object>());
        }

        [HttpGet("recommendations")]
        public IActionResult RoomRecommendations()
        {
            return Ok(new List<object>());
        }

        [HttpGet("curated_playlists")]
        public IActionResult RoomCuratedPlaylists()
        {
            return Ok(new List<object>());
        }

        [HttpGet("topcreators")]
        public IActionResult RoomTopCreators()
        {
            return Ok(new List<object>());
        }

        [HttpGet("/api/rooms/v1/filters")]
        [HttpGet("filters")]
        public IActionResult RoomFilters()
        {
            return Ok(new
            {
                PinnedFilters = new[] { "rro", "community", "featured", "quest", "pvp", "hangout", "game", "art", "horror" },
                PopularFilters = new[] { "pvp", "quest", "game", "hangout", "art" },
                TrendingFilters = new[] { "featured", "game", "horror", "quest" }
            });
        }

        [HttpGet("/api/roomcurrencies/v1/betaEnabled")]
        public IActionResult RoomCurrenciesBetaEnabled()
        {
            return Ok(true);
        }

        [HttpGet("{roomId}/playerdata/me")]
        public IActionResult RoomPlayerDataMe(long roomId)
        {
            return Ok(new { Data = "" });
        }

        [HttpPost("{roomId}/subrooms/{subroomId}/data")]
        [HttpPut("{roomId}/subrooms/{subroomId}/data")]
        public async Task<IActionResult> SubRoomData(long roomId, long subroomId)
        {
            var filename = Request.HasFormContentType && Request.Form.TryGetValue("filename", out var fn) ? fn.ToString() : "";
            var roomDataFilename = Request.HasFormContentType && Request.Form.TryGetValue("roomDataFilename", out var rn) ? rn.ToString() : "";

            Console.WriteLine($"[ROOMS] Saving room data for Room={roomId}, SubRoom={subroomId}, Filename={filename}, RoomDataFilename={roomDataFilename}");

            var room = await _context.Rooms
                .FirstOrDefaultAsync(r => r.RoomId == roomId);

            if (room != null && !string.IsNullOrEmpty(roomDataFilename))
            {
                room.DataBlob = roomDataFilename;
            }
            
            var subroom = await _context.SubRooms.FirstOrDefaultAsync(s => s.RoomId == roomId && s.SubRoomId == subroomId);
            if (subroom != null && !string.IsNullOrEmpty(filename))
            {
                subroom.DataBlob = filename;
                
                var accountId = GetAccountId();
                if (accountId != null)
                {
                    subroom.SavedByAccountId = (int)accountId.Value;
                    _context.SubRoomDataHistorys.Add(new SubRoomDataHistory
                    {
                        SubRoomId = (int)subroomId,
                        DataBlob = filename,
                        SavedByAccountId = (int)accountId.Value,
                        CreatedAt = DateTime.UtcNow
                    });
                }
                
                _context.SubRooms.Update(subroom);
            }
            
            if (room != null || subroom != null) await _context.SaveChangesAsync();
            
            if (room != null) await InitRoomSlices(new List<Room> { room });

            return Ok(new { error = "", success = true, value = room });
        }

        [HttpGet("/api/roomkeys/v1/room")]
        public IActionResult RoomKeys([FromQuery] long roomId)
        {
            return Ok(new object[0]);
        }

        [HttpGet("/room/{filename}")]
        public IActionResult RoomData(string filename)
        {
            var path = Path.Combine(_env.ContentRootPath, "uploads", "room", filename);
            if (System.IO.File.Exists(path))
            {
                return PhysicalFile(path, "application/octet-stream");
            }
            return NotFound();
        }

        private RoomInteraction GetOrCreateInteraction(long roomId, long accountId)
        {
            var interaction = _context.RoomInteractions.FirstOrDefault(i => i.RoomId == roomId && i.AccountId == accountId);
            if (interaction == null)
            {
                interaction = new RoomInteraction { RoomId = roomId, AccountId = accountId };
                _context.RoomInteractions.Add(interaction);
            }
            return interaction;
        }

        [HttpGet("{roomId}/interactionby/me")]
        public IActionResult GetInteraction(long roomId)
        {
            var accountId = GetAccountId();
            if (accountId == null) return Unauthorized();

            var interaction = _context.RoomInteractions.FirstOrDefault(i => i.RoomId == roomId && i.AccountId == accountId.Value);
            return Ok(new
            {
                Cheered = interaction?.Cheered ?? false,
                Favorited = interaction?.Favorited ?? false
            });
        }

        [HttpPut("{roomId}/interactionby/me/cheer")]
        public async Task<IActionResult> Cheer(long roomId)
        {
            var accountId = GetAccountId();
            if (accountId == null) return Unauthorized();

            var interaction = GetOrCreateInteraction(roomId, accountId.Value);
            interaction.Cheered = true;
            await _context.SaveChangesAsync();
            return Ok(new { Cheered = true, Favorited = interaction.Favorited });
        }

        [HttpDelete("{roomId}/interactionby/me/cheer")]
        public async Task<IActionResult> Uncheer(long roomId)
        {
            var accountId = GetAccountId();
            if (accountId == null) return Unauthorized();

            var interaction = GetOrCreateInteraction(roomId, accountId.Value);
            interaction.Cheered = false;
            await _context.SaveChangesAsync();
            return Ok(new { Cheered = false, Favorited = interaction.Favorited });
        }

        [HttpPut("{roomId}/interactionby/me/favorite")]
        public async Task<IActionResult> Favorite(long roomId)
        {
            var accountId = GetAccountId();
            if (accountId == null) return Unauthorized();

            var interaction = GetOrCreateInteraction(roomId, accountId.Value);
            interaction.Favorited = true;
            await _context.SaveChangesAsync();
            return Ok(new { Cheered = interaction.Cheered, Favorited = true });
        }

        [HttpDelete("{roomId}/interactionby/me/favorite")]
        public async Task<IActionResult> Unfavorite(long roomId)
        {
            var accountId = GetAccountId();
            if (accountId == null) return Unauthorized();

            var interaction = GetOrCreateInteraction(roomId, accountId.Value);
            interaction.Favorited = false;
            await _context.SaveChangesAsync();
            return Ok(new { Cheered = interaction.Cheered, Favorited = false });
        }
    }
}
