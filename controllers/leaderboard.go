package controllers

import (
	"encoding/json"
	"net/http"

	"meow.net/db"
	"meow.net/models"
)

type leaderboardEntry struct {
	PlayerId uint `json:"playerId"`
	Rank     int  `json:"rank"`
	Score    int  `json:"score"`
}

func computeRank(roomId int, statChannel int, accountId uint, score int) int {
	var higher int64
	db.DB.Model(&models.LeaderboardStat{}).
		Where("room_id = ? AND stat_channel = ? AND (score > ? OR (score = ? AND account_id < ?))",
			roomId, statChannel, score, score, accountId).
		Count(&higher)
	return int(higher)
}

// POST /leaderboard/GetRanks
func LeaderboardGetRanks(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	var body struct {
		FilterType    int  `json:"FilterType"`
		PlayerId      uint `json:"PlayerId"`
		RankStart     int  `json:"RankStart"`
		RankEnd       int  `json:"RankEnd"`
		RoomId        int  `json:"RoomId"`
		SortAscending bool `json:"SortAscending"`
		StatChannel   int  `json:"StatChannel"`
		Timeframe     int  `json:"Timeframe"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		http.Error(w, "bad request", http.StatusBadRequest)
		return
	}

	if body.RankStart < 0 {
		body.RankStart = 0
	}
	limit := body.RankEnd - body.RankStart + 1
	if limit <= 0 {
		json.NewEncoder(w).Encode(map[string]interface{}{"rows": []leaderboardEntry{}})
		return
	}

	order := "score DESC, account_id ASC"
	if body.SortAscending {
		order = "score ASC, account_id ASC"
	}

	var stats []models.LeaderboardStat
	db.DB.Where("room_id = ? AND stat_channel = ?", body.RoomId, body.StatChannel).
		Order(order).
		Offset(body.RankStart).
		Limit(limit).
		Find(&stats)

	rows := make([]leaderboardEntry, 0, len(stats))
	for i, s := range stats {
		rows = append(rows, leaderboardEntry{
			PlayerId: s.AccountID,
			Rank:     body.RankStart + i,
			Score:    s.Score,
		})
	}

	json.NewEncoder(w).Encode(map[string]interface{}{"rows": rows})
}

// POST /leaderboard/GetPlayerRank
func LeaderboardGetPlayerRank(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	var body struct {
		FilterType    int  `json:"FilterType"`
		PlayerId      uint `json:"PlayerId"`
		RoomId        int  `json:"RoomId"`
		SortAscending bool `json:"SortAscending"`
		StatChannel   int  `json:"StatChannel"`
		Timeframe     int  `json:"Timeframe"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		http.Error(w, "bad request", http.StatusBadRequest)
		return
	}

	var stat models.LeaderboardStat
	err := db.DB.Where("account_id = ? AND room_id = ? AND stat_channel = ?",
		body.PlayerId, body.RoomId, body.StatChannel).First(&stat).Error

	resp := leaderboardEntry{PlayerId: body.PlayerId}
	if err == nil {
		resp.Score = stat.Score
		resp.Rank = computeRank(body.RoomId, body.StatChannel, body.PlayerId, stat.Score)
	} else {
		// no stat, rank below everyone else
		var total int64
		db.DB.Model(&models.LeaderboardStat{}).
			Where("room_id = ? AND stat_channel = ?", body.RoomId, body.StatChannel).
			Count(&total)
		resp.Rank = int(total)
		resp.Score = 0
	}

	json.NewEncoder(w).Encode(resp)
}

// POST /leaderboard/GetNearbyScores
func LeaderboardGetNearbyScores(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	var body struct {
		FilterType    int  `json:"FilterType"`
		PlayerId      uint `json:"PlayerId"`
		RoomId        int  `json:"RoomId"`
		SortAscending bool `json:"SortAscending"`
		StatChannel   int  `json:"StatChannel"`
		Timeframe     int  `json:"Timeframe"`
		WindowSize    int  `json:"WindowSize"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		http.Error(w, "bad request", http.StatusBadRequest)
		return
	}
	if body.WindowSize <= 0 {
		body.WindowSize = 10
	}

	order := "score DESC, account_id ASC"
	if body.SortAscending {
		order = "score ASC, account_id ASC"
	}

	var all []models.LeaderboardStat
	db.DB.Where("room_id = ? AND stat_channel = ?", body.RoomId, body.StatChannel).
		Order(order).Find(&all)

	if len(all) == 0 {
		json.NewEncoder(w).Encode(map[string]interface{}{"rows": []leaderboardEntry{}})
		return
	}

	// find player index, default center to bottom if missing
	center := len(all)
	for i, s := range all {
		if s.AccountID == body.PlayerId {
			center = i
			break
		}
	}

	half := body.WindowSize / 2
	start := center - half
	end := start + body.WindowSize
	if start < 0 {
		end -= start
		start = 0
	}
	if end > len(all) {
		end = len(all)
		start = end - body.WindowSize
		if start < 0 {
			start = 0
		}
	}

	rows := make([]leaderboardEntry, 0, end-start)
	for i := start; i < end; i++ {
		rows = append(rows, leaderboardEntry{
			PlayerId: all[i].AccountID,
			Rank:     i,
			Score:    all[i].Score,
		})
	}

	json.NewEncoder(w).Encode(map[string]interface{}{"rows": rows})
}

// POST /leaderboard/CheckAndSetStat
func LeaderboardCheckAndSetStat(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	accountID, ok := AccountIDFromRequest(r)
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	var body struct {
		CurrentStatValue int `json:"CurrentStatValue"`
		RoomId           int `json:"RoomId"`
		StatChannel      int `json:"StatChannel"`
		StatValue        int `json:"StatValue"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		http.Error(w, "bad request", http.StatusBadRequest)
		return
	}

	var stat models.LeaderboardStat
	db.DB.Where("account_id = ? AND room_id = ? AND stat_channel = ?",
		accountID, body.RoomId, body.StatChannel).
		Attrs(models.LeaderboardStat{Score: 0}).
		FirstOrCreate(&stat, models.LeaderboardStat{
			AccountID:   accountID,
			RoomID:      body.RoomId,
			StatChannel: body.StatChannel,
		})

	stat.Score = body.StatValue
	db.DB.Save(&stat)

	rank := computeRank(body.RoomId, body.StatChannel, accountID, stat.Score)

	json.NewEncoder(w).Encode(map[string]interface{}{
		"error":   nil,
		"success": true,
		"value": leaderboardEntry{
			PlayerId: accountID,
			Rank:     rank,
			Score:    stat.Score,
		},
	})
}
