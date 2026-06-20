using MeowNet.API.Data;
using Microsoft.AspNetCore.Mvc;

namespace MeowNet.API.Controllers
{
    [ApiController]
    public class SocialController : ControllerBase
    {
        private readonly MeowNetDbContext _context;

        public SocialController(MeowNetDbContext context)
        {
            _context = context;
        }

        [HttpGet("api/relationships/v2/get")]
        public IActionResult RelationshipsGet()
        {
            return Ok(Array.Empty<object>());
        }

        [HttpPost("api/relationships/v2/sendfriendrequest")]
        [HttpPost("api/relationships/v2/acceptfriendrequest")]
        [HttpPost("api/relationships/v2/removefriend")]
        [HttpPost("api/relationships/v2/addfriend")]
        [HttpPost("api/relationships/v1/favorite")]
        [HttpPost("api/relationships/v1/unfavorite")]
        [HttpPost("api/relationships/v1/mute")]
        [HttpPost("api/relationships/v1/unmute")]
        [HttpPost("api/relationships/v1/ignore")]
        [HttpPost("api/relationships/v1/unignore")]
        public IActionResult GenericRelationshipAction()
        {
            return Ok(new { Success = true });
        }

        [HttpGet("api/messages/v2/get")]
        public IActionResult MessagesGet()
        {
            return Ok(Array.Empty<object>());
        }

        [HttpPost("api/messages/v2/send")]
        public IActionResult SendMessage()
        {
            return Ok(new { Success = true });
        }

        [HttpPost("api/PlayerCheer/v1/create")]
        public IActionResult SendCheer()
        {
            return Ok(new { Success = true });
        }

        [HttpPost("api/gamesight/event")]
        public IActionResult GamesightEvent()
        {
            return Ok(new { Success = true });
        }

        [HttpGet("api/communityboard/v2/current")]
        public IActionResult CommunityBoard()
        {
            return Ok(Array.Empty<object>());
        }
        [HttpPost("api/PlayerCheer/v1/SetSelectedCheer")]
        public IActionResult SetSelectedCheer() => Ok(new { });


        [HttpPost("api/messages/v3/delete")]
        public IActionResult DeleteMessages() => Ok(new { });

        [HttpPost("invite")]
        public IActionResult Invite() => Ok(new { });

        [HttpGet("api/messages/v1/favoriteFriendOnlineStatus")]
        public IActionResult FavoriteFriendOnlineStatus() => Ok(new List<object>());

        [HttpPost("api/sanitize/v1")]
        public IActionResult Sanitize() => Ok(new { });

        [HttpGet("chat/thread")]
        public IActionResult ChatThread() => Ok(new List<object>());

        [HttpGet("announcements")]
        [HttpPost("announcements")]
        public IActionResult AnnouncementsUnread() => Ok(new List<object>());

        [HttpGet("api/announcement/v1/get")]
        public IActionResult AnnouncementGet() => Ok(new List<object>());

        [HttpGet("subscription/mine")]
        public IActionResult SubscriptionMine() => Ok(new List<object>());

        [HttpGet("subscription/details")]
        public IActionResult SubscriptionDetails() => Ok(new List<object>());

        [HttpGet("subscription/{**catchAll}")]
        [HttpPost("subscription/{**catchAll}")]
        public IActionResult SubscriptionDispatch() => Ok(new List<object>());
        
        [HttpGet("/thread")]
        public IActionResult GetThread([FromQuery] int maxCount, [FromQuery] int mode)
        {
            return Ok(new List<object>());
        }
    }
}
