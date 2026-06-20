using System;
using System.Collections.Generic;
using System.ComponentModel.DataAnnotations;
using System.ComponentModel.DataAnnotations.Schema;
using System.Text.Json.Serialization;

namespace MeowNet.API.Models
{
    public class Account
    {
    [Key]
    [DatabaseGenerated(DatabaseGeneratedOption.Identity)]
    [Column("account_id")]
    [JsonPropertyName("accountId")]
    public long AccountID { get; set; }
    [Column("created_at")]
    [JsonPropertyName("createdAt")]
    public DateTime CreatedAt { get; set; }
    [Column("display_name")]
    [JsonPropertyName("displayName")]
    public string DisplayName { get; set; }
    [Column("is_junior")]
    [JsonPropertyName("isJunior")]
    public bool? IsJunior { get; set; }
    [Column("platforms")]
    [JsonPropertyName("platforms")]
    public int Platforms { get; set; }
    [Column("profile_image")]
    [JsonPropertyName("profileImage")]
    public string ProfileImage { get; set; }
    [Column("username")]
    [JsonPropertyName("username")]
    public string Username { get; set; }
    [Column("raw_username")]
    [JsonIgnore]
    public string RawUsername { get; set; }
    [Column("treat_as_junior")]
    [JsonIgnore]
    public bool TreatAsJunior { get; set; }
    [Column("has_birthday")]
    [JsonIgnore]
    public bool HasBirthday { get; set; }
    [Column("is_developer")]
    [JsonIgnore]
    public bool IsDeveloper { get; set; }
    [Column("is_moderator")]
    [JsonIgnore]
    public bool IsModerator { get; set; }
    [Column("password_hash")]
    [JsonIgnore]
    public string? PasswordHash { get; set; }
    [Column("home_club_id")]
    [JsonIgnore]
    public long? HomeClubId { get; set; }
    [Column("selected_cheer")]
    [JsonIgnore]
    public int SelectedCheer { get; set; }
    [Column("last_online")]
    [JsonIgnore]
    public DateTime? LastOnline { get; set; }
    [Column("no_token")]
    [JsonIgnore]
    public bool NoToken { get; set; }
    [NotMapped]
    [JsonPropertyName("availableUsernameChanges")]
    public int AvailableUsernameChanges { get; set; }
    [Column("email")]
    [JsonPropertyName("email")]
    public string? Email { get; set; }
    [Column("phone")]
    [JsonPropertyName("phone")]
    public string? Phone { get; set; }
    [Column("birthday")]
    [JsonPropertyName("birthday")]
    public string? Birthday { get; set; }
    [Column("junior_state")]
    [JsonPropertyName("juniorState")]
    public int JuniorState { get; set; }
    [Column("parent_account_id")]
    [JsonPropertyName("parentAccountId")]
    public long? ParentAccountID { get; set; }
    }
}