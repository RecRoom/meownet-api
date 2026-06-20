using System;
using System.Collections.Generic;
using System.ComponentModel.DataAnnotations;
using System.ComponentModel.DataAnnotations.Schema;
using System.Text.Json.Serialization;

namespace MeowNet.API.Models
{
    public class DailyXpLedger
    {
    [Key]
    [Column("id")]
    public long ID { get; set; }
    [Column("account_id")]
    public long AccountID { get; set; }
    [Column("day")]
    public DateTime Day { get; set; }
    [Column("xp")]
    public int Xp { get; set; }
    [Column("updated_at")]
    public DateTime UpdatedAt { get; set; }
    }
}