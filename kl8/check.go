// Check tickets against draw result and compute payouts.
package main

import (
	"fmt"
	"strings"
)

type BetResult struct {
	TicketID   string
	Play       string
	Mode       string
	Multiplier int
	BetIndex   int
	Numbers    []int
	Hits       int
	HitNums    []int
	Level      string
	UnitPrize  float64
	TotalPrize float64
	Won        bool
}

type CheckSummary struct {
	Draw         *DrawResult
	Results      []BetResult
	TotalPrize   float64
	WinningBets  int
	CheckedBets  int
	TicketCount  int
}

func checkAll(tickets []Ticket, draw *DrawResult) *CheckSummary {
	summary := &CheckSummary{
		Draw:        draw,
		TicketCount: len(tickets),
	}
	drawSet := map[int]struct{}{}
	for _, n := range draw.Numbers {
		drawSet[n] = struct{}{}
	}

	for _, ticket := range tickets {
		for i, bet := range ticket.Bets {
			hits, hitNums := countHits(bet, drawSet)
			amount, level, ok := unitPrize(ticket.Play, hits, draw.APIPrizes)
			res := BetResult{
				TicketID:   ticket.ID,
				Play:       ticket.Play,
				Mode:       ticket.Mode,
				Multiplier: ticket.Multiplier,
				BetIndex:   i + 1,
				Numbers:    append([]int{}, bet...),
				Hits:       hits,
				HitNums:    hitNums,
				Level:      level,
				UnitPrize:  amount,
				TotalPrize: roundMoney(amount * float64(ticket.Multiplier)),
				Won:        ok && amount > 0,
			}
			if ok && strings.Contains(level, "浮动") && amount == 0 {
				// Floating award without published amount yet: still mark as won.
				res.Won = true
			}
			summary.Results = append(summary.Results, res)
			summary.CheckedBets++
			if res.Won {
				summary.WinningBets++
				summary.TotalPrize += res.TotalPrize
			}
		}
	}
	summary.TotalPrize = roundMoney(summary.TotalPrize)
	return summary
}

func countHits(bet []int, drawSet map[int]struct{}) (int, []int) {
	hitNums := make([]int, 0)
	for _, n := range bet {
		if _, ok := drawSet[n]; ok {
			hitNums = append(hitNums, n)
		}
	}
	return len(hitNums), hitNums
}

func formatSummary(summary *CheckSummary) string {
	var b strings.Builder
	draw := summary.Draw
	fmt.Fprintf(&b, "快乐8 查奖报告\n")
	fmt.Fprintf(&b, "期号: %s\n", draw.Code)
	fmt.Fprintf(&b, "开奖日期: %s\n", draw.Date)
	fmt.Fprintf(&b, "开奖号码: %s\n", draw.RawNumbers)
	fmt.Fprintf(&b, "票数: %d | 注数: %d | 中奖注数: %d\n", summary.TicketCount, summary.CheckedBets, summary.WinningBets)
	fmt.Fprintf(&b, "预计奖金合计: %.2f 元\n\n", summary.TotalPrize)

	currentTicket := ""
	for _, r := range summary.Results {
		if r.TicketID != currentTicket {
			currentTicket = r.TicketID
			fmt.Fprintf(&b, "==== 票 %s（%s/%s，%d倍）====\n", r.TicketID, r.Play, r.Mode, r.Multiplier)
		}
		status := "未中"
		if r.Won {
			status = "中奖"
		}
		fmt.Fprintf(
			&b,
			"[%s] 注#%d 命中%d个(%s) 奖级=%s 单注=%.2f 实得=%.2f 号码=%s\n",
			status,
			r.BetIndex,
			r.Hits,
			joinInts(r.HitNums),
			r.Level,
			r.UnitPrize,
			r.TotalPrize,
			joinInts(r.Numbers),
		)
	}

	fmt.Fprintf(&b, "\n说明:\n")
	fmt.Fprintf(&b, "1) 奖金优先使用当期官方 prizegrades；缺失时回退固定奖表。\n")
	fmt.Fprintf(&b, "2) 选十中十为浮动奖，封顶 500 万；未公布金额时会标注待官方公布。\n")
	fmt.Fprintf(&b, "3) 复式按展开后的每一注单式分别计奖。\n")
	return b.String()
}

func joinInts(nums []int) string {
	if len(nums) == 0 {
		return "-"
	}
	parts := make([]string, 0, len(nums))
	for _, n := range nums {
		parts = append(parts, fmt.Sprintf("%02d", n))
	}
	return strings.Join(parts, ",")
}
