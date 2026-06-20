using MeowNet.API.Data;
using Microsoft.AspNetCore.Mvc;

namespace MeowNet.API.Controllers
{
    [ApiController]
    public class StoreController : ControllerBase
    {
        private readonly MeowNetDbContext _context;

        public StoreController(MeowNetDbContext context)
        {
            _context = context;
        }

        [HttpGet("api/storefronts/v4/balance/{id?}")]
        [HttpGet("api/storefronts/v4/balance")]
        public IActionResult BalanceGet()
        {
            return Ok(new[] { new { CurrencyType = 1, Balance = 50000, BalanceType = 1 } });
        }

        [HttpGet("api/equipment/v2/getUnlocked")]
        public IActionResult EquipmentUnlocked()
        {
            return Ok(Array.Empty<object>());
        }

        [HttpGet("api/consumables/v2/getUnlocked")]
        public IActionResult ConsumablesUnlocked()
        {
            return Ok(Array.Empty<object>());
        }

        [HttpPost("api/equipment/v1/update")]
        public IActionResult EquipmentUpdate()
        {
            return Ok(Array.Empty<object>());
        }

        [HttpPost("api/storefronts/v2/buyItem")]
        public IActionResult BuyItem()
        {
            return Ok(new { Success = true });
        }

        [HttpPost("api/storefronts/v2/buyInvention")]
        public IActionResult BuyInvention()
        {
            return Ok(new { Success = true });
        }

        [HttpGet("api/storefronts/v1/adcarouselitems")]
        public IActionResult AdCarouselItems()
        {
            return Ok(Array.Empty<object>());
        }

        [HttpGet("api/storefronts/{**type}")]
        public IActionResult StorefrontByType(string type)
        {
            var parts = type?.Split('/') ?? Array.Empty<string>();
            var lastPart = parts.LastOrDefault() ?? "";
            
            if (int.TryParse(lastPart, out var id)) {
                return Ok(new { StorefrontType = id, Items = Array.Empty<object>() });
            }
            return Ok(new { StorefrontType = 0, Items = Array.Empty<object>() });
        }
    }
}
