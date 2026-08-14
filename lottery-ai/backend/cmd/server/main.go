// 彩票 AI 预测后端入口。
// 免责声明：预测结果仅供个人学习研究，不构成投注建议。
package main

import (
	"context"
	"flag"
	"log"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"

	"youthlab/lottery-ai/internal/config"
	"youthlab/lottery-ai/internal/consts"
	"youthlab/lottery-ai/internal/controller"
	"youthlab/lottery-ai/internal/middleware"
	"youthlab/lottery-ai/internal/service/lottery"
	"youthlab/lottery-ai/internal/service/predictor"
	"youthlab/lottery-ai/internal/service/scheduler"
	"youthlab/lottery-ai/internal/store"
)

func main() {
	migrateOnly := flag.Bool("migrate", false, "only run sql migrations")
	sqlPath := flag.String("sql", "manifest/sql/001_init.sql", "sql migration path")
	flag.Parse()

	cfg := config.Load()
	ctx := context.Background()
	st, err := store.New(ctx, cfg.DatabaseURL)
	if err != nil {
		log.Fatalf("db connect failed: %v", err)
	}
	defer st.Close()

	absSQL := *sqlPath
	if !filepath.IsAbs(absSQL) {
		if _, err := os.Stat(absSQL); err != nil {
			absSQL = filepath.Join("backend", absSQL)
		}
	}
	if err := st.Migrate(ctx, absSQL); err != nil {
		log.Fatalf("migrate failed: %v", err)
	}
	log.Printf("migrate ok; disclaimer=%s", consts.Disclaimer)
	if *migrateOnly {
		return
	}

	syncer := lottery.NewSyncer(st, cfg.HTTPTimeout, cfg.HistoryN)
	pred := predictor.New(st)
	sched := scheduler.New(syncer, pred)
	sched.Start()
	defer sched.Stop()

	api := &controller.API{Store: st, Syncer: syncer, Predictor: pred}
	r := gin.Default()
	r.GET("/api/v1/health", api.Health)

	v1 := r.Group("/api/v1")
	v1.Use(middleware.OptionalBearer(cfg.APIToken))
	{
		v1.GET("/lottery-types", api.LotteryTypes)
		v1.GET("/predictions/today", api.PredictionsToday)
		v1.GET("/predictions", api.Predictions)
		v1.GET("/draw-results", api.DrawResults)
		v1.GET("/accuracy", api.Accuracy)
	}
	admin := v1.Group("/admin")
	admin.Use(middleware.AdminToken(cfg.AdminToken))
	{
		admin.POST("/sync", api.AdminSync)
		admin.POST("/generate", api.AdminGenerate)
		admin.POST("/evaluate", api.AdminEvaluate)
	}

	go func() {
		log.Printf("http listening on %s", cfg.HTTPAddr)
		if err := r.Run(cfg.HTTPAddr); err != nil {
			log.Fatalf("http server error: %v", err)
		}
	}()

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGINT, syscall.SIGTERM)
	<-stop
	log.Printf("shutting down...")
	time.Sleep(300 * time.Millisecond)
}
