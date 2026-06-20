using Microsoft.AspNetCore.Mvc;

namespace MeowNet.API.Controllers
{
    [ApiController]
    [Route("api/sanitize/v1")]
    public class SanitizeController : ControllerBase
    {
        [HttpGet("isPure")]
        [HttpPost("isPure")]
        public IActionResult IsPure()
        {
            return Ok(new { isPure = true });
        }
    }
}
