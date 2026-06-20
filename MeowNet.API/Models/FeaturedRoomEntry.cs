using System;
using System.Collections.Generic;
using System.ComponentModel.DataAnnotations;
using System.ComponentModel.DataAnnotations.Schema;
using System.Text.Json.Serialization;

namespace MeowNet.API.Models
{
    public class FeaturedRoomEntry
    {
    [Key]
    [Column("id")]
    [JsonIgnore]
    public long Id { get; set; }
    [Column("group_id")]
    [JsonIgnore]
    public long GroupId { get; set; }
    [Column("room_id")]
    [JsonIgnore]
    public long RoomId { get; set; }
    [Column("sort_order")]
    [JsonIgnore]
    public int SortOrder { get; set; }
    }
}