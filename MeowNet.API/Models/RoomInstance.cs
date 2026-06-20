using System;
using System.Collections.Generic;
using System.ComponentModel.DataAnnotations;
using System.ComponentModel.DataAnnotations.Schema;
using System.Text.Json.Serialization;

namespace MeowNet.API.Models
{
    public class RoomInstance
    {
    [Key]
    [JsonPropertyName("roomInstanceId")]
    public long Id { get; set; }
    [Column("owner_account_id")]
    [JsonIgnore]
    public int OwnerAccountId { get; set; }
    [Column("room_id")]
    [JsonPropertyName("roomId")]
    public long RoomId { get; set; }
    [Column("sub_room_id")]
    [JsonPropertyName("subRoomId")]
    public long SubRoomId { get; set; }
    [Column("location")]
    [JsonPropertyName("location")]
    public string Location { get; set; }
    [Column("data_blob")]
    [JsonIgnore]
    public string DataBlob { get; set; }
    [Column("event_id")]
    [JsonPropertyName("eventId")]
    public long EventId { get; set; }
    [Column("photon_region_id")]
    [JsonPropertyName("photonRegionId")]
    public string PhotonRegionId { get; set; }
    [Column("photon_room_id")]
    [JsonPropertyName("photonRoomId")]
    public string PhotonRoomId { get; set; }
    [Column("name")]
    [JsonPropertyName("name")]
    public string Name { get; set; }
    [Column("max_capacity")]
    [JsonPropertyName("maxCapacity")]
    public int MaxCapacity { get; set; }
    [Column("is_full")]
    [JsonPropertyName("isFull")]
    public bool IsFull { get; set; }
    [Column("is_private")]
    [JsonPropertyName("isPrivate")]
    public bool IsPrivate { get; set; }
    [Column("is_in_progress")]
    [JsonPropertyName("isInProgress")]
    public bool IsInProgress { get; set; }
    [Column("room_code")]
    [JsonPropertyName("roomCode")]
    public string RoomCode { get; set; }
    [Column("room_instance_type")]
    [JsonPropertyName("roomInstanceType")]
    public int RoomInstanceType { get; set; }
    [Column("club_id")]
    [JsonPropertyName("clubId")]
    public long ClubId { get; set; }
    [Column("encrypt_voice_chat")]
    [JsonPropertyName("EncryptVoiceChat")]
    public bool EncryptVoiceChat { get; set; }
    [Column("matchmaking_policy")]
    [JsonPropertyName("matchmakingPolicy")]
    public int MatchmakingPolicy { get; set; }
    [Column("allow_new_users")]
    [JsonIgnore]
    public bool AllowNewUsers { get; set; }
    [Column("join_disabled")]
    [JsonIgnore]
    public bool JoinDisabled { get; set; }
    [Column("created_at")]
    [JsonPropertyName("createdAt")]
    public DateTime CreatedAt { get; set; }
    }
}