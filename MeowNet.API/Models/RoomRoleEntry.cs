using System;
using System.Collections.Generic;
using System.ComponentModel.DataAnnotations;
using System.ComponentModel.DataAnnotations.Schema;
using System.Text.Json.Serialization;

namespace MeowNet.API.Models
{
    public class RoomRoleEntry
    {
    [Key]
    [JsonIgnore]
    public long Id { get; set; }
    [Column("room_id")]
    [JsonIgnore]
    public long RoomId { get; set; }
    [Column("account_id")]
    [JsonPropertyName("AccountId")]
    public int AccountId { get; set; }
    [Column("invited_role")]
    [JsonPropertyName("InvitedRole")]
    public int InvitedRole { get; set; }
    [Column("role")]
    [JsonPropertyName("Role")]
    public int Role { get; set; }
    }
}