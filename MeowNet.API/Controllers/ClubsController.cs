using MeowNet.API.Data;
using MeowNet.API.Models;
using Microsoft.AspNetCore.Mvc;
using Microsoft.EntityFrameworkCore;
using System.Security.Claims;
using System.Text.Json;

namespace MeowNet.API.Controllers
{
    [ApiController]
    [Route("club")]
    public class ClubsController : ControllerBase
    {
        private readonly MeowNetDbContext _context;

        public ClubsController(MeowNetDbContext context)
        {
            _context = context;
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

        [HttpGet("categoryTags")]
        public IActionResult ClubCategoryTags()
        {
            return Ok(new[]
            {
                "Social",
                "Creative",
                "Competitive",
                "Casual",
                "Entertainment"
            });
        }

        [HttpGet("{clubId:long}/details")]
        public async Task<IActionResult> ClubDetails(long clubId)
        {
            var club = await _context.Clubs.FirstOrDefaultAsync(c => c.ClubId == clubId);
            if (club == null) return NotFound("Not Found");

            return Ok(new { error = "", success = true, value = club });
        }

        [HttpGet("v1/myClubsWithUnreadAnnouncements")]
        public IActionResult MyClubsWithUnreadAnnouncements()
        {
            return Ok(new { error = "", success = true, value = Array.Empty<object>() });
        }

        [HttpGet("v2/get")]
        public async Task<IActionResult> ClubHomeMe()
        {
            var accountId = GetAccountId();
            if (accountId == null) return Unauthorized();

            var account = await _context.Accounts.FirstOrDefaultAsync(a => a.AccountID == accountId.Value);
            if (account == null || account.HomeClubId == null) return NotFound();

            var club = await _context.Clubs.FirstOrDefaultAsync(c => c.ClubId == account.HomeClubId.Value);
            if (club == null || club.ClubhouseRoomId == null) return NotFound();

            return Ok(club);
        }

        [HttpGet("home/me")]
        public async Task<IActionResult> ClubHomeGet()
        {
            var accountId = GetAccountId();
            if (accountId == null) return Unauthorized();

            var account = await _context.Accounts.FirstOrDefaultAsync(a => a.AccountID == accountId.Value);
            if (account == null || account.HomeClubId == null) return NotFound("Not Found");

            var club = await _context.Clubs.FirstOrDefaultAsync(c => c.ClubId == account.HomeClubId.Value);
            if (club == null) return NotFound("Not Found");

            return Ok(new { error = "", success = true, value = club });
        }

        [HttpPut("home/me")]
        public async Task<IActionResult> ClubHomeSet([FromForm] long clubId)
        {
            var accountId = GetAccountId();
            if (accountId == null) return Unauthorized();

            var club = await _context.Clubs.FirstOrDefaultAsync(c => c.ClubId == clubId);
            if (club == null) return NotFound("Not Found");

            var member = await _context.ClubMembers.FirstOrDefaultAsync(m => m.ClubId == clubId && m.AccountId == accountId.Value);
            if (member == null || member.MembershipType < 1) // Member
            {
                return StatusCode(403, new { error = "not a member", success = false });
            }

            var account = await _context.Accounts.FirstOrDefaultAsync(a => a.AccountID == accountId.Value);
            if (account != null)
            {
                account.HomeClubId = clubId;
                _context.Accounts.Update(account);
                await _context.SaveChangesAsync();
            }

            return Ok(new { error = "", success = true, value = club });
        }

        [HttpGet("mine/created")]
        public async Task<IActionResult> ClubMineCreated()
        {
            var accountId = GetAccountId();
            if (accountId == null) return Ok(new List<Club>());

            var clubs = await _context.Clubs
                .Where(c => c.CreatorAccountId == accountId.Value && c.ClubType != 1)
                .OrderBy(c => c.CreatedAt)
                .ToListAsync();

            return Ok(clubs);
        }

        [HttpGet("mine/member")]
        public async Task<IActionResult> ClubMineMember()
        {
            var accountId = GetAccountId();
            if (accountId == null) return Ok(new List<Club>());

            var clubs = await _context.Clubs
                .Join(_context.ClubMembers, 
                      c => c.ClubId, 
                      m => m.ClubId, 
                      (c, m) => new { Club = c, Member = m })
                .Where(cm => cm.Member.AccountId == accountId.Value && cm.Member.MembershipType >= 1 && cm.Club.ClubType != 1)
                .OrderBy(cm => cm.Club.CreatedAt)
                .Select(cm => cm.Club)
                .ToListAsync();

            return Ok(clubs);
        }

        [HttpGet("/api/clubs/v1/my_membership_clubs")]
        public IActionResult MyMembershipClubs()
        {
            return Ok(new List<object>());
        }

        [HttpGet("search")]
        public async Task<IActionResult> ClubSearch([FromQuery] string? category, [FromQuery] string? query, [FromQuery] int count = 30, [FromQuery] string? sort = "")
        {
            if (count <= 0 || count > 100) count = 30;

            var q = _context.Clubs.Where(c => c.Visibility == 1 && c.ClubType != 1);

            if (!string.IsNullOrEmpty(category))
            {
                q = q.Where(c => EF.Functions.ILike(c.Category, category));
            }

            if (!string.IsNullOrEmpty(query))
            {
                var like = $"%{query.ToLower()}%";
                q = q.Where(c => EF.Functions.ILike(c.Name, like) || EF.Functions.ILike(c.Description, like));
            }

            if (sort == "1")
                q = q.OrderByDescending(c => c.CreatedAt);
            else if (sort == "2")
                q = q.OrderBy(c => c.Name);
            else
                q = q.OrderByDescending(c => c.MemberCount).ThenByDescending(c => c.CreatedAt);

            var total = await q.CountAsync();
            var clubs = await q.Take(count).ToListAsync();

            return Ok(new
            {
                Clubs = clubs,
                ContinuationToken = (object)null,
                TotalClubs = total
            });
        }
    }
}
