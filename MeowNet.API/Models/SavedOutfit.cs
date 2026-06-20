using System;
using System.Collections.Generic;
using System.ComponentModel.DataAnnotations;
using System.ComponentModel.DataAnnotations.Schema;
using System.Text.Json.Serialization;

using Microsoft.EntityFrameworkCore;

namespace MeowNet.API.Models
{
    [PrimaryKey(nameof(AccountID), nameof(Slot))]
    public class SavedOutfit
    {
    [Column("account_id")]
    [JsonIgnore]
    public long AccountID { get; set; }
    [Column("slot")]
    [JsonPropertyName("Slot")]
    public string Slot { get; set; }
    [JsonIgnore]
    public Account Account { get; set; }
    [Column("preview_image_name")]
    [JsonPropertyName("PreviewImageName")]
    public string PreviewImageName { get; set; }
    [Column("outfit_selections")]
    [JsonPropertyName("OutfitSelections")]
    public string OutfitSelections { get; set; }
    [Column("face_features")]
    [JsonPropertyName("FaceFeatures")]
    public string FaceFeatures { get; set; }
    [Column("skin_color")]
    [JsonPropertyName("SkinColor")]
    public string SkinColor { get; set; }
    [Column("hair_color")]
    [JsonPropertyName("HairColor")]
    public string HairColor { get; set; }
    }
}