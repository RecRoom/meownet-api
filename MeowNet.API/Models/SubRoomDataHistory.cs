using System;
using System.Collections.Generic;
using System.ComponentModel.DataAnnotations;
using System.ComponentModel.DataAnnotations.Schema;
using System.Text.Json.Serialization;

namespace MeowNet.API.Models
{
    public class SubRoomDataHistory
    {
    [Key]
    [Column("id")]
    [JsonIgnore]
    public long Id { get; set; }
    [Column("sub_room_id")]
    [JsonPropertyName("SubRoomId")]
    public int SubRoomId { get; set; }
    [Column("data_blob")]
    [JsonPropertyName("DataBlob")]
    public string DataBlob { get; set; }
    [Column("saved_by_account_id")]
    [JsonPropertyName("SavedByAccountId")]
    public int SavedByAccountId { get; set; }
    [Column("created_at")]
    [JsonPropertyName("CreatedAt")]
    public DateTime CreatedAt { get; set; }
    }
}