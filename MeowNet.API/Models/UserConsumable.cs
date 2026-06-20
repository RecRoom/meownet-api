using System;
using System.Collections.Generic;
using System.ComponentModel.DataAnnotations;
using System.ComponentModel.DataAnnotations.Schema;
using System.Text.Json.Serialization;

namespace MeowNet.API.Models
{
    public class UserConsumable
    {
    [Key]
    [Column("id")]
    public long ID { get; set; }
    [Column("account_id")]
    public long AccountID { get; set; }
    [Column("consumable_item_desc")]
    public string ConsumableItemDesc { get; set; }
    [Column("active_duration_minutes")]
    public int ActiveDurationMinutes { get; set; }
    [Column("initial_count")]
    public int InitialCount { get; set; }
    [Column("is_active")]
    public bool IsActive { get; set; }
    [Column("is_transferable")]
    public bool IsTransferable { get; set; }
    [Column("created_at")]
    public DateTime CreatedAt { get; set; }
    }
}