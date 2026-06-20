using System;
using System.Collections.Generic;
using System.ComponentModel.DataAnnotations;
using System.ComponentModel.DataAnnotations.Schema;
using System.Text.Json.Serialization;

using Microsoft.EntityFrameworkCore;

namespace MeowNet.API.Models
{
    [PrimaryKey(nameof(InstanceID), nameof(AccountID))]
    public class InstanceInvite
    {
    [Column("instance_id")]
    public long InstanceID { get; set; }
    [Column("account_id")]
    public int AccountID { get; set; }
    [Column("invited_by")]
    public int InvitedBy { get; set; }
    [Column("created_at")]
    public DateTime CreatedAt { get; set; }
    }
}