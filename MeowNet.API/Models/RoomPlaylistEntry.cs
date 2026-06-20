using System;
using System.Collections.Generic;
using System.ComponentModel.DataAnnotations;
using System.ComponentModel.DataAnnotations.Schema;
using System.Text.Json.Serialization;

namespace MeowNet.API.Models
{
    public class RoomPlaylistEntry
    {
    [Key]
    [Column("id")]
    [JsonIgnore]
    public long Id { get; set; }
    [Column("playlist_id")]
    [JsonIgnore]
    public long PlaylistId { get; set; }
    [Column("room_id")]
    [JsonPropertyName("RoomId")]
    public long RoomId { get; set; }
    [Column("sort_order")]
    [JsonIgnore]
    public int SortOrder { get; set; }
    }
}