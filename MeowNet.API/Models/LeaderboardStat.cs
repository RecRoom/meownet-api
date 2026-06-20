using System;
using System.Collections.Generic;
using System.ComponentModel.DataAnnotations;
using System.ComponentModel.DataAnnotations.Schema;
using System.Text.Json.Serialization;

namespace MeowNet.API.Models
{
    public class LeaderboardStat
    {
    [Key]
    [Column("id")]
    public long ID { get; set; }
    [Column("account_id")]
    public long AccountID { get; set; }
    [Column("room_id")]
    public int RoomID { get; set; }
    [Column("stat_channel")]
    public int StatChannel { get; set; }
    [Column("score")]
    public int Score { get; set; }
    [Column("updated_at")]
    public DateTime UpdatedAt { get; set; }
    }
}