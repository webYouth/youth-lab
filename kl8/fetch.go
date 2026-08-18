// Fetch 快乐8 draw results from China Welfare Lottery public API.
package main

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"
)

const defaultDrawAPI = "https://www.cwl.gov.cn/cwl_admin/front/cwlkj/search/kjxx/findDrawNotice"

type DrawResult struct {
	Code       string
	Date       string
	Numbers    []int
	APIPrizes  map[string]float64 // type -> typemoney
	RawNumbers string
}

type cwlResponse struct {
	State   int         `json:"state"`
	Message string      `json:"message"`
	Result  []cwlRecord `json:"result"`
}

type cwlRecord struct {
	Name        string     `json:"name"`
	Code        string     `json:"code"`
	Date        string     `json:"date"`
	Red         string     `json:"red"`
	Prizegrades []cwlPrize `json:"prizegrades"`
}

type cwlPrize struct {
	Type      string `json:"type"`
	TypeNum   string `json:"typenum"`
	TypeMoney string `json:"typemoney"`
}

func fetchLatestDraw(apiURL string) (*DrawResult, error) {
	if apiURL == "" {
		apiURL = defaultDrawAPI
	}

	q := url.Values{}
	q.Set("name", "kl8")
	q.Set("issueCount", "1")
	q.Set("pageNo", "1")
	q.Set("pageSize", "1")
	q.Set("systemType", "PC")

	req, err := http.NewRequest(http.MethodGet, apiURL+"?"+q.Encode(), nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("User-Agent", "Mozilla/5.0 (compatible; youth-lab-kl8/1.0)")
	req.Header.Set("Referer", "https://www.cwl.gov.cn/")
	req.Header.Set("Accept", "application/json,text/plain,*/*")

	client := &http.Client{Timeout: 20 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status %d: %s", resp.StatusCode, string(body))
	}

	var parsed cwlResponse
	if err := json.Unmarshal(body, &parsed); err != nil {
		return nil, err
	}
	if parsed.State != 0 || len(parsed.Result) == 0 {
		return nil, fmt.Errorf("draw not ready: state=%d message=%s", parsed.State, parsed.Message)
	}

	rec := parsed.Result[0]
	nums, err := parseNumbers(rec.Red)
	if err != nil {
		return nil, err
	}
	if len(nums) != 20 {
		return nil, fmt.Errorf("expected 20 draw numbers, got %d", len(nums))
	}

	prizes := map[string]float64{}
	for _, p := range rec.Prizegrades {
		if p.Type == "" {
			continue
		}
		v, err := strconv.ParseFloat(p.TypeMoney, 64)
		if err != nil {
			continue
		}
		prizes[p.Type] = v
	}

	return &DrawResult{
		Code:       rec.Code,
		Date:       rec.Date,
		Numbers:    nums,
		APIPrizes:  prizes,
		RawNumbers: rec.Red,
	}, nil
}

func parseNumbers(raw string) ([]int, error) {
	parts := strings.Split(raw, ",")
	out := make([]int, 0, len(parts))
	for _, p := range parts {
		p = strings.TrimSpace(p)
		if p == "" {
			continue
		}
		n, err := strconv.Atoi(p)
		if err != nil {
			return nil, fmt.Errorf("invalid number %q", p)
		}
		out = append(out, n)
	}
	sortInts(out)
	return out, nil
}

func sortInts(nums []int) {
	for i := 0; i < len(nums); i++ {
		for j := i + 1; j < len(nums); j++ {
			if nums[j] < nums[i] {
				nums[i], nums[j] = nums[j], nums[i]
			}
		}
	}
}

func isTodayDraw(draw *DrawResult, now time.Time) bool {
	// date field looks like: 2026-08-10(一)
	if draw == nil || draw.Date == "" {
		return false
	}
	datePart := draw.Date
	if idx := strings.Index(datePart, "("); idx > 0 {
		datePart = datePart[:idx]
	}
	day := now.In(shanghaiLoc()).Format("2006-01-02")
	return datePart == day
}

func expectedIssueCode(now time.Time) string {
	// Issue code format: YYYY + day-of-year, e.g. 2026213
	t := now.In(shanghaiLoc())
	return fmt.Sprintf("%d%03d", t.Year(), t.YearDay())
}
