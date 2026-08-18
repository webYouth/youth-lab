// Package predictor 开奖后评估命中率。
package predictor

import (
	"context"
	"encoding/json"
	"log"

	"youthlab/lottery-ai/internal/consts"
	"youthlab/lottery-ai/internal/model"
)

func (s *Service) EvaluateAll(ctx context.Context) error {
	for _, code := range []string{consts.LotteryDLT, consts.LotteryP3, consts.LotteryKL8} {
		if err := s.EvaluateOne(ctx, code); err != nil {
			log.Printf("[evaluate] %s failed: %v", code, err)
		}
	}
	return nil
}

func (s *Service) EvaluateOne(ctx context.Context, lotteryCode string) error {
	for {
		preds, err := s.Store.UnevaluatedPredictions(ctx, lotteryCode, ScoreVersion)
		if err != nil {
			return err
		}
		if len(preds) == 0 {
			break
		}
		for _, p := range preds {
			draw, err := s.Store.GetDrawByIssue(ctx, lotteryCode, p.Issue)
			if err != nil {
				return err
			}
			if draw == nil {
				continue
			}
			matched, hitCount, overlap, level, isWin, prize, floating, weightYuan, score := compare(lotteryCode, p.PredictedNumbers, draw.Result)
			matchedJSON, _ := json.Marshal(matched)
			if err := s.Store.InsertPredictionResult(ctx, model.PredictionResult{
				PredictionID:   p.ID,
				DrawResultID:   draw.ID,
				MatchedNumbers: matchedJSON,
				HitCount:       hitCount,
				HitRate:        overlap,
				Level:          level,
				IsWin:          isWin,
				StakeYuan:      StakeYuan,
				PrizeYuan:      prize,
				ProfitYuan:     prize - StakeYuan,
				PrizeFloating:  floating,
				WeightYuan:     weightYuan,
				WeightScore:    score,
				ScoreVersion:   ScoreVersion,
			}); err != nil {
				return err
			}
		}
	}
	if err := s.refreshAccuracy(ctx, lotteryCode); err != nil {
		return err
	}
	return s.Recalibrate(ctx, lotteryCode)
}

func compare(lotteryCode string, predicted, actual json.RawMessage) (matched []int, hitCount int, overlap float64, level string, isWin bool, prize float64, floating bool, weightYuan, score float64) {
	var pred model.LLMPredictPayload
	_ = json.Unmarshal(predicted, &pred)
	drawNums := extractNumbers(lotteryCode, actual)
	drawSet := map[int]struct{}{}
	for _, n := range drawNums {
		drawSet[n] = struct{}{}
	}

	var out prizeOutcome
	switch lotteryCode {
	case consts.LotteryDLT:
		var m map[string]any
		_ = json.Unmarshal(actual, &m)
		front := toInts(m["front"])
		back := toInts(m["back"])
		frontSet, backSet := map[int]struct{}{}, map[int]struct{}{}
		for _, n := range front {
			frontSet[n] = struct{}{}
		}
		for _, n := range back {
			backSet[n] = struct{}{}
		}
		fHit, bHit := 0, 0
		for _, n := range pred.Numbers {
			if _, ok := frontSet[n]; ok {
				matched = append(matched, n)
				fHit++
			}
		}
		for _, n := range pred.BackNumbers {
			if _, ok := backSet[n]; ok {
				matched = append(matched, n)
				bHit++
			}
		}
		hitCount = fHit + bHit
		overlap = float64(hitCount) / 7.0
		out = dltPrize(fHit, bHit)
	case consts.LotteryP3:
		for i, n := range pred.Numbers {
			if i < len(drawNums) && n == drawNums[i] {
				matched = append(matched, n)
			}
		}
		hitCount = len(matched)
		overlap = float64(hitCount) / 3.0
		out = p3Prize(hitCount)
	default:
		nums := pred.Numbers
		if len(pred.Pick10) == 10 && len(nums) != 10 {
			nums = pred.Pick10
		}
		for _, n := range nums {
			if _, ok := drawSet[n]; ok {
				matched = append(matched, n)
			}
		}
		hitCount = len(matched)
		denom := 10.0
		if len(nums) > 0 {
			denom = float64(len(nums))
		}
		overlap = float64(hitCount) / denom
		out = kl8Pick10Prize(hitCount)
	}
	level, isWin, prize, floating = out.Level, out.Win, out.Prize, out.Floating
	weightYuan = out.weightYuan()
	score = weightScore(weightYuan, maxWeightYuan(lotteryCode))
	return
}

func (s *Service) refreshAccuracy(ctx context.Context, lotteryCode string) error {
	rows, err := s.Store.Pool().Query(ctx, `
		SELECT p.model_code,
			COUNT(*)::int,
			COALESCE(SUM(CASE WHEN pr.is_win THEN 1 ELSE 0 END),0)::int,
			COALESCE(AVG(pr.weight_score),0),
			COALESCE(SUM(pr.stake_yuan),0),
			COALESCE(SUM(pr.prize_yuan),0),
			COALESCE(SUM(pr.profit_yuan),0)
		FROM predictions p
		JOIN prediction_results pr ON pr.prediction_id=p.id
		WHERE p.lottery_code=$1 AND p.success=TRUE
		GROUP BY p.model_code`, lotteryCode)
	if err != nil {
		return err
	}
	defer rows.Close()
	for rows.Next() {
		var a model.AccuracyStat
		a.LotteryCode = lotteryCode
		if err := rows.Scan(&a.ModelCode, &a.TotalPredictions, &a.TotalWins, &a.AvgHitRate, &a.TotalStake, &a.TotalPrize, &a.TotalProfit); err != nil {
			return err
		}
		a.TotalHits = a.TotalWins
		_ = s.Store.Pool().QueryRow(ctx, `
			SELECT
				COALESCE(AVG(pr.weight_score),0),
				COALESCE(SUM(CASE WHEN pr.is_win THEN 1 ELSE 0 END),0)::int,
				COALESCE(SUM(pr.profit_yuan),0)
			FROM predictions p
			JOIN prediction_results pr ON pr.prediction_id=p.id
			WHERE p.lottery_code=$1 AND p.model_code=$2 AND p.predict_date >= CURRENT_DATE - INTERVAL '30 days'`,
			lotteryCode, a.ModelCode).Scan(&a.Last30HitRate, &a.Last30Wins, &a.Last30Profit)
		if err := s.Store.UpsertAccuracy(ctx, a); err != nil {
			return err
		}
	}
	return rows.Err()
}
