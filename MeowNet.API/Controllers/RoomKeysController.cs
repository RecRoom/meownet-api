using Microsoft.AspNetCore.Mvc;

namespace MeowNet.API.Controllers
{
    [ApiController]
    [Route("api/roomkeys")]
    public class RoomKeysController : ControllerBase
    {
        [HttpGet]
        [HttpPost]
        public IActionResult Dispatch() => Ok(new List<object>());
    }
}
