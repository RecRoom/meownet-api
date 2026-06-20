using Microsoft.AspNetCore.Mvc;

namespace MeowNet.API.Controllers
{
    [ApiController]
    [Route("api/itemWishlists/v1")]
    public class ItemWishlistsController : ControllerBase
    {
        [HttpGet("wishlist")]
        [HttpPost("wishlist")]
        [HttpDelete("wishlist")]
        public IActionResult WishlistDispatch() => Ok(new List<object>());
    }
}
