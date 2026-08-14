// 微调数据导出工具。
// 用法：go run ./cmd/export-finetune --lottery=DLT --format=jsonl --output=./finetune/dlt.jsonl --min-hit-rate=0.2
package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"os"
	"time"

	"youthlab/lottery-ai/internal/config"
	"youthlab/lottery-ai/internal/store"
)

func main() {
	lotteryCode := flag.String("lottery", "DLT", "lottery code")
	output := flag.String("output", "./finetune/out.jsonl", "output jsonl path")
	minHit := flag.Float64("min-hit-rate", 0, "minimum hit rate filter")
	flag.Parse()

	cfg := config.Load()
	ctx := context.Background()
	st, err := store.New(ctx, cfg.DatabaseURL)
	if err != nil {
		log.Fatal(err)
	}
	defer st.Close()

	rows, err := st.Pool().Query(ctx, `
		SELECT p.predicted_numbers, p.reason, pr.hit_rate, d.result, d.issue
		FROM predictions p
		JOIN prediction_results pr ON pr.prediction_id=p.id
		JOIN draw_results d ON d.id=pr.draw_result_id
		WHERE p.lottery_code=$1 AND p.success=TRUE AND pr.hit_rate >= $2
		ORDER BY p.id DESC LIMIT 5000`, *lotteryCode, *minHit)
	if err != nil {
		log.Fatal(err)
	}
	defer rows.Close()

	if err := os.MkdirAll(filepathDir(*output), 0o755); err != nil {
		log.Fatal(err)
	}
	f, err := os.Create(*output)
	if err != nil {
		log.Fatal(err)
	}
	defer f.Close()

	count := 0
	for rows.Next() {
		var predJSON, reason string
		var hitRate float64
		var resultJSON []byte
		var issue string
		if err := rows.Scan(&predJSON, &reason, &hitRate, &resultJSON, &issue); err != nil {
			log.Fatal(err)
		}
		sample := map[string]any{
			"messages": []map[string]string{
				{"role": "system", "content": "你是彩票数据分析助手。"},
				{"role": "user", "content": fmt.Sprintf("请基于历史特征预测 %s 期号 %s。实际开奖参考仅用于训练：%s", *lotteryCode, issue, string(resultJSON))},
				{"role": "assistant", "content": predJSON},
			},
			"meta": map[string]any{"hit_rate": hitRate, "exported_at": time.Now().Format(time.RFC3339)},
		}
		b, _ := json.Marshal(sample)
		_, _ = f.Write(append(b, '\n'))
		count++
	}
	log.Printf("exported %d samples to %s", count, *output)
}

func filepathDir(p string) string {
	for i := len(p) - 1; i >= 0; i-- {
		if p[i] == '/' {
			return p[:i]
		}
	}
	return "."
}
