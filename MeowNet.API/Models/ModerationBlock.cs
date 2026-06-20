using System;
using System.Collections.Generic;
using System.ComponentModel.DataAnnotations;
using System.ComponentModel.DataAnnotations.Schema;
using System.Text.Json.Serialization;

namespace MeowNet.API.Models
{
    public class ModerationBlock
    {
    [Key]
    [JsonIgnore]
    public long ID { get; set; }
    [Column("account_id")]
    [JsonIgnore]
    public long AccountID { get; set; }
    [Column("reporter_id")]
    [JsonPropertyName("PlayerIdReporter")]
    public long? ReporterID { get; set; }
    [Column("game_session_id")]
    [JsonPropertyName("GameSessionId")]
    public long GameSessionID { get; set; }
    [Column("is_ban")]
    [JsonPropertyName("IsBan")]
    public bool IsBan { get; set; }
    [Column("is_host_kick")]
    [JsonPropertyName("IsHostKick")]
    public bool IsHostKick { get; set; }
    [Column("message")]
    [JsonPropertyName("Message")]
    public string? Message { get; set; }
    [Column("report_category")]
    [JsonPropertyName("ReportCategory")]
    public int ReportCategory { get; set; }
    [Column("duration")]
    [JsonPropertyName("Duration")]
    public int Duration { get; set; }
    [Column("created_at")]
    [JsonIgnore]
    public DateTime CreatedAt { get; set; }
    [Column("expires_at")]
    [JsonIgnore]
    public DateTime? ExpiresAt { get; set; }
    }
}