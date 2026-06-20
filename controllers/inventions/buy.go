package inventions

import (
	"log"
	"net/http"

	"gorm.io/gorm"
	"meow.net/controllers"
	"meow.net/controllers/player"
	"meow.net/db"
	"meow.net/models"
)

const inventionCurrencyType = 2

type buyBalanceUpdate struct {
	Data           models.Invention `json:"Data"`
	UpdateResponse int              `json:"UpdateResponse"`
}

type buyBalanceResponse struct {
	Balance        int                `json:"Balance"`
	BalanceType    int                `json:"BalanceType"`
	BalanceUpdates []buyBalanceUpdate `json:"BalanceUpdates"`
	CurrencyType   int                `json:"CurrencyType"`
}

type buyInventionResponse struct {
	BalanceUpdateResponse buyBalanceResponse `json:"BalanceUpdateResponse"`
	InventionResponse     inventionEnvelope  `json:"InventionResponse"`
}

func BuyInvention(w http.ResponseWriter, r *http.Request) {
	accountID, ok := controllers.AccountIDFromRequest(r)
	if !ok {
		writeError(w, http.StatusUnauthorized, "unauthorized")
		return
	}
	id, ok := parseInt64Param(r, "inventionId")
	if !ok {
		writeError(w, http.StatusBadRequest, "inventionId required")
		return
	}
	requestedPrice, _ := parseIntParam(r, "requestedPrice")

	inv, ok := loadInvention(id)
	if !ok {
		writeError(w, http.StatusNotFound, "invention not found")
		return
	}
	if !inv.IsPublished {
		writeError(w, http.StatusForbidden, "invention not for sale")
		return
	}
	if isCreator(accountID, inv) {
		writeError(w, http.StatusBadRequest, "cannot buy own invention")
		return
	}

	var existing models.InventionOwnership
	if err := db.DB.Where("invention_id = ? AND account_id = ?", inv.InventionId, accountID).First(&existing).Error; err == nil {
		writeError(w, http.StatusConflict, "already owned")
		return
	}

	price := inv.Price
	if requestedPrice != price {
		log.Printf("[INVENTIONS] buy price mismatch: requested=%d actual=%d", requestedPrice, price)
	}

	bal := player.GetOrCreateBalance(accountID, inventionCurrencyType)
	if bal.Amount < price {
		writeError(w, http.StatusPaymentRequired, "insufficient funds")
		return
	}

	tx := db.DB.Begin()
	if price > 0 {
		bal.Amount -= price
		if err := tx.Save(&bal).Error; err != nil {
			tx.Rollback()
			log.Printf("[INVENTIONS] buy save balance error: %v", err)
			writeError(w, http.StatusInternalServerError, "purchase failed")
			return
		}
		if inv.CreatorPlayerId > 0 {
			creatorBal := player.GetOrCreateBalance(uint(inv.CreatorPlayerId), inventionCurrencyType)
			creatorBal.Amount += price
			if err := tx.Save(&creatorBal).Error; err != nil {
				tx.Rollback()
				log.Printf("[INVENTIONS] buy save creator balance error: %v", err)
				writeError(w, http.StatusInternalServerError, "purchase failed")
				return
			}
		}
	}
	ownership := models.InventionOwnership{
		InventionId: inv.InventionId,
		AccountId:   accountID,
	}
	if err := tx.Create(&ownership).Error; err != nil {
		tx.Rollback()
		log.Printf("[INVENTIONS] buy create ownership error: %v", err)
		writeError(w, http.StatusInternalServerError, "purchase failed")
		return
	}
	if err := tx.Model(&models.Invention{}).Where("invention_id = ?", inv.InventionId).
		UpdateColumn("num_downloads", gorm.Expr("num_downloads + ?", 1)).Error; err != nil {
		log.Printf("[INVENTIONS] buy bump downloads error: %v", err)
	}
	if err := tx.Commit().Error; err != nil {
		writeError(w, http.StatusInternalServerError, "purchase failed")
		return
	}
	db.DB.First(&inv, inv.InventionId)
	version, _ := loadCurrentVersion(inv)

	writeJSON(w, http.StatusOK, buyInventionResponse{
		BalanceUpdateResponse: buyBalanceResponse{
			Balance:      bal.Amount,
			BalanceType:  bal.BalanceType,
			CurrencyType: inventionCurrencyType,
			BalanceUpdates: []buyBalanceUpdate{{
				Data:           inv,
				UpdateResponse: 0,
			}},
		},
		InventionResponse: inventionEnvelope{
			Invention:        &inv,
			InventionVersion: &version,
			Status:           0,
		},
	})
}
