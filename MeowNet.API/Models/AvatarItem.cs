using System;
using System.Collections.Generic;
using System.ComponentModel.DataAnnotations;
using System.ComponentModel.DataAnnotations.Schema;
using System.Text.Json.Serialization;

namespace MeowNet.API.Models
{
    public class AvatarItem
    {
    [Key]
    [Column("avatar_item_desc")]
    public string AvatarItemDesc { get; set; }
    [Column("avatar_item_type")]
    public int AvatarItemType { get; set; }
    [Column("friendly_name")]
    public string FriendlyName { get; set; }
    [Column("tool_tip")]
    public string ToolTip { get; set; }
    [Column("rarity")]
    public int Rarity { get; set; }
    }
}