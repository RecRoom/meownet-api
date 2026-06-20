package player

import (
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"os"
	"strconv"
	"time"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"

	"meow.net/controllers/hub"
	"meow.net/db"
	"meow.net/models"
	"meow.net/utils"
)

const objectivesCooldown = 2 * time.Minute

const defaultMaxDailyActivityXp = 1000

type configV2 struct {
	LevelProgressionMaps []struct {
		Level      int `json:"Level"`
		RequiredXp int `json:"RequiredXp"`
	} `json:"LevelProgressionMaps"`
	DailyObjectiveCompletionXp int `json:"DailyObjectiveCompletionXp"`
	MaxDailyActivityXp         int `json:"MaxDailyActivityXp"`
}

func loadConfigV2() (*configV2, error) {
	raw, err := os.ReadFile("data/jsons/configv2.json")
	if err != nil {
		return nil, err
	}
	var cfg configV2
	if err := json.Unmarshal(raw, &cfg); err != nil {
		return nil, err
	}
	return &cfg, nil
}

func applyProgressionXP(cfg *configV2, playerId uint, gainedXP int) {
	if gainedXP <= 0 {
		return
	}

	var prog models.Progression
	db.DB.Where(models.Progression{AccountID: playerId}).
		Attrs(models.Progression{Level: 1, XP: 0}).
		FirstOrCreate(&prog)

	prog.XP += gainedXP
	maxLevel := len(cfg.LevelProgressionMaps) - 1
	for prog.Level < maxLevel {
		required := cfg.LevelProgressionMaps[prog.Level].RequiredXp
		if required <= 0 || prog.XP < required {
			break
		}
		prog.XP -= required
		prog.Level++
	}

	db.DB.Save(&prog)
	hub.HubSendProgressionUpdate(int(playerId), prog.Level, prog.XP)
}

func applyCappedActivityXP(cfg *configV2, playerId uint, requestedXP int) {
	if requestedXP <= 0 {
		return
	}
	dailyCap := cfg.MaxDailyActivityXp
	if dailyCap <= 0 {
		dailyCap = defaultMaxDailyActivityXp
	}
	day := time.Now().UTC().Truncate(24 * time.Hour)

	granted := 0
	err := db.DB.Transaction(func(tx *gorm.DB) error {
		var ledger models.DailyXpLedger
		err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).
			Where("account_id = ? AND day = ?", playerId, day).
			First(&ledger).Error
		if errors.Is(err, gorm.ErrRecordNotFound) {
			ledger = models.DailyXpLedger{AccountID: playerId, Day: day}
			if err := tx.Create(&ledger).Error; err != nil {
				return err
			}
		} else if err != nil {
			return err
		}

		remaining := dailyCap - ledger.Xp
		if remaining <= 0 {
			return nil
		}
		granted = requestedXP
		if granted > remaining {
			granted = remaining
		}
		ledger.Xp += granted
		return tx.Save(&ledger).Error
	})
	if err != nil {
		log.Printf("[OBJECTIVES] daily xp ledger error for player %d: %v", playerId, err)
		return
	}
	if granted > 0 {
		applyProgressionXP(cfg, playerId, granted)
	}
}

func ObjectivesV2(w http.ResponseWriter, r *http.Request) {
	log.Printf("[OBJECTIVES] v2 progress")

	tokenStr := utils.GetBearerToken(r)
	if tokenStr == "" {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}
	accountId, err := utils.ParseSubFromJWT(tokenStr)
	if err != nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	var body []struct {
		AdditionalXp  int  `json:"additionalXp"`
		InParty       bool `json:"inParty"`
		ObjectiveType int  `json:"objectiveType"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil || len(body) == 0 {
		w.WriteHeader(http.StatusOK)
		return
	}
	if len(body) > 25 {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	accId, _ := strconv.ParseUint(accountId, 10, 32)
	playerId := uint(accId)

	if !utils.AccountActionAllow("players_v2_objectives", playerId, objectivesCooldown) {
		w.WriteHeader(http.StatusOK)
		return
	}

	totalXP := 0
	for _, e := range body {
		if e.AdditionalXp > 25 {
			continue
		}
		totalXP += e.AdditionalXp
	}
	if totalXP <= 0 {
		w.WriteHeader(http.StatusOK)
		return
	}

	cfg, err := loadConfigV2()
	if err != nil {
		log.Printf("[OBJECTIVES] failed to load config: %v", err)
		w.WriteHeader(http.StatusOK)
		return
	}

	applyCappedActivityXP(cfg, playerId, totalXP)

	w.WriteHeader(http.StatusOK)
}

func UpdateObjective(w http.ResponseWriter, r *http.Request) {
	log.Printf("[OBJECTIVES] updateobjective")

	tokenStr := utils.GetBearerToken(r)
	if tokenStr == "" {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}
	accountId, err := utils.ParseSubFromJWT(tokenStr)
	if err != nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	var body struct {
		Group          int     `json:"Group"`
		Index          int     `json:"Index"`
		IsCompleted    bool    `json:"IsCompleted"`
		IsRewarded     bool    `json:"IsRewarded"`
		Progress       float64 `json:"Progress"`
		VisualProgress float64 `json:"VisualProgress"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		http.Error(w, "Bad request", http.StatusBadRequest)
		return
	}

	accId, _ := strconv.ParseUint(accountId, 10, 32)
	playerId := uint(accId)

	var obj models.Objective
	db.DB.Where(models.Objective{
		AccountID: playerId,
		Group:     body.Group,
		Index:     body.Index,
	}).First(&obj)

	newlyCompleted := body.IsCompleted && !obj.IsCompleted

	obj.AccountID = playerId
	obj.Group = body.Group
	obj.Index = body.Index
	obj.IsCompleted = body.IsCompleted
	obj.IsRewarded = body.IsRewarded
	obj.Progress = body.Progress
	obj.VisualProgress = body.VisualProgress

	if newlyCompleted && !obj.HasClaimedReward {
		obj.HasClaimedReward = true
		if cfg, err := loadConfigV2(); err != nil {
			log.Printf("[OBJECTIVES] failed to load config: %v", err)
		} else {
			applyProgressionXP(cfg, playerId, cfg.DailyObjectiveCompletionXp)
		}
	}

	db.DB.Save(&obj)

	w.WriteHeader(http.StatusOK)
}

func Objectives(w http.ResponseWriter, r *http.Request) {
	log.Printf("[OBJECTIVES] myprogress")

	tokenStr := utils.GetBearerToken(r)
	if tokenStr == "" {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}
	accountId, err := utils.ParseSubFromJWT(tokenStr)
	if err != nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	accId, _ := strconv.ParseUint(accountId, 10, 32)
	playerId := uint(accId)

	var groups []models.ObjectiveGroup
	db.DB.Where("account_id = ?", playerId).Find(&groups)
	if groups == nil {
		groups = []models.ObjectiveGroup{}
	}

	var objectives []models.Objective
	db.DB.Where("account_id = ?", playerId).Find(&objectives)
	if objectives == nil {
		objectives = []models.Objective{}
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"ObjectiveGroups": groups,
		"Objectives":      objectives,
	})
}
