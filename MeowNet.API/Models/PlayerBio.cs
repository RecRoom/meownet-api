using System;
using System.Collections.Generic;
using System.ComponentModel.DataAnnotations;
using System.ComponentModel.DataAnnotations.Schema;
using System.Text.Json.Serialization;

namespace MeowNet.API.Models
{
    public class PlayerBio
    {
    [Key]
    [DatabaseGenerated(DatabaseGeneratedOption.None)]
    [Column("account_id")]
    [JsonPropertyName("accountId")]
    public long AccountID { get; set; }
    [Column("bio")]
    [JsonPropertyName("bio")]
    public string Bio { get; set; }
    }
}