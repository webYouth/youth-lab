// Package controller HTTP API。
package controller

import (
	"log"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"

	"youthlab/lottery-ai/internal/consts"
	"youthlab/lottery-ai/internal/service/lottery"
	"youthlab/lottery-ai/internal/service/notify"
	"youthlab/lottery-ai/internal/service/predictor"
	"youthlab/lottery-ai/internal/store"
)

type API struct {
	Store     *store.Store
	Syncer    *lottery.Syncer
	Predictor *predictor.Service
	Notify    *notify.Service
	JWTSecret string
}

func OK(c *gin.Context, data any) {
	c.JSON(http.StatusOK, gin.H{"code": 0, "message": "ok", "data": data})
}

func Fail(c *gin.Context, code int, msg string) {
	c.JSON(http.StatusOK, gin.H{"code": code, "message": msg, "data": nil})
}

func (a *API) Health(c *gin.Context) {
	OK(c, gin.H{"status": "ok"})
}

func (a *API) LotteryTypes(c *gin.Context) {
	list, err := a.Store.ListLotteryTypes(c.Request.Context())
	if err != nil {
		Fail(c, 500, err.Error())
		return
	}
	OK(c, list)
}

func (a *API) PredictionsToday(c *gin.Context) {
	code := c.DefaultQuery("lottery_code", consts.LotteryDLT)
	day := time.Now().In(loc())
	list, err := a.Store.TodayPredictions(c.Request.Context(), code, day)
	if err != nil {
		Fail(c, 500, err.Error())
		return
	}
	var final any
	models := make([]any, 0)
	for _, p := range list {
		if p.FinalFlag {
			final = p
		} else {
			models = append(models, p)
		}
	}
	OK(c, gin.H{"final": final, "models": models})
}

func (a *API) Predictions(c *gin.Context) {
	code := c.DefaultQuery("lottery_code", consts.LotteryDLT)
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("pageSize", "20"))
	var datePtr *time.Time
	if d := c.Query("date"); d != "" {
		t, err := time.ParseInLocation("2006-01-02", d, loc())
		if err == nil {
			datePtr = &t
		}
	}
	list, total, err := a.Store.ListPredictions(c.Request.Context(), code, datePtr, page, pageSize)
	if err != nil {
		Fail(c, 500, err.Error())
		return
	}
	OK(c, gin.H{"list": list, "total": total, "page": page, "pageSize": pageSize})
}

func (a *API) DrawResults(c *gin.Context) {
	code := c.DefaultQuery("lottery_code", consts.LotteryDLT)
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("pageSize", "20"))
	list, total, err := a.Store.ListDraws(c.Request.Context(), code, page, pageSize)
	if err != nil {
		Fail(c, 500, err.Error())
		return
	}
	OK(c, gin.H{"list": list, "total": total, "page": page, "pageSize": pageSize})
}

func (a *API) Accuracy(c *gin.Context) {
	code := c.DefaultQuery("lottery_code", consts.LotteryDLT)
	list, err := a.Store.GetAccuracy(c.Request.Context(), code)
	if err != nil {
		Fail(c, 500, err.Error())
		return
	}
	strategies, _ := a.Store.LatestStrategies(c.Request.Context(), code)
	history, _ := a.Store.ListAccuracyHistory(c.Request.Context(), code, 40)
	OK(c, gin.H{
		"list":       list,
		"strategies": strategies,
		"history":    history,
		"stake_yuan": 2,
		"days":       c.DefaultQuery("days", "30"),
	})
}

func (a *API) AdminSync(c *gin.Context) {
	code := c.Query("lottery_code")
	var err error
	if code == "" {
		err = a.Syncer.SyncAll(c.Request.Context())
	} else {
		err = a.Syncer.SyncOne(c.Request.Context(), code)
	}
	if err != nil {
		Fail(c, 500, err.Error())
		return
	}
	OK(c, gin.H{"synced": true})
}

func (a *API) AdminGenerate(c *gin.Context) {
	code := c.DefaultQuery("lottery_code", "")
	codes := []string{consts.LotteryDLT, consts.LotteryP3, consts.LotteryKL8}
	if code != "" {
		codes = []string{code}
	}
	for _, x := range codes {
		if err := a.Predictor.GenerateToday(c.Request.Context(), x); err != nil {
			Fail(c, 500, err.Error())
			return
		}
	}
	OK(c, gin.H{"generated": true})
}

func (a *API) AdminEvaluate(c *gin.Context) {
	if err := a.Predictor.EvaluateAll(c.Request.Context()); err != nil {
		Fail(c, 500, err.Error())
		return
	}
	OK(c, gin.H{"evaluated": true})
}

func (a *API) RunPredict(c *gin.Context) {
	code := c.DefaultQuery("lottery_code", consts.LotteryDLT)
	ctx := c.Request.Context()
	_ = a.Syncer.SyncOne(ctx, code)
	if err := a.Predictor.EvaluateOne(ctx, code); err != nil {
		log.Printf("[run-predict] evaluate %s: %v", code, err)
	}
	if err := a.Predictor.Generate(ctx, code, true); err != nil {
		Fail(c, 500, publicFail(err))
		return
	}
	list, err := a.Store.TodayPredictions(ctx, code, time.Now())
	if err != nil {
		Fail(c, 500, "读取预测结果失败")
		return
	}
	var final any
	models := make([]any, 0)
	for _, p := range list {
		if p.FinalFlag {
			final = p
		} else {
			models = append(models, p)
		}
	}
	strategies, _ := a.Store.LatestStrategies(ctx, code)
	if a.Notify != nil {
		a.Notify.PublishBestEffort(ctx, "predict", "预测完成 "+code, "最新一期预测已生成，打开 App 查看。", gin.H{"lottery_code": code})
	}
	OK(c, gin.H{"final": final, "models": models, "strategies": strategies, "generated": true})
}

func loc() *time.Location {
	l, err := time.LoadLocation("Asia/Shanghai")
	if err != nil {
		return time.FixedZone("CST", 8*3600)
	}
	return l
}

// publicFail 接口只返回可读原因，不回传上游 API/堆栈原文。
func publicFail(err error) string {
	if err == nil {
		return "操作失败"
	}
	msg := err.Error()
	if msg == "已有预测任务在执行，请稍后再试" {
		return msg
	}
	return "预测失败，请稍后重试"
}
