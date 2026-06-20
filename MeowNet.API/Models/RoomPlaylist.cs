using System;
using System.Collections.Generic;
using System.ComponentModel.DataAnnotations;
using System.ComponentModel.DataAnnotations.Schema;
using System.Text.Json.Serialization;

namespace MeowNet.API.Models
{
    public class RoomPlaylist
    {
    [Column("name")]
    [JsonPropertyName("Name")]
    public string Name { get; set; }
    [Column("description")]
    [JsonPropertyName("Description")]
    public string Description { get; set; }
    [Column("image_name")]
    [JsonPropertyName("ImageName")]
    public string ImageName { get; set; }
    [Column("warning_mask")]
    [JsonPropertyName("WarningMask")]
    public int WarningMask { get; set; }
    [Column("custom_warning")]
    [JsonPropertyName("CustomWarning")]
    public string CustomWarning { get; set; }
    [Column("creator_account_id")]
    [JsonPropertyName("CreatorAccountId")]
    public int CreatorAccountId { get; set; }
    [Column("state")]
    [JsonPropertyName("State")]
    public int State { get; set; }
    [Column("accessibility")]
    [JsonPropertyName("Accessibility")]
    public int Accessibility { get; set; }
    [Column("supports_level_voting")]
    [JsonPropertyName("SupportsLevelVoting")]
    public bool SupportsLevelVoting { get; set; }
    [Column("is_rro")]
    [JsonPropertyName("IsRRO")]
    public bool IsRRO { get; set; }
    [Column("supports_screens")]
    [JsonPropertyName("SupportsScreens")]
    public bool SupportsScreens { get; set; }
    [Column("supports_walk_vr")]
    [JsonPropertyName("SupportsWalkVR")]
    public bool SupportsWalkVR { get; set; }
    [Column("supports_teleport_vr")]
    [JsonPropertyName("SupportsTeleportVR")]
    public bool SupportsTeleportVR { get; set; }
    [Column("supports_vr_low")]
    [JsonPropertyName("SupportsVRLow")]
    public bool SupportsVRLow { get; set; }
    [Column("supports_quest_2")]
    [JsonPropertyName("SupportsQuest2")]
    public bool SupportsQuest2 { get; set; }
    [Column("supports_mobile")]
    [JsonPropertyName("SupportsMobile")]
    public bool SupportsMobile { get; set; }
    [Column("supports_juniors")]
    [JsonPropertyName("SupportsJuniors")]
    public bool SupportsJuniors { get; set; }
    [Column("min_level")]
    [JsonPropertyName("MinLevel")]
    public int MinLevel { get; set; }
    [Column("created_at")]
    [JsonPropertyName("CreatedAt")]
    public DateTime CreatedAt { get; set; }
    [JsonPropertyName("Stats")]
    public RoomStats Stats { get; set; }
    [Key]
    [Column("id")]
    [JsonPropertyName("PlaylistId")]
    public long PlaylistId { get; set; }
    [Column("sort_order")]
    [JsonIgnore]
    public int SortOrder { get; set; }
    [JsonIgnore]
    public List<RoomPlaylistEntry> Entries { get; set; }
    }
}