using System;
using System.Collections.Generic;
using System.ComponentModel.DataAnnotations;
using System.ComponentModel.DataAnnotations.Schema;
using System.Text.Json.Serialization;

namespace MeowNet.API.Models
{
    public class PlayerReport
    {
    [Key]
    public long ID { get; set; }
    [Column("reporter_id")]
    public long ReporterID { get; set; }
    [Column("reported_player_id")]
    public long ReportedPlayerID { get; set; }
    [Column("report_category")]
    public int ReportCategory { get; set; }
    [Column("details")]
    public string Details { get; set; }
    [Column("height_reporter")]
    public double HeightReporter { get; set; }
    [Column("height_reported")]
    public double HeightReported { get; set; }
    [Column("room_id")]
    public long RoomID { get; set; }
    [Column("room_instance_type")]
    public int RoomInstanceType { get; set; }
    [Column("created_at")]
    public DateTime CreatedAt { get; set; }
    [Column("resolved")]
    public bool Resolved { get; set; }
    }
}