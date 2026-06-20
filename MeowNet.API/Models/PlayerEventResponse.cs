using System;
using System.Collections.Generic;
using System.ComponentModel.DataAnnotations;
using System.ComponentModel.DataAnnotations.Schema;
using System.Text.Json.Serialization;

namespace MeowNet.API.Models
{
    public class PlayerEventResponse
    {
    [Key]
    [Column("player_event_response_id")]
    [JsonPropertyName("PlayerEventResponseId")]
    public long PlayerEventResponseId { get; set; }
    [Column("player_event_id")]
    [JsonPropertyName("PlayerEventId")]
    public long PlayerEventId { get; set; }
    [Column("player_id")]
    [JsonPropertyName("PlayerId")]
    public long PlayerId { get; set; }
    [Column("type")]
    [JsonPropertyName("Type")]
    public int Type { get; set; }
    [Column("created_at")]
    [JsonPropertyName("CreatedAt")]
    public DateTime CreatedAt { get; set; }
    }
}