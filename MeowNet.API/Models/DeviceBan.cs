using System;
using System.Collections.Generic;
using System.ComponentModel.DataAnnotations;
using System.ComponentModel.DataAnnotations.Schema;
using System.Text.Json.Serialization;

namespace MeowNet.API.Models
{
    public class DeviceBan
    {
    [Key]
    [JsonPropertyName("id")]
    public long ID { get; set; }
    [JsonPropertyName("device_id")]
    public string DeviceID { get; set; }
    [JsonPropertyName("account_id")]
    public long AccountID { get; set; }
    [JsonPropertyName("ban_id")]
    public long BanID { get; set; }
    [JsonPropertyName("reason")]
    public string Reason { get; set; }
    [JsonPropertyName("banned_by")]
    public string BannedBy { get; set; }
    [JsonPropertyName("created_at")]
    public DateTime CreatedAt { get; set; }
    [JsonPropertyName("expires_at")]
    public DateTime? ExpiresAt { get; set; }
    }
}