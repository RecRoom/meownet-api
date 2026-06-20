using Microsoft.AspNetCore.Mvc;
using MeowNet.API.Data;

namespace MeowNet.API.Controllers
{
    [ApiController]
    [Route("playlists")]
    public class PlaylistsController : ControllerBase
    {
        private readonly MeowNetDbContext _context;

        public PlaylistsController(MeowNetDbContext context)
        {
            _context = context;
        }

        [HttpGet("{id}")]
        public async Task<IActionResult> GetPlaylist(long id)
        {
            var playlist = await _context.RoomPlaylists.FindAsync(id);
            if (playlist == null) return NotFound();

            return Ok(playlist);
        }
    }
}
