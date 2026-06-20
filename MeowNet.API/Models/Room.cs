using System;
using System.Collections.Generic;
using System.ComponentModel.DataAnnotations;
using System.ComponentModel.DataAnnotations.Schema;
using System.Text.Json.Serialization;

namespace MeowNet.API.Models
{
    public class Room
    {
    [Key]
    [Column("room_id")]
    [JsonPropertyName("RoomId")]
    public long RoomId { get; set; }
    [Column("name")]
    [JsonPropertyName("Name")]
    public string Name { get; set; }
    [Column("description")]
    [JsonPropertyName("Description")]
    public string Description { get; set; }
    [Column("image_name")]
    [JsonPropertyName("ImageName")]
    public string ImageName { get; set; }
    [Column("creator_account_id")]
    [JsonPropertyName("CreatorAccountId")]
    public int CreatorAccountId { get; set; }
    [Column("state")]
    [JsonPropertyName("State")]
    public int State { get; set; }
    [Column("accessibility")]
    [JsonPropertyName("Accessibility")]
    public int Accessibility { get; set; }
    [Column("auto_localize_room")]
    [JsonPropertyName("AutoLocalizeRoom")]
    public bool AutoLocalizeRoom { get; set; }
    [Column("cloning_allowed")]
    [JsonPropertyName("CloningAllowed")]
    public bool CloningAllowed { get; set; }
    [Column("custom_warning")]
    [JsonPropertyName("CustomWarning")]
    public string CustomWarning { get; set; }
    [Column("disable_mic_auto_mute")]
    [JsonPropertyName("DisableMicAutoMute")]
    public bool DisableMicAutoMute { get; set; }
    [Column("disable_room_comments")]
    [JsonPropertyName("DisableRoomComments")]
    public bool DisableRoomComments { get; set; }
    [Column("encrypt_voice_chat")]
    [JsonPropertyName("EncryptVoiceChat")]
    public bool EncryptVoiceChat { get; set; }
    [Column("is_developer_owned")]
    [JsonPropertyName("IsDeveloperOwned")]
    public bool IsDeveloperOwned { get; set; }
    [Column("is_dorm")]
    [JsonPropertyName("IsDorm")]
    public bool IsDorm { get; set; }
    [Column("is_rro")]
    [JsonPropertyName("IsRRO")]
    public bool IsRRO { get; set; }
    [Column("load_screen_locked")]
    [JsonPropertyName("LoadScreenLocked")]
    public bool LoadScreenLocked { get; set; }
    [Column("max_player_calculation_mode")]
    [JsonPropertyName("MaxPlayerCalculationMode")]
    public int MaxPlayerCalculationMode { get; set; }
    [Column("max_players")]
    [JsonPropertyName("MaxPlayers")]
    public int MaxPlayers { get; set; }
    [Column("min_level")]
    [JsonPropertyName("MinLevel")]
    public int MinLevel { get; set; }
    [Column("persistence_version")]
    [JsonPropertyName("PersistenceVersion")]
    public int PersistenceVersion { get; set; }
    [Column("ranked_entity_id")]
    [JsonPropertyName("RankedEntityId")]
    public string RankedEntityId { get; set; }
    [Column("ranking_context")]
    [JsonPropertyName("RankingContext")]
    public int RankingContext { get; set; }
    [Column("supports_juniors")]
    [JsonPropertyName("SupportsJuniors")]
    public bool SupportsJuniors { get; set; } = true;
    [Column("supports_level_voting")]
    [JsonPropertyName("SupportsLevelVoting")]
    public bool SupportsLevelVoting { get; set; }
    [Column("supports_mobile")]
    [JsonPropertyName("SupportsMobile")]
    public bool SupportsMobile { get; set; } = true;
    [Column("supports_quest_2")]
    [JsonPropertyName("SupportsQuest2")]
    public bool SupportsQuest2 { get; set; } = true;
    [Column("supports_screens")]
    [JsonPropertyName("SupportsScreens")]
    public bool SupportsScreens { get; set; } = true;
    [Column("supports_teleport_vr")]
    [JsonPropertyName("SupportsTeleportVR")]
    public bool SupportsTeleportVR { get; set; } = true;
    [Column("supports_vr_low")]
    [JsonPropertyName("SupportsVRLow")]
    public bool SupportsVRLow { get; set; } = true;
    [Column("supports_walk_vr")]
    [JsonPropertyName("SupportsWalkVR")]
    public bool SupportsWalkVR { get; set; } = true;
    [Column("toxmod_enabled")]
    [JsonPropertyName("ToxmodEnabled")]
    public bool ToxmodEnabled { get; set; }
    [Column("ugc_version")]
    [JsonPropertyName("UgcVersion")]
    public int UgcVersion { get; set; }
    [Column("warning_mask")]
    [JsonPropertyName("WarningMask")]
    public int WarningMask { get; set; }
    [Column("data_blob")]
    [JsonPropertyName("DataBlob")]
    public string? DataBlob { get; set; }
    [Column("created_at")]
    [JsonPropertyName("CreatedAt")]
    public DateTime CreatedAt { get; set; }
    [NotMapped]
    [JsonPropertyName("Score")]
    public int Score { get; set; } = 0;
    [NotMapped]
    [JsonPropertyName("Stats")]
    public RoomStats Stats { get; set; }
    [NotMapped]
    [JsonPropertyName("SubRooms")]
    public List<SubRoom> SubRooms { get; set; }
    [NotMapped]
    [JsonPropertyName("Roles")]
    public List<RoomRoleEntry> Roles { get; set; }
    [NotMapped]
    [JsonPropertyName("Tags")]
    public List<RoomTag> Tags { get; set; }
    [NotMapped]
    [JsonPropertyName("LoadScreens")]
    public List<object> LoadScreens { get; set; }
    [NotMapped]
    [JsonPropertyName("PromoImages")]
    public List<object> PromoImages { get; set; }
    [NotMapped]
    [JsonPropertyName("PromoExternalContent")]
    public List<object> PromoExternalContent { get; set; }
    }
}