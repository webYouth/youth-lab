// Package predictor 统计特征、大模型调用与聚合预测。
package predictor

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"math"
	"net/http"
	"os"
	"sort"
	"strings"
	"sync"
	"time"

	"youthlab/lottery-ai/internal/consts"
	"youthlab/lottery-ai/internal/model"
	"youthlab/lottery-ai/internal/service/lottery"
	"youthlab/lottery-ai/internal/store"
)

type Service struct {
	Store *store.Store
	mu    sync.Mutex
}

func New(st *store.Store) *Service { return &Service{Store: st} }

type Features struct {
	Summary  string         `json:"summary"`
	Hot      []int          `json:"hot"`
	Cold     []int          `json:"cold"`
	Missing  map[string]int `json:"missing"`
	Extras   map[string]any `json:"extras"`
	Strategy string         `json:"strategy,omitempty"`
}

func (s *Service) GenerateToday(ctx context.Context, lotteryCode string) error {
	return s.Generate(ctx, lotteryCode, false)
}

func (s *Service) Generate(ctx context.Context, lotteryCode string, force bool) error {
	if !s.mu.TryLock() {
		return fmt.Errorf("已有预测任务在执行，请稍后再试")
	}
	defer s.mu.Unlock()

	now := time.Now().In(shanghai())
	if !force && !lottery.HasDrawToday(lotteryCode, now) {
		log.Printf("[predict] skip %s: no draw today", lotteryCode)
		return nil
	}
	history, err := s.Store.ListRecentDraws(ctx, lotteryCode, 80)
	if err != nil {
		return err
	}
	feat := ComputeFeatures(lotteryCode, history)
	feat.Strategy = s.strategyPrompt(ctx, lotteryCode)
	latest, _ := s.Store.LatestDraw(ctx, lotteryCode)
	issue := lottery.NextIssueHint(latest, now)
	models, err := s.Store.ListEnabledModels(ctx)
	if err != nil {
		return err
	}

	ch := make(chan aggItem, len(models))
	var wg sync.WaitGroup
	for _, m := range models {
		wg.Add(1)
		go func(cfg model.ModelConfig) {
			defer wg.Done()
			payload, raw, err := callModel(ctx, cfg, lotteryCode, feat)
			ch <- aggItem{cfg: cfg, pred: payload, raw: raw, err: err}
		}(m)
	}
	wg.Wait()
	close(ch)

	var success []aggItem
	for it := range ch {
		nums, _ := json.Marshal(it.pred)
		p := model.Prediction{
			LotteryCode:      lotteryCode,
			Issue:            issue,
			PredictDate:      now,
			ModelCode:        it.cfg.ModelCode,
			PredictedNumbers: nums,
			Confidence:       it.pred.Confidence,
			Reason:           it.pred.Reason,
			RawResponse:      it.raw,
			FinalFlag:        false,
			Success:          it.err == nil,
		}
		if it.err != nil {
			p.ErrorMessage = it.err.Error()
			p.PredictedNumbers = json.RawMessage(`{}`)
			log.Printf("[predict] model %s failed: %v", it.cfg.ModelCode, it.err)
		} else {
			success = append(success, it)
		}
		if _, err := s.Store.InsertPrediction(ctx, p); err != nil {
			log.Printf("[predict] save failed: %v", err)
		}
	}

	final := Aggregate(lotteryCode, success, func(code string) float64 {
		return s.Store.ModelHitRate(ctx, lotteryCode, code)
	})
	finalJSON, _ := json.Marshal(final)
	_, err = s.Store.InsertPrediction(ctx, model.Prediction{
		LotteryCode:      lotteryCode,
		Issue:            issue,
		PredictDate:      now,
		ModelCode:        consts.ModelFinal,
		PredictedNumbers: finalJSON,
		Confidence:       final.Confidence,
		Reason:           final.Reason,
		FinalFlag:        true,
		Success:          true,
	})
	return err
}

func ComputeFeatures(lotteryCode string, draws []model.DrawResult) Features {
	freq := map[int]int{}
	missing := map[int]int{}
	var last []int
	for i, d := range draws {
		nums := extractNumbers(lotteryCode, d.Result)
		if i == 0 {
			last = nums
		}
		seen := map[int]bool{}
		for _, n := range nums {
			freq[n]++
			seen[n] = true
		}
		maxN := 80
		if lotteryCode == consts.LotteryDLT {
			maxN = 35
		}
		if lotteryCode == consts.LotteryP3 {
			maxN = 9
		}
		for n := 0; n <= maxN; n++ {
			if lotteryCode != consts.LotteryP3 && n == 0 {
				continue
			}
			if !seen[n] {
				missing[n]++
			} else if i == 0 {
				missing[n] = 0
			}
		}
	}
	type kv struct{ n, c int }
	var arr []kv
	for n, c := range freq {
		arr = append(arr, kv{n, c})
	}
	sort.Slice(arr, func(i, j int) bool { return arr[i].c > arr[j].c })
	hot, cold := []int{}, []int{}
	for i := 0; i < len(arr) && i < 10; i++ {
		hot = append(hot, arr[i].n)
	}
	sort.Slice(arr, func(i, j int) bool { return arr[i].c < arr[j].c })
	for i := 0; i < len(arr) && i < 10; i++ {
		cold = append(cold, arr[i].n)
	}
	sum := 0
	odd, even, small, big := 0, 0, 0, 0
	for _, n := range last {
		sum += n
		if n%2 == 0 {
			even++
		} else {
			odd++
		}
		mid := 40
		if lotteryCode == consts.LotteryDLT {
			mid = 18
		}
		if lotteryCode == consts.LotteryP3 {
			mid = 5
		}
		if n < mid {
			small++
		} else {
			big++
		}
	}
	missStr := map[string]int{}
	for k, v := range missing {
		missStr[fmt.Sprintf("%d", k)] = v
	}
	summary := fmt.Sprintf("彩种=%s 历史样本=%d 上期号码=%v 和值=%d 奇偶=%d:%d 大小=%d:%d 热号=%v 冷号=%v",
		lotteryCode, len(draws), last, sum, odd, even, small, big, hot, cold)
	return Features{
		Summary: summary,
		Hot:     hot,
		Cold:    cold,
		Missing: missStr,
		Extras: map[string]any{
			"last": last,
			"sum":  sum,
		},
	}
}

func extractNumbers(lotteryCode string, raw json.RawMessage) []int {
	var m map[string]any
	_ = json.Unmarshal(raw, &m)
	switch lotteryCode {
	case consts.LotteryDLT:
		return append(toInts(m["front"]), toInts(m["back"])...)
	case consts.LotteryP3:
		return toInts(m["digits"])
	default:
		return toInts(m["numbers"])
	}
}

func toInts(v any) []int {
	arr, ok := v.([]any)
	if !ok {
		return nil
	}
	out := make([]int, 0, len(arr))
	for _, x := range arr {
		switch t := x.(type) {
		case float64:
			out = append(out, int(t))
		case json.Number:
			n, _ := t.Int64()
			out = append(out, int(n))
		}
	}
	return out
}

func callModel(ctx context.Context, cfg model.ModelConfig, lotteryCode string, feat Features) (model.LLMPredictPayload, string, error) {
	apiKey := os.Getenv(cfg.APIKeyEnv)
	if apiKey == "" {
		// 兼容 DeepSeek 常见命名
		if cfg.APIKeyEnv == "DEEP_SEEK_API_KEY" {
			apiKey = os.Getenv("DEEPSEEK_API_KEY")
		}
	}
	if apiKey == "" {
		return model.LLMPredictPayload{}, "", fmt.Errorf("missing api key env %s", cfg.APIKeyEnv)
	}
	prompt := buildPrompt(lotteryCode, feat)
	reqBody := map[string]any{
		"model": cfg.ModelName,
		"messages": []map[string]string{
			{"role": "system", "content": "你是彩票历史数据分析助手，只输出指定 JSON。"},
			{"role": "user", "content": prompt},
		},
		"temperature": 0.4,
	}
	b, _ := json.Marshal(reqBody)
	url := strings.TrimRight(cfg.BaseURL, "/") + "/chat/completions"
	timeout := time.Duration(cfg.TimeoutSec) * time.Second
	if timeout <= 0 {
		timeout = 60 * time.Second
	}
	client := &http.Client{Timeout: timeout}
	var lastErr error
	var raw string
	for i := 0; i < 2; i++ {
		req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(b))
		if err != nil {
			return model.LLMPredictPayload{}, "", err
		}
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Authorization", "Bearer "+apiKey)
		resp, err := client.Do(req)
		if err != nil {
			lastErr = err
			continue
		}
		body, _ := io.ReadAll(resp.Body)
		resp.Body.Close()
		raw = string(body)
		if resp.StatusCode >= 300 {
			lastErr = fmt.Errorf("status %d: %s", resp.StatusCode, truncate(raw, 300))
			continue
		}
		content, err := extractContent(body)
		if err != nil {
			return model.LLMPredictPayload{}, raw, err
		}
		payload, err := parsePayload(content)
		if err != nil {
			return model.LLMPredictPayload{}, raw, err
		}
		return payload, raw, nil
	}
	return model.LLMPredictPayload{}, raw, lastErr
}

func buildPrompt(lotteryCode string, feat Features) string {
	extra := ""
	if feat.Strategy != "" {
		extra = "\n" + feat.Strategy + "\n请按上述权重侧重表现更好的模型思路，并避免重复近期低命中组合。"
	}
	switch lotteryCode {
	case consts.LotteryDLT:
		return feat.Summary + extra + "\n请预测下一期大乐透：前区5个(1-35)、后区2个(1-12)。只返回JSON：{\"numbers\":[...5],\"back_numbers\":[...2],\"confidence\":0.0,\"reason\":\"...\"}"
	case consts.LotteryP3:
		return feat.Summary + extra + "\n请预测下一期排列三三位数字(0-9，有序)。只返回JSON：{\"numbers\":[d1,d2,d3],\"confidence\":0.0,\"reason\":\"...\"}"
	default:
		return feat.Summary + extra + "\n请预测下一期快乐8：20个开奖号码(1-80)，并给出选十推荐10个。只返回JSON：{\"numbers\":[...20],\"pick10\":[...10],\"confidence\":0.0,\"reason\":\"...\"}"
	}
}

func extractContent(body []byte) (string, error) {
	var resp struct {
		Choices []struct {
			Message struct {
				Content string `json:"content"`
			} `json:"message"`
		} `json:"choices"`
	}
	if err := json.Unmarshal(body, &resp); err != nil {
		return "", err
	}
	if len(resp.Choices) == 0 {
		return "", fmt.Errorf("empty choices")
	}
	return resp.Choices[0].Message.Content, nil
}

func parsePayload(content string) (model.LLMPredictPayload, error) {
	content = strings.TrimSpace(content)
	if i := strings.Index(content, "{"); i >= 0 {
		if j := strings.LastIndex(content, "}"); j > i {
			content = content[i : j+1]
		}
	}
	var p model.LLMPredictPayload
	if err := json.Unmarshal([]byte(content), &p); err != nil {
		return p, fmt.Errorf("invalid json payload: %w", err)
	}
	return p, nil
}

type aggItem struct {
	cfg  model.ModelConfig
	pred model.LLMPredictPayload
	raw  string
	err  error
}

func Aggregate(lotteryCode string, items []aggItem, hitRateFn func(string) float64) model.LLMPredictPayload {
	if len(items) == 0 {
		return model.LLMPredictPayload{Confidence: 0, Reason: "无可用模型结果，已优雅降级"}
	}
	score := map[int]float64{}
	backScore := map[int]float64{}
	pickScore := map[int]float64{}
	var confSum, weightSum float64
	var reasons []string
	for _, it := range items {
		w := it.cfg.Weight * (0.5 + hitRateFn(it.cfg.ModelCode))
		if w <= 0 {
			w = 1
		}
		for _, n := range it.pred.Numbers {
			score[n] += w
		}
		for _, n := range it.pred.BackNumbers {
			backScore[n] += w
		}
		for _, n := range it.pred.Pick10 {
			pickScore[n] += w
		}
		confSum += it.pred.Confidence * w
		weightSum += w
		reasons = append(reasons, it.cfg.ModelCode+":"+it.pred.Reason)
	}
	need := 5
	switch lotteryCode {
	case consts.LotteryP3:
		need = 3
	case consts.LotteryKL8:
		need = 20
	}
	final := model.LLMPredictPayload{
		Numbers:    topN(score, need),
		Confidence: safeDiv(confSum, weightSum),
		Reason:     "多模型加权投票。" + strings.Join(reasons, " | "),
	}
	if lotteryCode == consts.LotteryDLT {
		final.BackNumbers = topN(backScore, 2)
	}
	if lotteryCode == consts.LotteryKL8 {
		final.Pick10 = topN(pickScore, 10)
		if len(final.Pick10) == 0 {
			final.Pick10 = topN(score, 10)
		}
	}
	return final
}

func topN(score map[int]float64, n int) []int {
	type kv struct {
		k int
		v float64
	}
	arr := make([]kv, 0, len(score))
	for k, v := range score {
		arr = append(arr, kv{k, v})
	}
	sort.Slice(arr, func(i, j int) bool {
		if arr[i].v == arr[j].v {
			return arr[i].k < arr[j].k
		}
		return arr[i].v > arr[j].v
	})
	out := make([]int, 0, n)
	for i := 0; i < len(arr) && i < n; i++ {
		out = append(out, arr[i].k)
	}
	sort.Ints(out)
	return out
}

func safeDiv(a, b float64) float64 {
	if b == 0 {
		return 0
	}
	return math.Round(a/b*1000) / 1000
}

func shanghai() *time.Location {
	loc, err := time.LoadLocation("Asia/Shanghai")
	if err != nil {
		return time.FixedZone("CST", 8*3600)
	}
	return loc
}

func truncate(s string, n int) string {
	if len(s) <= n {
		return s
	}
	return s[:n]
}
