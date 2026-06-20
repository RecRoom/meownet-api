using System;
using System.Collections.Generic;
using System.ComponentModel.DataAnnotations;
using System.ComponentModel.DataAnnotations.Schema;
using System.Text.Json.Serialization;

namespace MeowNet.API.Models
{
    public class ClubMember
    {
    [Key]
    [Column("club_member_id")]
    [JsonPropertyName("ClubMemberId")]
    public long ClubMemberId { get; set; }
    [Column("club_id")]
    [JsonPropertyName("ClubId")]
    public long ClubId { get; set; }
    [Column("account_id")]
    [JsonPropertyName("AccountId")]
    public int AccountId { get; set; }
    [Column("membership_type")]
    [JsonPropertyName("MembershipType")]
    public int MembershipType { get; set; }
    [Column("created_at")]
    [JsonPropertyName("CreatedAt")]
    public DateTime CreatedAt { get; set; }
    }
}