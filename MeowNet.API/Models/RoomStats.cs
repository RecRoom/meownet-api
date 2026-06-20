using System;
using System.Collections.Generic;
using System.ComponentModel.DataAnnotations;
using System.ComponentModel.DataAnnotations.Schema;
using System.Text.Json.Serialization;

using Microsoft.EntityFrameworkCore;

namespace MeowNet.API.Models
{
    [NotMapped]
    public class RoomStats
    {
    [Column("cheer_count")]
    [JsonPropertyName("CheerCount")]
    public int CheerCount { get; set; }
    [Column("favorite_count")]
    [JsonPropertyName("FavoriteCount")]
    public int FavoriteCount { get; set; }
    [Column("visitor_count")]
    [JsonPropertyName("VisitorCount")]
    public int VisitorCount { get; set; }
    [Column("visit_count")]
    [JsonPropertyName("VisitCount")]
    public int VisitCount { get; set; }
    }
}