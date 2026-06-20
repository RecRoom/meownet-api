using System;
using System.Collections.Generic;
using System.ComponentModel.DataAnnotations;
using System.ComponentModel.DataAnnotations.Schema;
using System.Text.Json.Serialization;

using Microsoft.EntityFrameworkCore;

namespace MeowNet.API.Models
{
    [PrimaryKey(nameof(AccountId), nameof(RoomId))]
    public class RoomInteraction
    {
    public long AccountId { get; set; }
    public long RoomId { get; set; }
    public bool Cheered { get; set; }
    public bool Favorited { get; set; }
    public bool Visited { get; set; }
    }
}