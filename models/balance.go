package models

type Balance struct {
	ID           uint    `gorm:"primaryKey" json:"-"`
	AccountID    uint    `gorm:"column:account_id;uniqueIndex:idx_balance_account_currency" json:"accountId"`
	Account      Account `gorm:"foreignKey:AccountID;references:AccountID;constraint:OnDelete:CASCADE" json:"-"`
	CurrencyType int     `gorm:"column:currency_type;uniqueIndex:idx_balance_account_currency" json:"currencyType"`
	Amount       int     `gorm:"column:amount;default:0" json:"balance"`
	BalanceType  int     `gorm:"column:balance_type;default:-2" json:"balanceType"`
}

func (Balance) TableName() string { return "balances" }
