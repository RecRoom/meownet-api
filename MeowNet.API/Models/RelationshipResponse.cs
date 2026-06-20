using System;
using System.Collections.Generic;
using System.ComponentModel.DataAnnotations;
using System.ComponentModel.DataAnnotations.Schema;
using System.Text.Json.Serialization;

using Microsoft.EntityFrameworkCore;

namespace MeowNet.API.Models
{
    [NotMapped]
    public class RelationshipResponse
    {
    [JsonPropertyName("Favorited")]
    public int Favorited { get; set; }
    [JsonPropertyName("Ignored")]
    public int Ignored { get; set; }
    [JsonPropertyName("Muted")]
    public int Muted { get; set; }
    [JsonPropertyName("PlayerID")]
    public long PlayerID { get; set; }
    [JsonPropertyName("RelationshipType")]
    public int Type { get; set; }
    }
}