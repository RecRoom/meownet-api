using MeowNet.API.Services;
using Microsoft.AspNetCore.Mvc;

namespace MeowNet.API.Controllers
{
    [ApiController]
    public class UploadController : ControllerBase
    {
        private readonly IStorageService _storage;

        public UploadController(IStorageService storage)
        {
            _storage = storage;
        }

        private string GetPrefixFor(int fileType)
        {
            return fileType switch
            {
                1 => "Room",
                6 => "RoomMetaData",
                2 => "Holotar",
                3 => "Image",
                4 => "Video",
                5 => "Invention",
                _ => ""
            };
        }

        private string GetSubFolderFor(int fileType)
        {
            return fileType switch
            {
                1 or 6 => "room",
                3 => "images",
                2 => "data",
                4 => "video",
                5 => "invention",
                _ => ""
            };
        }

        [HttpPost("upload")]
        [RequestSizeLimit(100 * 1024 * 1024)] // 100MB limit for Room blobs
        public async Task<IActionResult> UploadData()
        {
            if (!Request.HasFormContentType) return BadRequest("Requires multipart/form-data");

            var form = await Request.ReadFormAsync();
            
            int fileType = 0;
            if (form.TryGetValue("FileType", out var ftStr) && int.TryParse(ftStr, out var ft))
            {
                fileType = ft;
            }

            string explicitName = "";
            if (form.TryGetValue("ImageName", out var n1)) explicitName = n1;
            else if (form.TryGetValue("FileName", out var n2)) explicitName = n2;
            else if (form.TryGetValue("Name", out var n3)) explicitName = n3;

            var file = form.Files.FirstOrDefault();
            string fileName = "";

            if (file != null)
            {
                var prefix = GetPrefixFor(fileType);
                if (string.IsNullOrEmpty(prefix)) return BadRequest("missing or unknown FileType");

                fileName = $"{prefix}{Guid.NewGuid()}{DateTime.UtcNow.Ticks}";
                var subFolder = GetSubFolderFor(fileType);

                using var stream = file.OpenReadStream();
                await _storage.SaveFileAsync(fileName, stream, subFolder);
            }
            else if (!string.IsNullOrEmpty(explicitName))
            {
                fileName = explicitName;
            }

            if (string.IsNullOrEmpty(fileName)) return BadRequest("missing filename or valid upload data");

            return Ok(new { filename = fileName });
        }
    }
}
