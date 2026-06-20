using System;
using System.Collections.Generic;
using System.ComponentModel.DataAnnotations;
using System.ComponentModel.DataAnnotations.Schema;
using System.Text.Json.Serialization;

namespace MeowNet.API.Models
{
    public class SubRoom
    {
    [Key]
    [Column("sub_room_id")]
    [JsonPropertyName("SubRoomId")]
    public long SubRoomId { get; set; }
    [Column("room_id")]
    [JsonPropertyName("RoomId")]
    public long RoomId { get; set; }
    [Column("accessibility")]
    [JsonPropertyName("Accessibility")]
    public int Accessibility { get; set; }
    [Column("data_blob")]
    [JsonPropertyName("DataBlob")]
    public string DataBlob { get; set; }
    [Column("is_sandbox")]
    [JsonPropertyName("IsSandbox")]
    public bool IsSandbox { get; set; }
    [Column("max_players")]
    [JsonPropertyName("MaxPlayers")]
    public int MaxPlayers { get; set; }
    [Column("name")]
    [JsonPropertyName("Name")]
    public string Name { get; set; }
    [Column("saved_by_account_id")]
    [JsonPropertyName("SavedByAccountId")]
    public int SavedByAccountId { get; set; }
    [Column("unity_scene_id")]
    [JsonPropertyName("UnitySceneId")]
    public string UnitySceneId { get; set; }
    }
}