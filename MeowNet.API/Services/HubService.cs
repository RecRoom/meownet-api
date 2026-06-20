using Microsoft.AspNetCore.SignalR;
using MeowNet.API.Hubs;
using System.Text.Json;
using System.Threading.Tasks;

namespace MeowNet.API.Services
{
    public class HubService
    {
        private readonly IHubContext<GameHub> _hubContext;
        private readonly System.Collections.Concurrent.ConcurrentDictionary<long, int> _onlinePlayers = new();
        private readonly System.Collections.Concurrent.ConcurrentDictionary<long, long> _playerInstances = new();

        public HubService(IHubContext<GameHub> hubContext)
        {
            _hubContext = hubContext;
        }

        public void SetPlayerInstance(long playerId, long instanceId)
        {
            _playerInstances[playerId] = instanceId;
        }

        public long? GetPlayerInstance(long playerId)
        {
            if (_playerInstances.TryGetValue(playerId, out var instanceId)) return instanceId;
            return null;
        }

        public void ClearPlayerInstance(long playerId)
        {
            _playerInstances.TryRemove(playerId, out _);
        }

        public int LivePlayerCountInInstance(long instanceId, long excludePlayerId = 0)
        {
            return _playerInstances.Values.Count(v => v == instanceId);
        }

        public void PlayerConnected(long playerId)
        {
            _onlinePlayers.AddOrUpdate(playerId, 1, (key, val) => val + 1);
        }

        public void PlayerDisconnected(long playerId)
        {
            if (_onlinePlayers.TryGetValue(playerId, out var count))
            {
                if (count <= 1)
                    _onlinePlayers.TryRemove(playerId, out _);
                else
                    _onlinePlayers[playerId] = count - 1;
            }
        }

        public bool IsOnline(long playerId)
        {
            return _onlinePlayers.ContainsKey(playerId);
        }

        public async Task BroadcastPresenceAsync(long playerId, object presence)
        {
            var msg = new
            {
                Id = "PresenceUpdate",
                Msg = presence
            };
            await _hubContext.Clients.All.SendAsync("Notification", JsonSerializer.Serialize(msg));
        }

        public async Task SendToPlayerAsync(long playerId, string target, object data)
        {
            var msg = new
            {
                Id = target,
                Msg = data
            };
            await _hubContext.Clients.Group($"player_{playerId}").SendAsync("Notification", JsonSerializer.Serialize(msg));
        }
        
        public async Task BroadcastToAllAsync(string target, object data)
        {
            var msg = new
            {
                Id = target,
                Msg = data
            };
            await _hubContext.Clients.All.SendAsync("Notification", JsonSerializer.Serialize(msg));
        }
    }
}
