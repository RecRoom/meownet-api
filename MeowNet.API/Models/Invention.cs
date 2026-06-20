using System;
using System.Collections.Generic;
using System.ComponentModel.DataAnnotations;
using System.ComponentModel.DataAnnotations.Schema;
using System.Text.Json.Serialization;

namespace MeowNet.API.Models
{
    public class Invention
    {
    [Key]
    [Column("invention_id")]
    [JsonPropertyName("InventionId")]
    public long InventionId { get; set; }
    [Column("name")]
    [JsonPropertyName("Name")]
    public string Name { get; set; }
    [Column("description")]
    [JsonPropertyName("Description")]
    public string Description { get; set; }
    [Column("image_name")]
    [JsonPropertyName("ImageName")]
    public string ImageName { get; set; }
    [Column("creator_player_id")]
    [JsonPropertyName("CreatorPlayerId")]
    public int CreatorPlayerId { get; set; }
    [Column("creator_permission")]
    [JsonPropertyName("CreatorPermission")]
    public int CreatorPermission { get; set; }
    [Column("general_permission")]
    [JsonPropertyName("GeneralPermission")]
    public int GeneralPermission { get; set; }
    [Column("allow_trial")]
    [JsonPropertyName("AllowTrial")]
    public bool AllowTrial { get; set; }
    [Column("hide_from_player")]
    [JsonPropertyName("HideFromPlayer")]
    public bool HideFromPlayer { get; set; }
    [Column("is_ag_invention")]
    [JsonPropertyName("IsAGInvention")]
    public bool IsAGInvention { get; set; }
    [Column("is_certified_invention")]
    [JsonPropertyName("IsCertifiedInvention")]
    public bool IsCertifiedInvention { get; set; }
    [Column("is_published")]
    [JsonPropertyName("IsPublished")]
    public bool IsPublished { get; set; }
    [Column("price")]
    [JsonPropertyName("Price")]
    public int Price { get; set; }
    [Column("cheer_count")]
    [JsonPropertyName("CheerCount")]
    public int CheerCount { get; set; }
    [Column("num_downloads")]
    [JsonPropertyName("NumDownloads")]
    public int NumDownloads { get; set; }
    [Column("num_players_used_in_room")]
    [JsonPropertyName("NumPlayersHaveUsedInRoom")]
    public int NumPlayersHaveUsedInRoom { get; set; }
    [Column("current_version_number")]
    [JsonPropertyName("CurrentVersionNumber")]
    public int CurrentVersionNumber { get; set; }
    [Column("replication_id")]
    [JsonPropertyName("ReplicationId")]
    public string ReplicationId { get; set; }
    [Column("created_at")]
    [JsonPropertyName("CreatedAt")]
    public DateTime CreatedAt { get; set; }
    [Column("modified_at")]
    [JsonPropertyName("ModifiedAt")]
    public DateTime ModifiedAt { get; set; }
    }
}