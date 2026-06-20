using Microsoft.AspNetCore.Mvc;

namespace MeowNet.API.Controllers
{
    [ApiController]
    public class ConfigController : ControllerBase
    {
        private readonly IWebHostEnvironment _env;

        public ConfigController(IWebHostEnvironment env)
        {
            _env = env;
        }

        [HttpGet("api/config/v1/amplitude")]
        public IActionResult Amplitude()
        {
            return Ok(new
            {
                AmplitudeKey = "",
                UseRudderStack = false,
                RudderStackKey = "",
                UseStatSig = false,
                StatSigKey = ""
            });
        }

        [HttpGet("api/versioncheck/{**path}")]
        public IActionResult VersionCheck()
        {
            return Ok(new { VersionStatus = 0 });
        }

        [HttpGet("api/gameconfigs/v1/all")]
        public async Task<IActionResult> GameConfigs()
        {
            var filePath = Path.Combine(_env.ContentRootPath, "..", "data", "jsons", "gameconfigs.json");
            if (!System.IO.File.Exists(filePath)) return NotFound();
            var json = await System.IO.File.ReadAllTextAsync(filePath);
            return Content(json, "application/json");
        }

        [HttpGet("api/config/v2")]
        public async Task<IActionResult> ConfigV2()
        {
            var filePath = Path.Combine(_env.ContentRootPath, "..", "data", "jsons", "configv2.json");
            if (!System.IO.File.Exists(filePath)) return Ok(new { error = "", success = true, value = new object() });
            var json = await System.IO.File.ReadAllTextAsync(filePath);
            
            // Assuming configv2 is a json object, we can append ServerMaintenance if needed.
            // For now, just serve the raw json.
            return Content(json, "application/json");
        }

        [HttpGet("config/LoadingScreenTipData")]
        public async Task<IActionResult> LoadingScreenTips()
        {
            var filePath = Path.Combine(_env.ContentRootPath, "..", "data", "jsons", "loadingscreentipdata.json");
            if (!System.IO.File.Exists(filePath)) return Ok(new List<object>());
            var json = await System.IO.File.ReadAllTextAsync(filePath);
            return Content(json, "application/json");
        }

        [HttpGet("/eac/challenge")]
        public IActionResult EACChallenge()
        {
            return Content($"\"{Guid.NewGuid()}\"");
        }
    }
}
