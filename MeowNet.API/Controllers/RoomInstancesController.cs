using Microsoft.AspNetCore.Mvc;

namespace MeowNet.API.Controllers
{
    [ApiController]
    [Route("roominstance")]
    public class RoomInstancesController : ControllerBase
    {
        [HttpPost("{instanceId}/reportjoinresult")]
        public IActionResult ReportJoinResult(long instanceId)
        {
            return Ok(new { success = true });
        }
    }
}
