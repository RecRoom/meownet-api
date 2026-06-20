package controllers

import (
	"encoding/json"
	"log"
	"math/rand"
	"net/http"
	"strconv"
	"strings"
	"time"

	"meow.net/controllers/hub"
	"meow.net/db"
	"meow.net/models"
	"meow.net/utils"
)

const rewardCooldown = 5 * time.Minute

const (
	rewardRequestWindow        = 3 * time.Minute
	maxRewardContextsPerWindow = 3
)

type rewardPeriod int

const (
	rewardPeriodDaily rewardPeriod = iota
	rewardPeriodWeekly
)

var rewardClaimLimits = map[int]struct {
	period rewardPeriod
	max    int
}{
	int(models.GiftContextFirstActivity):              {rewardPeriodDaily, 1},
	int(models.GiftContextAllDailyChallengesComplete): {rewardPeriodDaily, 1},
	int(models.GiftContextDailyChallengeComplete):     {rewardPeriodDaily, 10},
	int(models.GiftContextAllWeeklyChallengeComplete): {rewardPeriodWeekly, 1},
	int(models.GiftContextWeeklyChallengeComplete):    {rewardPeriodWeekly, 10},
}

func rewardPeriodStart(p rewardPeriod, now time.Time) time.Time {
	day := now.UTC().Truncate(24 * time.Hour)
	if p == rewardPeriodWeekly {
		offset := (int(day.Weekday()) + 6) % 7
		return day.AddDate(0, 0, -offset)
	}
	return day
}

func rewardClaimAllowed(accountID uint, giftContext int) bool {
	limit, ok := rewardClaimLimits[giftContext]
	if !ok {
		return true
	}
	since := rewardPeriodStart(limit.period, time.Now())
	var claimed int64
	db.DB.Model(&models.RewardSelection{}).
		Where("account_id = ? AND gift_context = ? AND created_at >= ?", accountID, giftContext, since).
		Count(&claimed)
	return claimed < int64(limit.max)
}

func writeRewardRateLimited(w http.ResponseWriter) {
	w.WriteHeader(http.StatusTooManyRequests)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"error":   "rate limit exceeded",
		"success": false,
		"value":   nil,
	})
}

var giftContextStringToInt = func() map[string]int {
	pairs := []struct {
		name string
		val  models.GiftContext
	}{
		{"FirstActivity", models.GiftContextFirstActivity},
		{"GameDrop", models.GiftContextGameDrop},
		{"AllDailyChallengesComplete", models.GiftContextAllDailyChallengesComplete},
		{"AllWeeklyChallengeComplete", models.GiftContextAllWeeklyChallengeComplete},
		{"DailyChallengeComplete", models.GiftContextDailyChallengeComplete},
		{"WeeklyChallengeComplete", models.GiftContextWeeklyChallengeComplete},
		{"LevelUp", models.GiftContextLevelUp},
		{"Paintball_ClearCut", models.GiftContextPaintballClearCut},
		{"Paintball_Homestead", models.GiftContextPaintballHomestead},
		{"Paintball_Quarry", models.GiftContextPaintballQuarry},
		{"Paintball_River", models.GiftContextPaintballRiver},
		{"Paintball_Dam", models.GiftContextPaintballDam},
		{"Paintball_DriveIn", models.GiftContextPaintballDriveIn},
		{"Discgolf_Propulsion", models.GiftContextDiscgolfPropulsion},
		{"Discgolf_Lake", models.GiftContextDiscgolfLake},
		{"Discgolf_ModeCoopCatch", models.GiftContextDiscgolfModeCoopCatch},
		{"Quest_Goblin_A", models.GiftContextQuestGoblinA},
		{"Quest_Goblin_B", models.GiftContextQuestGoblinB},
		{"Quest_Goblin_C", models.GiftContextQuestGoblinC},
		{"Quest_Goblin_S", models.GiftContextQuestGoblinS},
		{"Quest_Goblin_Consumable", models.GiftContextQuestGoblinConsumable},
		{"Quest_Cauldron_A", models.GiftContextQuestCauldronA},
		{"Quest_Cauldron_B", models.GiftContextQuestCauldronB},
		{"Quest_Cauldron_C", models.GiftContextQuestCauldronC},
		{"Quest_Cauldron_S", models.GiftContextQuestCauldronS},
		{"Quest_Cauldron_Consumable", models.GiftContextQuestCauldronConsumable},
		{"Quest_Pirate1_A", models.GiftContextQuestPirate1A},
		{"Quest_Pirate1_B", models.GiftContextQuestPirate1B},
		{"Quest_Pirate1_C", models.GiftContextQuestPirate1C},
		{"Quest_Pirate1_S", models.GiftContextQuestPirate1S},
		{"Quest_Pirate1_X", models.GiftContextQuestPirate1X},
		{"Quest_Pirate1_Consumable", models.GiftContextQuestPirate1Consumable},
		{"Quest_Dracula1_A", models.GiftContextQuestDracula1A},
		{"Quest_Dracula1_B", models.GiftContextQuestDracula1B},
		{"Quest_Dracula1_C", models.GiftContextQuestDracula1C},
		{"Quest_Dracula1_S", models.GiftContextQuestDracula1S},
		{"Quest_Dracula1_X", models.GiftContextQuestDracula1X},
		{"Quest_Dracula1_Consumable", models.GiftContextQuestDracula1Consumable},
		{"Quest_Dracula1_SS", models.GiftContextQuestDracula1SS},
		{"Quest_SciFi_A", models.GiftContextQuestSciFiA},
		{"Quest_SciFi_B", models.GiftContextQuestSciFiB},
		{"Quest_SciFi_C", models.GiftContextQuestSciFiC},
		{"Quest_SciFi_S", models.GiftContextQuestSciFiS},
		{"Quest_SciFi_Consumable", models.GiftContextQuestSciFiConsumable},
		{"StuntRunner_TheMainEvent_A", models.GiftContextStuntRunnerTheMainEventA},
		{"StuntRunner_TheMainEvent_B", models.GiftContextStuntRunnerTheMainEventB},
		{"StuntRunner_TheMainEvent_C", models.GiftContextStuntRunnerTheMainEventC},
		{"StuntRunner_TheMainEvent_D", models.GiftContextStuntRunnerTheMainEventD},
		{"StuntRunner_TheMainEvent_S", models.GiftContextStuntRunnerTheMainEventS},
		{"StuntRunner_TheMainEvent_X", models.GiftContextStuntRunnerTheMainEventX},
		{"StuntRunner_TheMainEvent_Consumable", models.GiftContextStuntRunnerTheMainEventConsumable},
		{"StuntRunner_TheMainEvent_SS", models.GiftContextStuntRunnerTheMainEventSS},
		{"Charades", models.GiftContextCharades},
		{"Soccer", models.GiftContextSoccer},
		{"Paddleball", models.GiftContextPaddleball},
		{"Dodgeball", models.GiftContextDodgeball},
		{"Lasertag", models.GiftContextLasertag},
		{"Bowling", models.GiftContextBowling},
		{"PunchcardChallengeComplete", models.GiftContextPunchcardChallengeComplete},
		{"AllPunchcardChallengesComplete", models.GiftContextAllPunchcardChallengesComplete},
	}
	m := make(map[string]int, len(pairs))
	for _, p := range pairs {
		m[normalizeContextKey(p.name)] = int(p.val)
	}
	m[normalizeContextKey("DailyChallengesComplete")] = int(models.GiftContextAllDailyChallengesComplete)
	m[normalizeContextKey("WeeklyChallengesComplete")] = int(models.GiftContextAllWeeklyChallengeComplete)
	return m
}()

func normalizeContextKey(s string) string {
	return strings.ToLower(strings.ReplaceAll(s, "_", ""))
}

func parseGiftContext(raw string) (int, bool) {
	if raw == "" {
		return 0, false
	}
	if n, err := strconv.Atoi(raw); err == nil {
		return n, true
	}
	if v, ok := giftContextStringToInt[normalizeContextKey(raw)]; ok {
		return v, true
	}
	return 0, false
}

type gameRewardDrop struct {
	GiftDropId                int    `json:"GiftDropId"`
	FriendlyName              string `json:"FriendlyName"`
	Tooltip                   string `json:"Tooltip"`
	ConsumableItemDesc        string `json:"ConsumableItemDesc"`
	AvatarItemDesc            string `json:"AvatarItemDesc"`
	AvatarItemType            int    `json:"AvatarItemType"`
	EquipmentPrefabName       string `json:"EquipmentPrefabName"`
	EquipmentModificationGuid string `json:"EquipmentModificationGuid"`
	IsQuery                   bool   `json:"IsQuery"`
	Unique                    bool   `json:"Unique"`
	SubscribersOnly           bool   `json:"SubscribersOnly"`
	Rarity                    int    `json:"Rarity"`
	CurrencyType              int    `json:"CurrencyType"`
	Currency                  int    `json:"Currency"`
	Context                   int    `json:"Context"`
	ItemSetId                 int    `json:"ItemSetId"`
	ItemSetFriendlyName       string `json:"ItemSetFriendlyName"`
}

func dropToGameReward(d models.GiftDrop) gameRewardDrop {
	itemSetId := 1
	if d.ItemSetId != nil {
		itemSetId = *d.ItemSetId
	}
	return gameRewardDrop{
		GiftDropId:                d.GiftDropId,
		FriendlyName:              d.FriendlyName,
		Tooltip:                   d.Tooltip,
		ConsumableItemDesc:        strOrEmpty(d.ConsumableItemDesc),
		AvatarItemDesc:            strOrEmpty(d.AvatarItemDesc),
		AvatarItemType:            d.AvatarItemType,
		EquipmentPrefabName:       strOrEmpty(d.EquipmentPrefabName),
		EquipmentModificationGuid: strOrEmpty(d.EquipmentModificationGuid),
		IsQuery:                   d.IsQuery,
		Unique:                    d.Unique,
		SubscribersOnly:           d.SubscribersOnly,
		Rarity:                    d.Rarity,
		CurrencyType:              d.CurrencyType,
		Currency:                  d.Currency,
		Context:                   d.Context,
		ItemSetId:                 itemSetId,
		ItemSetFriendlyName:       strOrEmpty(d.ItemSetFriendlyName),
	}
}

func tokenRewardDrop(giftDropId int, amount int, context int) gameRewardDrop {
	return gameRewardDrop{
		GiftDropId:   giftDropId,
		FriendlyName: strconv.Itoa(amount) + " Tokens!",
		Tooltip:      "Winner!",
		CurrencyType: 2,
		Currency:     amount,
		Context:      context,
		ItemSetId:    1,
	}
}

func accountOwnsRewardDrop(accountID uint, d models.RewardDrop) bool {
	if accountID == 0 {
		return false
	}
	if desc := strOrEmpty(d.AvatarItemDesc); desc != "" {
		var count int64
		db.DB.Model(&models.UserAvatarItem{}).
			Where("account_id = ? AND avatar_item_desc = ?", accountID, desc).
			Count(&count)
		if count > 0 {
			return true
		}
	}
	if guid := strOrEmpty(d.EquipmentModificationGuid); guid != "" {
		var count int64
		db.DB.Model(&models.UserEquipment{}).
			Where("account_id = ? AND modification_guid = ?", accountID, guid).
			Count(&count)
		if count > 0 {
			return true
		}
	}
	return false
}

var tokenRewardTiers = []struct {
	amount int
	weight int
}{
	{100, 30},
	{150, 25},
	{200, 20},
	{250, 12},
	{300, 8},
	{400, 4},
	{1000, 1},
}

func randomTokenAmount(rng *rand.Rand) int {
	total := 0
	for _, t := range tokenRewardTiers {
		total += t.weight
	}
	roll := rng.Intn(total)
	for _, t := range tokenRewardTiers {
		if roll < t.weight {
			return t.amount
		}
		roll -= t.weight
	}
	return tokenRewardTiers[0].amount
}

func pickRandomRewardDrops(accountID uint, n int, context int) []models.GiftDrop {
	var all []models.RewardDrop
	db.DB.Where("context = ?", context).Find(&all)
	if len(all) == 0 {
		return nil
	}
	rng := rand.New(rand.NewSource(time.Now().UnixNano()))
	rng.Shuffle(len(all), func(i, j int) { all[i], all[j] = all[j], all[i] })
	out := make([]models.GiftDrop, 0, n)
	for i := 0; i < len(all) && len(out) < n; i++ {
		if accountOwnsRewardDrop(accountID, all[i]) {
			continue
		}
		out = append(out, all[i].ToGiftDrop())
	}
	return out
}

func GameRewardsRequest(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	accountID, ok := AccountIDFromRequest(r)
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	if err := r.ParseForm(); err != nil {
		http.Error(w, "bad request", http.StatusBadRequest)
		return
	}
	message := r.FormValue("Message")
	rewardType := firstNonEmpty(r.FormValue("rewardType"), r.FormValue("RewardType"))
	giftContextRaw := firstNonEmpty(r.FormValue("giftContext"), r.FormValue("GiftContext"))
	giftContext, giftContextOK := parseGiftContext(giftContextRaw)

	if giftContextOK && !rewardClaimAllowed(accountID, giftContext) {
		json.NewEncoder(w).Encode(map[string]interface{}{
			"error":   "",
			"success": true,
			"value":   nil,
		})
		return
	}

	if !utils.AccountActionAllow("gamerewards_request_"+strconv.Itoa(giftContext), accountID, rewardRequestWindow) {
		writeRewardRateLimited(w)
		return
	}
	if !utils.AccountActionAllowBurst("gamerewards_global", accountID, rewardRequestWindow, maxRewardContextsPerWindow) {
		writeRewardRateLimited(w)
		return
	}

	var drops []models.GiftDrop
	if giftContextOK {
		drops = pickRandomRewardDrops(accountID, 3, giftContext)
	}
	if len(drops) < 3 {
		log.Printf("[GAMEREWARDS] only %d drops available for context=%d (raw=%q rewardType=%q), filling remaining slots with token choices",
			len(drops), giftContext, giftContextRaw, rewardType)
	}

	rng := rand.New(rand.NewSource(time.Now().UnixNano()))
	usedAmounts := make(map[int]bool)
	gameDrops := make([]gameRewardDrop, 3)
	dropIDs := make([]int, 3)
	for i := 0; i < 3; i++ {
		if i < len(drops) {
			gameDrops[i] = dropToGameReward(drops[i])
			dropIDs[i] = drops[i].GiftDropId
			continue
		}
		amt := randomTokenAmount(rng)
		for usedAmounts[amt] {
			amt = randomTokenAmount(rng)
		}
		usedAmounts[amt] = true
		gameDrops[i] = tokenRewardDrop(-amt, amt, giftContext)
		dropIDs[i] = -amt
	}
	sel := models.RewardSelection{
		AccountID:   accountID,
		Message:     message,
		GiftContext: giftContext,
		RewardType:  0,
		GiftDrop1Id: dropIDs[0],
		GiftDrop2Id: dropIDs[1],
		GiftDrop3Id: dropIDs[2],
	}
	gameDrop1 := gameDrops[0]
	gameDrop2 := gameDrops[1]
	gameDrop3 := gameDrops[2]
	if err := db.DB.Create(&sel).Error; err != nil {
		log.Printf("[GAMEREWARDS] create selection error: %v", err)
		http.Error(w, "internal", http.StatusInternalServerError)
		return
	}

	notif := map[string]interface{}{
		"RewardSelectionId": sel.ID,
		"Message":           message,
		"GiftContext":       sel.GiftContext,
		"RewardType":        sel.RewardType,
		"GiftDrop1":         gameDrop1,
		"GiftDrop2":         gameDrop2,
		"GiftDrop3":         gameDrop3,
		"CreatedAt":         sel.CreatedAt.UTC().Format("2006-01-02T15:04:05.0000000Z"),
		"PlayerId":          0,
	}
	hub.HubSendToPlayer(int(accountID), hub.NotifFrame("RewardSelectionReceived", notif))

	json.NewEncoder(w).Encode(map[string]interface{}{
		"error":   "",
		"success": true,
		"value":   nil,
	})
}

// POST /api/gamerewards/v1/select
func GameRewardsSelect(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	accountID, ok := AccountIDFromRequest(r)
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	if err := r.ParseForm(); err != nil {
		http.Error(w, "bad request", http.StatusBadRequest)
		return
	}
	rewardSelectionId, _ := strconv.Atoi(r.FormValue("rewardSelectionId"))
	giftDropId, _ := strconv.Atoi(r.FormValue("giftDropId"))
	if giftDropId == 0 {
		http.Error(w, "bad request", http.StatusBadRequest)
		return
	}

	if rewardSelectionId <= 0 {
		http.Error(w, "Forbidden", http.StatusForbidden)
		return
	}
	var sel models.RewardSelection
	if err := db.DB.First(&sel, rewardSelectionId).Error; err != nil {
		http.Error(w, "Forbidden", http.StatusForbidden)
		return
	}
	if sel.AccountID != accountID || sel.Consumed ||
		(sel.GiftDrop1Id != giftDropId && sel.GiftDrop2Id != giftDropId && sel.GiftDrop3Id != giftDropId) {
		http.Error(w, "Forbidden", http.StatusForbidden)
		return
	}

	var drop models.GiftDrop
	if giftDropId < 0 {
		amount := -giftDropId
		drop = models.GiftDrop{
			GiftDropId:   giftDropId,
			FriendlyName: strconv.Itoa(amount) + " Tokens!",
			Tooltip:      "Winner!",
			CurrencyType: 2,
			Currency:     amount,
			Context:      sel.GiftContext,
		}
	} else {
		var rewardDrop models.RewardDrop
		if err := db.DB.Where("gift_drop_id = ?", giftDropId).First(&rewardDrop).Error; err != nil {
			log.Printf("[GAMEREWARDS] reward drop %d not found: %v", giftDropId, err)
			http.Error(w, "not found", http.StatusNotFound)
			return
		}
		drop = rewardDrop.ToGiftDrop()
	}

	res := db.DB.Model(&models.RewardSelection{}).
		Where("id = ? AND consumed = ?", sel.ID, false).
		Update("consumed", true)
	if res.Error != nil || res.RowsAffected == 0 {
		http.Error(w, "Forbidden", http.StatusForbidden)
		return
	}

	gift := models.Gift{
		AccountID:                 accountID,
		FromPlayerId:              1,
		Message:                   sel.Message,
		AvatarItemDesc:            strOrEmpty(drop.AvatarItemDesc),
		AvatarItemType:            drop.AvatarItemType,
		ConsumableItemDesc:        strOrEmpty(drop.ConsumableItemDesc),
		EquipmentPrefabName:       strOrEmpty(drop.EquipmentPrefabName),
		EquipmentModificationGuid: strOrEmpty(drop.EquipmentModificationGuid),
		Currency:                  drop.Currency,
		CurrencyType:              drop.CurrencyType,
		BalanceType:               -2,
		Level:                     drop.Level,
		GiftContext:               drop.Context,
		GiftRarity:                drop.Rarity,
		Platform:                  -1,
		PlatformsToSpawnOn:        -1,
	}
	db.DB.Create(&gift)

	notifMsg := map[string]interface{}{
		"Id":                        gift.ID,
		"FromGiftDropId":            drop.GiftDropId,
		"FromPlayerId":              gift.FromPlayerId,
		"ConsumableItemDesc":        gift.ConsumableItemDesc,
		"AvatarItemDesc":            gift.AvatarItemDesc,
		"EquipmentPrefabName":       gift.EquipmentPrefabName,
		"EquipmentModificationGuid": gift.EquipmentModificationGuid,
		"CurrencyType":              gift.CurrencyType,
		"Currency":                  gift.Currency,
		"Xp":                        gift.Xp,
		"Level":                     gift.Level,
		"Platform":                  gift.Platform,
		"PlatformsToSpawnOn":        gift.PlatformsToSpawnOn,
		"BalanceType":               gift.BalanceType,
		"GiftContext":               gift.GiftContext,
		"GiftRarity":                gift.GiftRarity,
		"Message":                   gift.Message,
		"AvatarItemType":            gift.AvatarItemType,
	}
	hub.HubSendToPlayer(int(accountID), hub.NotifFrame(int(models.GiftPackageRewardSelectionReceived), notifMsg))

	json.NewEncoder(w).Encode(dropToGameReward(drop))
}

func GiftsGenerate(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	accountID, ok := AccountIDFromRequest(r)
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	if !utils.AccountActionAllow("gifts_generate", accountID, rewardCooldown) {
		writeRewardRateLimited(w)
		return
	}

	if err := r.ParseForm(); err != nil {
		http.Error(w, "bad request", http.StatusBadRequest)
		return
	}
	message := r.FormValue("Message")
	giftContextRaw := firstNonEmpty(r.FormValue("GiftContext"), r.FormValue("giftContext"))
	giftContext, giftContextOK := parseGiftContext(giftContextRaw)

	var drops []models.GiftDrop
	if giftContextOK {
		drops = pickRandomRewardDrops(accountID, 1, giftContext)
	}
	var gift models.Gift
	if len(drops) == 0 {
		log.Printf("[GIFTSGENERATE] no drops for context=%d (raw=%q), falling back to random token gift", giftContext, giftContextRaw)
		rng := rand.New(rand.NewSource(time.Now().UnixNano()))
		gift = models.Gift{
			AccountID:          accountID,
			FromPlayerId:       1,
			Message:            message,
			Currency:           randomTokenAmount(rng),
			CurrencyType:       2,
			BalanceType:        -2,
			GiftContext:        giftContext,
			Platform:           -1,
			PlatformsToSpawnOn: -1,
		}
	} else {
		drop := drops[0]
		gift = models.Gift{
			AccountID:                 accountID,
			FromPlayerId:              1,
			Message:                   message,
			AvatarItemDesc:            strOrEmpty(drop.AvatarItemDesc),
			AvatarItemType:            drop.AvatarItemType,
			ConsumableItemDesc:        strOrEmpty(drop.ConsumableItemDesc),
			EquipmentPrefabName:       strOrEmpty(drop.EquipmentPrefabName),
			EquipmentModificationGuid: strOrEmpty(drop.EquipmentModificationGuid),
			Currency:                  drop.Currency,
			CurrencyType:              drop.CurrencyType,
			BalanceType:               -2,
			Level:                     drop.Level,
			GiftContext:               giftContext,
			GiftRarity:                drop.Rarity,
			Platform:                  -1,
			PlatformsToSpawnOn:        -1,
		}
	}
	if err := db.DB.Create(&gift).Error; err != nil {
		log.Printf("[GIFTSGENERATE] create gift error: %v", err)
		http.Error(w, "internal", http.StatusInternalServerError)
		return
	}

	json.NewEncoder(w).Encode(gift)
}

func strOrEmpty(p *string) string {
	if p == nil {
		return ""
	}
	return *p
}
