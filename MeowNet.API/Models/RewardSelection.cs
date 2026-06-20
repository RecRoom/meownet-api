using System;
using System.Collections.Generic;
using System.ComponentModel.DataAnnotations;
using System.ComponentModel.DataAnnotations.Schema;
using System.Text.Json.Serialization;

namespace MeowNet.API.Models
{
    public class RewardSelection
    {
    [Key]
    [Column("id")]
    [JsonPropertyName("RewardSelectionId")]
    public long ID { get; set; }
    [Column("account_id")]
    [JsonIgnore]
    public long AccountID { get; set; }
    [Column("message")]
    [JsonPropertyName("Message")]
    public string Message { get; set; }
    [Column("gift_context")]
    [JsonPropertyName("GiftContext")]
    public int GiftContext { get; set; }
    [Column("reward_type")]
    [JsonPropertyName("RewardType")]
    public int RewardType { get; set; }
    [Column("gift_drop_1_id")]
    [JsonIgnore]
    public int GiftDrop1Id { get; set; }
    [Column("gift_drop_2_id")]
    [JsonIgnore]
    public int GiftDrop2Id { get; set; }
    [Column("gift_drop_3_id")]
    [JsonIgnore]
    public int GiftDrop3Id { get; set; }
    [Column("consumed")]
    [JsonIgnore]
    public bool Consumed { get; set; }
    [Column("created_at")]
    [JsonPropertyName("CreatedAt")]
    public DateTime CreatedAt { get; set; }
    }
}