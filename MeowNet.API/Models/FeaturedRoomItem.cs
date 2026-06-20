using System;
using System.Collections.Generic;
using System.ComponentModel.DataAnnotations;
using System.ComponentModel.DataAnnotations.Schema;
using System.Text.Json.Serialization;

using Microsoft.EntityFrameworkCore;

namespace MeowNet.API.Models
{
    [NotMapped]
    public class FeaturedRoomItem
    {
    [JsonPropertyName("RoomId")]
    public long RoomId { get; set; }
    [JsonPropertyName("RoomName")]
    public string RoomName { get; set; }
    [JsonPropertyName("ImageName")]
    public string ImageName { get; set; }
    }
}