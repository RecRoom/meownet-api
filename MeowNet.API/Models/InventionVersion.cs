using System;
using System.Collections.Generic;
using System.ComponentModel.DataAnnotations;
using System.ComponentModel.DataAnnotations.Schema;
using System.Text.Json.Serialization;

namespace MeowNet.API.Models
{
    public class InventionVersion
    {
    [Key]
    [Column("id")]
    [JsonIgnore]
    public long Id { get; set; }
    [Column("invention_id")]
    [JsonPropertyName("InventionId")]
    public long InventionId { get; set; }
    [Column("version_number")]
    [JsonPropertyName("VersionNumber")]
    public int VersionNumber { get; set; }
    [Column("blob_name")]
    [JsonPropertyName("BlobName")]
    public string BlobName { get; set; }
    [Column("chips_cost")]
    [JsonPropertyName("ChipsCost")]
    public int ChipsCost { get; set; }
    [Column("cloud_variables_cost")]
    [JsonPropertyName("CloudVariablesCost")]
    public int CloudVariablesCost { get; set; }
    [Column("instantiation_cost")]
    [JsonPropertyName("InstantiationCost")]
    public int InstantiationCost { get; set; }
    [Column("lights_cost")]
    [JsonPropertyName("LightsCost")]
    public int LightsCost { get; set; }
    [Column("replication_id")]
    [JsonPropertyName("ReplicationId")]
    public string ReplicationId { get; set; }
    }
}