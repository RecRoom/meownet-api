using System;
using System.Collections.Generic;
using System.ComponentModel.DataAnnotations;
using System.ComponentModel.DataAnnotations.Schema;
using System.Text.Json.Serialization;

namespace MeowNet.API.Models
{
    public class UserEquipment
    {
    [Key]
    [Column("id")]
    public long ID { get; set; }
    [Column("account_id")]
    public long AccountID { get; set; }
    [Column("modification_guid")]
    public string ModificationGuid { get; set; }
    [Column("prefab_name")]
    public string PrefabName { get; set; }
    [Column("friendly_name")]
    public string FriendlyName { get; set; }
    [Column("tooltip")]
    public string Tooltip { get; set; }
    [Column("rarity")]
    public int Rarity { get; set; }
    [Column("favorited")]
    public bool Favorited { get; set; }
    }
}