package main

import "testing"

func TestCombinationsPick9From10(t *testing.T) {
	nums := []int{3, 7, 18, 19, 20, 26, 27, 44, 46, 56}
	got := combinations(nums, 9)
	if len(got) != 10 {
		t.Fatalf("expected 10 bets, got %d", len(got))
	}
}

func TestUnitPrizeSelect10(t *testing.T) {
	amount, level, ok := unitPrize("选十", 8, nil)
	if !ok || amount != 720 || level != "选十中8" {
		t.Fatalf("unexpected prize: ok=%v amount=%v level=%s", ok, amount, level)
	}
}

func TestCountHits(t *testing.T) {
	draw := map[int]struct{}{1: {}, 2: {}, 3: {}}
	hits, hitNums := countHits([]int{1, 8, 3}, draw)
	if hits != 2 || len(hitNums) != 2 {
		t.Fatalf("hits=%d hitNums=%v", hits, hitNums)
	}
}

func TestCheckProfitStakeTwoYuan(t *testing.T) {
	draw := &DrawResult{Numbers: []int{1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16, 17, 18, 19, 20}}
	tickets := []Ticket{
		{
			ID:         "t1",
			Play:       "选十",
			Mode:       "单式",
			Multiplier: 1,
			Bets:       [][]int{{1, 2, 3, 4, 21, 22, 23, 24, 25, 26}},
		},
	}
	sum := checkAll(tickets, draw)
	if sum.CheckedBets != 1 || sum.TotalStake != 2 || sum.TotalPrize != 0 || sum.TotalProfit != -2 {
		t.Fatalf("got stake=%v prize=%v profit=%v bets=%d hits=%d", sum.TotalStake, sum.TotalPrize, sum.TotalProfit, sum.CheckedBets, sum.Results[0].Hits)
	}
}

func TestCheckProfitSelect10Hit0(t *testing.T) {
	draw := &DrawResult{Numbers: []int{1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16, 17, 18, 19, 20}}
	tickets := []Ticket{
		{
			ID:         "t1",
			Play:       "选十",
			Mode:       "单式",
			Multiplier: 2,
			Bets:       [][]int{{21, 22, 23, 24, 25, 26, 27, 28, 29, 30}},
		},
	}
	sum := checkAll(tickets, draw)
	if !sum.Results[0].Won || sum.TotalPrize != 4 || sum.TotalStake != 4 || sum.TotalProfit != 0 {
		t.Fatalf("hit0: won=%v prize=%v stake=%v profit=%v", sum.Results[0].Won, sum.TotalPrize, sum.TotalStake, sum.TotalProfit)
	}
}

func TestChaseLedgerUpsertAndFingerprint(t *testing.T) {
	dir := t.TempDir()
	path := dir + "/ledger.json"
	t.Setenv("KL8_LEDGER_FILE", path)
	fp := "abc"
	d1 := ChaseDay{Period: "1", Fingerprint: fp, Stake: 2, Prize: 0, Profit: -2}
	led, err := upsertChaseDay(path, d1)
	if err != nil {
		t.Fatal(err)
	}
	d1.Prize = 5
	d1.Profit = 3
	led, err = upsertChaseDay(path, d1)
	if err != nil {
		t.Fatal(err)
	}
	tot := chaseTotals(led, fp)
	if tot.Days != 1 || tot.Profit != 3 {
		t.Fatalf("expected 1 day profit 3, got %+v", tot)
	}
	other := ChaseDay{Period: "2", Fingerprint: "other", Stake: 2, Prize: 0, Profit: -2}
	led, err = upsertChaseDay(path, other)
	if err != nil {
		t.Fatal(err)
	}
	tot = chaseTotals(led, fp)
	if tot.Days != 1 {
		t.Fatalf("fingerprint filter failed: %+v", tot)
	}
}
