using Microsoft.AspNetCore.SignalR;
using MeowNet.API.Data;
using MeowNet.API.Models;
using Microsoft.EntityFrameworkCore;
using System.Security.Claims;
using System.Text.Json;

namespace MeowNet.API.Hubs
{
    public class GameHub : Hub
    {
        private readonly IServiceProvider _serviceProvider;

        public GameHub(IServiceProvider serviceProvider)
        {
            _serviceProvider = serviceProvider;
        }

        private long? GetAccountId()
        {
            var tokenStr = Context.GetHttpContext()?.Request.Query["access_token"].ToString();
            
            var idClaim = Context.User?.FindFirstValue(ClaimTypes.NameIdentifier);
            if (long.TryParse(idClaim, out var id)) return id;
            
            // Temporary token decoder stub for dev
            if (!string.IsNullOrEmpty(tokenStr) && tokenStr.Split('.').Length == 3)
            {
                try
                {
                    var payload = tokenStr.Split('.')[1];
                    // Pad base64
                    switch (payload.Length % 4)
                    {
                        case 2: payload += "=="; break;
                        case 3: payload += "="; break;
                    }
                    var decoded = System.Text.Encoding.UTF8.GetString(Convert.FromBase64String(payload));
                    var doc = JsonDocument.Parse(decoded);
                    if (doc.RootElement.TryGetProperty("sub", out var subProp) && long.TryParse(subProp.GetString(), out var subId))
                    {
                        return subId;
                    }
                }
                catch { }
            }

            return null;
        }

        public override async Task OnConnectedAsync()
        {
            var accountId = GetAccountId();
            if (accountId != null)
            {
                var hubService = _serviceProvider.GetRequiredService<MeowNet.API.Services.HubService>();
                hubService.PlayerConnected(accountId.Value);

                await Groups.AddToGroupAsync(Context.ConnectionId, $"player_{accountId}");
                
                using var scope = _serviceProvider.CreateScope();
                var db = scope.ServiceProvider.GetRequiredService<MeowNetDbContext>();
                
                var selfAccount = await db.Accounts.FirstOrDefaultAsync(a => a.AccountID == accountId);
                if (selfAccount != null)
                {
                    selfAccount.AvailableUsernameChanges = 3;
                    await Clients.Caller.SendAsync("Notification", JsonSerializer.Serialize(new { Id = "SelfAccountUpdate", Msg = selfAccount }));
                    await Clients.Caller.SendAsync("Notification", JsonSerializer.Serialize(new { Id = "AccountUpdate", Msg = selfAccount }));
                }

                var presence = new {
                    playerId = accountId,
                    isOnline = true,
                    roomInstance = (object)null,
                    appVersion = "20210827"
                };
                await Clients.Caller.SendAsync("Notification", JsonSerializer.Serialize(new { Id = "PresenceUpdate", Msg = presence }));
            }
            
            await base.OnConnectedAsync();
        }

        public override async Task OnDisconnectedAsync(Exception? exception)
        {
            var accountId = GetAccountId();
            if (accountId != null)
            {
                var hubService = _serviceProvider.GetRequiredService<MeowNet.API.Services.HubService>();
                hubService.PlayerDisconnected(accountId.Value);

                using var scope = _serviceProvider.CreateScope();
                var db = scope.ServiceProvider.GetRequiredService<MeowNetDbContext>();
                
                var acc = await db.Accounts.FirstOrDefaultAsync(a => a.AccountID == accountId);
                if (acc != null)
                {
                    acc.LastOnline = DateTime.UtcNow;
                    db.Accounts.Update(acc);
                    await db.SaveChangesAsync();
                }

                var st = await db.PlayerStates.FirstOrDefaultAsync(p => p.AccountID == accountId);
                if (st != null)
                {
                    st.LoginLockToken = null;
                    db.PlayerStates.Update(st);
                    await db.SaveChangesAsync();
                }
            }

            await base.OnDisconnectedAsync(exception);
        }

        public async Task<object> heartbeat2()
        {
            var accountId = GetAccountId() ?? 0;
            var presence = new {
                playerId = accountId,
                isOnline = true,
                roomInstance = (object)null,
                appVersion = "20210827"
            };

            await Clients.Caller.SendAsync("Notification", JsonSerializer.Serialize(new { Id = "PresenceUpdate", Msg = presence }));
            
            return new { Id = accountId, Msg = presence };
        }

        [HubMethodName("playerSubscriptions/v1/update")]
        public async Task SubscribeToPlayersUpdate(object args)
        {
            await SubscribeToPlayers(args);
        }

        public async Task SubscribeToPlayers(object args)
        {
            var accountId = GetAccountId();
            if (accountId == null) return;

            // Typically we'd parse args to get the PlayerIds array and send their presences.
            // As a stub, we just return the caller's presence.
            await Clients.Caller.SendAsync("Notification", JsonSerializer.Serialize(new { 
                Id = "PresenceUpdate", 
                Msg = new {
                    playerId = accountId,
                    isOnline = true,
                    appVersion = "20210827"
                } 
            }));
        }
    }
}
