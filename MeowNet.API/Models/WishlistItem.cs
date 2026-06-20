using System;
using System.Collections.Generic;
using System.ComponentModel.DataAnnotations;
using System.ComponentModel.DataAnnotations.Schema;
using System.Text.Json.Serialization;

namespace MeowNet.API.Models
{
    public class WishlistItem
    {
    [Key]
    [Column("wishlist_item_id")]
    [JsonPropertyName("WishlistItemId")]
    public string WishlistItemId { get; set; }
    [Column("account_id")]
    [JsonPropertyName("AccountId")]
    public int AccountId { get; set; }
    [Column("purchasable_item_id")]
    [JsonPropertyName("PurchasableItemId")]
    public int PurchasableItemId { get; set; }
    [Column("created_at")]
    [JsonPropertyName("CreatedAt")]
    public DateTime CreatedAt { get; set; }
    }
}