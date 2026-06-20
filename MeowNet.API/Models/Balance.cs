using System;
using System.Collections.Generic;
using System.ComponentModel.DataAnnotations;
using System.ComponentModel.DataAnnotations.Schema;
using System.Text.Json.Serialization;

namespace MeowNet.API.Models
{
    public class Balance
    {
    [Key]
    [JsonIgnore]
    public long ID { get; set; }
    [Column("account_id")]
    [JsonPropertyName("accountId")]
    public long AccountID { get; set; }
    [JsonIgnore]
    public Account Account { get; set; }
    [Column("currency_type")]
    [JsonPropertyName("currencyType")]
    public int CurrencyType { get; set; }
    [Column("amount")]
    [JsonPropertyName("balance")]
    public int Amount { get; set; }
    [Column("balance_type")]
    [JsonPropertyName("balanceType")]
    public int BalanceType { get; set; }
    }
}