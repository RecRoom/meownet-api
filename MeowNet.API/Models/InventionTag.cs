using System;
using System.Collections.Generic;
using System.ComponentModel.DataAnnotations;
using System.ComponentModel.DataAnnotations.Schema;
using System.Text.Json.Serialization;

namespace MeowNet.API.Models
{
    public class InventionTag
    {
    [Key]
    [Column("id")]
    [JsonIgnore]
    public long Id { get; set; }
    [Column("invention_id")]
    [JsonIgnore]
    public long InventionId { get; set; }
    [Column("tag")]
    [JsonPropertyName("Tag")]
    public string Tag { get; set; }
    [Column("type")]
    [JsonPropertyName("Type")]
    public int Type { get; set; }
    }
}