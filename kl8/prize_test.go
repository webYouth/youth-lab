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
