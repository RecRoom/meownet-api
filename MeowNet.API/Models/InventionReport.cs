using System;
using System.Collections.Generic;
using System.ComponentModel.DataAnnotations;
using System.ComponentModel.DataAnnotations.Schema;
using System.Text.Json.Serialization;

namespace MeowNet.API.Models
{
    public class InventionReport
    {
    [Key]
    public long ID { get; set; }
    [Column("reporter_id")]
    public long ReporterID { get; set; }
    [Column("invention_id")]
    public long InventionID { get; set; }
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