using System;
using System.Collections.Generic;
using System.ComponentModel.DataAnnotations;
using System.ComponentModel.DataAnnotations.Schema;
using System.Text.Json.Serialization;

namespace MeowNet.API.Models
{
    public class ClubPermission
    {
    [Key]
    [Column("club_permissions_id")]
    [JsonPropertyName("ClubPermissionsId")]
    public long ClubPermissionsId { get; set; }
    [Column("club_id")]
    [JsonPropertyName("ClubId")]
    public long ClubId { get; set; }
    [Column("type")]
    [JsonPropertyName("Type")]
    public int Type { get; set; }
    [Column("approve_member")]
    [JsonPropertyName("ApproveMember")]
    public bool ApproveMember { get; set; }
    [Column("ban_unban")]
    [JsonPropertyName("BanUnban")]
    public bool BanUnban { get; set; }
    [Column("create_event")]
    [JsonPropertyName("CreateEvent")]
    public bool CreateEvent { get; set; }
    [Column("edit_details")]
    [JsonPropertyName("EditDetails")]
    public bool EditDetails { get; set; }
    [Column("edit_permission_settings")]
    [JsonPropertyName("EditPermissionSettings")]
    public bool EditPermissionSettings { get; set; }
    [Column("post_announcement")]
    [JsonPropertyName("PostAnnouncement")]
    public bool PostAnnouncement { get; set; }
    }
}