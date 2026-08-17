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
			p.ErrorMessage = publicErr(it.err)
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
			log.Printf("[predict] model %s request: %v", cfg.ModelCode, err)
			lastErr = fmt.Errorf("模型服务暂时不可用")
			continue
		}
		body, _ := io.ReadAll(resp.Body)
		resp.Body.Close()
		raw = string(body)
		if resp.StatusCode >= 300 {
			log.Printf("[predict] model %s http %d: %s", cfg.ModelCode, resp.StatusCode, truncate(raw, 300))
			lastErr = fmt.Errorf("模型服务暂时不可用")
			continue
		}
		content, err := extractContent(body)
		if err != nil {
			log.Printf("[predict] model %s extract: %v raw=%s", cfg.ModelCode, err, truncate(raw, 200))
			return model.LLMPredictPayload{}, raw, fmt.Errorf("模型返回内容无效")
		}
		payload, err := parsePayload(lotteryCode, content)
		if err != nil {
			log.Printf("[predict] model %s parse: %v content=%s", cfg.ModelCode, err, truncate(content, 200))
			return model.LLMPredictPayload{}, raw, err
		}
		return payload, raw, nil
	}
	if lastErr == nil {
		lastErr = fmt.Errorf("模型调用失败")
	}
	return model.LLMPredictPayload{}, raw, lastErr
}

func buildPrompt(lotteryCode string, feat Features) string {
	extra := ""
	if feat.Strategy != "" {
		extra = "\n" + feat.Strategy + "\n请按上述权重侧重表现更好的模型思路，并避免重复近期低命中组合。"
	}
	confRule := "必须同时给出 sum(和值) 与 span(跨度=最大号-最小号)；confidence 必须是 (0,1] 的小数（例如 0.62），禁止为 0，禁止省略。"
	switch lotteryCode {
	case consts.LotteryDLT:
		return feat.Summary + extra + "\n请预测下一期大乐透：前区5个(1-35)、后区2个(1-12)。只返回JSON：{\"numbers\":[...5],\"back_numbers\":[...2],\"sum\":0,\"span\":0,\"back_sum\":0,\"back_span\":0,\"confidence\":0.62,\"reason\":\"...\"}。sum/span 为前区，back_sum/back_span 为后区。" + confRule
	case consts.LotteryP3:
		return feat.Summary + extra + "\n请预测下一期排列三三位数字(0-9，有序)。只返回JSON：{\"numbers\":[d1,d2,d3],\"sum\":0,\"span\":0,\"confidence\":0.62,\"reason\":\"...\"}。" + confRule
	default:
		return feat.Summary + extra + "\n请预测下一期快乐8「选十」：从1-80中选出恰好10个不重复号码（不是20个）。只返回JSON：{\"numbers\":[...10],\"sum\":0,\"span\":0,\"confidence\":0.62,\"reason\":\"...\"}。" + confRule
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

func parsePayload(lotteryCode, content string) (model.LLMPredictPayload, error) {
	content = strings.TrimSpace(content)
	if i := strings.Index(content, "{"); i >= 0 {
		if j := strings.LastIndex(content, "}"); j > i {
			content = content[i : j+1]
		}
	}
	var p model.LLMPredictPayload
	if err := json.Unmarshal([]byte(content), &p); err != nil {
		return p, fmt.Errorf("模型输出不是合法 JSON")
	}
	if p.Confidence <= 0 || p.Confidence > 1 {
		return p, fmt.Errorf("置信度无效，必须在 (0,1] 且不能为 0")
	}
	switch lotteryCode {
	case consts.LotteryDLT:
		if len(p.Numbers) != 5 || len(p.BackNumbers) != 2 {
			return p, fmt.Errorf("大乐透号码数量不正确（需前区5后区2）")
		}
	case consts.LotteryP3:
		if len(p.Numbers) != 3 {
			return p, fmt.Errorf("排列三号码数量不正确（需3位）")
		}
	case consts.LotteryKL8:
		// 兼容旧输出：若 numbers 不是 10 个但给了 pick10，则采用选十
		if len(p.Numbers) != 10 && len(p.Pick10) == 10 {
			p.Numbers = p.Pick10
		}
		if len(p.Numbers) != 10 {
			return p, fmt.Errorf("快乐8须选出恰好10个号码")
		}
		p.Pick10 = nil
	}
	fillSumSpan(&p)
	return p, nil
}

func fillSumSpan(p *model.LLMPredictPayload) {
	if len(p.Numbers) > 0 {
		p.Sum, p.Span = sumAndSpan(p.Numbers)
	}
	if len(p.BackNumbers) > 0 {
		p.BackSum, p.BackSpan = sumAndSpan(p.BackNumbers)
	}
}

func sumAndSpan(nums []int) (sum, span int) {
	if len(nums) == 0 {
		return 0, 0
	}
	minN, maxN := nums[0], nums[0]
	for _, n := range nums {
		sum += n
		if n < minN {
			minN = n
		}
		if n > maxN {
			maxN = n
		}
	}
	return sum, maxN - minN
}

// publicErr 对外只给可读原因，不回传上游 API 原文。
func publicErr(err error) string {
	if err == nil {
		return "未知错误"
	}
	msg := strings.TrimSpace(err.Error())
	switch {
	case strings.Contains(msg, "置信度"):
		return "置信度无效，必须在 (0,1] 且不能为 0"
	case strings.Contains(msg, "快乐8"):
		return "快乐8须选出恰好10个号码"
	case strings.Contains(msg, "大乐透"):
		return "大乐透号码数量不正确（需前区5后区2）"
	case strings.Contains(msg, "排列三"):
		return "排列三号码数量不正确（需3位）"
	case strings.Contains(msg, "api key") || strings.Contains(msg, "API_KEY") || strings.Contains(msg, "密钥"):
		return "模型密钥未配置"
	case strings.HasPrefix(msg, "模型"):
		return msg
	default:
		return "模型调用失败"
	}
}

type aggItem struct {
	cfg  model.ModelConfig
	pred model.LLMPredictPayload
	raw  string
	err  error
}

func Aggregate(lotteryCode string, items []aggItem, hitRateFn func(string) float64) model.LLMPredictPayload {
	if len(items) == 0 {
		return model.LLMPredictPayload{Confidence: 0.01, Reason: "无可用模型结果，已优雅降级"}
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
		need = 10
	}
	// 快乐8：优先用选十票；若模型只填了 numbers，则 numbers 已计入 score
	if lotteryCode == consts.LotteryKL8 && len(pickScore) > 0 {
		for n, w := range pickScore {
			score[n] += w
		}
	}
	conf := safeDiv(confSum, weightSum)
	if conf <= 0 {
		conf = 0.01
	}
	final := model.LLMPredictPayload{
		Numbers:    topN(score, need),
		Confidence: conf,
		Reason:     "多模型加权投票。" + strings.Join(reasons, " | "),
	}
	if lotteryCode == consts.LotteryDLT {
		final.BackNumbers = topN(backScore, 2)
	}
	fillSumSpan(&final)
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
