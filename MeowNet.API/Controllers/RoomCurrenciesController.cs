using Microsoft.AspNetCore.Mvc;

namespace MeowNet.API.Controllers
{
    [ApiController]
    [Route("api/roomcurrencies")]
    public class RoomCurrenciesController : ControllerBase
    {
        [HttpGet("")]
        public IActionResult GetCurrencies([FromQuery] long roomId)
        {
            return Ok(new List<object>());
        }
    }
}
