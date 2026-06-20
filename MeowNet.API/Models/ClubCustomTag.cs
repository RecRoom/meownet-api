using System;
using System.Collections.Generic;
using System.ComponentModel.DataAnnotations;
using System.ComponentModel.DataAnnotations.Schema;
using System.Text.Json.Serialization;

namespace MeowNet.API.Models
{
    public class ClubCustomTag
    {
    [Key]
    [Column("id")]
    [JsonIgnore]
    public long Id { get; set; }
    [Column("club_id")]
    [JsonIgnore]
    public long ClubId { get; set; }
    [Column("tag")]
    [JsonIgnore]
    public string Tag { get; set; }
    }
}