using MeowNet.API.Data;
using Microsoft.EntityFrameworkCore;
using Microsoft.AspNetCore.RateLimiting;
using System.Threading.RateLimiting;
using System.Linq;

var builder = WebApplication.CreateBuilder(args);

// Add services to the container.
builder.Services.AddControllers()
    .AddJsonOptions(options =>
    {
        options.JsonSerializerOptions.PropertyNamingPolicy = null;
    });
builder.Services.AddEndpointsApiExplorer();
builder.Services.AddSignalR();
builder.Services.AddSingleton<MeowNet.API.Services.HubService>();
builder.Services.AddSingleton<MeowNet.API.Services.IStorageService, MeowNet.API.Services.LocalStorageService>();

// Add Database Context
var connectionString = builder.Configuration.GetConnectionString("DefaultConnection") ?? "Data Source=meownet.db";
builder.Services.AddDbContext<MeowNetDbContext>(options =>
    options.UseSqlite(connectionString));

// Add Response Compression (Gzip)
builder.Services.AddResponseCompression(options =>
{
    options.EnableForHttps = true;
});

// Add Rate Limiting
builder.Services.AddRateLimiter(options =>
{
    options.AddFixedWindowLimiter("Fixed", opt =>
    {
        opt.Window = TimeSpan.FromSeconds(1);
        opt.PermitLimit = 10;
        opt.QueueProcessingOrder = QueueProcessingOrder.OldestFirst;
        opt.QueueLimit = 2;
    });
});

var app = builder.Build();
            using (var scope = app.Services.CreateScope())
{
    var context = scope.ServiceProvider.GetRequiredService<MeowNet.API.Data.MeowNetDbContext>();
    context.Database.EnsureCreated();

    // Data migration: ensure all rooms support all platforms and DormRoom is marked as dorm
    var allRooms = context.Rooms.ToList();
    foreach (var r in allRooms)
    {
        r.SupportsScreens = true;
        r.SupportsWalkVR = true;
        r.SupportsTeleportVR = true;
        r.SupportsQuest2 = true;
        r.SupportsMobile = true;
        r.SupportsVRLow = true;
        r.SupportsJuniors = true;
        if (r.Name.ToLower() == "dormroom") r.IsDorm = true;
    }
    context.SaveChanges();

    if (!context.Accounts.Any(a => a.AccountID == 1))
    {
        context.Accounts.Add(new MeowNet.API.Models.Account
        {
            AccountID = 1,
            Username = "Coach",
            DisplayName = "Coach",
            RawUsername = "Coach",
            CreatedAt = DateTime.UtcNow,
            PasswordHash = "coach123",
            ProfileImage = "DefaultImage.png"
        });
        context.SaveChanges();
    }

    var defaultRoomsPath = System.IO.Path.Combine(AppContext.BaseDirectory, "Data", "defaultrooms.json");
    if (System.IO.File.Exists(defaultRoomsPath))
    {
        var jsonText = System.IO.File.ReadAllText(defaultRoomsPath);
        var rooms = System.Text.Json.JsonSerializer.Deserialize<List<MeowNet.API.Models.Room>>(jsonText);
        if (rooms != null)
        {
            foreach (var r in rooms)
            {
                if (!context.Rooms.Any(existing => existing.Name == r.Name))
                {
                    r.RoomId = 0; // Let SQLite autogenerate the ID
                    if (r.SubRooms != null)
                    {
                        foreach (var sr in r.SubRooms)
                        {
                            sr.RoomId = 0;
                            sr.SubRoomId = 0;
                        }
                    }

                    // SubRooms is NotMapped in EF, so we add manually
                    context.Rooms.Add(r);
                    context.SaveChanges();
                    
                    if (r.SubRooms != null)
                    {
                        foreach (var sr in r.SubRooms)
                        {
                            sr.RoomId = r.RoomId;
                            sr.SubRoomId = 0;
                            context.SubRooms.Add(sr);
                        }
                        context.SaveChanges();
                    }

                    // Tags is NotMapped in EF, so we add manually
                    if (r.Tags != null)
                    {
                        foreach(var tag in r.Tags)
                        {
                            tag.RoomId = r.RoomId;
                            context.RoomTags.Add(tag);
                        }
                        context.SaveChanges();
                    }
                }
            }
        }
    }
            }
app.Urls.Add("http://localhost:8080");

// Configure the HTTP request pipeline.
app.UseResponseCompression();
app.UseRateLimiter();

app.Use(async (context, next) =>
{
    if (context.Request.Path.StartsWithSegments("/hub/v1/negotiate") && context.Request.Method == "GET")
    {
        context.Request.Method = "POST";
    }
    await next();
});

app.Use(async (context, next) =>
{
    var token = context.Request.Headers["Authorization"].FirstOrDefault()?.Split(" ").Last();
    if (!string.IsNullOrEmpty(token))
    {
        var tokenHandler = new System.IdentityModel.Tokens.Jwt.JwtSecurityTokenHandler();
        try
        {
            var jwtToken = tokenHandler.ReadJwtToken(token);
            var claimsIdentity = new System.Security.Claims.ClaimsIdentity(jwtToken.Claims, "jwt");
            context.User = new System.Security.Claims.ClaimsPrincipal(claimsIdentity);
        }
        catch { }
    }
    await next();
});

app.UseAuthorization();

var uploadsPath = Path.Combine(builder.Environment.ContentRootPath, "uploads");
if (!Directory.Exists(uploadsPath)) Directory.CreateDirectory(uploadsPath);

var imagesPath = Path.Combine(uploadsPath, "images");
if (!Directory.Exists(imagesPath)) Directory.CreateDirectory(imagesPath);

app.UseStaticFiles(new StaticFileOptions
{
    FileProvider = new Microsoft.Extensions.FileProviders.PhysicalFileProvider(imagesPath),
    RequestPath = "",
    ServeUnknownFileTypes = true
});

var roomPath = Path.Combine(uploadsPath, "room");
if (!Directory.Exists(roomPath)) Directory.CreateDirectory(roomPath);

app.UseStaticFiles(new StaticFileOptions
{
    FileProvider = new Microsoft.Extensions.FileProviders.PhysicalFileProvider(roomPath),
    RequestPath = "",
    ServeUnknownFileTypes = true
});



app.UseStaticFiles(new StaticFileOptions
{
    FileProvider = new Microsoft.Extensions.FileProviders.PhysicalFileProvider(uploadsPath),
    RequestPath = "/api/images/v2",
    ServeUnknownFileTypes = true
});

app.MapControllers();
app.MapHub<MeowNet.API.Hubs.GameHub>("/hub/v1");

// Name server endpoint (root)
app.MapGet("/", () => 
{
    var host = builder.Configuration["Host"] ?? "http://localhost:8080";
    var cdnHost = builder.Configuration["CdnHost"] ?? host;

    return Results.Ok(new Dictionary<string, string>
    {
        { "Accounts", host },
        { "API", host },
        { "Auth", host },
        { "BugReporting", host },
        { "CDN", cdnHost },
        { "Chat", host },
        { "Clubs", host },
        { "CMS", host },
        { "Commerce", host },
        { "DataCollection", host },
        { "Discovery", host },
        { "Econ", host },
        { "GameLogs", host },
        { "Images", cdnHost },
        { "Leaderboard", host },
        { "Link", host },
        { "Lists", host },
        { "Matchmaking", host },
        { "Moderation", host },
        { "Notifications", host },
        { "PlatformNotifications", host },
        { "PlayerSettings", host },
        { "RoomComments", host },
        { "Rooms", host },
        { "Storage", host },
        { "Strings", host },
        { "StringsCDN", cdnHost },
        { "Studio", host },
        { "WWW", host }
    });
});

app.Run();
