using System;
using System.Collections.Generic;
using System.ComponentModel.DataAnnotations;
using System.ComponentModel.DataAnnotations.Schema;
using System.Text.Json.Serialization;

namespace MeowNet.API.Models
{
    public class Gift
    {
    [Key]
    [Column("id")]
    [JsonPropertyName("Id")]
    public long ID { get; set; }
    [Column("account_id")]
    [JsonIgnore]
    public long AccountID { get; set; }
    [Column("from_player_id")]
    [JsonPropertyName("FromPlayerId")]
    public long FromPlayerId { get; set; }
    [Column("message")]
    [JsonPropertyName("Message")]
    public string Message { get; set; }
    [Column("avatar_item_desc")]
    [JsonPropertyName("AvatarItemDesc")]
    public string AvatarItemDesc { get; set; }
    [Column("avatar_item_type")]
    [JsonPropertyName("AvatarItemType")]
    public int AvatarItemType { get; set; }
    [Column("consumable_item_desc")]
    [JsonPropertyName("ConsumableItemDesc")]
    public string ConsumableItemDesc { get; set; }
    [Column("equipment_prefab_name")]
    [JsonPropertyName("EquipmentPrefabName")]
    public string EquipmentPrefabName { get; set; }
    [Column("equipment_modification_guid")]
    [JsonPropertyName("EquipmentModificationGuid")]
    public string EquipmentModificationGuid { get; set; }
    [Column("currency")]
    [JsonPropertyName("Currency")]
    public int Currency { get; set; }
    [Column("currency_type")]
    [JsonPropertyName("CurrencyType")]
    public int CurrencyType { get; set; }
    [Column("balance_type")]
    [JsonPropertyName("BalanceType")]
    public int BalanceType { get; set; }
    [Column("level")]
    [JsonPropertyName("Level")]
    public int Level { get; set; }
    [Column("xp")]
    [JsonPropertyName("Xp")]
    public int Xp { get; set; }
    [Column("gift_context")]
    [JsonPropertyName("GiftContext")]
    public int GiftContext { get; set; }
    [Column("gift_rarity")]
    [JsonPropertyName("GiftRarity")]
    public int GiftRarity { get; set; }
    [Column("platform")]
    [JsonPropertyName("Platform")]
    public int Platform { get; set; }
    [Column("platforms_to_spawn_on")]
    [JsonPropertyName("PlatformsToSpawnOn")]
    public int PlatformsToSpawnOn { get; set; }
    [Column("consumed")]
    [JsonIgnore]
    public bool Consumed { get; set; }
    [Column("created_at")]
    [JsonIgnore]
    public DateTime CreatedAt { get; set; }
    }
}