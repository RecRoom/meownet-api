using System;
using System.Collections.Generic;
using System.ComponentModel.DataAnnotations;
using System.ComponentModel.DataAnnotations.Schema;
using System.Text.Json.Serialization;

namespace MeowNet.API.Models
{
    public class RoomReport
    {
    [Key]
    public long ID { get; set; }
    [Column("reporter_id")]
    public long ReporterID { get; set; }
    [Column("room_id")]
    public long RoomID { get; set; }
    [Column("report_category")]
    public int ReportCategory { get; set; }
    [Column("details")]
    public string Details { get; set; }
    [Column("created_at")]
    public DateTime CreatedAt { get; set; }
    [Column("resolved")]
    public bool Resolved { get; set; }
    }
}