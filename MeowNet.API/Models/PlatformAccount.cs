using System;
using System.Collections.Generic;
using System.ComponentModel.DataAnnotations;
using System.ComponentModel.DataAnnotations.Schema;
using System.Text.Json.Serialization;

namespace MeowNet.API.Models
{
    public class PlatformAccount
    {
    [Key]
    public long ID { get; set; }
    [Column("account_id")]
    public long AccountID { get; set; }
    public Account Account { get; set; }
    [Column("platform")]
    public int Platform { get; set; }
    [Column("platform_id")]
    public string PlatformID { get; set; }
    }
}