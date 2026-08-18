package predictor

import (
	"context"
	"fmt"
	"log"
	"math"
	"strings"

	"youthlab/lottery-ai/internal/consts"
)

// Recalibrate 根据近30天加权准确率改写各模型投票权重。
func (s *Service) Recalibrate(ctx context.Context, lotteryCode string) error {
	stats, err := s.Store.GetAccuracy(ctx, lotteryCode)
	if err != nil {
		return err
	}
	if len(stats) == 0 {
		return nil
	}
	hits, _ := s.Store.RecentFinalHits(ctx, lotteryCode, 5)
	var hitNotes []string
	for _, h := range hits {
		hitNotes = append(hitNotes, fmt.Sprintf("%s奖%.0f元", h.Level, h.PrizeYuan))
	}
	recent := "尚无开奖回测"
	if len(hitNotes) > 0 {
		recent = strings.Join(hitNotes, "；")
	}

	for _, a := range stats {
		if a.ModelCode == consts.ModelFinal {
			continue
		}
		w := 0.35 + a.Last30HitRate*1.6
		if a.TotalPredictions < 3 {
			w = 1.0
		}
		w = math.Max(0.25, math.Min(2.5, w))
		w = math.Round(w*100) / 100
		note := fmt.Sprintf("近30天加权准确率 %.1f%%，累计盈亏 %.0f 元，投票权重 %.2f。近期回测：%s", a.Last30HitRate*100, a.TotalProfit, w, recent)
		if err := s.Store.UpdateModelWeight(ctx, a.ModelCode, w); err != nil {
			return err
		}
		if err := s.Store.InsertStrategy(ctx, lotteryCode, a.ModelCode, w, a.Last30HitRate, note); err != nil {
			return err
		}
		log.Printf("[strategy] %s/%s weight=%.2f hit30=%.3f", lotteryCode, a.ModelCode, w, a.Last30HitRate)
	}
	return nil
}

func (s *Service) strategyPrompt(ctx context.Context, lotteryCode string) string {
	list, err := s.Store.LatestStrategies(ctx, lotteryCode)
	if err != nil || len(list) == 0 {
		return ""
	}
	var b strings.Builder
	b.WriteString("根据历史预测回测，当前投票策略：")
	for _, st := range list {
		fmt.Fprintf(&b, " %s 权重%.2f(近30天加权%.1f%%)；", st.ModelCode, st.Weight, st.Last30HitRate*100)
	}
	if list[0].Notes != "" {
		b.WriteString(" ")
		b.WriteString(list[0].Notes)
	}
	return b.String()
}
