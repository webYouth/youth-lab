// Package predictor 开奖后评估命中率。
package predictor

import (
	"context"
	"encoding/json"
	"fmt"
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
	preds, err := s.Store.UnevaluatedPredictions(ctx, lotteryCode)
	if err != nil {
		return err
	}
	for _, p := range preds {
		draw, err := s.Store.GetDrawByIssue(ctx, lotteryCode, p.Issue)
		if err != nil {
			return err
		}
		if draw == nil {
			continue
		}
		matched, hitCount, hitRate, level, isWin := compare(lotteryCode, p.PredictedNumbers, draw.Result)
		matchedJSON, _ := json.Marshal(matched)
		if err := s.Store.InsertPredictionResult(ctx, model.PredictionResult{
			PredictionID:   p.ID,
			DrawResultID:   draw.ID,
			MatchedNumbers: matchedJSON,
			HitCount:       hitCount,
			HitRate:        hitRate,
			Level:          level,
			IsWin:          isWin,
		}); err != nil {
			return err
		}
	}
	if err := s.refreshAccuracy(ctx, lotteryCode); err != nil {
		return err
	}
	return s.Recalibrate(ctx, lotteryCode)
}

func compare(lotteryCode string, predicted, actual json.RawMessage) (matched []int, hitCount int, hitRate float64, level string, isWin bool) {
	var pred model.LLMPredictPayload
	_ = json.Unmarshal(predicted, &pred)
	drawNums := extractNumbers(lotteryCode, actual)
	drawSet := map[int]struct{}{}
	for _, n := range drawNums {
		drawSet[n] = struct{}{}
	}

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
		hitRate = float64(hitCount) / 7.0
		level = formatDLTLevel(fHit, bHit)
		isWin = fHit+bHit >= 3 || (fHit >= 2 && bHit >= 1) || bHit == 2
		return
	case consts.LotteryP3:
		for i, n := range pred.Numbers {
			if i < len(drawNums) && n == drawNums[i] {
				matched = append(matched, n)
			}
		}
		hitCount = len(matched)
		hitRate = float64(hitCount) / 3.0
		level = "直选命中位"
		isWin = hitCount == 3
		return
	default:
		for _, n := range pred.Numbers {
			if _, ok := drawSet[n]; ok {
				matched = append(matched, n)
			}
		}
		hitCount = len(matched)
		denom := 20.0
		if len(pred.Numbers) > 0 {
			denom = float64(len(pred.Numbers))
		}
		hitRate = float64(hitCount) / denom
		level = "号码命中"
		isWin = hitCount >= 5
		return
	}
}

func formatDLTLevel(front, back int) string {
	return fmt.Sprintf("%d+%d", front, back)
}

func (s *Service) refreshAccuracy(ctx context.Context, lotteryCode string) error {
	rows, err := s.Store.Pool().Query(ctx, `
		SELECT p.model_code,
			COUNT(*)::int,
			COALESCE(SUM(CASE WHEN pr.hit_count > 0 THEN 1 ELSE 0 END),0)::int,
			COALESCE(AVG(pr.hit_rate),0)
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
		if err := rows.Scan(&a.ModelCode, &a.TotalPredictions, &a.TotalHits, &a.AvgHitRate); err != nil {
			return err
		}
		_ = s.Store.Pool().QueryRow(ctx, `
			SELECT COALESCE(AVG(pr.hit_rate),0) FROM predictions p
			JOIN prediction_results pr ON pr.prediction_id=p.id
			WHERE p.lottery_code=$1 AND p.model_code=$2 AND p.predict_date >= CURRENT_DATE - INTERVAL '30 days'`,
			lotteryCode, a.ModelCode).Scan(&a.Last30HitRate)
		if err := s.Store.UpsertAccuracy(ctx, a); err != nil {
			return err
		}
	}
	return rows.Err()
}
