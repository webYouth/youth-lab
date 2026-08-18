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
	// 早晨全量同步
	_, _ = s.cron.AddFunc("0 0 6 * * *", func() { s.runSync() })

	// 预测按开奖先后：快乐8(21:15) → 排列三(21:30) → 大乐透(周一三/六约21:25)
	_, _ = s.cron.AddFunc("0 10 8 * * *", func() { s.runPredictOne(consts.LotteryKL8) })
	_, _ = s.cron.AddFunc("0 25 8 * * *", func() { s.runPredictOne(consts.LotteryP3) })
	_, _ = s.cron.AddFunc("0 40 8 * * *", func() { s.runPredictOne(consts.LotteryDLT) })

	// 开奖后按顺序同步并评估
	_, _ = s.cron.AddFunc("0 20 21 * * *", func() { s.runSyncOne(consts.LotteryKL8) })     // 快乐8 21:15 后
	_, _ = s.cron.AddFunc("0 35 21 * * *", func() { s.runSyncOne(consts.LotteryP3) })      // 排列三 21:30 后
	_, _ = s.cron.AddFunc("0 40 21 * * 1,3,6", func() { s.runSyncOne(consts.LotteryDLT) }) // 大乐透开奖日
	_, _ = s.cron.AddFunc("0 50 21 * * *", func() { s.runSync() })                         // 兜底再拉一轮
	_, _ = s.cron.AddFunc("0 5 22 * * *", func() { s.runEvaluate() })
	_, _ = s.cron.AddFunc("0 20 22 * * *", func() { s.runSync() }) // 官网晚公布时补拉
	_, _ = s.cron.AddFunc("0 35 22 * * *", func() { s.runEvaluate() })
	_, _ = s.cron.AddFunc("0 15 23 * * *", func() { s.runSync(); s.runEvaluate() })

	s.cron.Start()
	log.Printf("[scheduler] started (Asia/Shanghai) predict=KL8@08:10 P3@08:25 DLT@08:40; draws KL8 21:15 / P3 21:30 / DLT MonWedSat")
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

func (s *Scheduler) runSyncOne(code string) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
	defer cancel()
	log.Printf("[scheduler] sync %s start", code)
	if err := s.Syncer.SyncOne(ctx, code); err != nil {
		log.Printf("[scheduler] sync %s error: %v", code, err)
	}
}

func (s *Scheduler) runPredictOne(code string) {
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Minute)
	defer cancel()
	log.Printf("[scheduler] predict %s start", code)
	_ = s.Syncer.SyncOne(ctx, code)
	if err := s.Predictor.GenerateToday(ctx, code); err != nil {
		log.Printf("[scheduler] predict %s error: %v", code, err)
		return
	}
	if s.Notify != nil {
		name := map[string]string{
			consts.LotteryKL8: "快乐8",
			consts.LotteryP3:  "排列三",
			consts.LotteryDLT: "大乐透",
		}[code]
		s.Notify.PublishBestEffort(ctx, "predict", name+"预测已完成", fmt.Sprintf("%s 下一期预测已生成。", name), map[string]any{"lottery_code": code})
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
