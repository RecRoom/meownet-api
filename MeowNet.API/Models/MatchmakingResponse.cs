using System;
using System.Collections.Generic;
using System.ComponentModel.DataAnnotations;
using System.ComponentModel.DataAnnotations.Schema;
using System.Text.Json.Serialization;

using Microsoft.EntityFrameworkCore;

namespace MeowNet.API.Models
{
    [NotMapped]
    public class MatchmakingResponse
    {
    [JsonPropertyName("ErrorCode")]
    public int ErrorCode { get; set; }
    [JsonPropertyName("RoomInstance")]
    public RoomInstance RoomInstance { get; set; }
    }
}