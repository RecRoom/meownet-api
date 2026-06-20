package player

import (
	"encoding/json"
	"net/http"
	"strconv"
	"strings"

	"meow.net/db"
	"meow.net/models"
	"meow.net/utils"
)

func GetOrCreateBalance(accountID uint, currencyType int) models.Balance {
	var bal models.Balance
	err := db.DB.
		Where("account_id = ? AND currency_type = ?", accountID, currencyType).
		First(&bal).Error
	if err == nil {
		return bal
	}
	bal = models.Balance{
		AccountID:    accountID,
		CurrencyType: currencyType,
		Amount:       0,
		BalanceType:  -2,
	}
	db.DB.Create(&bal)
	return bal
}

func BalanceGet(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	parts := strings.Split(strings.TrimRight(r.URL.Path, "/"), "/")
	currencyType := 0
	if id, err := strconv.Atoi(parts[len(parts)-1]); err == nil {
		currencyType = id
	}

	accountId := 0
	if tokenStr := utils.GetBearerToken(r); tokenStr != "" {
		if sub, err := utils.ParseSubFromJWT(tokenStr); err == nil {
			accountId, _ = strconv.Atoi(sub)
		}
	}

	var bal models.Balance
	if accountId > 0 {
		bal = GetOrCreateBalance(uint(accountId), currencyType)
	} else {
		bal.CurrencyType = currencyType
		bal.Amount = 0
		bal.BalanceType = -2
	}

	json.NewEncoder(w).Encode([]map[string]interface{}{
		{
			"Balance":      bal.Amount,
			"BalanceType":  bal.BalanceType,
			"CurrencyType": currencyType,
		},
	})
}
