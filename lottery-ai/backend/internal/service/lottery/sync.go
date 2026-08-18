// Package lottery 实现历史开奖抓取与解析。
package lottery

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"strconv"
	"strings"
	"time"

	"youthlab/lottery-ai/internal/consts"
	"youthlab/lottery-ai/internal/model"
	"youthlab/lottery-ai/internal/store"
)

type Syncer struct {
	Store   *store.Store
	Client  *http.Client
	History int
}

func NewSyncer(st *store.Store, timeout time.Duration, historyN int) *Syncer {
	if historyN <= 0 {
		historyN = consts.DefaultHistoryN
	}
	return &Syncer{
		Store:   st,
		Client:  &http.Client{Timeout: timeout},
		History: historyN,
	}
}

func (s *Syncer) SyncAll(ctx context.Context) error {
	for _, code := range []string{consts.LotteryDLT, consts.LotteryP3, consts.LotteryKL8} {
		if err := s.SyncOne(ctx, code); err != nil {
			log.Printf("[sync] %s failed: %v", code, err)
		}
	}
	return nil
}

func (s *Syncer) SyncOne(ctx context.Context, lotteryCode string) error {
	switch lotteryCode {
	case consts.LotteryDLT:
		return s.syncSporttery(ctx, lotteryCode, "85")
	case consts.LotteryP3:
		return s.syncSporttery(ctx, lotteryCode, "35")
	case consts.LotteryKL8:
		return s.syncKL8(ctx)
	default:
		return fmt.Errorf("unsupported lottery: %s", lotteryCode)
	}
}

func (s *Syncer) syncSporttery(ctx context.Context, lotteryCode, gameNo string) error {
	latest, _ := s.Store.LatestDraw(ctx, lotteryCode)
	pageSize := 100
	maxPages := (s.History + pageSize - 1) / pageSize
	if latest != nil {
		maxPages = 5 // 增量多拉几页，避免漏掉最新期
	}
	for page := 1; page <= maxPages; page++ {
		url := fmt.Sprintf("https://webapi.sporttery.cn/gateway/lottery/getHistoryPageListV1.qry?gameNo=%s&provinceId=0&pageSize=%d&isVerify=1&pageNo=%d", gameNo, pageSize, page)
		body, status, err := s.get(ctx, url)
		s.Store.SaveRawLog(ctx, lotteryCode, url, status, string(body))
		if err != nil {
			return err
		}
		items, err := parseSporttery(body, lotteryCode)
		if err != nil {
			return err
		}
		if len(items) == 0 {
			break
		}
		hitKnown := false
		for _, item := range items {
			if item.DrawDate.IsZero() {
				continue
			}
			if err := s.Store.UpsertDraw(ctx, item); err != nil {
				return err
			}
			if latest != nil && !issueNewer(item.Issue, latest.Issue) {
				hitKnown = true
			}
		}
		if latest != nil && hitKnown {
			return nil
		}
		if latest == nil && page*pageSize >= s.History {
			break
		}
	}
	return nil
}

func (s *Syncer) syncKL8(ctx context.Context) error {
	url := fmt.Sprintf("https://www.cwl.gov.cn/cwl_admin/front/cwlkj/search/kjxx/findDrawNotice?name=kl8&issueCount=%d&pageNo=1&pageSize=%d&systemType=PC", s.History, min(s.History, 100))
	body, status, err := s.get(ctx, url)
	s.Store.SaveRawLog(ctx, consts.LotteryKL8, url, status, string(body))
	if err != nil {
		return err
	}
	items, err := parseKL8(body)
	if err != nil {
		return err
	}
	for _, item := range items {
		if err := s.Store.UpsertDraw(ctx, item); err != nil {
			return err
		}
	}
	return nil
}

func (s *Syncer) get(ctx context.Context, url string) ([]byte, int, error) {
	var lastErr error
	for i := 0; i < 3; i++ {
		req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
		if err != nil {
			return nil, 0, err
		}
		req.Header.Set("User-Agent", "Mozilla/5.0 (compatible; youth-lab-lottery-ai/1.0)")
		req.Header.Set("Referer", "https://www.cwl.gov.cn/")
		req.Header.Set("Accept", "application/json,text/plain,*/*")
		resp, err := s.Client.Do(req)
		if err != nil {
			lastErr = err
			time.Sleep(time.Duration(i+1) * time.Second)
			continue
		}
		body, err := io.ReadAll(resp.Body)
		resp.Body.Close()
		if err != nil {
			lastErr = err
			continue
		}
		if resp.StatusCode >= 500 {
			lastErr = fmt.Errorf("status %d", resp.StatusCode)
			time.Sleep(time.Duration(i+1) * time.Second)
			continue
		}
		return body, resp.StatusCode, nil
	}
	return nil, 0, lastErr
}

func parseSporttery(body []byte, lotteryCode string) ([]model.DrawResult, error) {
	var root map[string]any
	if err := json.Unmarshal(body, &root); err != nil {
		return nil, err
	}
	value, _ := root["value"].(map[string]any)
	if value == nil {
		return nil, fmt.Errorf("sporttery response missing value")
	}
	list, _ := value["list"].([]any)
	out := make([]model.DrawResult, 0, len(list))
	for _, item := range list {
		m, _ := item.(map[string]any)
		if m == nil {
			continue
		}
		issue := fmt.Sprint(m["lotteryDrawNum"])
		dateStr := fmt.Sprint(m["lotteryDrawTime"])
		drawDate := ParseCivilDate(dateStr)
		resultJSON, err := sportteryResult(lotteryCode, fmt.Sprint(m["lotteryDrawResult"]))
		if err != nil {
			continue
		}
		raw, _ := json.Marshal(m)
		out = append(out, model.DrawResult{
			LotteryCode: lotteryCode,
			Issue:       issue,
			DrawDate:    drawDate,
			Result:      resultJSON,
			RawData:     raw,
		})
	}
	return out, nil
}

func sportteryResult(lotteryCode, drawResult string) (json.RawMessage, error) {
	parts := strings.Fields(strings.ReplaceAll(drawResult, "+", " "))
	nums := make([]int, 0, len(parts))
	for _, p := range parts {
		p = strings.TrimSpace(p)
		if p == "" {
			continue
		}
		n, err := strconv.Atoi(p)
		if err != nil {
			continue
		}
		nums = append(nums, n)
	}
	switch lotteryCode {
	case consts.LotteryDLT:
		if len(nums) < 7 {
			return nil, fmt.Errorf("invalid dlt result")
		}
		payload := map[string]any{"front": nums[:5], "back": nums[5:7]}
		return json.Marshal(payload)
	case consts.LotteryP3:
		if len(nums) < 3 {
			return nil, fmt.Errorf("invalid p3 result")
		}
		payload := map[string]any{"digits": nums[:3]}
		return json.Marshal(payload)
	default:
		return nil, fmt.Errorf("unsupported")
	}
}

func parseKL8(body []byte) ([]model.DrawResult, error) {
	var root struct {
		State  int `json:"state"`
		Result []struct {
			Code string `json:"code"`
			Date string `json:"date"`
			Red  string `json:"red"`
		} `json:"result"`
	}
	if err := json.Unmarshal(body, &root); err != nil {
		return nil, err
	}
	out := make([]model.DrawResult, 0, len(root.Result))
	for _, r := range root.Result {
		drawDate := ParseCivilDate(r.Date)
		parts := strings.Split(r.Red, ",")
		nums := make([]int, 0, len(parts))
		for _, p := range parts {
			n, err := strconv.Atoi(strings.TrimSpace(p))
			if err == nil {
				nums = append(nums, n)
			}
		}
		result, _ := json.Marshal(map[string]any{"numbers": nums})
		raw, _ := json.Marshal(r)
		out = append(out, model.DrawResult{
			LotteryCode: consts.LotteryKL8,
			Issue:       r.Code,
			DrawDate:    drawDate,
			Result:      result,
			RawData:     raw,
		})
	}
	return out, nil
}

// HasDrawToday 判断彩种当天是否开奖。
// 大乐透：周一/三/六；排列三、快乐8：每天。
func HasDrawToday(lotteryCode string, now time.Time) bool {
	wd := int(now.Weekday()) // 0=Sunday
	switch lotteryCode {
	case consts.LotteryDLT:
		return wd == 1 || wd == 3 || wd == 6
	default:
		return true
	}
}

// NextIssueHint 基于库内最新开奖期号生成下一期预测期号。
func NextIssueHint(latest *model.DrawResult, now time.Time) string {
	if latest == nil {
		return now.Format("20060102")
	}
	if n, err := strconv.Atoi(strings.TrimSpace(latest.Issue)); err == nil {
		return strconv.Itoa(n + 1)
	}
	return now.Format("20060102") + "-pred"
}

func issueNewer(a, b string) bool {
	na, ea := strconv.Atoi(strings.TrimSpace(a))
	nb, eb := strconv.Atoi(strings.TrimSpace(b))
	if ea == nil && eb == nil {
		return na > nb
	}
	return strings.TrimSpace(a) > strings.TrimSpace(b)
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
