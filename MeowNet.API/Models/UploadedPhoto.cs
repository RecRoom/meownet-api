using System;
using System.Collections.Generic;
using System.ComponentModel.DataAnnotations;
using System.ComponentModel.DataAnnotations.Schema;
using System.Text.Json.Serialization;

namespace MeowNet.API.Models
{
    public class UploadedPhoto
    {
    [Key]
    [Column("id")]
    public long ID { get; set; }
    [Column("account_id")]
    public long AccountID { get; set; }
    [Column("image_name")]
    public string ImageName { get; set; }
    [Column("player_ids")]
    public List<int> PlayerIDs { get; set; }
    [Column("saved_image_type")]
    public int Type { get; set; }
    [Column("room_id")]
    public int RoomID { get; set; }
    [Column("player_event_id")]
    public int PlayerEventID { get; set; }
    [Column("accessibility")]
    public int Accessibility { get; set; }
    [Column("cheer_count")]
    public int CheerCount { get; set; }
    [Column("created_at")]
    public DateTime CreatedAt { get; set; }
    }
}