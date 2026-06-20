using System;
using System.Collections.Generic;
using System.ComponentModel.DataAnnotations;
using System.ComponentModel.DataAnnotations.Schema;
using System.Text.Json.Serialization;

namespace MeowNet.API.Models
{
    public class Message
    {
    [Key]
    [JsonPropertyName("Id")]
    public long Id { get; set; }
    [Column("from_player_id")]
    [JsonPropertyName("FromPlayerId")]
    public long FromPlayerId { get; set; }
    [Column("to_player_id")]
    [JsonPropertyName("ToPlayerId")]
    public long ToPlayerId { get; set; }
    [Column("sent_time")]
    [JsonPropertyName("SentTime")]
    public DateTime SentTime { get; set; }
    [Column("type")]
    [JsonPropertyName("Type")]
    public int Type { get; set; }
    [Column("room_id")]
    [JsonPropertyName("RoomId")]
    public long? RoomId { get; set; }
    [Column("invention_id")]
    [JsonPropertyName("InventionId")]
    public long? InventionId { get; set; }
    [Column("player_event_id")]
    [JsonPropertyName("PlayerEventId")]
    public long? PlayerEventId { get; set; }
    [Column("data")]
    [JsonPropertyName("Data")]
    public string Data { get; set; }
    }
}