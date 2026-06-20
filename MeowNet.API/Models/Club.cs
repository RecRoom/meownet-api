using System;
using System.Collections.Generic;
using System.ComponentModel.DataAnnotations;
using System.ComponentModel.DataAnnotations.Schema;
using System.Text.Json.Serialization;

namespace MeowNet.API.Models
{
    public class Club
    {
    [Key]
    [Column("club_id")]
    [JsonPropertyName("ClubId")]
    public long ClubId { get; set; }
    [Column("name")]
    [JsonPropertyName("Name")]
    public string Name { get; set; }
    [Column("description")]
    [JsonPropertyName("Description")]
    public string Description { get; set; }
    [Column("category")]
    [JsonPropertyName("Category")]
    public string Category { get; set; }
    [Column("visibility")]
    [JsonPropertyName("Visibility")]
    public int Visibility { get; set; }
    [Column("joinability")]
    [JsonPropertyName("Joinability")]
    public int Joinability { get; set; }
    [Column("allow_juniors")]
    [JsonPropertyName("AllowJuniors")]
    public bool AllowJuniors { get; set; }
    [Column("main_image_name")]
    [JsonPropertyName("MainImageName")]
    public string MainImageName { get; set; }
    [Column("club_type")]
    [JsonPropertyName("ClubType")]
    public int ClubType { get; set; }
    [Column("clubhouse_room_id")]
    [JsonPropertyName("ClubhouseRoomId")]
    public long? ClubhouseRoomId { get; set; }
    [Column("creator_account_id")]
    [JsonPropertyName("CreatorAccountId")]
    public int CreatorAccountId { get; set; }
    [Column("is_rro")]
    [JsonPropertyName("IsRRO")]
    public bool IsRRO { get; set; }
    [Column("min_level")]
    [JsonPropertyName("MinLevel")]
    public int MinLevel { get; set; }
    [Column("state")]
    [JsonPropertyName("State")]
    public int State { get; set; }
    [Column("member_count")]
    [JsonPropertyName("MemberCount")]
    public int MemberCount { get; set; }
    [Column("created_at")]
    [JsonIgnore]
    public DateTime CreatedAt { get; set; }
    }
}