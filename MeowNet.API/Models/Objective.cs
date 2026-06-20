using System;
using System.Collections.Generic;
using System.ComponentModel.DataAnnotations;
using System.ComponentModel.DataAnnotations.Schema;
using System.Text.Json.Serialization;

namespace MeowNet.API.Models
{
    public class Objective
    {
    [Key]
    [JsonIgnore]
    public long Id { get; set; }
    [Column("account_id")]
    [JsonIgnore]
    public long AccountID { get; set; }
    [JsonIgnore]
    public Account Account { get; set; }
    [Column("group_index")]
    [JsonPropertyName("Group")]
    public int Group { get; set; }
    [Column("type")]
    [JsonPropertyName("Type")]
    public int Type { get; set; }
    [Column("has_claimed_reward")]
    [JsonPropertyName("HasClaimedReward")]
    public bool HasClaimedReward { get; set; }
    [Column("obj_index")]
    [JsonPropertyName("Index")]
    public int Index { get; set; }
    [Column("is_completed")]
    [JsonPropertyName("IsCompleted")]
    public bool IsCompleted { get; set; }
    [Column("is_rewarded")]
    [JsonPropertyName("IsRewarded")]
    public bool IsRewarded { get; set; }
    [Column("progress")]
    [JsonPropertyName("Progress")]
    public double Progress { get; set; }
    [Column("visual_progress")]
    [JsonPropertyName("VisualProgress")]
    public double VisualProgress { get; set; }
    }
}