using MeowNet.API.Data;
using MeowNet.API.Models;
using Microsoft.AspNetCore.Mvc;
using Microsoft.EntityFrameworkCore;
using System.Security.Claims;
using System.Text.Json;
using SkiaSharp;

namespace MeowNet.API.Controllers
{
    [ApiController]
    [Route("account")]
    public class AccountController : ControllerBase
    {
        private readonly MeowNetDbContext _context;
        private readonly IWebHostEnvironment _env;

        public AccountController(MeowNetDbContext context, IWebHostEnvironment env)
        {
            _context = context;
            _env = env;
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

        [HttpGet("test_claims")]
        public IActionResult TestClaims()
        {
            var authHeader = Request.Headers.Authorization.ToString();
            var claims = User.Claims.Select(c => new { c.Type, c.Value }).ToList();
            return Ok(new { authHeader, claims });
        }

        [HttpGet("me")]
        public async Task<IActionResult> AccountMe()
        {
            var accountId = GetAccountId();
            if (accountId == null) return Unauthorized("Unauthorized");

            var acc = await _context.Accounts.FirstOrDefaultAsync(a => a.AccountID == accountId);
            if (acc == null) return NotFound("Account not found");

            if (string.IsNullOrEmpty(acc.ProfileImage))
                acc.ProfileImage = "DefaultImage.png";

            acc.AvailableUsernameChanges = 3;

            return Ok(acc);
        }

        [HttpGet("/parentalcontrol/me")]
        public IActionResult ParentalControlMe()
        {
            var accountId = GetAccountId();
            if (accountId == null) return Unauthorized("Unauthorized");

            return Ok(new { accountId = accountId, disallowInAppPurchases = false });
        }

        [HttpPost("me/birthday")]
        [HttpPut("me/birthday")]
        public async Task<IActionResult> AccountUpdateBirthday()
        {
            var accountId = GetAccountId();
            if (accountId == null) return Unauthorized("Unauthorized");

            var acc = await _context.Accounts.FirstOrDefaultAsync(a => a.AccountID == accountId);
            if (acc == null) return NotFound("Not Found");

            string birthday = null;
            if (Request.HasFormContentType && Request.Form.ContainsKey("birthday"))
            {
                birthday = Request.Form["birthday"];
            }
            else
            {
                using var reader = new StreamReader(Request.Body);
                var bodyString = await reader.ReadToEndAsync();
                try
                {
                    var j = JsonSerializer.Deserialize<Dictionary<string, string>>(bodyString);
                    if (j != null && j.ContainsKey("birthday"))
                    {
                        birthday = j["birthday"];
                    }
                }
                catch
                {
                    // Ignore json parse error, try url-encoded fallback
                    var parts = bodyString.Split('&');
                    foreach(var part in parts)
                    {
                        var kv = part.Split('=');
                        if (kv.Length == 2 && kv[0] == "birthday")
                        {
                            birthday = System.Net.WebUtility.UrlDecode(kv[1]);
                        }
                    }
                }
            }

            if (string.IsNullOrEmpty(birthday))
            {
                return Ok(new { error = "Invalid date format", success = false, value = (object)null });
            }

            DateTime parsed;
            if (!DateTime.TryParseExact(birthday, "yyyy-MM-dd", null, System.Globalization.DateTimeStyles.None, out parsed))
            {
                if (!DateTime.TryParse(birthday, out parsed))
                {
                    return Ok(new { error = "Invalid date format", success = false, value = (object)null });
                }
            }

            int age = DateTime.UtcNow.Year - parsed.Year;
            if (DateTime.UtcNow.DayOfYear < parsed.DayOfYear) age--;

            if (age < 13)
            {
                return Ok(new { error = "You must be at least 13 years old to create a non-junior account", success = false, value = (object)null });
            }

            acc.Birthday = birthday;
            acc.HasBirthday = true;
            acc.IsJunior = false;
            acc.AvailableUsernameChanges = 3;
            if (string.IsNullOrEmpty(acc.ProfileImage))
                acc.ProfileImage = "DefaultImage.png";

            await _context.SaveChangesAsync();

            return Ok(new { error = (object)null, success = true, value = acc });
        }

        private async Task<string> ParseFieldFromBodyOrForm(string fieldName)
        {
            if (Request.HasFormContentType && Request.Form.ContainsKey(fieldName))
            {
                return Request.Form[fieldName];
            }
            using var reader = new StreamReader(Request.Body);
            var bodyString = await reader.ReadToEndAsync();
            try
            {
                var j = JsonSerializer.Deserialize<Dictionary<string, string>>(bodyString);
                if (j != null && j.ContainsKey(fieldName))
                {
                    return j[fieldName];
                }
            }
            catch
            {
                var parts = bodyString.Split('&');
                foreach (var part in parts)
                {
                    var kv = part.Split('=');
                    if (kv.Length == 2 && kv[0] == fieldName)
                    {
                        return System.Net.WebUtility.UrlDecode(kv[1]);
                    }
                }
            }
            return null;
        }

        [HttpPost("me/displayname")]
        [HttpPut("me/displayname")]
        public async Task<IActionResult> AccountUpdateDisplayName()
        {
            var accountId = GetAccountId();
            if (accountId == null) return Unauthorized("Unauthorized");

            var acc = await _context.Accounts.FirstOrDefaultAsync(a => a.AccountID == accountId);
            if (acc == null) return NotFound("Not Found");

            var name = await ParseFieldFromBodyOrForm("displayName");
            if (string.IsNullOrEmpty(name)) return BadRequest("displayName required");

            acc.DisplayName = name;
            await _context.SaveChangesAsync();

            return Ok(new { error = (object)null, success = true, value = acc });
        }

        [HttpPost("me/username")]
        [HttpPut("me/username")]
        public async Task<IActionResult> AccountUpdateUsername()
        {
            var accountId = GetAccountId();
            if (accountId == null) return Unauthorized("Unauthorized");

            var acc = await _context.Accounts.FirstOrDefaultAsync(a => a.AccountID == accountId);
            if (acc == null) return NotFound("Not Found");

            var name = await ParseFieldFromBodyOrForm("username");
            if (string.IsNullOrEmpty(name)) return BadRequest("Username required");

            acc.Username = name;
            acc.RawUsername = name;
            await _context.SaveChangesAsync();

            return Ok(new { error = (object)null, success = true, value = acc });
        }

        [HttpPost("me/changepassword")]
        [HttpPut("me/changepassword")]
        public async Task<IActionResult> AccountChangePassword()
        {
            var accountId = GetAccountId();
            if (accountId == null) return Unauthorized("Unauthorized");

            var acc = await _context.Accounts.FirstOrDefaultAsync(a => a.AccountID == accountId);
            if (acc == null) return NotFound("Not Found");

            var newPassword = await ParseFieldFromBodyOrForm("newPassword");
            if (string.IsNullOrEmpty(newPassword)) return BadRequest("Bad Request");

            acc.PasswordHash = "bcrypt_mock_" + newPassword;
            await _context.SaveChangesAsync();

            return Ok(new { error = (object)null, success = true, value = acc });
        }

        [HttpGet("search")]
        public async Task<IActionResult> AccountSearch([FromQuery] string name)
        {
            if (string.IsNullOrWhiteSpace(name) || name.Length < 2)
            {
                return Ok(new List<Account>());
            }

            var accounts = await _context.Accounts
                .Where(a => a.Username.ToLower().Contains(name.ToLower()))
                .Take(50)
                .ToListAsync();

            foreach (var acc in accounts)
            {
                if (string.IsNullOrEmpty(acc.ProfileImage))
                    acc.ProfileImage = "DefaultImage.png";
            }

            return Ok(accounts);
        }

        [HttpGet("bulk")]
        public async Task<IActionResult> AccountBulk([FromQuery] List<long> id)
        {
            if (id == null || id.Count == 0) return Ok(new List<Account>());

            var accounts = await _context.Accounts
                .Where(a => id.Contains(a.AccountID))
                .ToListAsync();

            foreach (var acc in accounts)
            {
                if (string.IsNullOrEmpty(acc.ProfileImage))
                    acc.ProfileImage = "DefaultImage.png";
            }

            return Ok(accounts);
        }

        [HttpGet("{id}")]
        public async Task<IActionResult> AccountGet(long id)
        {
            var acc = await _context.Accounts.FirstOrDefaultAsync(a => a.AccountID == id);
            if (acc == null) return NotFound("Not Found");

            if (string.IsNullOrEmpty(acc.ProfileImage))
                acc.ProfileImage = "DefaultImage.png";

            return Ok(acc);
        }

        [HttpPost("me/profileimage")]
        [HttpPut("me/profileimage")]
        public async Task<IActionResult> AccountProfileImage()
        {
            var accountId = GetAccountId();
            if (accountId == null) return Unauthorized();

            var account = await _context.Accounts.FirstOrDefaultAsync(a => a.AccountID == accountId);
            if (account == null) return NotFound();

            string imageName = "";
            if (Request.HasFormContentType && Request.Form.TryGetValue("imageName", out var val))
            {
                imageName = val.ToString();
            }
            else
            {
                using var reader = new StreamReader(Request.Body);
                var body = await reader.ReadToEndAsync();
                
                try {
                    var json = JsonDocument.Parse(body);
                    if (json.RootElement.TryGetProperty("imageName", out var jsonVal))
                        imageName = jsonVal.GetString();
                } catch {
                    var parts = body.Split('&');
                    foreach(var p in parts) {
                        var kv = p.Split('=');
                        if (kv.Length == 2 && kv[0] == "imageName")
                            imageName = System.Net.WebUtility.UrlDecode(kv[1]);
                    }
                }
            }

            if (!string.IsNullOrEmpty(imageName))
            {
                var originalPath = Path.Combine(_env.ContentRootPath, "uploads", "images", imageName);
                if (System.IO.File.Exists(originalPath))
                {
                    try
                    {
                        using var originalBitmap = SKBitmap.Decode(originalPath);
                        if (originalBitmap != null)
                        {
                            var size = Math.Min(originalBitmap.Width, originalBitmap.Height);
                            var x = (originalBitmap.Width - size) / 2;
                            var y = (originalBitmap.Height - size) / 2;

                            using var surface = SKSurface.Create(new SKImageInfo(512, 512));
                            var canvas = surface.Canvas;
                            canvas.Clear(SKColors.Transparent);

                            var srcRect = new SKRect(x, y, x + size, y + size);
                            var destRect = new SKRect(0, 0, 512, 512);

                            using var paint = new SKPaint { IsAntialias = true, FilterQuality = SKFilterQuality.High };
                            canvas.DrawBitmap(originalBitmap, srcRect, destRect, paint);

                            using var image = surface.Snapshot();
                            using var data = image.Encode(SKEncodedImageFormat.Png, 100);

                            var profileName = $"profile_{accountId}_{DateTimeOffset.UtcNow.ToUnixTimeMilliseconds()}_{Guid.NewGuid()}.png";
                            var profilePath = Path.Combine(_env.ContentRootPath, "uploads", "images", profileName);

                            using var stream = System.IO.File.OpenWrite(profilePath);
                            data.SaveTo(stream);
                            stream.Close();

                            account.ProfileImage = profileName;
                            await _context.SaveChangesAsync();

                            if (imageName != profileName)
                            {
                                try { System.IO.File.Delete(originalPath); } catch { }
                            }
                        }
                        else
                        {
                            account.ProfileImage = imageName;
                            await _context.SaveChangesAsync();
                        }
                    }
                    catch (Exception ex)
                    {
                        Console.WriteLine($"[ACCOUNT] profileimage process error: {ex}");
                        return BadRequest("invalid image");
                    }
                }
                else
                {
                    account.ProfileImage = imageName;
                    await _context.SaveChangesAsync();
                }
            }
            return Ok(new { success = true, ProfileImage = account.ProfileImage });
        }

        [HttpGet("developer/{id}")]
        public async Task<IActionResult> RoleCheck(long id)
        {
            var acc = await _context.Accounts.FirstOrDefaultAsync(a => a.AccountID == id);
            if (acc != null && (acc.IsDeveloper || acc.IsModerator))
            {
                return Content("true");
            }
            return Content("false");
        }

        [HttpGet("has_password")]
        public async Task<IActionResult> AccountHasPassword()
        {
            var accountId = GetAccountId();
            if (accountId == null) return Unauthorized("Unauthorized");

            var acc = await _context.Accounts.FirstOrDefaultAsync(a => a.AccountID == accountId);
            if (acc == null) return NotFound("Not Found");

            if (!string.IsNullOrEmpty(acc.PasswordHash))
                return Content("true");

            return Content("false");
        }

        [HttpPost("create")]
        public async Task<IActionResult> AccountCreate([FromForm] int platform, [FromForm] string platformId)
        {
            var pa = await _context.PlatformAccounts
                .Include(p => p.Account)
                .FirstOrDefaultAsync(p => p.Platform == platform && p.PlatformID == platformId);

            Account account;
            if (pa != null && pa.Account != null)
            {
                account = pa.Account;
            }
            else
            {
                account = new Account
                {
                    AccountID = new Random().Next(1000000, 999999999),
                    RawUsername = "Player_" + platformId,
                    Username = "player_" + platformId,
                    DisplayName = "Player",
                    Platforms = platform,
                    CreatedAt = DateTime.UtcNow,
                    ProfileImage = "DefaultImage.png"
                };
                _context.Accounts.Add(account);
                await _context.SaveChangesAsync();

                await Helpers.AccountHelpers.SetupNewAccountDefaults(_context, account);

                var newPa = new PlatformAccount
                {
                    AccountID = account.AccountID,
                    Platform = platform,
                    PlatformID = platformId
                };
                _context.PlatformAccounts.Add(newPa);
                await _context.SaveChangesAsync();
            }

            return Ok(new
            {
                error = (object)null,
                success = true,
                value = account
            });
        }

        [HttpGet("{id}/bio")]
        public async Task<IActionResult> AccountGetBio(long id)
        {
            var bio = await _context.PlayerBios.FirstOrDefaultAsync(b => b.AccountID == id);
            if (bio == null)
            {
                return Ok(new { accountId = id, bio = "" });
            }
            return Ok(new { accountId = bio.AccountID, bio = bio.Bio });
        }

        [HttpGet("me/bio")]
        public async Task<IActionResult> AccountGetBioMe()
        {
            var accountId = GetAccountId();
            if (accountId == null) return Unauthorized();
            return await AccountGetBio(accountId.Value);
        }

        [HttpPost("me/bio")]
        public async Task<IActionResult> AccountUpdateBio([FromForm] string bio)
        {
            var accountId = GetAccountId();
            if (accountId == null) return Unauthorized();

            var playerBio = await _context.PlayerBios.FirstOrDefaultAsync(b => b.AccountID == accountId.Value);
            if (playerBio == null)
            {
                playerBio = new PlayerBio { AccountID = accountId.Value, Bio = bio ?? "" };
                _context.PlayerBios.Add(playerBio);
            }
            else
            {
                playerBio.Bio = bio ?? "";
            }
            await _context.SaveChangesAsync();
            return Ok(new { accountId = playerBio.AccountID, bio = playerBio.Bio });
        }

        [HttpGet("me/haspassword")]
        public async Task<IActionResult> AccountHasPasswordMe()
        {
            var accountId = GetAccountId();
            if (accountId == null) return Unauthorized();

            var account = await _context.Accounts.FirstOrDefaultAsync(a => a.AccountID == accountId.Value);
            if (account == null) return NotFound();

            bool hasPassword = !string.IsNullOrEmpty(account.PasswordHash);
            return Ok(hasPassword);
        }

        [HttpGet("/namegen/options")]
        public IActionResult NamegenOptions()
        {
            return Ok(new
            {
                Adjectives = new[] { "Purrfect", "Fluffy", "Whiskered", "Sleek", "Playful", "Cuddly", "Mysterious" },
                Nouns = new[] { "Meow", "Cat", "Kitty", "Purr", "Whisker", "Claw", "Tail", "Fur" }
            });
        }
    }
}
