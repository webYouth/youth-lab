// Internal Happy8 (快乐8) checker daemon.
// Not exposed publicly. Runs daily after 22:00 Asia/Shanghai,
// retries every 5 minutes until today's draw is available, then emails results.
package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"time"
)

func shanghaiLoc() *time.Location {
	loc, err := time.LoadLocation("Asia/Shanghai")
	if err != nil {
		return time.FixedZone("CST", 8*3600)
	}
	return loc
}

func main() {
	once := flag.Bool("once", false, "run one check immediately and exit")
	dryRun := flag.Bool("dry-run", false, "print report without sending email")
	ticketsPath := flag.String("tickets", getenv("KL8_TICKETS_FILE", "/data/tickets.yaml"), "path to tickets yaml")
	apiURL := flag.String("api", getenv("KL8_API_URL", defaultDrawAPI), "draw API url")
	startHour := flag.Int("start-hour", getenvInt("KL8_START_HOUR", 22), "daily start hour in Asia/Shanghai")
	retryMinutes := flag.Int("retry-minutes", getenvInt("KL8_RETRY_MINUTES", 5), "retry interval minutes")
	flag.Parse()

	log.Printf("kl8 checker starting; tickets=%s once=%v dryRun=%v startHour=%d retry=%dm", *ticketsPath, *once, *dryRun, *startHour, *retryMinutes)

	if *once {
		if err := runOnce(*ticketsPath, *apiURL, true, *dryRun); err != nil {
			log.Fatalf("check failed: %v", err)
		}
		return
	}

	var lastSuccessDay string
	for {
		waitUntilRunnable(*startHour, lastSuccessDay)
		if err := runWithRetry(*ticketsPath, *apiURL, *retryMinutes, *dryRun); err != nil {
			log.Printf("daily check ended with error: %v", err)
		} else {
			lastSuccessDay = time.Now().In(shanghaiLoc()).Format("2006-01-02")
		}
		// Avoid immediate re-entry in the same minute.
		time.Sleep(time.Minute)
	}
}

func waitUntilRunnable(hour int, lastSuccessDay string) {
	for {
		now := time.Now().In(shanghaiLoc())
		today := now.Format("2006-01-02")
		if now.Hour() >= hour && lastSuccessDay != today {
			log.Printf("reached daily window, start checking for %s", today)
			return
		}
		next := time.Date(now.Year(), now.Month(), now.Day(), hour, 0, 0, 0, shanghaiLoc())
		if !now.Before(next) {
			next = next.Add(24 * time.Hour)
		}
		log.Printf("sleeping until %s", next.Format(time.RFC3339))
		time.Sleep(time.Until(next))
	}
}

func runWithRetry(ticketsPath, apiURL string, retryMinutes int, dryRun bool) error {
	for {
		err := runOnce(ticketsPath, apiURL, false, dryRun)
		if err == nil {
			return nil
		}
		log.Printf("draw not ready or check failed: %v; retry in %d minutes", err, retryMinutes)
		time.Sleep(time.Duration(retryMinutes) * time.Minute)
	}
}

func runOnce(ticketsPath, apiURL string, allowAnyDay, dryRun bool) error {
	tickets, err := loadTickets(ticketsPath)
	if err != nil {
		return fmt.Errorf("load tickets: %w", err)
	}

	draw, err := fetchLatestDraw(apiURL)
	if err != nil {
		return fmt.Errorf("fetch draw: %w", err)
	}

	now := time.Now().In(shanghaiLoc())
	if !allowAnyDay && !isTodayDraw(draw, now) {
		return fmt.Errorf("latest draw is %s (%s), waiting for today %s", draw.Code, draw.Date, now.Format("2006-01-02"))
	}

	summary := checkAll(tickets.Tickets, draw)
	fp := ticketFingerprint(tickets.Tickets)
	day := ChaseDay{
		Date:        draw.Date,
		Period:      draw.Code,
		Fingerprint: fp,
		Checked:     summary.CheckedBets,
		Winning:     summary.WinningBets,
		Stake:       summary.TotalStake,
		Prize:       summary.TotalPrize,
		Profit:      summary.TotalProfit,
		Floating:    summary.FloatingBets,
	}
	led, err := upsertChaseDay(ledgerPath(), day)
	if err != nil {
		log.Printf("chase ledger: %v", err)
		led = &ChaseLedger{Days: []ChaseDay{day}}
	}
	chase := chaseTotals(led, fp)

	body := formatSummary(summary)
	body += formatChaseFooter(chase)
	log.Print(body)

	subject := fmt.Sprintf("快乐8查奖 %s 本期%s 累计%s", draw.Code, formatProfit(summary.TotalProfit), formatProfit(chase.Profit))
	if !dryRun {
		notifyApp(subject, kl8NotifyBody(summary, chase), map[string]any{
			"period":       draw.Code,
			"date":         draw.Date,
			"numbers":      draw.RawNumbers,
			"winning":      summary.WinningBets,
			"checked":      summary.CheckedBets,
			"stake":        summary.TotalStake,
			"prize":        summary.TotalPrize,
			"profit":       summary.TotalProfit,
			"floating":     summary.FloatingBets,
			"chase_days":   chase.Days,
			"chase_stake":  chase.Stake,
			"chase_prize":  chase.Prize,
			"chase_profit": chase.Profit,
			"report":       body,
		})
	}

	if dryRun {
		log.Printf("dry-run enabled, skip sending mail")
		return nil
	}

	mailCfg, err := loadMailConfigFromEnv()
	if err != nil {
		return fmt.Errorf("mail config: %w", err)
	}
	if err := sendMail(mailCfg, subject, body); err != nil {
		return fmt.Errorf("send mail: %w", err)
	}
	log.Printf("mail sent to %s", mailCfg.To)
	return nil
}

func getenvInt(key string, fallback int) int {
	v := os.Getenv(key)
	if v == "" {
		return fallback
	}
	var n int
	_, err := fmt.Sscanf(v, "%d", &n)
	if err != nil {
		return fallback
	}
	return n
}
