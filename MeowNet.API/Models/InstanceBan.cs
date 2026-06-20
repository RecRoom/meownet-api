using System;
using System.Collections.Generic;
using System.ComponentModel.DataAnnotations;
using System.ComponentModel.DataAnnotations.Schema;
using System.Text.Json.Serialization;

namespace MeowNet.API.Models
{
    public class InstanceBan
    {
    [Key]
    public long ID { get; set; }
    [Column("instance_id")]
    public long InstanceID { get; set; }
    [Column("account_id")]
    public long AccountID { get; set; }
    [Column("issued_by")]
    public long IssuedBy { get; set; }
    [Column("expires_at")]
    public DateTime ExpiresAt { get; set; }
    [Column("created_at")]
    public DateTime CreatedAt { get; set; }
    }
}