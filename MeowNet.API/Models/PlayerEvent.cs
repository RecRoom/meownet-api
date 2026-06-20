using System;
using System.Collections.Generic;
using System.ComponentModel.DataAnnotations;
using System.ComponentModel.DataAnnotations.Schema;
using System.Text.Json.Serialization;

namespace MeowNet.API.Models
{
    public class PlayerEvent
    {
    [Key]
    [Column("player_event_id")]
    [JsonPropertyName("PlayerEventId")]
    public long PlayerEventId { get; set; }
    [Column("creator_player_id")]
    [JsonPropertyName("CreatorPlayerId")]
    public long CreatorPlayerId { get; set; }
    [Column("room_id")]
    [JsonPropertyName("RoomId")]
    public long RoomId { get; set; }
    [Column("sub_room_id")]
    [JsonPropertyName("SubRoomId")]
    public long? SubRoomId { get; set; }
    [Column("club_id")]
    [JsonPropertyName("ClubId")]
    public long? ClubId { get; set; }
    [Column("name")]
    [JsonPropertyName("Name")]
    public string Name { get; set; }
    [Column("description")]
    [JsonPropertyName("Description")]
    public string Description { get; set; }
    [Column("image_name")]
    [JsonPropertyName("ImageName")]
    public string? ImageName { get; set; }
    [Column("start_time")]
    [JsonPropertyName("StartTime")]
    public DateTime StartTime { get; set; }
    [Column("end_time")]
    [JsonPropertyName("EndTime")]
    public DateTime EndTime { get; set; }
    [Column("accessibility")]
    [JsonPropertyName("Accessibility")]
    public int Accessibility { get; set; }
    [Column("state")]
    [JsonPropertyName("State")]
    public int State { get; set; }
    [Column("attendee_count")]
    [JsonPropertyName("AttendeeCount")]
    public int AttendeeCount { get; set; }
    [JsonPropertyName("Tags")]
    public List<PlayerEventTag> Tags { get; set; }
    }
}