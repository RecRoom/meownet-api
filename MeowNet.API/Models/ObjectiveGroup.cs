using System;
using System.Collections.Generic;
using System.ComponentModel.DataAnnotations;
using System.ComponentModel.DataAnnotations.Schema;
using System.Text.Json.Serialization;

namespace MeowNet.API.Models
{
    public class ObjectiveGroup
    {
    [Key]
    [JsonIgnore]
    public long Id { get; set; }
    [Column("account_id")]
    [JsonIgnore]
    public long AccountID { get; set; }
    [JsonIgnore]
    public Account Account { get; set; }
    [Column("cleared_at")]
    [JsonPropertyName("ClearedAt")]
    public DateTime ClearedAt { get; set; }
    [Column("group_index")]
    [JsonPropertyName("Group")]
    public int Group { get; set; }
    [Column("is_completed")]
    [JsonPropertyName("IsCompleted")]
    public bool IsCompleted { get; set; }
    [Column("requires_complete_on_server")]
    [JsonPropertyName("RequiresCompleteOnServer")]
    public bool RequiresCompleteOnServer { get; set; }
    }
}