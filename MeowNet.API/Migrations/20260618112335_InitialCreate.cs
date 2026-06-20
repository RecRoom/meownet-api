using System;
using Microsoft.EntityFrameworkCore.Migrations;

#nullable disable

namespace MeowNet.API.Migrations
{
    /// <inheritdoc />
    public partial class InitialCreate : Migration
    {
        /// <inheritdoc />
        protected override void Up(MigrationBuilder migrationBuilder)
        {
            migrationBuilder.CreateTable(
                name: "AccountBans",
                columns: table => new
                {
                    ID = table.Column<long>(type: "INTEGER", nullable: false)
                        .Annotation("Sqlite:Autoincrement", true),
                    AccountID = table.Column<long>(type: "INTEGER", nullable: false),
                    Reason = table.Column<string>(type: "TEXT", nullable: false),
                    Message = table.Column<string>(type: "TEXT", nullable: false),
                    is_ban = table.Column<bool>(type: "BOOLEAN", nullable: false),
                    BannedBy = table.Column<string>(type: "TEXT", nullable: false),
                    CreatedAt = table.Column<DateTime>(type: "DATETIME", nullable: false),
                    ExpiresAt = table.Column<DateTime>(type: "DATETIME", nullable: true)
                },
                constraints: table =>
                {
                    table.PrimaryKey("PK_AccountBans", x => x.ID);
                });

            migrationBuilder.CreateTable(
                name: "Accounts",
                columns: table => new
                {
                    account_id = table.Column<long>(type: "INTEGER", nullable: false)
                        .Annotation("Sqlite:Autoincrement", true),
                    created_at = table.Column<DateTime>(type: "DATETIME", nullable: false),
                    display_name = table.Column<string>(type: "TEXT", nullable: false),
                    is_junior = table.Column<bool>(type: "BOOLEAN", nullable: true),
                    platforms = table.Column<int>(type: "INTEGER", nullable: false),
                    profile_image = table.Column<string>(type: "TEXT", nullable: false),
                    username = table.Column<string>(type: "TEXT", nullable: false),
                    raw_username = table.Column<string>(type: "TEXT", nullable: false),
                    treat_as_junior = table.Column<bool>(type: "BOOLEAN", nullable: false),
                    has_birthday = table.Column<bool>(type: "BOOLEAN", nullable: false),
                    is_developer = table.Column<bool>(type: "BOOLEAN", nullable: false),
                    is_moderator = table.Column<bool>(type: "BOOLEAN", nullable: false),
                    password_hash = table.Column<string>(type: "TEXT", nullable: true),
                    home_club_id = table.Column<long>(type: "INTEGER", nullable: true),
                    selected_cheer = table.Column<int>(type: "INTEGER", nullable: false),
                    last_online = table.Column<DateTime>(type: "DATETIME", nullable: true),
                    no_token = table.Column<bool>(type: "BOOLEAN", nullable: false),
                    email = table.Column<string>(type: "TEXT", nullable: true),
                    phone = table.Column<string>(type: "TEXT", nullable: true),
                    birthday = table.Column<string>(type: "TEXT", nullable: true),
                    junior_state = table.Column<int>(type: "INTEGER", nullable: false),
                    parent_account_id = table.Column<long>(type: "INTEGER", nullable: true)
                },
                constraints: table =>
                {
                    table.PrimaryKey("PK_Accounts", x => x.account_id);
                });

            migrationBuilder.CreateTable(
                name: "Announcements",
                columns: table => new
                {
                    announcement_id = table.Column<long>(type: "INTEGER", nullable: false)
                        .Annotation("Sqlite:Autoincrement", true),
                    announcement_type = table.Column<int>(type: "INTEGER", nullable: false),
                    body = table.Column<string>(type: "TEXT", nullable: false),
                    created_at = table.Column<DateTime>(type: "DATETIME", nullable: false),
                    image_name = table.Column<string>(type: "TEXT", nullable: false),
                    link_name = table.Column<string>(type: "TEXT", nullable: false),
                    link_type = table.Column<int>(type: "INTEGER", nullable: false),
                    link_uri = table.Column<string>(type: "TEXT", nullable: false),
                    platform = table.Column<int>(type: "INTEGER", nullable: false),
                    title = table.Column<string>(type: "TEXT", nullable: false)
                },
                constraints: table =>
                {
                    table.PrimaryKey("PK_Announcements", x => x.announcement_id);
                });

            migrationBuilder.CreateTable(
                name: "AvatarItems",
                columns: table => new
                {
                    avatar_item_desc = table.Column<string>(type: "TEXT", nullable: false),
                    avatar_item_type = table.Column<int>(type: "INTEGER", nullable: false),
                    friendly_name = table.Column<string>(type: "TEXT", nullable: false),
                    tool_tip = table.Column<string>(type: "TEXT", nullable: false),
                    rarity = table.Column<int>(type: "INTEGER", nullable: false)
                },
                constraints: table =>
                {
                    table.PrimaryKey("PK_AvatarItems", x => x.avatar_item_desc);
                });

            migrationBuilder.CreateTable(
                name: "Avatars",
                columns: table => new
                {
                    account_id = table.Column<long>(type: "INTEGER", nullable: false),
                    face_features = table.Column<string>(type: "TEXT", nullable: false),
                    hair_color = table.Column<string>(type: "TEXT", nullable: false),
                    outfit_selections = table.Column<string>(type: "TEXT", nullable: false),
                    skin_color = table.Column<string>(type: "TEXT", nullable: false)
                },
                constraints: table =>
                {
                    table.PrimaryKey("PK_Avatars", x => x.account_id);
                });

            migrationBuilder.CreateTable(
                name: "ClubAnnouncements",
                columns: table => new
                {
                    announcement_id = table.Column<long>(type: "INTEGER", nullable: false)
                        .Annotation("Sqlite:Autoincrement", true),
                    club_id = table.Column<long>(type: "INTEGER", nullable: false),
                    account_id = table.Column<int>(type: "INTEGER", nullable: false),
                    title = table.Column<string>(type: "TEXT", nullable: false),
                    body = table.Column<string>(type: "TEXT", nullable: false),
                    image_name = table.Column<string>(type: "TEXT", nullable: false),
                    meta = table.Column<string>(type: "TEXT", nullable: false),
                    created_at = table.Column<DateTime>(type: "DATETIME", nullable: false)
                },
                constraints: table =>
                {
                    table.PrimaryKey("PK_ClubAnnouncements", x => x.announcement_id);
                });

            migrationBuilder.CreateTable(
                name: "ClubCustomTags",
                columns: table => new
                {
                    id = table.Column<long>(type: "INTEGER", nullable: false)
                        .Annotation("Sqlite:Autoincrement", true),
                    club_id = table.Column<long>(type: "INTEGER", nullable: false),
                    tag = table.Column<string>(type: "TEXT", nullable: false)
                },
                constraints: table =>
                {
                    table.PrimaryKey("PK_ClubCustomTags", x => x.id);
                });

            migrationBuilder.CreateTable(
                name: "ClubMembers",
                columns: table => new
                {
                    club_member_id = table.Column<long>(type: "INTEGER", nullable: false)
                        .Annotation("Sqlite:Autoincrement", true),
                    club_id = table.Column<long>(type: "INTEGER", nullable: false),
                    account_id = table.Column<int>(type: "INTEGER", nullable: false),
                    membership_type = table.Column<int>(type: "INTEGER", nullable: false),
                    created_at = table.Column<DateTime>(type: "DATETIME", nullable: false)
                },
                constraints: table =>
                {
                    table.PrimaryKey("PK_ClubMembers", x => x.club_member_id);
                });

            migrationBuilder.CreateTable(
                name: "ClubPermissions",
                columns: table => new
                {
                    club_permissions_id = table.Column<long>(type: "INTEGER", nullable: false)
                        .Annotation("Sqlite:Autoincrement", true),
                    club_id = table.Column<long>(type: "INTEGER", nullable: false),
                    type = table.Column<int>(type: "INTEGER", nullable: false),
                    approve_member = table.Column<bool>(type: "BOOLEAN", nullable: false),
                    ban_unban = table.Column<bool>(type: "BOOLEAN", nullable: false),
                    create_event = table.Column<bool>(type: "BOOLEAN", nullable: false),
                    edit_details = table.Column<bool>(type: "BOOLEAN", nullable: false),
                    edit_permission_settings = table.Column<bool>(type: "BOOLEAN", nullable: false),
                    post_announcement = table.Column<bool>(type: "BOOLEAN", nullable: false)
                },
                constraints: table =>
                {
                    table.PrimaryKey("PK_ClubPermissions", x => x.club_permissions_id);
                });

            migrationBuilder.CreateTable(
                name: "ClubReports",
                columns: table => new
                {
                    ID = table.Column<long>(type: "INTEGER", nullable: false)
                        .Annotation("Sqlite:Autoincrement", true),
                    reporter_id = table.Column<long>(type: "INTEGER", nullable: false),
                    club_id = table.Column<long>(type: "INTEGER", nullable: false),
                    report_category = table.Column<int>(type: "INTEGER", nullable: false),
                    details = table.Column<string>(type: "TEXT", nullable: false),
                    created_at = table.Column<DateTime>(type: "DATETIME", nullable: false),
                    resolved = table.Column<bool>(type: "BOOLEAN", nullable: false)
                },
                constraints: table =>
                {
                    table.PrimaryKey("PK_ClubReports", x => x.ID);
                });

            migrationBuilder.CreateTable(
                name: "Clubs",
                columns: table => new
                {
                    club_id = table.Column<long>(type: "INTEGER", nullable: false)
                        .Annotation("Sqlite:Autoincrement", true),
                    name = table.Column<string>(type: "TEXT", nullable: false),
                    description = table.Column<string>(type: "TEXT", nullable: false),
                    category = table.Column<string>(type: "TEXT", nullable: false),
                    visibility = table.Column<int>(type: "INTEGER", nullable: false),
                    joinability = table.Column<int>(type: "INTEGER", nullable: false),
                    allow_juniors = table.Column<bool>(type: "BOOLEAN", nullable: false),
                    main_image_name = table.Column<string>(type: "TEXT", nullable: false),
                    club_type = table.Column<int>(type: "INTEGER", nullable: false),
                    clubhouse_room_id = table.Column<long>(type: "INTEGER", nullable: true),
                    creator_account_id = table.Column<int>(type: "INTEGER", nullable: false),
                    is_rro = table.Column<bool>(type: "BOOLEAN", nullable: false),
                    min_level = table.Column<int>(type: "INTEGER", nullable: false),
                    state = table.Column<int>(type: "INTEGER", nullable: false),
                    member_count = table.Column<int>(type: "INTEGER", nullable: false),
                    created_at = table.Column<DateTime>(type: "DATETIME", nullable: false)
                },
                constraints: table =>
                {
                    table.PrimaryKey("PK_Clubs", x => x.club_id);
                });

            migrationBuilder.CreateTable(
                name: "DailyXpLedgers",
                columns: table => new
                {
                    id = table.Column<long>(type: "INTEGER", nullable: false)
                        .Annotation("Sqlite:Autoincrement", true),
                    account_id = table.Column<long>(type: "INTEGER", nullable: false),
                    day = table.Column<DateTime>(type: "DATETIME", nullable: false),
                    xp = table.Column<int>(type: "INTEGER", nullable: false),
                    updated_at = table.Column<DateTime>(type: "DATETIME", nullable: false)
                },
                constraints: table =>
                {
                    table.PrimaryKey("PK_DailyXpLedgers", x => x.id);
                });

            migrationBuilder.CreateTable(
                name: "DeviceBans",
                columns: table => new
                {
                    ID = table.Column<long>(type: "INTEGER", nullable: false)
                        .Annotation("Sqlite:Autoincrement", true),
                    DeviceID = table.Column<string>(type: "TEXT", nullable: false),
                    AccountID = table.Column<long>(type: "INTEGER", nullable: false),
                    BanID = table.Column<long>(type: "INTEGER", nullable: false),
                    Reason = table.Column<string>(type: "TEXT", nullable: false),
                    BannedBy = table.Column<string>(type: "TEXT", nullable: false),
                    CreatedAt = table.Column<DateTime>(type: "DATETIME", nullable: false),
                    ExpiresAt = table.Column<DateTime>(type: "DATETIME", nullable: true)
                },
                constraints: table =>
                {
                    table.PrimaryKey("PK_DeviceBans", x => x.ID);
                });

            migrationBuilder.CreateTable(
                name: "DeviceLogins",
                columns: table => new
                {
                    ID = table.Column<long>(type: "INTEGER", nullable: false)
                        .Annotation("Sqlite:Autoincrement", true),
                    AccountID = table.Column<long>(type: "INTEGER", nullable: false),
                    DeviceID = table.Column<string>(type: "TEXT", nullable: false),
                    DeviceClass = table.Column<int>(type: "INTEGER", nullable: false),
                    platform_id = table.Column<string>(type: "TEXT", nullable: false),
                    platform = table.Column<string>(type: "TEXT", nullable: false),
                    ip = table.Column<string>(type: "TEXT", nullable: false),
                    LoginCount = table.Column<int>(type: "INTEGER", nullable: false),
                    FirstSeen = table.Column<DateTime>(type: "DATETIME", nullable: false),
                    LastSeen = table.Column<DateTime>(type: "DATETIME", nullable: false)
                },
                constraints: table =>
                {
                    table.PrimaryKey("PK_DeviceLogins", x => x.ID);
                });

            migrationBuilder.CreateTable(
                name: "FeaturedRoomGroups",
                columns: table => new
                {
                    id = table.Column<long>(type: "INTEGER", nullable: false)
                        .Annotation("Sqlite:Autoincrement", true),
                    name = table.Column<string>(type: "TEXT", nullable: false),
                    sort_order = table.Column<int>(type: "INTEGER", nullable: false)
                },
                constraints: table =>
                {
                    table.PrimaryKey("PK_FeaturedRoomGroups", x => x.id);
                });

            migrationBuilder.CreateTable(
                name: "Gifts",
                columns: table => new
                {
                    id = table.Column<long>(type: "INTEGER", nullable: false)
                        .Annotation("Sqlite:Autoincrement", true),
                    account_id = table.Column<long>(type: "INTEGER", nullable: false),
                    from_player_id = table.Column<long>(type: "INTEGER", nullable: false),
                    message = table.Column<string>(type: "TEXT", nullable: false),
                    avatar_item_desc = table.Column<string>(type: "TEXT", nullable: false),
                    avatar_item_type = table.Column<int>(type: "INTEGER", nullable: false),
                    consumable_item_desc = table.Column<string>(type: "TEXT", nullable: false),
                    equipment_prefab_name = table.Column<string>(type: "TEXT", nullable: false),
                    equipment_modification_guid = table.Column<string>(type: "TEXT", nullable: false),
                    currency = table.Column<int>(type: "INTEGER", nullable: false),
                    currency_type = table.Column<int>(type: "INTEGER", nullable: false),
                    balance_type = table.Column<int>(type: "INTEGER", nullable: false),
                    level = table.Column<int>(type: "INTEGER", nullable: false),
                    xp = table.Column<int>(type: "INTEGER", nullable: false),
                    gift_context = table.Column<int>(type: "INTEGER", nullable: false),
                    gift_rarity = table.Column<int>(type: "INTEGER", nullable: false),
                    platform = table.Column<int>(type: "INTEGER", nullable: false),
                    platforms_to_spawn_on = table.Column<int>(type: "INTEGER", nullable: false),
                    consumed = table.Column<bool>(type: "BOOLEAN", nullable: false),
                    created_at = table.Column<DateTime>(type: "DATETIME", nullable: false)
                },
                constraints: table =>
                {
                    table.PrimaryKey("PK_Gifts", x => x.id);
                });

            migrationBuilder.CreateTable(
                name: "InstanceBans",
                columns: table => new
                {
                    ID = table.Column<long>(type: "INTEGER", nullable: false)
                        .Annotation("Sqlite:Autoincrement", true),
                    instance_id = table.Column<long>(type: "INTEGER", nullable: false),
                    account_id = table.Column<long>(type: "INTEGER", nullable: false),
                    issued_by = table.Column<long>(type: "INTEGER", nullable: false),
                    expires_at = table.Column<DateTime>(type: "DATETIME", nullable: false),
                    created_at = table.Column<DateTime>(type: "DATETIME", nullable: false)
                },
                constraints: table =>
                {
                    table.PrimaryKey("PK_InstanceBans", x => x.ID);
                });

            migrationBuilder.CreateTable(
                name: "InstanceInvites",
                columns: table => new
                {
                    instance_id = table.Column<long>(type: "INTEGER", nullable: false),
                    account_id = table.Column<int>(type: "INTEGER", nullable: false),
                    invited_by = table.Column<int>(type: "INTEGER", nullable: false),
                    created_at = table.Column<DateTime>(type: "DATETIME", nullable: false)
                },
                constraints: table =>
                {
                    table.PrimaryKey("PK_InstanceInvites", x => new { x.instance_id, x.account_id });
                });

            migrationBuilder.CreateTable(
                name: "InventionCheers",
                columns: table => new
                {
                    id = table.Column<long>(type: "INTEGER", nullable: false)
                        .Annotation("Sqlite:Autoincrement", true),
                    invention_id = table.Column<long>(type: "INTEGER", nullable: false),
                    account_id = table.Column<long>(type: "INTEGER", nullable: false),
                    created_at = table.Column<DateTime>(type: "DATETIME", nullable: false)
                },
                constraints: table =>
                {
                    table.PrimaryKey("PK_InventionCheers", x => x.id);
                });

            migrationBuilder.CreateTable(
                name: "InventionOwnerships",
                columns: table => new
                {
                    id = table.Column<long>(type: "INTEGER", nullable: false)
                        .Annotation("Sqlite:Autoincrement", true),
                    invention_id = table.Column<long>(type: "INTEGER", nullable: false),
                    account_id = table.Column<long>(type: "INTEGER", nullable: false),
                    acquired_at = table.Column<DateTime>(type: "DATETIME", nullable: false)
                },
                constraints: table =>
                {
                    table.PrimaryKey("PK_InventionOwnerships", x => x.id);
                });

            migrationBuilder.CreateTable(
                name: "InventionReports",
                columns: table => new
                {
                    ID = table.Column<long>(type: "INTEGER", nullable: false)
                        .Annotation("Sqlite:Autoincrement", true),
                    reporter_id = table.Column<long>(type: "INTEGER", nullable: false),
                    invention_id = table.Column<long>(type: "INTEGER", nullable: false),
                    report_category = table.Column<int>(type: "INTEGER", nullable: false),
                    details = table.Column<string>(type: "TEXT", nullable: false),
                    created_at = table.Column<DateTime>(type: "DATETIME", nullable: false),
                    resolved = table.Column<bool>(type: "BOOLEAN", nullable: false)
                },
                constraints: table =>
                {
                    table.PrimaryKey("PK_InventionReports", x => x.ID);
                });

            migrationBuilder.CreateTable(
                name: "Inventions",
                columns: table => new
                {
                    invention_id = table.Column<long>(type: "INTEGER", nullable: false)
                        .Annotation("Sqlite:Autoincrement", true),
                    name = table.Column<string>(type: "TEXT", nullable: false),
                    description = table.Column<string>(type: "TEXT", nullable: false),
                    image_name = table.Column<string>(type: "TEXT", nullable: false),
                    creator_player_id = table.Column<int>(type: "INTEGER", nullable: false),
                    creator_permission = table.Column<int>(type: "INTEGER", nullable: false),
                    general_permission = table.Column<int>(type: "INTEGER", nullable: false),
                    allow_trial = table.Column<bool>(type: "BOOLEAN", nullable: false),
                    hide_from_player = table.Column<bool>(type: "BOOLEAN", nullable: false),
                    is_ag_invention = table.Column<bool>(type: "BOOLEAN", nullable: false),
                    is_certified_invention = table.Column<bool>(type: "BOOLEAN", nullable: false),
                    is_published = table.Column<bool>(type: "BOOLEAN", nullable: false),
                    price = table.Column<int>(type: "INTEGER", nullable: false),
                    cheer_count = table.Column<int>(type: "INTEGER", nullable: false),
                    num_downloads = table.Column<int>(type: "INTEGER", nullable: false),
                    num_players_used_in_room = table.Column<int>(type: "INTEGER", nullable: false),
                    current_version_number = table.Column<int>(type: "INTEGER", nullable: false),
                    replication_id = table.Column<string>(type: "TEXT", nullable: false),
                    created_at = table.Column<DateTime>(type: "DATETIME", nullable: false),
                    modified_at = table.Column<DateTime>(type: "DATETIME", nullable: false)
                },
                constraints: table =>
                {
                    table.PrimaryKey("PK_Inventions", x => x.invention_id);
                });

            migrationBuilder.CreateTable(
                name: "InventionTags",
                columns: table => new
                {
                    id = table.Column<long>(type: "INTEGER", nullable: false)
                        .Annotation("Sqlite:Autoincrement", true),
                    invention_id = table.Column<long>(type: "INTEGER", nullable: false),
                    tag = table.Column<string>(type: "TEXT", nullable: false),
                    type = table.Column<int>(type: "INTEGER", nullable: false)
                },
                constraints: table =>
                {
                    table.PrimaryKey("PK_InventionTags", x => x.id);
                });

            migrationBuilder.CreateTable(
                name: "InventionVersions",
                columns: table => new
                {
                    id = table.Column<long>(type: "INTEGER", nullable: false)
                        .Annotation("Sqlite:Autoincrement", true),
                    invention_id = table.Column<long>(type: "INTEGER", nullable: false),
                    version_number = table.Column<int>(type: "INTEGER", nullable: false),
                    blob_name = table.Column<string>(type: "TEXT", nullable: false),
                    chips_cost = table.Column<int>(type: "INTEGER", nullable: false),
                    cloud_variables_cost = table.Column<int>(type: "INTEGER", nullable: false),
                    instantiation_cost = table.Column<int>(type: "INTEGER", nullable: false),
                    lights_cost = table.Column<int>(type: "INTEGER", nullable: false),
                    replication_id = table.Column<string>(type: "TEXT", nullable: false)
                },
                constraints: table =>
                {
                    table.PrimaryKey("PK_InventionVersions", x => x.id);
                });

            migrationBuilder.CreateTable(
                name: "LeaderboardStats",
                columns: table => new
                {
                    id = table.Column<long>(type: "INTEGER", nullable: false)
                        .Annotation("Sqlite:Autoincrement", true),
                    account_id = table.Column<long>(type: "INTEGER", nullable: false),
                    room_id = table.Column<int>(type: "INTEGER", nullable: false),
                    stat_channel = table.Column<int>(type: "INTEGER", nullable: false),
                    score = table.Column<int>(type: "INTEGER", nullable: false),
                    updated_at = table.Column<DateTime>(type: "DATETIME", nullable: false)
                },
                constraints: table =>
                {
                    table.PrimaryKey("PK_LeaderboardStats", x => x.id);
                });

            migrationBuilder.CreateTable(
                name: "Messages",
                columns: table => new
                {
                    Id = table.Column<long>(type: "INTEGER", nullable: false)
                        .Annotation("Sqlite:Autoincrement", true),
                    from_player_id = table.Column<long>(type: "INTEGER", nullable: false),
                    to_player_id = table.Column<long>(type: "INTEGER", nullable: false),
                    sent_time = table.Column<DateTime>(type: "DATETIME", nullable: false),
                    type = table.Column<int>(type: "INTEGER", nullable: false),
                    room_id = table.Column<long>(type: "INTEGER", nullable: true),
                    invention_id = table.Column<long>(type: "INTEGER", nullable: true),
                    player_event_id = table.Column<long>(type: "INTEGER", nullable: true),
                    data = table.Column<string>(type: "TEXT", nullable: false)
                },
                constraints: table =>
                {
                    table.PrimaryKey("PK_Messages", x => x.Id);
                });

            migrationBuilder.CreateTable(
                name: "ModerationBlocks",
                columns: table => new
                {
                    ID = table.Column<long>(type: "INTEGER", nullable: false)
                        .Annotation("Sqlite:Autoincrement", true),
                    account_id = table.Column<long>(type: "INTEGER", nullable: false),
                    reporter_id = table.Column<long>(type: "INTEGER", nullable: true),
                    game_session_id = table.Column<long>(type: "INTEGER", nullable: false),
                    is_ban = table.Column<bool>(type: "BOOLEAN", nullable: false),
                    is_host_kick = table.Column<bool>(type: "BOOLEAN", nullable: false),
                    message = table.Column<string>(type: "TEXT", nullable: true),
                    report_category = table.Column<int>(type: "INTEGER", nullable: false),
                    duration = table.Column<int>(type: "INTEGER", nullable: false),
                    created_at = table.Column<DateTime>(type: "DATETIME", nullable: false),
                    expires_at = table.Column<DateTime>(type: "DATETIME", nullable: true)
                },
                constraints: table =>
                {
                    table.PrimaryKey("PK_ModerationBlocks", x => x.ID);
                });

            migrationBuilder.CreateTable(
                name: "ModerationReports",
                columns: table => new
                {
                    ID = table.Column<long>(type: "INTEGER", nullable: false)
                        .Annotation("Sqlite:Autoincrement", true),
                    reporter_id = table.Column<long>(type: "INTEGER", nullable: false),
                    target_id = table.Column<long>(type: "INTEGER", nullable: false),
                    report_category = table.Column<int>(type: "INTEGER", nullable: false),
                    message = table.Column<string>(type: "TEXT", nullable: false),
                    game_session_id = table.Column<long>(type: "INTEGER", nullable: false),
                    source = table.Column<string>(type: "TEXT", nullable: false),
                    created_at = table.Column<DateTime>(type: "DATETIME", nullable: false),
                    resolved = table.Column<bool>(type: "BOOLEAN", nullable: false)
                },
                constraints: table =>
                {
                    table.PrimaryKey("PK_ModerationReports", x => x.ID);
                });

            migrationBuilder.CreateTable(
                name: "PlatformAccountLimits",
                columns: table => new
                {
                    platform = table.Column<int>(type: "INTEGER", nullable: false),
                    platform_id = table.Column<string>(type: "TEXT", nullable: false),
                    max_accounts = table.Column<int>(type: "INTEGER", nullable: false)
                },
                constraints: table =>
                {
                    table.PrimaryKey("PK_PlatformAccountLimits", x => new { x.platform, x.platform_id });
                });

            migrationBuilder.CreateTable(
                name: "PlayerBios",
                columns: table => new
                {
                    account_id = table.Column<long>(type: "INTEGER", nullable: false),
                    bio = table.Column<string>(type: "TEXT", nullable: false)
                },
                constraints: table =>
                {
                    table.PrimaryKey("PK_PlayerBios", x => x.account_id);
                });

            migrationBuilder.CreateTable(
                name: "PlayerCheers",
                columns: table => new
                {
                    id = table.Column<long>(type: "INTEGER", nullable: false)
                        .Annotation("Sqlite:Autoincrement", true),
                    from_account_id = table.Column<long>(type: "INTEGER", nullable: false),
                    to_account_id = table.Column<long>(type: "INTEGER", nullable: false),
                    category = table.Column<int>(type: "INTEGER", nullable: false),
                    room_id = table.Column<long>(type: "INTEGER", nullable: true),
                    anonymous = table.Column<bool>(type: "BOOLEAN", nullable: false),
                    created_at = table.Column<DateTime>(type: "DATETIME", nullable: false)
                },
                constraints: table =>
                {
                    table.PrimaryKey("PK_PlayerCheers", x => x.id);
                });

            migrationBuilder.CreateTable(
                name: "PlayerEventResponses",
                columns: table => new
                {
                    player_event_response_id = table.Column<long>(type: "INTEGER", nullable: false)
                        .Annotation("Sqlite:Autoincrement", true),
                    player_event_id = table.Column<long>(type: "INTEGER", nullable: false),
                    player_id = table.Column<long>(type: "INTEGER", nullable: false),
                    type = table.Column<int>(type: "INTEGER", nullable: false),
                    created_at = table.Column<DateTime>(type: "DATETIME", nullable: false)
                },
                constraints: table =>
                {
                    table.PrimaryKey("PK_PlayerEventResponses", x => x.player_event_response_id);
                });

            migrationBuilder.CreateTable(
                name: "PlayerEvents",
                columns: table => new
                {
                    player_event_id = table.Column<long>(type: "INTEGER", nullable: false)
                        .Annotation("Sqlite:Autoincrement", true),
                    creator_player_id = table.Column<long>(type: "INTEGER", nullable: false),
                    room_id = table.Column<long>(type: "INTEGER", nullable: false),
                    sub_room_id = table.Column<long>(type: "INTEGER", nullable: true),
                    club_id = table.Column<long>(type: "INTEGER", nullable: true),
                    name = table.Column<string>(type: "TEXT", nullable: false),
                    description = table.Column<string>(type: "TEXT", nullable: false),
                    image_name = table.Column<string>(type: "TEXT", nullable: true),
                    start_time = table.Column<DateTime>(type: "DATETIME", nullable: false),
                    end_time = table.Column<DateTime>(type: "DATETIME", nullable: false),
                    accessibility = table.Column<int>(type: "INTEGER", nullable: false),
                    state = table.Column<int>(type: "INTEGER", nullable: false),
                    attendee_count = table.Column<int>(type: "INTEGER", nullable: false)
                },
                constraints: table =>
                {
                    table.PrimaryKey("PK_PlayerEvents", x => x.player_event_id);
                });

            migrationBuilder.CreateTable(
                name: "PlayerReports",
                columns: table => new
                {
                    ID = table.Column<long>(type: "INTEGER", nullable: false)
                        .Annotation("Sqlite:Autoincrement", true),
                    reporter_id = table.Column<long>(type: "INTEGER", nullable: false),
                    reported_player_id = table.Column<long>(type: "INTEGER", nullable: false),
                    report_category = table.Column<int>(type: "INTEGER", nullable: false),
                    details = table.Column<string>(type: "TEXT", nullable: false),
                    height_reporter = table.Column<double>(type: "REAL", nullable: false),
                    height_reported = table.Column<double>(type: "REAL", nullable: false),
                    room_id = table.Column<long>(type: "INTEGER", nullable: false),
                    room_instance_type = table.Column<int>(type: "INTEGER", nullable: false),
                    created_at = table.Column<DateTime>(type: "DATETIME", nullable: false),
                    resolved = table.Column<bool>(type: "BOOLEAN", nullable: false)
                },
                constraints: table =>
                {
                    table.PrimaryKey("PK_PlayerReports", x => x.ID);
                });

            migrationBuilder.CreateTable(
                name: "PlayerStates",
                columns: table => new
                {
                    account_id = table.Column<long>(type: "INTEGER", nullable: false),
                    status_visibility = table.Column<int>(type: "INTEGER", nullable: false),
                    vr_movement_mode = table.Column<int>(type: "INTEGER", nullable: false),
                    avoid_juniors = table.Column<bool>(type: "BOOLEAN", nullable: false),
                    login_lock_token = table.Column<string>(type: "TEXT", nullable: true)
                },
                constraints: table =>
                {
                    table.PrimaryKey("PK_PlayerStates", x => x.account_id);
                });

            migrationBuilder.CreateTable(
                name: "Progressions",
                columns: table => new
                {
                    account_id = table.Column<long>(type: "INTEGER", nullable: false),
                    level = table.Column<int>(type: "INTEGER", nullable: false),
                    xp = table.Column<int>(type: "INTEGER", nullable: false)
                },
                constraints: table =>
                {
                    table.PrimaryKey("PK_Progressions", x => x.account_id);
                });

            migrationBuilder.CreateTable(
                name: "RefreshTokens",
                columns: table => new
                {
                    ID = table.Column<long>(type: "INTEGER", nullable: false)
                        .Annotation("Sqlite:Autoincrement", true),
                    Token = table.Column<string>(type: "TEXT", nullable: false),
                    AccountID = table.Column<long>(type: "INTEGER", nullable: false),
                    platform_id = table.Column<string>(type: "TEXT", nullable: false),
                    platform = table.Column<string>(type: "TEXT", nullable: false),
                    ExpiresAt = table.Column<DateTime>(type: "DATETIME", nullable: false),
                    used_at = table.Column<DateTime>(type: "DATETIME", nullable: true),
                    CreatedAt = table.Column<DateTime>(type: "DATETIME", nullable: false)
                },
                constraints: table =>
                {
                    table.PrimaryKey("PK_RefreshTokens", x => x.ID);
                });

            migrationBuilder.CreateTable(
                name: "Relationships",
                columns: table => new
                {
                    ID = table.Column<long>(type: "INTEGER", nullable: false)
                        .Annotation("Sqlite:Autoincrement", true),
                    requester_id = table.Column<long>(type: "INTEGER", nullable: false),
                    target_id = table.Column<long>(type: "INTEGER", nullable: false),
                    relationship_type = table.Column<int>(type: "INTEGER", nullable: false),
                    requester_favorited = table.Column<int>(type: "INTEGER", nullable: false),
                    requester_ignored = table.Column<int>(type: "INTEGER", nullable: false),
                    requester_muted = table.Column<int>(type: "INTEGER", nullable: false),
                    target_favorited = table.Column<int>(type: "INTEGER", nullable: false),
                    target_ignored = table.Column<int>(type: "INTEGER", nullable: false),
                    target_muted = table.Column<int>(type: "INTEGER", nullable: false)
                },
                constraints: table =>
                {
                    table.PrimaryKey("PK_Relationships", x => x.ID);
                });

            migrationBuilder.CreateTable(
                name: "RewardDrops",
                columns: table => new
                {
                    gift_drop_id = table.Column<int>(type: "INTEGER", nullable: false)
                        .Annotation("Sqlite:Autoincrement", true),
                    friendly_name = table.Column<string>(type: "TEXT", nullable: false),
                    tooltip = table.Column<string>(type: "TEXT", nullable: false),
                    avatar_item_desc = table.Column<string>(type: "TEXT", nullable: true),
                    avatar_item_type = table.Column<int>(type: "INTEGER", nullable: false),
                    consumable_item_desc = table.Column<string>(type: "TEXT", nullable: true),
                    equipment_prefab_name = table.Column<string>(type: "TEXT", nullable: true),
                    equipment_modification_guid = table.Column<string>(type: "TEXT", nullable: true),
                    is_query = table.Column<bool>(type: "BOOLEAN", nullable: false),
                    unique = table.Column<bool>(type: "BOOLEAN", nullable: false),
                    subscribers_only = table.Column<bool>(type: "BOOLEAN", nullable: false),
                    level = table.Column<int>(type: "INTEGER", nullable: false),
                    rarity = table.Column<int>(type: "INTEGER", nullable: false),
                    currency_type = table.Column<int>(type: "INTEGER", nullable: false),
                    currency = table.Column<int>(type: "INTEGER", nullable: false),
                    context = table.Column<int>(type: "INTEGER", nullable: false),
                    item_set_id = table.Column<int>(type: "INTEGER", nullable: true),
                    item_set_friendly_name = table.Column<string>(type: "TEXT", nullable: true)
                },
                constraints: table =>
                {
                    table.PrimaryKey("PK_RewardDrops", x => x.gift_drop_id);
                });

            migrationBuilder.CreateTable(
                name: "RewardSelections",
                columns: table => new
                {
                    id = table.Column<long>(type: "INTEGER", nullable: false)
                        .Annotation("Sqlite:Autoincrement", true),
                    account_id = table.Column<long>(type: "INTEGER", nullable: false),
                    message = table.Column<string>(type: "TEXT", nullable: false),
                    gift_context = table.Column<int>(type: "INTEGER", nullable: false),
                    reward_type = table.Column<int>(type: "INTEGER", nullable: false),
                    gift_drop_1_id = table.Column<int>(type: "INTEGER", nullable: false),
                    gift_drop_2_id = table.Column<int>(type: "INTEGER", nullable: false),
                    gift_drop_3_id = table.Column<int>(type: "INTEGER", nullable: false),
                    consumed = table.Column<bool>(type: "BOOLEAN", nullable: false),
                    created_at = table.Column<DateTime>(type: "DATETIME", nullable: false)
                },
                constraints: table =>
                {
                    table.PrimaryKey("PK_RewardSelections", x => x.id);
                });

            migrationBuilder.CreateTable(
                name: "RoomInstances",
                columns: table => new
                {
                    Id = table.Column<long>(type: "INTEGER", nullable: false)
                        .Annotation("Sqlite:Autoincrement", true),
                    owner_account_id = table.Column<int>(type: "INTEGER", nullable: false),
                    room_id = table.Column<long>(type: "INTEGER", nullable: false),
                    sub_room_id = table.Column<long>(type: "INTEGER", nullable: false),
                    location = table.Column<string>(type: "TEXT", nullable: false),
                    data_blob = table.Column<string>(type: "TEXT", nullable: false),
                    event_id = table.Column<long>(type: "INTEGER", nullable: false),
                    photon_region_id = table.Column<string>(type: "TEXT", nullable: false),
                    photon_room_id = table.Column<string>(type: "TEXT", nullable: false),
                    name = table.Column<string>(type: "TEXT", nullable: false),
                    max_capacity = table.Column<int>(type: "INTEGER", nullable: false),
                    is_full = table.Column<bool>(type: "BOOLEAN", nullable: false),
                    is_private = table.Column<bool>(type: "BOOLEAN", nullable: false),
                    is_in_progress = table.Column<bool>(type: "BOOLEAN", nullable: false),
                    room_code = table.Column<string>(type: "TEXT", nullable: false),
                    room_instance_type = table.Column<int>(type: "INTEGER", nullable: false),
                    club_id = table.Column<long>(type: "INTEGER", nullable: false),
                    encrypt_voice_chat = table.Column<bool>(type: "BOOLEAN", nullable: false),
                    matchmaking_policy = table.Column<int>(type: "INTEGER", nullable: false),
                    allow_new_users = table.Column<bool>(type: "BOOLEAN", nullable: false),
                    join_disabled = table.Column<bool>(type: "BOOLEAN", nullable: false),
                    created_at = table.Column<DateTime>(type: "DATETIME", nullable: false)
                },
                constraints: table =>
                {
                    table.PrimaryKey("PK_RoomInstances", x => x.Id);
                });

            migrationBuilder.CreateTable(
                name: "RoomInteractions",
                columns: table => new
                {
                    AccountId = table.Column<long>(type: "INTEGER", nullable: false),
                    RoomId = table.Column<long>(type: "INTEGER", nullable: false),
                    Cheered = table.Column<bool>(type: "BOOLEAN", nullable: false),
                    Favorited = table.Column<bool>(type: "BOOLEAN", nullable: false),
                    Visited = table.Column<bool>(type: "BOOLEAN", nullable: false)
                },
                constraints: table =>
                {
                    table.PrimaryKey("PK_RoomInteractions", x => new { x.AccountId, x.RoomId });
                });

            migrationBuilder.CreateTable(
                name: "RoomPlaylists",
                columns: table => new
                {
                    id = table.Column<long>(type: "INTEGER", nullable: false)
                        .Annotation("Sqlite:Autoincrement", true),
                    name = table.Column<string>(type: "TEXT", nullable: false),
                    description = table.Column<string>(type: "TEXT", nullable: false),
                    image_name = table.Column<string>(type: "TEXT", nullable: false),
                    warning_mask = table.Column<int>(type: "INTEGER", nullable: false),
                    custom_warning = table.Column<string>(type: "TEXT", nullable: false),
                    creator_account_id = table.Column<int>(type: "INTEGER", nullable: false),
                    state = table.Column<int>(type: "INTEGER", nullable: false),
                    accessibility = table.Column<int>(type: "INTEGER", nullable: false),
                    supports_level_voting = table.Column<bool>(type: "BOOLEAN", nullable: false),
                    is_rro = table.Column<bool>(type: "BOOLEAN", nullable: false),
                    supports_screens = table.Column<bool>(type: "BOOLEAN", nullable: false),
                    supports_walk_vr = table.Column<bool>(type: "BOOLEAN", nullable: false),
                    supports_teleport_vr = table.Column<bool>(type: "BOOLEAN", nullable: false),
                    supports_vr_low = table.Column<bool>(type: "BOOLEAN", nullable: false),
                    supports_quest_2 = table.Column<bool>(type: "BOOLEAN", nullable: false),
                    supports_mobile = table.Column<bool>(type: "BOOLEAN", nullable: false),
                    supports_juniors = table.Column<bool>(type: "BOOLEAN", nullable: false),
                    min_level = table.Column<int>(type: "INTEGER", nullable: false),
                    created_at = table.Column<DateTime>(type: "DATETIME", nullable: false),
                    sort_order = table.Column<int>(type: "INTEGER", nullable: false)
                },
                constraints: table =>
                {
                    table.PrimaryKey("PK_RoomPlaylists", x => x.id);
                });

            migrationBuilder.CreateTable(
                name: "RoomReports",
                columns: table => new
                {
                    ID = table.Column<long>(type: "INTEGER", nullable: false)
                        .Annotation("Sqlite:Autoincrement", true),
                    reporter_id = table.Column<long>(type: "INTEGER", nullable: false),
                    room_id = table.Column<long>(type: "INTEGER", nullable: false),
                    report_category = table.Column<int>(type: "INTEGER", nullable: false),
                    details = table.Column<string>(type: "TEXT", nullable: false),
                    created_at = table.Column<DateTime>(type: "DATETIME", nullable: false),
                    resolved = table.Column<bool>(type: "BOOLEAN", nullable: false)
                },
                constraints: table =>
                {
                    table.PrimaryKey("PK_RoomReports", x => x.ID);
                });

            migrationBuilder.CreateTable(
                name: "RoomRoleEntrys",
                columns: table => new
                {
                    Id = table.Column<long>(type: "INTEGER", nullable: false)
                        .Annotation("Sqlite:Autoincrement", true),
                    room_id = table.Column<long>(type: "INTEGER", nullable: false),
                    account_id = table.Column<int>(type: "INTEGER", nullable: false),
                    invited_role = table.Column<int>(type: "INTEGER", nullable: false),
                    role = table.Column<int>(type: "INTEGER", nullable: false)
                },
                constraints: table =>
                {
                    table.PrimaryKey("PK_RoomRoleEntrys", x => x.Id);
                });

            migrationBuilder.CreateTable(
                name: "Rooms",
                columns: table => new
                {
                    room_id = table.Column<long>(type: "INTEGER", nullable: false)
                        .Annotation("Sqlite:Autoincrement", true),
                    name = table.Column<string>(type: "TEXT", nullable: false),
                    description = table.Column<string>(type: "TEXT", nullable: false),
                    image_name = table.Column<string>(type: "TEXT", nullable: false),
                    creator_account_id = table.Column<int>(type: "INTEGER", nullable: false),
                    state = table.Column<int>(type: "INTEGER", nullable: false),
                    accessibility = table.Column<int>(type: "INTEGER", nullable: false),
                    auto_localize_room = table.Column<bool>(type: "BOOLEAN", nullable: false),
                    cloning_allowed = table.Column<bool>(type: "BOOLEAN", nullable: false),
                    custom_warning = table.Column<string>(type: "TEXT", nullable: false),
                    disable_mic_auto_mute = table.Column<bool>(type: "BOOLEAN", nullable: false),
                    disable_room_comments = table.Column<bool>(type: "BOOLEAN", nullable: false),
                    encrypt_voice_chat = table.Column<bool>(type: "BOOLEAN", nullable: false),
                    is_developer_owned = table.Column<bool>(type: "BOOLEAN", nullable: false),
                    is_dorm = table.Column<bool>(type: "BOOLEAN", nullable: false),
                    is_rro = table.Column<bool>(type: "BOOLEAN", nullable: false),
                    load_screen_locked = table.Column<bool>(type: "BOOLEAN", nullable: false),
                    max_player_calculation_mode = table.Column<int>(type: "INTEGER", nullable: false),
                    max_players = table.Column<int>(type: "INTEGER", nullable: false),
                    min_level = table.Column<int>(type: "INTEGER", nullable: false),
                    persistence_version = table.Column<int>(type: "INTEGER", nullable: false),
                    ranked_entity_id = table.Column<string>(type: "TEXT", nullable: false),
                    ranking_context = table.Column<int>(type: "INTEGER", nullable: false),
                    supports_juniors = table.Column<bool>(type: "BOOLEAN", nullable: false),
                    supports_level_voting = table.Column<bool>(type: "BOOLEAN", nullable: false),
                    supports_mobile = table.Column<bool>(type: "BOOLEAN", nullable: false),
                    supports_quest_2 = table.Column<bool>(type: "BOOLEAN", nullable: false),
                    supports_screens = table.Column<bool>(type: "BOOLEAN", nullable: false),
                    supports_teleport_vr = table.Column<bool>(type: "BOOLEAN", nullable: false),
                    supports_vr_low = table.Column<bool>(type: "BOOLEAN", nullable: false),
                    supports_walk_vr = table.Column<bool>(type: "BOOLEAN", nullable: false),
                    toxmod_enabled = table.Column<bool>(type: "BOOLEAN", nullable: false),
                    ugc_version = table.Column<int>(type: "INTEGER", nullable: false),
                    warning_mask = table.Column<int>(type: "INTEGER", nullable: false),
                    data_blob = table.Column<string>(type: "TEXT", nullable: true),
                    created_at = table.Column<DateTime>(type: "DATETIME", nullable: false)
                },
                constraints: table =>
                {
                    table.PrimaryKey("PK_Rooms", x => x.room_id);
                });

            migrationBuilder.CreateTable(
                name: "RoomTags",
                columns: table => new
                {
                    Id = table.Column<long>(type: "INTEGER", nullable: false)
                        .Annotation("Sqlite:Autoincrement", true),
                    room_id = table.Column<long>(type: "INTEGER", nullable: false),
                    tag = table.Column<string>(type: "TEXT", nullable: false),
                    type = table.Column<int>(type: "INTEGER", nullable: false)
                },
                constraints: table =>
                {
                    table.PrimaryKey("PK_RoomTags", x => x.Id);
                });

            migrationBuilder.CreateTable(
                name: "ScreenShareReports",
                columns: table => new
                {
                    ID = table.Column<long>(type: "INTEGER", nullable: false)
                        .Annotation("Sqlite:Autoincrement", true),
                    reporter_id = table.Column<long>(type: "INTEGER", nullable: false),
                    reported_player_id = table.Column<long>(type: "INTEGER", nullable: false),
                    room_id = table.Column<long>(type: "INTEGER", nullable: false),
                    room_instance_id = table.Column<long>(type: "INTEGER", nullable: false),
                    room_instance_type = table.Column<int>(type: "INTEGER", nullable: false),
                    image_name = table.Column<string>(type: "TEXT", nullable: false),
                    details = table.Column<string>(type: "TEXT", nullable: false),
                    created_at = table.Column<DateTime>(type: "DATETIME", nullable: false),
                    resolved = table.Column<bool>(type: "BOOLEAN", nullable: false)
                },
                constraints: table =>
                {
                    table.PrimaryKey("PK_ScreenShareReports", x => x.ID);
                });

            migrationBuilder.CreateTable(
                name: "SubRoomDataHistorys",
                columns: table => new
                {
                    id = table.Column<long>(type: "INTEGER", nullable: false)
                        .Annotation("Sqlite:Autoincrement", true),
                    sub_room_id = table.Column<int>(type: "INTEGER", nullable: false),
                    data_blob = table.Column<string>(type: "TEXT", nullable: false),
                    saved_by_account_id = table.Column<int>(type: "INTEGER", nullable: false),
                    created_at = table.Column<DateTime>(type: "DATETIME", nullable: false)
                },
                constraints: table =>
                {
                    table.PrimaryKey("PK_SubRoomDataHistorys", x => x.id);
                });

            migrationBuilder.CreateTable(
                name: "SubRooms",
                columns: table => new
                {
                    sub_room_id = table.Column<long>(type: "INTEGER", nullable: false)
                        .Annotation("Sqlite:Autoincrement", true),
                    room_id = table.Column<long>(type: "INTEGER", nullable: false),
                    accessibility = table.Column<int>(type: "INTEGER", nullable: false),
                    data_blob = table.Column<string>(type: "TEXT", nullable: false),
                    is_sandbox = table.Column<bool>(type: "BOOLEAN", nullable: false),
                    max_players = table.Column<int>(type: "INTEGER", nullable: false),
                    name = table.Column<string>(type: "TEXT", nullable: false),
                    saved_by_account_id = table.Column<int>(type: "INTEGER", nullable: false),
                    unity_scene_id = table.Column<string>(type: "TEXT", nullable: false)
                },
                constraints: table =>
                {
                    table.PrimaryKey("PK_SubRooms", x => x.sub_room_id);
                });

            migrationBuilder.CreateTable(
                name: "UploadedPhotoCheers",
                columns: table => new
                {
                    id = table.Column<long>(type: "INTEGER", nullable: false)
                        .Annotation("Sqlite:Autoincrement", true),
                    photo_id = table.Column<long>(type: "INTEGER", nullable: false),
                    account_id = table.Column<long>(type: "INTEGER", nullable: false),
                    created_at = table.Column<DateTime>(type: "DATETIME", nullable: false)
                },
                constraints: table =>
                {
                    table.PrimaryKey("PK_UploadedPhotoCheers", x => x.id);
                });

            migrationBuilder.CreateTable(
                name: "UploadedPhotos",
                columns: table => new
                {
                    id = table.Column<long>(type: "INTEGER", nullable: false)
                        .Annotation("Sqlite:Autoincrement", true),
                    account_id = table.Column<long>(type: "INTEGER", nullable: false),
                    image_name = table.Column<string>(type: "TEXT", nullable: false),
                    player_ids = table.Column<string>(type: "TEXT", nullable: false),
                    saved_image_type = table.Column<int>(type: "INTEGER", nullable: false),
                    room_id = table.Column<int>(type: "INTEGER", nullable: false),
                    player_event_id = table.Column<int>(type: "INTEGER", nullable: false),
                    accessibility = table.Column<int>(type: "INTEGER", nullable: false),
                    cheer_count = table.Column<int>(type: "INTEGER", nullable: false),
                    created_at = table.Column<DateTime>(type: "DATETIME", nullable: false)
                },
                constraints: table =>
                {
                    table.PrimaryKey("PK_UploadedPhotos", x => x.id);
                });

            migrationBuilder.CreateTable(
                name: "UserAvatarItems",
                columns: table => new
                {
                    account_id = table.Column<long>(type: "INTEGER", nullable: false),
                    avatar_item_desc = table.Column<string>(type: "TEXT", nullable: false),
                    created_at = table.Column<DateTime>(type: "DATETIME", nullable: false)
                },
                constraints: table =>
                {
                    table.PrimaryKey("PK_UserAvatarItems", x => new { x.account_id, x.avatar_item_desc });
                });

            migrationBuilder.CreateTable(
                name: "UserConsumables",
                columns: table => new
                {
                    id = table.Column<long>(type: "INTEGER", nullable: false)
                        .Annotation("Sqlite:Autoincrement", true),
                    account_id = table.Column<long>(type: "INTEGER", nullable: false),
                    consumable_item_desc = table.Column<string>(type: "TEXT", nullable: false),
                    active_duration_minutes = table.Column<int>(type: "INTEGER", nullable: false),
                    initial_count = table.Column<int>(type: "INTEGER", nullable: false),
                    is_active = table.Column<bool>(type: "BOOLEAN", nullable: false),
                    is_transferable = table.Column<bool>(type: "BOOLEAN", nullable: false),
                    created_at = table.Column<DateTime>(type: "DATETIME", nullable: false)
                },
                constraints: table =>
                {
                    table.PrimaryKey("PK_UserConsumables", x => x.id);
                });

            migrationBuilder.CreateTable(
                name: "UserEquipments",
                columns: table => new
                {
                    id = table.Column<long>(type: "INTEGER", nullable: false)
                        .Annotation("Sqlite:Autoincrement", true),
                    account_id = table.Column<long>(type: "INTEGER", nullable: false),
                    modification_guid = table.Column<string>(type: "TEXT", nullable: false),
                    prefab_name = table.Column<string>(type: "TEXT", nullable: false),
                    friendly_name = table.Column<string>(type: "TEXT", nullable: false),
                    tooltip = table.Column<string>(type: "TEXT", nullable: false),
                    rarity = table.Column<int>(type: "INTEGER", nullable: false),
                    favorited = table.Column<bool>(type: "BOOLEAN", nullable: false)
                },
                constraints: table =>
                {
                    table.PrimaryKey("PK_UserEquipments", x => x.id);
                });

            migrationBuilder.CreateTable(
                name: "UserImages",
                columns: table => new
                {
                    id = table.Column<long>(type: "INTEGER", nullable: false)
                        .Annotation("Sqlite:Autoincrement", true),
                    account_id = table.Column<long>(type: "INTEGER", nullable: false),
                    image_name = table.Column<string>(type: "TEXT", nullable: false),
                    is_saved = table.Column<bool>(type: "BOOLEAN", nullable: false),
                    created_at = table.Column<DateTime>(type: "DATETIME", nullable: false)
                },
                constraints: table =>
                {
                    table.PrimaryKey("PK_UserImages", x => x.id);
                });

            migrationBuilder.CreateTable(
                name: "WishlistItems",
                columns: table => new
                {
                    wishlist_item_id = table.Column<string>(type: "TEXT", nullable: false),
                    account_id = table.Column<int>(type: "INTEGER", nullable: false),
                    purchasable_item_id = table.Column<int>(type: "INTEGER", nullable: false),
                    created_at = table.Column<DateTime>(type: "DATETIME", nullable: false)
                },
                constraints: table =>
                {
                    table.PrimaryKey("PK_WishlistItems", x => x.wishlist_item_id);
                });

            migrationBuilder.CreateTable(
                name: "Balances",
                columns: table => new
                {
                    ID = table.Column<long>(type: "INTEGER", nullable: false)
                        .Annotation("Sqlite:Autoincrement", true),
                    account_id = table.Column<long>(type: "INTEGER", nullable: false),
                    currency_type = table.Column<int>(type: "INTEGER", nullable: false),
                    amount = table.Column<int>(type: "INTEGER", nullable: false),
                    balance_type = table.Column<int>(type: "INTEGER", nullable: false)
                },
                constraints: table =>
                {
                    table.PrimaryKey("PK_Balances", x => x.ID);
                    table.ForeignKey(
                        name: "FK_Balances_Accounts_account_id",
                        column: x => x.account_id,
                        principalTable: "Accounts",
                        principalColumn: "account_id",
                        onDelete: ReferentialAction.Cascade);
                });

            migrationBuilder.CreateTable(
                name: "ObjectiveGroups",
                columns: table => new
                {
                    Id = table.Column<long>(type: "INTEGER", nullable: false)
                        .Annotation("Sqlite:Autoincrement", true),
                    account_id = table.Column<long>(type: "INTEGER", nullable: false),
                    cleared_at = table.Column<DateTime>(type: "DATETIME", nullable: false),
                    group_index = table.Column<int>(type: "INTEGER", nullable: false),
                    is_completed = table.Column<bool>(type: "BOOLEAN", nullable: false),
                    requires_complete_on_server = table.Column<bool>(type: "BOOLEAN", nullable: false)
                },
                constraints: table =>
                {
                    table.PrimaryKey("PK_ObjectiveGroups", x => x.Id);
                    table.ForeignKey(
                        name: "FK_ObjectiveGroups_Accounts_account_id",
                        column: x => x.account_id,
                        principalTable: "Accounts",
                        principalColumn: "account_id",
                        onDelete: ReferentialAction.Cascade);
                });

            migrationBuilder.CreateTable(
                name: "Objectives",
                columns: table => new
                {
                    Id = table.Column<long>(type: "INTEGER", nullable: false)
                        .Annotation("Sqlite:Autoincrement", true),
                    account_id = table.Column<long>(type: "INTEGER", nullable: false),
                    group_index = table.Column<int>(type: "INTEGER", nullable: false),
                    type = table.Column<int>(type: "INTEGER", nullable: false),
                    has_claimed_reward = table.Column<bool>(type: "BOOLEAN", nullable: false),
                    obj_index = table.Column<int>(type: "INTEGER", nullable: false),
                    is_completed = table.Column<bool>(type: "BOOLEAN", nullable: false),
                    is_rewarded = table.Column<bool>(type: "BOOLEAN", nullable: false),
                    progress = table.Column<double>(type: "REAL", nullable: false),
                    visual_progress = table.Column<double>(type: "REAL", nullable: false)
                },
                constraints: table =>
                {
                    table.PrimaryKey("PK_Objectives", x => x.Id);
                    table.ForeignKey(
                        name: "FK_Objectives_Accounts_account_id",
                        column: x => x.account_id,
                        principalTable: "Accounts",
                        principalColumn: "account_id",
                        onDelete: ReferentialAction.Cascade);
                });

            migrationBuilder.CreateTable(
                name: "PlatformAccounts",
                columns: table => new
                {
                    ID = table.Column<long>(type: "INTEGER", nullable: false)
                        .Annotation("Sqlite:Autoincrement", true),
                    account_id = table.Column<long>(type: "INTEGER", nullable: false),
                    platform = table.Column<int>(type: "INTEGER", nullable: false),
                    platform_id = table.Column<string>(type: "TEXT", nullable: false)
                },
                constraints: table =>
                {
                    table.PrimaryKey("PK_PlatformAccounts", x => x.ID);
                    table.ForeignKey(
                        name: "FK_PlatformAccounts_Accounts_account_id",
                        column: x => x.account_id,
                        principalTable: "Accounts",
                        principalColumn: "account_id",
                        onDelete: ReferentialAction.Cascade);
                });

            migrationBuilder.CreateTable(
                name: "PlayerSettings",
                columns: table => new
                {
                    ID = table.Column<long>(type: "INTEGER", nullable: false)
                        .Annotation("Sqlite:Autoincrement", true),
                    account_id = table.Column<long>(type: "INTEGER", nullable: false),
                    Key = table.Column<string>(type: "TEXT", nullable: false),
                    Value = table.Column<string>(type: "TEXT", nullable: false)
                },
                constraints: table =>
                {
                    table.PrimaryKey("PK_PlayerSettings", x => x.ID);
                    table.ForeignKey(
                        name: "FK_PlayerSettings_Accounts_account_id",
                        column: x => x.account_id,
                        principalTable: "Accounts",
                        principalColumn: "account_id",
                        onDelete: ReferentialAction.Cascade);
                });

            migrationBuilder.CreateTable(
                name: "SavedOutfits",
                columns: table => new
                {
                    account_id = table.Column<long>(type: "INTEGER", nullable: false),
                    slot = table.Column<string>(type: "TEXT", nullable: false),
                    preview_image_name = table.Column<string>(type: "TEXT", nullable: false),
                    outfit_selections = table.Column<string>(type: "TEXT", nullable: false),
                    face_features = table.Column<string>(type: "TEXT", nullable: false),
                    skin_color = table.Column<string>(type: "TEXT", nullable: false),
                    hair_color = table.Column<string>(type: "TEXT", nullable: false)
                },
                constraints: table =>
                {
                    table.PrimaryKey("PK_SavedOutfits", x => new { x.account_id, x.slot });
                    table.ForeignKey(
                        name: "FK_SavedOutfits_Accounts_account_id",
                        column: x => x.account_id,
                        principalTable: "Accounts",
                        principalColumn: "account_id",
                        onDelete: ReferentialAction.Cascade);
                });

            migrationBuilder.CreateTable(
                name: "FeaturedRoomEntrys",
                columns: table => new
                {
                    id = table.Column<long>(type: "INTEGER", nullable: false)
                        .Annotation("Sqlite:Autoincrement", true),
                    group_id = table.Column<long>(type: "INTEGER", nullable: false),
                    room_id = table.Column<long>(type: "INTEGER", nullable: false),
                    sort_order = table.Column<int>(type: "INTEGER", nullable: false),
                    FeaturedRoomGroupId = table.Column<long>(type: "INTEGER", nullable: true)
                },
                constraints: table =>
                {
                    table.PrimaryKey("PK_FeaturedRoomEntrys", x => x.id);
                    table.ForeignKey(
                        name: "FK_FeaturedRoomEntrys_FeaturedRoomGroups_FeaturedRoomGroupId",
                        column: x => x.FeaturedRoomGroupId,
                        principalTable: "FeaturedRoomGroups",
                        principalColumn: "id");
                });

            migrationBuilder.CreateTable(
                name: "PlayerEventTags",
                columns: table => new
                {
                    Id = table.Column<long>(type: "INTEGER", nullable: false)
                        .Annotation("Sqlite:Autoincrement", true),
                    player_event_id = table.Column<long>(type: "INTEGER", nullable: false),
                    tag = table.Column<string>(type: "TEXT", nullable: false),
                    type = table.Column<int>(type: "INTEGER", nullable: false)
                },
                constraints: table =>
                {
                    table.PrimaryKey("PK_PlayerEventTags", x => x.Id);
                    table.ForeignKey(
                        name: "FK_PlayerEventTags_PlayerEvents_player_event_id",
                        column: x => x.player_event_id,
                        principalTable: "PlayerEvents",
                        principalColumn: "player_event_id",
                        onDelete: ReferentialAction.Cascade);
                });

            migrationBuilder.CreateTable(
                name: "RoomPlaylistEntrys",
                columns: table => new
                {
                    id = table.Column<long>(type: "INTEGER", nullable: false)
                        .Annotation("Sqlite:Autoincrement", true),
                    playlist_id = table.Column<long>(type: "INTEGER", nullable: false),
                    room_id = table.Column<long>(type: "INTEGER", nullable: false),
                    sort_order = table.Column<int>(type: "INTEGER", nullable: false),
                    RoomPlaylistPlaylistId = table.Column<long>(type: "INTEGER", nullable: true)
                },
                constraints: table =>
                {
                    table.PrimaryKey("PK_RoomPlaylistEntrys", x => x.id);
                    table.ForeignKey(
                        name: "FK_RoomPlaylistEntrys_RoomPlaylists_RoomPlaylistPlaylistId",
                        column: x => x.RoomPlaylistPlaylistId,
                        principalTable: "RoomPlaylists",
                        principalColumn: "id");
                });

            migrationBuilder.CreateIndex(
                name: "IX_Balances_account_id",
                table: "Balances",
                column: "account_id");

            migrationBuilder.CreateIndex(
                name: "IX_FeaturedRoomEntrys_FeaturedRoomGroupId",
                table: "FeaturedRoomEntrys",
                column: "FeaturedRoomGroupId");

            migrationBuilder.CreateIndex(
                name: "IX_ObjectiveGroups_account_id",
                table: "ObjectiveGroups",
                column: "account_id");

            migrationBuilder.CreateIndex(
                name: "IX_Objectives_account_id",
                table: "Objectives",
                column: "account_id");

            migrationBuilder.CreateIndex(
                name: "IX_PlatformAccounts_account_id",
                table: "PlatformAccounts",
                column: "account_id");

            migrationBuilder.CreateIndex(
                name: "IX_PlayerEventTags_player_event_id",
                table: "PlayerEventTags",
                column: "player_event_id");

            migrationBuilder.CreateIndex(
                name: "IX_PlayerSettings_account_id",
                table: "PlayerSettings",
                column: "account_id");

            migrationBuilder.CreateIndex(
                name: "IX_RoomPlaylistEntrys_RoomPlaylistPlaylistId",
                table: "RoomPlaylistEntrys",
                column: "RoomPlaylistPlaylistId");
        }

        /// <inheritdoc />
        protected override void Down(MigrationBuilder migrationBuilder)
        {
            migrationBuilder.DropTable(
                name: "AccountBans");

            migrationBuilder.DropTable(
                name: "Announcements");

            migrationBuilder.DropTable(
                name: "AvatarItems");

            migrationBuilder.DropTable(
                name: "Avatars");

            migrationBuilder.DropTable(
                name: "Balances");

            migrationBuilder.DropTable(
                name: "ClubAnnouncements");

            migrationBuilder.DropTable(
                name: "ClubCustomTags");

            migrationBuilder.DropTable(
                name: "ClubMembers");

            migrationBuilder.DropTable(
                name: "ClubPermissions");

            migrationBuilder.DropTable(
                name: "ClubReports");

            migrationBuilder.DropTable(
                name: "Clubs");

            migrationBuilder.DropTable(
                name: "DailyXpLedgers");

            migrationBuilder.DropTable(
                name: "DeviceBans");

            migrationBuilder.DropTable(
                name: "DeviceLogins");

            migrationBuilder.DropTable(
                name: "FeaturedRoomEntrys");

            migrationBuilder.DropTable(
                name: "Gifts");

            migrationBuilder.DropTable(
                name: "InstanceBans");

            migrationBuilder.DropTable(
                name: "InstanceInvites");

            migrationBuilder.DropTable(
                name: "InventionCheers");

            migrationBuilder.DropTable(
                name: "InventionOwnerships");

            migrationBuilder.DropTable(
                name: "InventionReports");

            migrationBuilder.DropTable(
                name: "Inventions");

            migrationBuilder.DropTable(
                name: "InventionTags");

            migrationBuilder.DropTable(
                name: "InventionVersions");

            migrationBuilder.DropTable(
                name: "LeaderboardStats");

            migrationBuilder.DropTable(
                name: "Messages");

            migrationBuilder.DropTable(
                name: "ModerationBlocks");

            migrationBuilder.DropTable(
                name: "ModerationReports");

            migrationBuilder.DropTable(
                name: "ObjectiveGroups");

            migrationBuilder.DropTable(
                name: "Objectives");

            migrationBuilder.DropTable(
                name: "PlatformAccountLimits");

            migrationBuilder.DropTable(
                name: "PlatformAccounts");

            migrationBuilder.DropTable(
                name: "PlayerBios");

            migrationBuilder.DropTable(
                name: "PlayerCheers");

            migrationBuilder.DropTable(
                name: "PlayerEventResponses");

            migrationBuilder.DropTable(
                name: "PlayerEventTags");

            migrationBuilder.DropTable(
                name: "PlayerReports");

            migrationBuilder.DropTable(
                name: "PlayerSettings");

            migrationBuilder.DropTable(
                name: "PlayerStates");

            migrationBuilder.DropTable(
                name: "Progressions");

            migrationBuilder.DropTable(
                name: "RefreshTokens");

            migrationBuilder.DropTable(
                name: "Relationships");

            migrationBuilder.DropTable(
                name: "RewardDrops");

            migrationBuilder.DropTable(
                name: "RewardSelections");

            migrationBuilder.DropTable(
                name: "RoomInstances");

            migrationBuilder.DropTable(
                name: "RoomInteractions");

            migrationBuilder.DropTable(
                name: "RoomPlaylistEntrys");

            migrationBuilder.DropTable(
                name: "RoomReports");

            migrationBuilder.DropTable(
                name: "RoomRoleEntrys");

            migrationBuilder.DropTable(
                name: "Rooms");

            migrationBuilder.DropTable(
                name: "RoomTags");

            migrationBuilder.DropTable(
                name: "SavedOutfits");

            migrationBuilder.DropTable(
                name: "ScreenShareReports");

            migrationBuilder.DropTable(
                name: "SubRoomDataHistorys");

            migrationBuilder.DropTable(
                name: "SubRooms");

            migrationBuilder.DropTable(
                name: "UploadedPhotoCheers");

            migrationBuilder.DropTable(
                name: "UploadedPhotos");

            migrationBuilder.DropTable(
                name: "UserAvatarItems");

            migrationBuilder.DropTable(
                name: "UserConsumables");

            migrationBuilder.DropTable(
                name: "UserEquipments");

            migrationBuilder.DropTable(
                name: "UserImages");

            migrationBuilder.DropTable(
                name: "WishlistItems");

            migrationBuilder.DropTable(
                name: "FeaturedRoomGroups");

            migrationBuilder.DropTable(
                name: "PlayerEvents");

            migrationBuilder.DropTable(
                name: "RoomPlaylists");

            migrationBuilder.DropTable(
                name: "Accounts");
        }
    }
}
