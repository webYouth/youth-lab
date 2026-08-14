// 彩票 AI 预测后端入口。
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
	"youthlab/lottery-ai/internal/controller"
	"youthlab/lottery-ai/internal/middleware"
	"youthlab/lottery-ai/internal/service/lottery"
	"youthlab/lottery-ai/internal/service/notify"
	"youthlab/lottery-ai/internal/service/predictor"
	"youthlab/lottery-ai/internal/service/scheduler"
	"youthlab/lottery-ai/internal/store"
)

func main() {
	migrateOnly := flag.Bool("migrate", false, "only run sql migrations")
	sqlPath := flag.String("sql", "manifest/sql/001_init.sql", "sql migration path")
	flag.Parse()

	cfg := config.Load()
	log.Printf("starting lottery-ai http=%s", cfg.HTTPAddr)
	if cfg.DatabaseURL == "" {
		log.Fatalf("DATABASE_URL is empty")
	}
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
	log.Printf("migrate ok")
	if *migrateOnly {
		return
	}

	syncer := lottery.NewSyncer(st, cfg.HTTPTimeout, cfg.HistoryN)
	pred := predictor.New(st)
	ntf := notify.New(st)
	sched := scheduler.New(syncer, pred, ntf)
	sched.Start()
	defer sched.Stop()

	api := &controller.API{Store: st, Syncer: syncer, Predictor: pred, Notify: ntf, JWTSecret: cfg.JWTSecret}
	r := gin.Default()
	r.GET("/api/v1/health", api.Health)

	v1 := r.Group("/api/v1")
	{
		v1.POST("/auth/register", api.Register)
		v1.POST("/auth/login", api.Login)
	}

	authed := v1.Group("")
	authed.Use(middleware.RequireUser(cfg.JWTSecret))
	{
		authed.GET("/auth/me", api.Me)
		authed.GET("/lottery-types", api.LotteryTypes)
		authed.GET("/predictions/today", api.PredictionsToday)
		authed.POST("/predictions/run", api.RunPredict)
		authed.GET("/predictions", api.Predictions)
		authed.GET("/draw-results", api.DrawResults)
		authed.GET("/accuracy", api.Accuracy)
		authed.GET("/notifications", api.ListNotifications)
		authed.GET("/notifications/unread", api.UnreadCount)
		authed.POST("/notifications/read-all", api.MarkAllRead)
		authed.POST("/notifications/batch-read", api.BatchSetNotificationRead)
		authed.POST("/notifications/:id/read", api.MarkNotificationRead)
		authed.POST("/notifications/:id/unread", api.MarkNotificationUnread)
		authed.POST("/devices", api.RegisterDevice)
	}

	admin := v1.Group("/admin")
	admin.Use(middleware.AdminToken(cfg.AdminToken))
	{
		admin.POST("/sync", api.AdminSync)
		admin.POST("/generate", api.AdminGenerate)
		admin.POST("/evaluate", api.AdminEvaluate)
		admin.POST("/notify", api.InternalNotify)
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
