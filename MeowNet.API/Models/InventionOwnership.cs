using System;
using System.Collections.Generic;
using System.ComponentModel.DataAnnotations;
using System.ComponentModel.DataAnnotations.Schema;
using System.Text.Json.Serialization;

namespace MeowNet.API.Models
{
    public class InventionOwnership
    {
    [Key]
    [Column("id")]
    [JsonIgnore]
    public long Id { get; set; }
    [Column("invention_id")]
    [JsonPropertyName("InventionId")]
    public long InventionId { get; set; }
    [Column("account_id")]
    [JsonPropertyName("AccountId")]
    public long AccountId { get; set; }
    [Column("acquired_at")]
    [JsonPropertyName("AcquiredAt")]
    public DateTime AcquiredAt { get; set; }
    }
}