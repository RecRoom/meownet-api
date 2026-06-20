using System;
using System.Collections.Generic;
using System.ComponentModel.DataAnnotations;
using System.ComponentModel.DataAnnotations.Schema;
using System.Text.Json.Serialization;

namespace MeowNet.API.Models
{
    public class Announcement
    {
    [Key]
    [Column("announcement_id")]
    [JsonPropertyName("AnnouncementId")]
    public long AnnouncementId { get; set; }
    [Column("announcement_type")]
    [JsonPropertyName("AnnouncementType")]
    public int AnnouncementType { get; set; }
    [Column("body")]
    [JsonPropertyName("Body")]
    public string Body { get; set; }
    [Column("created_at")]
    [JsonPropertyName("CreatedAt")]
    public DateTime CreatedAt { get; set; }
    [Column("image_name")]
    [JsonPropertyName("ImageName")]
    public string ImageName { get; set; }
    [Column("link_name")]
    [JsonPropertyName("LinkName")]
    public string LinkName { get; set; }
    [Column("link_type")]
    [JsonPropertyName("LinkType")]
    public int LinkType { get; set; }
    [Column("link_uri")]
    [JsonPropertyName("LinkUri")]
    public string LinkUri { get; set; }
    [Column("platform")]
    [JsonPropertyName("Platform")]
    public int Platform { get; set; }
    [Column("title")]
    [JsonPropertyName("Title")]
    public string Title { get; set; }
    }
}