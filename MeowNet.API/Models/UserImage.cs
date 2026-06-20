using System;
using System.Collections.Generic;
using System.ComponentModel.DataAnnotations;
using System.ComponentModel.DataAnnotations.Schema;
using System.Text.Json.Serialization;

namespace MeowNet.API.Models
{
    public class UserImage
    {
    [Key]
    [Column("id")]
    public long ID { get; set; }
    [Column("account_id")]
    public long AccountID { get; set; }
    [Column("image_name")]
    public string ImageName { get; set; }
    [Column("is_saved")]
    public bool IsSaved { get; set; }
    [Column("created_at")]
    public DateTime CreatedAt { get; set; }
    }
}