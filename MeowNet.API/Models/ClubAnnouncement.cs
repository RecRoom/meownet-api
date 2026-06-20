using System;
using System.Collections.Generic;
using System.ComponentModel.DataAnnotations;
using System.ComponentModel.DataAnnotations.Schema;
using System.Text.Json.Serialization;

namespace MeowNet.API.Models
{
    public class ClubAnnouncement
    {
    [Key]
    [Column("announcement_id")]
    [JsonPropertyName("AnnouncementId")]
    public long AnnouncementId { get; set; }
    [Column("club_id")]
    [JsonPropertyName("ClubId")]
    public long ClubId { get; set; }
    [Column("account_id")]
    [JsonPropertyName("AccountId")]
    public int AccountId { get; set; }
    [Column("title")]
    [JsonPropertyName("Title")]
    public string Title { get; set; }
    [Column("body")]
    [JsonPropertyName("Body")]
    public string Body { get; set; }
    [Column("image_name")]
    [JsonPropertyName("ImageName")]
    public string ImageName { get; set; }
    [Column("meta")]
    [JsonPropertyName("Meta")]
    public string Meta { get; set; }
    [Column("created_at")]
    [JsonPropertyName("CreatedAt")]
    public DateTime CreatedAt { get; set; }
    }
}