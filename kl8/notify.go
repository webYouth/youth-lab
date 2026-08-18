package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"
	"time"
)

func notifyApp(title, body string, payload any) {
	url := strings.TrimSpace(os.Getenv("LOTTERY_NOTIFY_URL"))
	token := strings.TrimSpace(os.Getenv("LOTTERY_ADMIN_TOKEN"))
	if url == "" || token == "" {
		log.Printf("skip app notify: LOTTERY_NOTIFY_URL/LOTTERY_ADMIN_TOKEN not set")
		return
	}
	raw, _ := json.Marshal(map[string]any{
		"type":    "kl8",
		"title":   title,
		"body":    body,
		"payload": payload,
	})
	req, err := http.NewRequest(http.MethodPost, url, bytes.NewReader(raw))
	if err != nil {
		log.Printf("app notify build request: %v", err)
		return
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Admin-Token", token)
	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		log.Printf("app notify failed: %v", err)
		return
	}
	defer resp.Body.Close()
	if resp.StatusCode >= 300 {
		log.Printf("app notify status=%d", resp.StatusCode)
		return
	}
	log.Printf("app notify sent")
}

func kl8NotifyBody(summary *CheckSummary, chase ChaseTotals) string {
	line := fmt.Sprintf(
		"期号 %s 投入 %s，奖金 %s，本期盈亏 %s。中奖 %d/%d 注。",
		summary.Draw.Code,
		formatYuanPlain(summary.TotalStake),
		formatYuanPlain(summary.TotalPrize),
		formatProfit(summary.TotalProfit),
		summary.WinningBets,
		summary.CheckedBets,
	)
	if summary.FloatingBets > 0 {
		line += fmt.Sprintf(" 浮动奖 %d 注未计入金额。", summary.FloatingBets)
	}
	if chase.Days > 0 {
		line += fmt.Sprintf(" 追号 %d 期累计投入 %s，奖金 %s，累计盈亏 %s。", chase.Days, formatYuanPlain(chase.Stake), formatYuanPlain(chase.Prize), formatProfit(chase.Profit))
	}
	return line
}

func formatYuanPlain(v float64) string {
	return fmt.Sprintf("%.2f元", v)
}
