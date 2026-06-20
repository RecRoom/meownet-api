using MeowNet.API.Data;
using MeowNet.API.Models;
using MeowNet.API.Services;
using Microsoft.AspNetCore.Mvc;
using Microsoft.EntityFrameworkCore;
using System.Security.Claims;

namespace MeowNet.API.Controllers
{
    [ApiController]
    [Route("api/images")]
    public class ImagesController : ControllerBase
    {
        private readonly MeowNetDbContext _context;
        private readonly IStorageService _storage;

        public ImagesController(MeowNetDbContext context, IStorageService storage)
        {
            _context = context;
            _storage = storage;
        }

        private long? GetAccountId()
        {
            var idClaim = User.FindFirstValue(ClaimTypes.NameIdentifier) ?? User.FindFirstValue("sub") ?? User.FindFirstValue("nameid");
            if (long.TryParse(idClaim, out var id)) return id;
            return null;
        }

        [HttpGet("v2/named")]
        public IActionResult ImagesNamed()
        {
            return NotFound();
        }

        [HttpPost("v4/uploadsaved")]
        public async Task<IActionResult> UploadSaved([FromForm] string? ImageName, [FromForm] string? imgmeta)
        {
            var accountId = GetAccountId();
            if (accountId == null) return Unauthorized();

            var file = Request.Form.Files.FirstOrDefault();
            if (file != null)
            {
                var ts = DateTime.UtcNow.ToString("yyyyMMddTHHmmss.fff");
                ImageName = $"img_{accountId}_{ts}_{Guid.NewGuid().ToString().Substring(0,8)}.png";

                using var stream = file.OpenReadStream();
                await _storage.SaveFileAsync(ImageName, stream, "images");
            }

            if (string.IsNullOrEmpty(ImageName)) return BadRequest("missing ImageName");

            // Normally parse imgmeta to check RoomID and tagged PlayerIDs. Stubbed here.
            var img = await _context.UserImages.FirstOrDefaultAsync(u => u.ImageName == ImageName);
            if (img == null)
            {
                img = new UserImage
                {
                    AccountID = (uint)accountId.Value,
                    ImageName = ImageName,
                    IsSaved = true,
                    CreatedAt = DateTime.UtcNow
                };
                _context.UserImages.Add(img);
            }
            else
            {
                img.AccountID = (uint)accountId.Value;
                img.IsSaved = true;
                _context.UserImages.Update(img);
            }

            var photo = new UploadedPhoto
            {
                AccountID = (uint)accountId.Value,
                ImageName = ImageName,
                Type = 1, // Default (1 = ShareCamera)
                PlayerIDs = new List<int>(),
                CreatedAt = DateTime.UtcNow
            };
            _context.UploadedPhotos.Add(photo);
            await _context.SaveChangesAsync();

            return Ok(new { ImageName = ImageName, Success = true });
        }

        [HttpGet("v4/room/{roomId:int}")]
        public async Task<IActionResult> RoomImages(int roomId, [FromQuery] int take = 100, [FromQuery] int skip = 0)
        {
            var photos = await _context.UploadedPhotos
                .Where(p => p.RoomID == roomId)
                .OrderByDescending(p => p.CreatedAt)
                .Skip(skip)
                .Take(take)
                .ToListAsync();

            var outPhotos = photos.Select(p => new
            {
                Accessibility = p.Accessibility,
                AccessibilityLocked = false,
                CheerCount = p.CheerCount,
                CommentCount = 0,
                CreatedAt = p.CreatedAt,
                Id = p.ID,
                ImageName = p.ImageName,
                PlayerEventId = p.PlayerEventID,
                PlayerId = p.AccountID,
                RoomId = p.RoomID,
                TaggedPlayerIds = p.PlayerIDs ?? new List<int>(),
                Type = p.Type
            });

            return Ok(outPhotos);
        }

        public class CheerReq
        {
            public bool Cheer { get; set; }
            public uint SavedImageId { get; set; }
        }

        [HttpPost("v1/cheer")]
        public async Task<IActionResult> ImageCheer([FromBody] CheerReq req)
        {
            var accountId = GetAccountId();
            if (accountId == null) return Unauthorized();

            var photo = await _context.UploadedPhotos.FirstOrDefaultAsync(p => p.ID == req.SavedImageId);
            if (photo == null) return Ok(new { Message = "image not found", Success = false });

            var existing = await _context.UploadedPhotoCheers.FirstOrDefaultAsync(c => c.PhotoId == req.SavedImageId && c.AccountId == accountId.Value);

            if (req.Cheer && existing == null)
            {
                _context.UploadedPhotoCheers.Add(new UploadedPhotoCheer { PhotoId = req.SavedImageId, AccountId = (uint)accountId.Value });
                photo.CheerCount++;
            }
            else if (!req.Cheer && existing != null)
            {
                _context.UploadedPhotoCheers.Remove(existing);
                if (photo.CheerCount > 0) photo.CheerCount--;
            }

            await _context.SaveChangesAsync();
            return Ok(new { Message = "", Success = true });
        }
    }
}
