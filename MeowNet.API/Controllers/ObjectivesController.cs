using Microsoft.AspNetCore.Mvc;

namespace MeowNet.API.Controllers
{
    [ApiController]
    [Route("api/objectives/v1")]
    public class ObjectivesController : ControllerBase
    {
        [HttpPost("cleargroup")]
        public IActionResult ClearGroup()
        {
            return Ok(new { success = true });
        }
    }
}
