using System;
using System.Collections.Generic;
using System.ComponentModel.DataAnnotations;
using System.ComponentModel.DataAnnotations.Schema;
using System.Text.Json.Serialization;

using Microsoft.EntityFrameworkCore;

namespace MeowNet.API.Models
{
    [PrimaryKey(nameof(Platform), nameof(PlatformID))]
    public class PlatformAccountLimit
    {
    [Column("platform")]
    public int Platform { get; set; }
    [Column("platform_id")]
    public string PlatformID { get; set; }
    [Column("max_accounts")]
    public int MaxAccounts { get; set; }
    }
}