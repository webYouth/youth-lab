// Package model 定义领域模型。
package model

import (
	"encoding/json"
	"time"
)

type LotteryType struct {
	Code     string          `json:"code"`
	Name     string          `json:"name"`
	Rules    json.RawMessage `json:"rules"`
	DrawDays json.RawMessage `json:"draw_days"`
}

type DrawResult struct {
	ID          int64           `json:"id"`
	LotteryCode string          `json:"lottery_code"`
	Issue       string          `json:"issue"`
	DrawDate    time.Time       `json:"-"`
	DrawDateStr string          `json:"draw_date"`
	Result      json.RawMessage `json:"result"`
	RawData     json.RawMessage `json:"raw_data,omitempty"`
}

func (d DrawResult) MarshalJSON() ([]byte, error) {
	type rest struct {
		ID          int64           `json:"id"`
		LotteryCode string          `json:"lottery_code"`
		Issue       string          `json:"issue"`
		DrawDate    string          `json:"draw_date"`
		Result      json.RawMessage `json:"result"`
		RawData     json.RawMessage `json:"raw_data,omitempty"`
	}
	date := d.DrawDateStr
	if date == "" && !d.DrawDate.IsZero() {
		date = d.DrawDate.UTC().Format("2006-01-02")
	}
	return json.Marshal(rest{
		ID:          d.ID,
		LotteryCode: d.LotteryCode,
		Issue:       d.Issue,
		DrawDate:    date,
		Result:      d.Result,
		RawData:     d.RawData,
	})
}

type ModelConfig struct {
	ID          int64   `json:"id"`
	ModelCode   string  `json:"model_code"`
	DisplayName string  `json:"display_name"`
	Provider    string  `json:"provider"`
	BaseURL     string  `json:"base_url"`
	ModelName   string  `json:"model_name"`
	APIKeyEnv   string  `json:"api_key_env"`
	Weight      float64 `json:"weight"`
	Enabled     bool    `json:"enabled"`
	TimeoutSec  int     `json:"timeout_sec"`
}

type Prediction struct {
	ID               int64           `json:"id"`
	LotteryCode      string          `json:"lottery_code"`
	Issue            string          `json:"issue"`
	PredictDate      time.Time       `json:"predict_date"`
	ModelCode        string          `json:"model_code"`
	PredictedNumbers json.RawMessage `json:"predicted_numbers"`
	Confidence       float64         `json:"confidence"`
	Reason           string          `json:"reason,omitempty"`
	RawResponse      string          `json:"-"` // 仅内部落库，不对外返回
	FinalFlag        bool            `json:"final_flag"`
	Success          bool            `json:"success"`
	ErrorMessage     string          `json:"error_message,omitempty"`
	CreatedAt        time.Time       `json:"created_at"`
}

type PredictionResult struct {
	ID             int64           `json:"id"`
	PredictionID   int64           `json:"prediction_id"`
	DrawResultID   int64           `json:"draw_result_id"`
	MatchedNumbers json.RawMessage `json:"matched_numbers"`
	HitCount       int             `json:"hit_count"`
	HitRate        float64         `json:"hit_rate"`
	Level          string          `json:"level,omitempty"`
	IsWin          bool            `json:"is_win"`
	StakeYuan      float64         `json:"stake_yuan"`
	PrizeYuan      float64         `json:"prize_yuan"`
	ProfitYuan     float64         `json:"profit_yuan"`
	PrizeFloating  bool            `json:"prize_floating"`
	WeightYuan     float64         `json:"weight_yuan"`
	WeightScore    float64         `json:"weight_score"`
	ScoreVersion   int             `json:"score_version"`
}

type AccuracyStat struct {
	LotteryCode      string    `json:"lottery_code"`
	ModelCode        string    `json:"model_code"`
	TotalPredictions int       `json:"total_predictions"`
	TotalHits        int       `json:"total_hits"`
	TotalWins        int       `json:"total_wins"`
	AvgHitRate       float64   `json:"avg_hit_rate"`
	Last30HitRate    float64   `json:"last_30_hit_rate"`
	TotalStake       float64   `json:"total_stake"`
	TotalPrize       float64   `json:"total_prize"`
	TotalProfit      float64   `json:"total_profit"`
	Last30Wins       int       `json:"last_30_wins"`
	Last30Profit     float64   `json:"last_30_profit"`
	UpdatedAt        time.Time `json:"updated_at"`
}

type AccuracyHistory struct {
	Issue         string  `json:"issue"`
	ModelCode     string  `json:"model_code"`
	FinalFlag     bool    `json:"final_flag"`
	DrawDate      string  `json:"draw_date"`
	IsWin         bool    `json:"is_win"`
	Level         string  `json:"level"`
	HitCount      int     `json:"hit_count"`
	StakeYuan     float64 `json:"stake_yuan"`
	PrizeYuan     float64 `json:"prize_yuan"`
	ProfitYuan    float64 `json:"profit_yuan"`
	PrizeFloating bool    `json:"prize_floating"`
}

type ModelStrategy struct {
	ID            int64     `json:"id"`
	LotteryCode   string    `json:"lottery_code"`
	ModelCode     string    `json:"model_code"`
	Weight        float64   `json:"weight"`
	Last30HitRate float64   `json:"last_30_hit_rate"`
	Notes         string    `json:"notes,omitempty"`
	CreatedAt     time.Time `json:"created_at"`
}

// LLMPredictPayload 大模型标准返回。
type LLMPredictPayload struct {
	Numbers     []int   `json:"numbers"`
	BackNumbers []int   `json:"back_numbers,omitempty"`
	Pick10      []int   `json:"pick10,omitempty"`
	Sum         int     `json:"sum"`                 // 号码和值
	Span        int     `json:"span"`                // 跨度 = max-min
	BackSum     int     `json:"back_sum,omitempty"`  // 大乐透后区和值
	BackSpan    int     `json:"back_span,omitempty"` // 大乐透后区跨度
	Confidence  float64 `json:"confidence"`
	Reason      string  `json:"reason"`
}

type AppNotification struct {
	ID        int64           `json:"id"`
	Type      string          `json:"type"`
	Title     string          `json:"title"`
	Body      string          `json:"body"`
	Payload   json.RawMessage `json:"payload"`
	Read      bool            `json:"read"`
	CreatedAt time.Time       `json:"created_at"`
}

type AppUser struct {
	ID           int64     `json:"id"`
	Username     string    `json:"username"`
	PasswordHash string    `json:"-"`
	CreatedAt    time.Time `json:"created_at"`
}
