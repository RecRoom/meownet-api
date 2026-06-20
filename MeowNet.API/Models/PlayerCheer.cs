using System;
using System.Collections.Generic;
using System.ComponentModel.DataAnnotations;
using System.ComponentModel.DataAnnotations.Schema;
using System.Text.Json.Serialization;

namespace MeowNet.API.Models
{
    public class PlayerCheer
    {
    [Key]
    [Column("id")]
    public long Id { get; set; }
    [Column("from_account_id")]
    public long FromAccountId { get; set; }
    [Column("to_account_id")]
    public long ToAccountId { get; set; }
    [Column("category")]
    public int Category { get; set; }
    [Column("room_id")]
    public long? RoomId { get; set; }
    [Column("anonymous")]
    public bool Anonymous { get; set; }
    [Column("created_at")]
    public DateTime CreatedAt { get; set; }
    }
}