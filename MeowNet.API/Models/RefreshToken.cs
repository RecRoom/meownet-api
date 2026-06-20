using System;
using System.Collections.Generic;
using System.ComponentModel.DataAnnotations;
using System.ComponentModel.DataAnnotations.Schema;
using System.Text.Json.Serialization;

namespace MeowNet.API.Models
{
    public class RefreshToken
    {
    [Key]
    public long ID { get; set; }
    public string Token { get; set; }
    public long AccountID { get; set; }
    [Column("platform_id")]
    public string PlatformID { get; set; }
    [Column("platform")]
    public string Platform { get; set; }
    public DateTime ExpiresAt { get; set; }
    [Column("used_at")]
    public DateTime? UsedAt { get; set; }
    public DateTime CreatedAt { get; set; }
    }
}