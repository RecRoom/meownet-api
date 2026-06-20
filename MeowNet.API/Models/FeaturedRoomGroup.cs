using System;
using System.Collections.Generic;
using System.ComponentModel.DataAnnotations;
using System.ComponentModel.DataAnnotations.Schema;
using System.Text.Json.Serialization;

namespace MeowNet.API.Models
{
    public class FeaturedRoomGroup
    {
    [Key]
    [Column("id")]
    [JsonPropertyName("FeaturedRoomGroupId")]
    public long Id { get; set; }
    [Column("name")]
    [JsonPropertyName("Name")]
    public string Name { get; set; }
    [Column("sort_order")]
    [JsonIgnore]
    public int SortOrder { get; set; }
    [JsonIgnore]
    public List<FeaturedRoomEntry> Entries { get; set; }
    [JsonPropertyName("Rooms")]
    public List<FeaturedRoomItem> Rooms { get; set; }
    }
}