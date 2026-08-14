// Package scheduler 定时任务：同步、预测、评估。
package scheduler

import (
	"context"
	"log"
	"time"

	"github.com/robfig/cron/v3"

	"youthlab/lottery-ai/internal/consts"
	"youthlab/lottery-ai/internal/service/lottery"
	"youthlab/lottery-ai/internal/service/predictor"
)

type Scheduler struct {
	cron      *cron.Cron
	Syncer    *lottery.Syncer
	Predictor *predictor.Service
}

func New(syncer *lottery.Syncer, pred *predictor.Service) *Scheduler {
	loc, err := time.LoadLocation("Asia/Shanghai")
	if err != nil {
		loc = time.FixedZone("CST", 8*3600)
	}
	return &Scheduler{
		cron:      cron.New(cron.WithLocation(loc), cron.WithSeconds()),
		Syncer:    syncer,
		Predictor: pred,
	}
}

func (s *Scheduler) Start() {
	// 每天 06:00 和 21:45 同步
	_, _ = s.cron.AddFunc("0 0 6 * * *", func() { s.runSync() })
	_, _ = s.cron.AddFunc("0 45 21 * * *", func() { s.runSync() })
	// 每天 06:10 生成预测
	_, _ = s.cron.AddFunc("0 10 6 * * *", func() { s.runPredict() })
	// 每天 22:05 评估
	_, _ = s.cron.AddFunc("0 5 22 * * *", func() { s.runEvaluate() })
	s.cron.Start()
	log.Printf("[scheduler] started (Asia/Shanghai)")
}

func (s *Scheduler) Stop() {
	ctx := s.cron.Stop()
	<-ctx.Done()
}

func (s *Scheduler) runSync() {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Minute)
	defer cancel()
	log.Printf("[scheduler] sync_history start")
	if err := s.Syncer.SyncAll(ctx); err != nil {
		log.Printf("[scheduler] sync_history error: %v", err)
	}
}

func (s *Scheduler) runPredict() {
	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Minute)
	defer cancel()
	log.Printf("[scheduler] generate_predictions start")
	for _, code := range []string{consts.LotteryDLT, consts.LotteryP3, consts.LotteryKL8} {
		if err := s.Predictor.GenerateToday(ctx, code); err != nil {
			log.Printf("[scheduler] predict %s error: %v", code, err)
		}
	}
}

func (s *Scheduler) runEvaluate() {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Minute)
	defer cancel()
	log.Printf("[scheduler] evaluate_predictions start")
	if err := s.Predictor.EvaluateAll(ctx); err != nil {
		log.Printf("[scheduler] evaluate error: %v", err)
	}
}
