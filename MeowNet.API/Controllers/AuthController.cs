using System.IdentityModel.Tokens.Jwt;
using System.Security.Claims;
using System.Text;
using MeowNet.API.Data;
using MeowNet.API.Models;
using Microsoft.AspNetCore.Mvc;
using Microsoft.EntityFrameworkCore;
using Microsoft.IdentityModel.Tokens;

namespace MeowNet.API.Controllers
{
    [ApiController]
    public class AuthController : ControllerBase
    {
        private readonly MeowNetDbContext _context;

        public AuthController(MeowNetDbContext context)
        {
            _context = context;
        }

        private string GenerateJwt(long accountId, int platform, string platformId)
        {
            var tokenHandler = new JwtSecurityTokenHandler();
            var key = new SymmetricSecurityKey(Encoding.UTF8.GetBytes("SuperSecretDummyKeyForLocalHost123!"));
            var creds = new SigningCredentials(key, SecurityAlgorithms.HmacSha256Signature);

            for (int i = 0; i < 10000; i++)
            {
                var claims = new List<Claim>
                {
                    new Claim(JwtRegisteredClaimNames.Sub, accountId.ToString()),
                    new Claim(ClaimTypes.NameIdentifier, accountId.ToString()),
                    new Claim("rn.plat", platform.ToString()),
                    new Claim("rn.platid", platformId),
                    new Claim("role", "gameClient"),
                    new Claim("pad", i.ToString())
                };

                var tokenDescriptor = new SecurityTokenDescriptor
                {
                    Subject = new ClaimsIdentity(claims),
                    Expires = DateTime.UtcNow.AddHours(12),
                    Issuer = "http://localhost:8080",
                    Audience = "http://localhost:8080",
                    SigningCredentials = creds
                };

                var token = tokenHandler.CreateJwtSecurityToken(tokenDescriptor);
                var tokenStr = tokenHandler.WriteToken(token);
                
                var parts = tokenStr.Split('.');
                if (parts.Length == 3 && tokenStr.Length < 800)
                {
                    return tokenStr;
                }
            }
            return ""; // Should never happen
        }

        [HttpGet("cachedlogin/forplatformids")]
        [HttpPost("cachedlogin/forplatformids")]
        public async Task<IActionResult> CachedLoginForPlatformIds([FromQuery, FromForm] string[] id)
        {
            var platformAccounts = await _context.PlatformAccounts
                .Where(pa => id.Contains(pa.PlatformID))
                .ToListAsync();

            var results = platformAccounts.Select(pa => new
            {
                platform = pa.Platform,
                platformId = pa.PlatformID,
                accountId = pa.AccountID,
                lastLoginTime = "2020-06-26T00:00:00Z",
                requirePassword = false
            }).ToList();

            return Ok(results);
        }

        [HttpGet("cachedlogin/forplatformid/{platform}/{platformId}")]
        public async Task<IActionResult> CachedLoginForPlatformId(int platform, string platformId)
        {
            var platformAccounts = await _context.PlatformAccounts
                .Where(pa => pa.Platform == platform && pa.PlatformID == platformId)
                .ToListAsync();

            var results = platformAccounts.Select(pa => new
            {
                accountId = pa.AccountID,
                lastLoginTime = "2020-06-26T00:00:00Z",
                platform = pa.Platform,
                platformId = pa.PlatformID,
                requirePassword = false
            }).ToList();

            return Ok(results);
        }

        [HttpPost("api/platformlogin")]
        [HttpPost("api/platformlogin/logintocachedaccount")]
        public async Task<IActionResult> PlatformLogin([FromForm] int platform, [FromForm] string platform_id, [FromForm] long account_id)
        {
            PlatformAccount pa = null;
            
            if (account_id > 0)
            {
                pa = await _context.PlatformAccounts.FirstOrDefaultAsync(p => p.Platform == platform && p.PlatformID == platform_id && p.AccountID == account_id);
            }
            
            // Fallback: If cached account_id is invalid (e.g. wiped DB), still check platform_id
            if (pa == null)
            {
                pa = await _context.PlatformAccounts.FirstOrDefaultAsync(p => p.Platform == platform && p.PlatformID == platform_id);
            }

            if (pa == null)
            {
                // Create a new account if it doesn't exist
                var account = new Account
                {
                    RawUsername = "Player_" + platform_id,
                    Username = "player_" + platform_id,
                    DisplayName = "Player",
                    Platforms = platform,
                    CreatedAt = DateTime.UtcNow,
                    ProfileImage = "DefaultImage.png"
                };
                
                _context.Accounts.Add(account);
                await _context.SaveChangesAsync();
                
                await Helpers.AccountHelpers.SetupNewAccountDefaults(_context, account);
                
                pa = new PlatformAccount
                {
                    AccountID = account.AccountID,
                    Platform = platform,
                    PlatformID = platform_id
                };
                
                _context.PlatformAccounts.Add(pa);
                await _context.SaveChangesAsync();
            }

            var token = GenerateJwt(pa.AccountID, pa.Platform, pa.PlatformID);

            return Ok(new
            {
                access_token = token,
                refresh_token = Guid.NewGuid().ToString("N"),
                error = (object)null,
                error_description = (object)null,
                key = "ZWZmNzk5ZGEtM2RmOC00NWQ5LTkwNjYtYTZmZWU1ZmIzMjI4"
            });
        }

        [HttpPost("connect/token")]
        [Consumes("application/x-www-form-urlencoded")]
        public async Task<IActionResult> ConnectToken([FromForm] string grant_type, [FromForm] int platform, [FromForm] string? platform_id, [FromForm] long account_id, [FromForm] string? username, [FromForm] string? password)
        {
            Console.WriteLine($"[ConnectToken] grant_type={grant_type}, platform={platform}, platform_id={platform_id}, account_id={account_id}, username={username}, password={password}");
            
            if (grant_type == "password" && !string.IsNullOrEmpty(username) && !string.IsNullOrEmpty(password))
            {
                var account = await _context.Accounts.FirstOrDefaultAsync(a => a.Username == username);
                if (account == null || account.PasswordHash != password)
                {
                    return Unauthorized();
                }

                if (!string.IsNullOrEmpty(platform_id))
                {
                    var link = await _context.PlatformAccounts.FirstOrDefaultAsync(p => p.AccountID == account.AccountID && p.Platform == platform && p.PlatformID == platform_id);
                    if (link == null)
                    {
                        link = new PlatformAccount
                        {
                            AccountID = account.AccountID,
                            Platform = platform,
                            PlatformID = platform_id
                        };
                        _context.PlatformAccounts.Add(link);
                        await _context.SaveChangesAsync();
                    }
                }

                var pa = await _context.PlatformAccounts.FirstOrDefaultAsync(p => p.AccountID == account.AccountID);
                int p_id = pa != null ? pa.Platform : 0;
                string p_str = pa != null ? pa.PlatformID : "0";

                var token = GenerateJwt(account.AccountID, p_id, p_str);
                return Ok(new
                {
                    access_token = token,
                    refresh_token = Guid.NewGuid().ToString("N"),
                    error = (object)null,
                    error_description = (object)null,
                    key = "ZWZmNzk5ZGEtM2RmOC00NWQ5LTkwNjYtYTZmZWU1ZmIzMjI4"
                });
            }
            else if (grant_type == "cached_login" || grant_type == "password" || grant_type == "refresh_token") // Fallthrough for password/refresh_token
            {
                return await PlatformLogin(platform, platform_id ?? "0", account_id);
            }
            
            return Unauthorized();
        }
    }
}
