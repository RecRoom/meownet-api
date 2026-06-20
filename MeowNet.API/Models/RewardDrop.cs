using System;
using System.Collections.Generic;
using System.ComponentModel.DataAnnotations;
using System.ComponentModel.DataAnnotations.Schema;
using System.Text.Json.Serialization;

namespace MeowNet.API.Models
{
    public class RewardDrop
    {
    [Key]
    [Column("gift_drop_id")]
    public int GiftDropId { get; set; }
    [Column("friendly_name")]
    public string FriendlyName { get; set; }
    [Column("tooltip")]
    public string Tooltip { get; set; }
    [Column("avatar_item_desc")]
    public string? AvatarItemDesc { get; set; }
    [Column("avatar_item_type")]
    public int AvatarItemType { get; set; }
    [Column("consumable_item_desc")]
    public string? ConsumableItemDesc { get; set; }
    [Column("equipment_prefab_name")]
    public string? EquipmentPrefabName { get; set; }
    [Column("equipment_modification_guid")]
    public string? EquipmentModificationGuid { get; set; }
    [Column("is_query")]
    public bool IsQuery { get; set; }
    [Column("unique")]
    public bool Unique { get; set; }
    [Column("subscribers_only")]
    public bool SubscribersOnly { get; set; }
    [Column("level")]
    public int Level { get; set; }
    [Column("rarity")]
    public int Rarity { get; set; }
    [Column("currency_type")]
    public int CurrencyType { get; set; }
    [Column("currency")]
    public int Currency { get; set; }
    [Column("context")]
    public int Context { get; set; }
    [Column("item_set_id")]
    public int? ItemSetId { get; set; }
    [Column("item_set_friendly_name")]
    public string? ItemSetFriendlyName { get; set; }
    }
}