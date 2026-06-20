using System;
using System.Collections.Generic;
using System.ComponentModel.DataAnnotations;
using System.ComponentModel.DataAnnotations.Schema;
using System.Text.Json.Serialization;

namespace MeowNet.API.Models
{
    public class PlayerSetting
    {
    [Key]
    [JsonIgnore]
    public long ID { get; set; }
    [Column("account_id")]
    [JsonIgnore]
    public long AccountID { get; set; }
    [JsonIgnore]
    public Account Account { get; set; }
    [JsonPropertyName("Key")]
    public string Key { get; set; }
    [JsonPropertyName("Value")]
    public string Value { get; set; }
    }
}