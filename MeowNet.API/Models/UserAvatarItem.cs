using System;
using System.Collections.Generic;
using System.ComponentModel.DataAnnotations;
using System.ComponentModel.DataAnnotations.Schema;
using System.Text.Json.Serialization;

using Microsoft.EntityFrameworkCore;

namespace MeowNet.API.Models
{
    [PrimaryKey(nameof(AccountID), nameof(AvatarItemDesc))]
    public class UserAvatarItem
    {
    [Column("account_id")]
    public long AccountID { get; set; }
    [Column("avatar_item_desc")]
    public string AvatarItemDesc { get; set; }
    [Column("created_at")]
    public DateTime CreatedAt { get; set; }
    }
}