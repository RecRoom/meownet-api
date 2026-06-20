using System;
using System.Collections.Generic;
using System.ComponentModel.DataAnnotations;
using System.ComponentModel.DataAnnotations.Schema;
using System.Text.Json.Serialization;

namespace MeowNet.API.Models
{
    public class DeviceLogin
    {
    [Key]
    [JsonPropertyName("id")]
    public long ID { get; set; }
    [JsonPropertyName("account_id")]
    public long AccountID { get; set; }
    [JsonPropertyName("device_id")]
    public string DeviceID { get; set; }
    [JsonPropertyName("device_class")]
    public int DeviceClass { get; set; }
    [Column("platform_id")]
    [JsonPropertyName("platform_id")]
    public string PlatformID { get; set; }
    [Column("platform")]
    [JsonPropertyName("platform")]
    public string Platform { get; set; }
    [Column("ip")]
    [JsonPropertyName("ip")]
    public string IP { get; set; }
    [JsonPropertyName("login_count")]
    public int LoginCount { get; set; }
    [JsonPropertyName("first_seen")]
    public DateTime FirstSeen { get; set; }
    [JsonPropertyName("last_seen")]
    public DateTime LastSeen { get; set; }
    }
}