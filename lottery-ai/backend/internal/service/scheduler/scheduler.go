package scheduler

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/robfig/cron/v3"

	"youthlab/lottery-ai/internal/consts"
	"youthlab/lottery-ai/internal/service/lottery"
	"youthlab/lottery-ai/internal/service/notify"
	"youthlab/lottery-ai/internal/service/predictor"
)

type Scheduler struct {
	cron      *cron.Cron
	Syncer    *lottery.Syncer
	Predictor *predictor.Service
	Notify    *notify.Service
}

func New(syncer *lottery.Syncer, pred *predictor.Service, ntf *notify.Service) *Scheduler {
	loc, err := time.LoadLocation("Asia/Shanghai")
	if err != nil {
		loc = time.FixedZone("CST", 8*3600)
	}
	return &Scheduler{
		cron:      cron.New(cron.WithLocation(loc), cron.WithSeconds()),
		Syncer:    syncer,
		Predictor: pred,
		Notify:    ntf,
	}
}

func (s *Scheduler) Start() {
	_, _ = s.cron.AddFunc("0 0 6 * * *", func() { s.runSync() })
	_, _ = s.cron.AddFunc("0 45 21 * * *", func() { s.runSync() })
	_, _ = s.cron.AddFunc("0 10 6 * * *", func() { s.runPredict() })
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
	ok := 0
	for _, code := range []string{consts.LotteryDLT, consts.LotteryP3, consts.LotteryKL8} {
		if err := s.Predictor.GenerateToday(ctx, code); err != nil {
			log.Printf("[scheduler] predict %s error: %v", code, err)
			continue
		}
		ok++
	}
	if s.Notify != nil && ok > 0 {
		s.Notify.PublishBestEffort(ctx, "predict", "定时预测已完成", fmt.Sprintf("已更新 %d 个彩种的预测。", ok), map[string]any{"count": ok})
	}
}

func (s *Scheduler) runEvaluate() {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Minute)
	defer cancel()
	log.Printf("[scheduler] evaluate_predictions start")
	if err := s.Predictor.EvaluateAll(ctx); err != nil {
		log.Printf("[scheduler] evaluate error: %v", err)
		return
	}
	if s.Notify != nil {
		s.Notify.PublishBestEffort(ctx, "evaluate", "开奖评估完成", "已根据开奖结果回测并调整模型权重。", map[string]any{"ok": true})
	}
}
