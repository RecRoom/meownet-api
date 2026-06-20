using System;
using System.Collections.Generic;
using System.ComponentModel.DataAnnotations;
using System.ComponentModel.DataAnnotations.Schema;
using System.Text.Json.Serialization;

namespace MeowNet.API.Models
{
    public class Relationship
    {
    [Key]
    public long ID { get; set; }
    [Column("requester_id")]
    public long RequesterID { get; set; }
    [Column("target_id")]
    public long TargetID { get; set; }
    [Column("relationship_type")]
    public int Type { get; set; }
    [Column("requester_favorited")]
    public int RequesterFavorited { get; set; }
    [Column("requester_ignored")]
    public int RequesterIgnored { get; set; }
    [Column("requester_muted")]
    public int RequesterMuted { get; set; }
    [Column("target_favorited")]
    public int TargetFavorited { get; set; }
    [Column("target_ignored")]
    public int TargetIgnored { get; set; }
    [Column("target_muted")]
    public int TargetMuted { get; set; }
    }
}