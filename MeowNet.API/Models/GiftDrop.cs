using System;
using System.Collections.Generic;
using System.ComponentModel.DataAnnotations;
using System.ComponentModel.DataAnnotations.Schema;
using System.Text.Json.Serialization;

using Microsoft.EntityFrameworkCore;

namespace MeowNet.API.Models
{
    [NotMapped]
    public class GiftDrop
    {
    public int GiftDropId { get; set; }
    public string FriendlyName { get; set; }
    public string Tooltip { get; set; }
    public string? AvatarItemDesc { get; set; }
    public int AvatarItemType { get; set; }
    public string? ConsumableItemDesc { get; set; }
    public string? EquipmentPrefabName { get; set; }
    public string? EquipmentModificationGuid { get; set; }
    public bool IsQuery { get; set; }
    public bool Unique { get; set; }
    public bool SubscribersOnly { get; set; }
    public int Level { get; set; }
    public int Rarity { get; set; }
    public int CurrencyType { get; set; }
    public int Currency { get; set; }
    public int Context { get; set; }
    public int? ItemSetId { get; set; }
    public string? ItemSetFriendlyName { get; set; }
    }
}