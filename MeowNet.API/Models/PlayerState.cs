using System;
using System.Collections.Generic;
using System.ComponentModel.DataAnnotations;
using System.ComponentModel.DataAnnotations.Schema;
using System.Text.Json.Serialization;

namespace MeowNet.API.Models
{
    public class PlayerState
    {
    [Key]
    [DatabaseGenerated(DatabaseGeneratedOption.None)]
    [Column("account_id")]
    [JsonIgnore]
    public long AccountID { get; set; }
    [Column("status_visibility")]
    [JsonIgnore]
    public int StatusVisibility { get; set; }
    [Column("vr_movement_mode")]
    [JsonIgnore]
    public int VrMovementMode { get; set; }
    [Column("avoid_juniors")]
    [JsonIgnore]
    public bool AvoidJuniors { get; set; }
    [Column("login_lock_token")]
    [JsonIgnore]
    public string? LoginLockToken { get; set; }
    }
}