using System;
using System.Collections.Generic;
using System.ComponentModel.DataAnnotations;
using System.ComponentModel.DataAnnotations.Schema;
using System.Text.Json.Serialization;

namespace MeowNet.API.Models
{
    public class InventionCheer
    {
    [Key]
    [Column("id")]
    [JsonIgnore]
    public long Id { get; set; }
    [Column("invention_id")]
    [JsonIgnore]
    public long InventionId { get; set; }
    [Column("account_id")]
    [JsonIgnore]
    public long AccountId { get; set; }
    [Column("created_at")]
    [JsonIgnore]
    public DateTime CreatedAt { get; set; }
    }
}