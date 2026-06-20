using System;
using System.Collections.Generic;
using System.ComponentModel.DataAnnotations;
using System.ComponentModel.DataAnnotations.Schema;
using System.Text.Json.Serialization;

namespace MeowNet.API.Models
{
    public class Progression
    {
    [Key]
    [DatabaseGenerated(DatabaseGeneratedOption.None)]
    [Column("account_id")]
    [JsonPropertyName("PlayerId")]
    public long AccountID { get; set; }
    [Column("level")]
    [JsonPropertyName("Level")]
    public long Level { get; set; }
    [Column("xp")]
    [JsonPropertyName("XP")]
    public long XP { get; set; }
    }
}