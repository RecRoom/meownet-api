using MeowNet.API.Data;
using Microsoft.AspNetCore.Mvc;
using Microsoft.EntityFrameworkCore;
using System.Security.Claims;

namespace MeowNet.API.Controllers
{
    [ApiController]
    public class SubscriptionsController : ControllerBase
    {
        private readonly MeowNetDbContext _context;

        public SubscriptionsController(MeowNetDbContext context)
        {
            _context = context;
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

        [HttpGet("/api/playersubscriptions/v1/my")]
        public IActionResult PlayerSubscriptionsMy()
        {
            return Ok(Array.Empty<object>());
        }

        [HttpGet("/subscription/mine")]
        [HttpGet("/subscription/mine/{accountId}")]
        public IActionResult SubscriptionMine()
        {
            return Ok(Array.Empty<object>());
        }

        [HttpGet("/subscription/details/{accountId}")]
        public async Task<IActionResult> SubscriptionDetails(long accountId)
        {
            // Simulate subscription details
            return Ok(new { accountId = accountId, clubId = 0, subscriberCount = 0 });
        }

        [HttpGet("/subscription/subscriberCount/{accountId}")]
        public IActionResult SubscriptionSubscriberCount(long accountId)
        {
            return Ok(0);
        }

        [HttpPost("/subscription/{accountId}")]
        public IActionResult SubscriptionSubscribe(long accountId)
        {
            return Ok(new { success = true });
        }

        [HttpDelete("/subscription/{accountId}")]
        public IActionResult SubscriptionUnsubscribe(long accountId)
        {
            return Ok(new { success = true });
        }
    }
}
