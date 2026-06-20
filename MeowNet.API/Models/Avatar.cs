using System;
using System.Collections.Generic;
using System.ComponentModel.DataAnnotations;
using System.ComponentModel.DataAnnotations.Schema;
using System.Text.Json.Serialization;

namespace MeowNet.API.Models
{
    public class Avatar
    {
    [Key]
    [DatabaseGenerated(DatabaseGeneratedOption.None)]
    [Column("account_id")]
    [JsonIgnore]
    public long AccountID { get; set; }
    [Column("face_features")]
    [JsonPropertyName("FaceFeatures")]
    public string FaceFeatures { get; set; }
    [Column("hair_color")]
    [JsonPropertyName("HairColor")]
    public string HairColor { get; set; }
    [Column("outfit_selections")]
    [JsonPropertyName("OutfitSelections")]
    public string OutfitSelections { get; set; }
    [Column("skin_color")]
    [JsonPropertyName("SkinColor")]
    public string SkinColor { get; set; }
    }
}