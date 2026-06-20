using System;
using System.Collections.Generic;
using System.ComponentModel.DataAnnotations;
using System.ComponentModel.DataAnnotations.Schema;
using System.Text.Json.Serialization;

namespace MeowNet.API.Models
{
    public class UploadedPhotoCheer
    {
    [Key]
    [Column("id")]
    public long Id { get; set; }
    [Column("photo_id")]
    public long PhotoId { get; set; }
    [Column("account_id")]
    public long AccountId { get; set; }
    [Column("created_at")]
    public DateTime CreatedAt { get; set; }
    }
}