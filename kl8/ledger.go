package main

import (
	"crypto/sha1"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"os"
	"sort"
	"strings"
	"sync"
)

const stakeYuan = 2.0

type ChaseDay struct {
	Date        string  `json:"date"`
	Period      string  `json:"period"`
	Fingerprint string  `json:"fingerprint"`
	Checked     int     `json:"checked"`
	Winning     int     `json:"winning"`
	Stake       float64 `json:"stake"`
	Prize       float64 `json:"prize"`
	Profit      float64 `json:"profit"`
	Floating    int     `json:"floating"`
}

type ChaseLedger struct {
	Days []ChaseDay `json:"days"`
}

type ChaseTotals struct {
	Days   int
	Stake  float64
	Prize  float64
	Profit float64
}

var ledgerMu sync.Mutex

func ticketFingerprint(tickets []Ticket) string {
	parts := make([]string, 0, len(tickets))
	for _, t := range tickets {
		raw, _ := json.Marshal(struct {
			ID         string  `json:"id"`
			Play       string  `json:"play"`
			Mode       string  `json:"mode"`
			Multiplier int     `json:"multiplier"`
			Numbers    []int   `json:"numbers"`
			Bets       [][]int `json:"bets"`
		}{t.ID, t.Play, t.Mode, t.Multiplier, t.Numbers, t.Bets})
		parts = append(parts, string(raw))
	}
	sort.Strings(parts)
	sum := sha1.Sum([]byte(strings.Join(parts, "\n")))
	return hex.EncodeToString(sum[:8])
}

func ledgerPath() string {
	if v := strings.TrimSpace(os.Getenv("KL8_LEDGER_FILE")); v != "" {
		return v
	}
	return "/data/ledger.json"
}

func loadLedger(path string) (*ChaseLedger, error) {
	raw, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return &ChaseLedger{}, nil
		}
		return nil, err
	}
	if len(strings.TrimSpace(string(raw))) == 0 {
		return &ChaseLedger{}, nil
	}
	var led ChaseLedger
	if err := json.Unmarshal(raw, &led); err != nil {
		return nil, err
	}
	return &led, nil
}

func saveLedger(path string, led *ChaseLedger) error {
	raw, err := json.MarshalIndent(led, "", "  ")
	if err != nil {
		return err
	}
	tmp := path + ".tmp"
	if err := os.WriteFile(tmp, raw, 0o644); err != nil {
		return err
	}
	return os.Rename(tmp, path)
}

func upsertChaseDay(path string, day ChaseDay) (*ChaseLedger, error) {
	ledgerMu.Lock()
	defer ledgerMu.Unlock()
	led, err := loadLedger(path)
	if err != nil {
		return nil, err
	}
	replaced := false
	for i, d := range led.Days {
		if d.Period == day.Period {
			led.Days[i] = day
			replaced = true
			break
		}
	}
	if !replaced {
		led.Days = append(led.Days, day)
	}
	if err := saveLedger(path, led); err != nil {
		return nil, err
	}
	return led, nil
}

func chaseTotals(led *ChaseLedger, fingerprint string) ChaseTotals {
	var tot ChaseTotals
	if led == nil {
		return tot
	}
	for _, d := range led.Days {
		if fingerprint != "" && d.Fingerprint != fingerprint {
			continue
		}
		tot.Days++
		tot.Stake += d.Stake
		tot.Prize += d.Prize
		tot.Profit += d.Profit
	}
	tot.Stake = roundMoney(tot.Stake)
	tot.Prize = roundMoney(tot.Prize)
	tot.Profit = roundMoney(tot.Profit)
	return tot
}

func formatProfit(v float64) string {
	if v > 0 {
		return fmt.Sprintf("+%.2f元", v)
	}
	return fmt.Sprintf("%.2f元", v)
}

func formatChaseFooter(chase ChaseTotals) string {
	if chase.Days == 0 {
		return ""
	}
	return fmt.Sprintf(
		"\n追号累计（当前固定号码 %d 期）\n投入: %.2f 元 | 奖金: %.2f 元 | 盈亏: %s\n",
		chase.Days, chase.Stake, chase.Prize, formatProfit(chase.Profit),
	)
}
