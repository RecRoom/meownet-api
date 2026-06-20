using MeowNet.API.Data;
using MeowNet.API.Models;
using Microsoft.AspNetCore.Mvc;
using Microsoft.EntityFrameworkCore;
using System.Security.Claims;

namespace MeowNet.API.Controllers
{
    [ApiController]
    [Route("leaderboard")]
    public class LeaderboardController : ControllerBase
    {
        private readonly MeowNetDbContext _context;

        public LeaderboardController(MeowNetDbContext context)
        {
            _context = context;
        }

        private long? GetAccountId()
        {
            var idClaim = User.FindFirstValue(ClaimTypes.NameIdentifier) ?? User.FindFirstValue("sub") ?? User.FindFirstValue("nameid");
            if (long.TryParse(idClaim, out var id)) return id;
            return null;
        }

        private async Task<int> ComputeRank(int roomId, int statChannel, long accountId, int score)
        {
            var higher = await _context.LeaderboardStats
                .Where(s => s.RoomID == roomId && s.StatChannel == statChannel && (s.Score > score || (s.Score == score && s.AccountID < accountId)))
                .CountAsync();
            return higher;
        }

        public class LbEntry
        {
            public long playerId { get; set; }
            public int rank { get; set; }
            public int score { get; set; }
        }

        public class RanksReq
        {
            public int FilterType { get; set; }
            public long PlayerId { get; set; }
            public int RankStart { get; set; }
            public int RankEnd { get; set; }
            public int RoomId { get; set; }
            public bool SortAscending { get; set; }
            public int StatChannel { get; set; }
            public int Timeframe { get; set; }
        }

        [HttpPost("GetRanks")]
        public async Task<IActionResult> GetRanks([FromBody] RanksReq body)
        {
            if (body.RankStart < 0) body.RankStart = 0;
            int limit = body.RankEnd - body.RankStart + 1;
            if (limit <= 0) return Ok(new { rows = Array.Empty<LbEntry>() });

            var query = _context.LeaderboardStats
                .Where(s => s.RoomID == body.RoomId && s.StatChannel == body.StatChannel);

            if (body.SortAscending)
                query = query.OrderBy(s => s.Score).ThenBy(s => s.AccountID);
            else
                query = query.OrderByDescending(s => s.Score).ThenBy(s => s.AccountID);

            var stats = await query.Skip(body.RankStart).Take(limit).ToListAsync();

            var rows = stats.Select((s, i) => new LbEntry
            {
                playerId = s.AccountID,
                rank = body.RankStart + i,
                score = s.Score
            }).ToList();

            return Ok(new { rows });
        }

        [HttpPost("GetPlayerRank")]
        public async Task<IActionResult> GetPlayerRank([FromBody] RanksReq body)
        {
            var stat = await _context.LeaderboardStats
                .FirstOrDefaultAsync(s => s.AccountID == body.PlayerId && s.RoomID == body.RoomId && s.StatChannel == body.StatChannel);

            var resp = new LbEntry { playerId = body.PlayerId };

            if (stat != null)
            {
                resp.score = stat.Score;
                resp.rank = await ComputeRank(body.RoomId, body.StatChannel, body.PlayerId, stat.Score);
            }
            else
            {
                var total = await _context.LeaderboardStats
                    .Where(s => s.RoomID == body.RoomId && s.StatChannel == body.StatChannel)
                    .CountAsync();
                resp.rank = total;
                resp.score = 0;
            }

            return Ok(resp);
        }

        public class NearbyReq : RanksReq
        {
            public int WindowSize { get; set; }
        }

        [HttpPost("GetNearbyScores")]
        public async Task<IActionResult> GetNearbyScores([FromBody] NearbyReq body)
        {
            if (body.WindowSize <= 0) body.WindowSize = 10;

            var query = _context.LeaderboardStats
                .Where(s => s.RoomID == body.RoomId && s.StatChannel == body.StatChannel);

            if (body.SortAscending)
                query = query.OrderBy(s => s.Score).ThenBy(s => s.AccountID);
            else
                query = query.OrderByDescending(s => s.Score).ThenBy(s => s.AccountID);

            var all = await query.ToListAsync();

            if (all.Count == 0) return Ok(new { rows = Array.Empty<LbEntry>() });

            int center = all.Count;
            for (int i = 0; i < all.Count; i++)
            {
                if (all[i].AccountID == body.PlayerId)
                {
                    center = i;
                    break;
                }
            }

            int half = body.WindowSize / 2;
            int start = center - half;
            int end = start + body.WindowSize;
            if (start < 0)
            {
                end -= start;
                start = 0;
            }
            if (end > all.Count)
            {
                end = all.Count;
                start = end - body.WindowSize;
                if (start < 0) start = 0;
            }

            var rows = new List<LbEntry>();
            for (int i = start; i < end; i++)
            {
                rows.Add(new LbEntry
                {
                    playerId = all[i].AccountID,
                    rank = i,
                    score = all[i].Score
                });
            }

            return Ok(new { rows });
        }

        public class SetStatReq
        {
            public int CurrentStatValue { get; set; }
            public int RoomId { get; set; }
            public int StatChannel { get; set; }
            public int StatValue { get; set; }
        }

        [HttpPost("CheckAndSetStat")]
        public async Task<IActionResult> CheckAndSetStat([FromBody] SetStatReq body)
        {
            var accountId = GetAccountId();
            if (accountId == null) return Unauthorized();

            var stat = await _context.LeaderboardStats
                .FirstOrDefaultAsync(s => s.AccountID == accountId.Value && s.RoomID == body.RoomId && s.StatChannel == body.StatChannel);

            if (stat == null)
            {
                stat = new LeaderboardStat
                {
                    AccountID = accountId.Value,
                    RoomID = body.RoomId,
                    StatChannel = body.StatChannel,
                    Score = body.StatValue,
                    UpdatedAt = DateTime.UtcNow
                };
                _context.LeaderboardStats.Add(stat);
            }
            else
            {
                stat.Score = body.StatValue;
                stat.UpdatedAt = DateTime.UtcNow;
                _context.LeaderboardStats.Update(stat);
            }

            await _context.SaveChangesAsync();

            int rank = await ComputeRank(body.RoomId, body.StatChannel, accountId.Value, stat.Score);

            return Ok(new
            {
                error = (object)null,
                success = true,
                value = new LbEntry
                {
                    playerId = accountId.Value,
                    rank = rank,
                    score = stat.Score
                }
            });
        }
    }
}
