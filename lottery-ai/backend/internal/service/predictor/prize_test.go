package predictor

import (
	"testing"

	"youthlab/lottery-ai/internal/consts"
)

func TestDLTPrize(t *testing.T) {
	cases := []struct {
		f, b     int
		prize    float64
		win      bool
		floating bool
	}{
		{5, 2, 0, true, true},
		{5, 0, 10000, true, false},
		{4, 2, 3000, true, false},
		{3, 1, 15, true, false},
		{0, 2, 5, true, false},
		{2, 0, 0, false, false},
	}
	for _, c := range cases {
		got := dltPrize(c.f, c.b)
		if got.Prize != c.prize || got.Win != c.win || got.Floating != c.floating {
			t.Fatalf("dlt %d+%d: got prize=%v win=%v float=%v", c.f, c.b, got.Prize, got.Win, got.Floating)
		}
	}
}

func TestP3AndKL8Prize(t *testing.T) {
	if p := p3Prize(3); !p.Win || p.Prize != 1040 {
		t.Fatalf("p3: %+v", p)
	}
	if p := p3Prize(2); p.Win || p.Prize != 0 {
		t.Fatalf("p3 miss: %+v", p)
	}
	if p := kl8Pick10Prize(9); !p.Win || p.Prize != 8000 {
		t.Fatalf("kl8 9: %+v", p)
	}
	if p := kl8Pick10Prize(0); !p.Win || p.Prize != 2 {
		t.Fatalf("kl8 0: %+v", p)
	}
	if p := kl8Pick10Prize(4); p.Win {
		t.Fatalf("kl8 4 should miss: %+v", p)
	}
}

func TestWeightScoreRanksPrizeLevels(t *testing.T) {
	maxW := maxWeightYuan(consts.LotteryDLT)
	ninth := weightScore(dltPrize(0, 2).weightYuan(), maxW)
	eighth := weightScore(dltPrize(3, 1).weightYuan(), maxW)
	third := weightScore(dltPrize(5, 0).weightYuan(), maxW)
	first := weightScore(dltPrize(5, 2).weightYuan(), maxW)
	miss := weightScore(dltPrize(2, 0).weightYuan(), maxW)
	if !(miss == 0 && ninth < eighth && eighth < third && third < first && first == 1) {
		t.Fatalf("rank miss=%.4f 9=%.4f 8=%.4f 3=%.4f 1=%.4f", miss, ninth, eighth, third, first)
	}
	if dltPrize(5, 2).weightYuan() != 5_000_000 || dltPrize(5, 1).weightYuan() != 200_000 {
		t.Fatalf("floating weights")
	}
}
