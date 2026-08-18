// Package main - prize tables and payout helpers for 快乐8.
package main

import (
	"fmt"
	"math"
)

// Fixed prize table based on current China Welfare Lottery 快乐8 structure
// (aligned with cwl.gov.cn prizegrades fixed awards).
// Floating awards:
//   - 选十中十: floating, cap 5,000,000
//   - 选九中九: treated as fixed 300,000 in game rules text; if API returns a
//     positive typemoney for the issue, that value is preferred.
var fixedPrizeTable = map[string]map[int]float64{
	"选十": {
		10: -1, // floating
		9:  8000,
		8:  720,
		7:  80,
		6:  5,
		5:  3,
		0:  2,
	},
	"选九": {
		9: 300000,
		8: 2000,
		7: 225,
		6: 22,
		5: 5,
		4: 3,
		0: 2,
	},
	"选八": {
		8: 50000,
		7: 800,
		6: 80,
		5: 10,
		4: 3,
		0: 2,
	},
	"选七": {
		7: 8500,
		6: 300,
		5: 30,
		4: 4,
		0: 2,
	},
	"选六": {
		6: 2880,
		5: 30,
		4: 10,
		3: 3,
	},
	"选五": {
		5: 1000,
		4: 20,
		3: 3,
	},
	"选四": {
		4: 93,
		3: 5,
		2: 3,
	},
	"选三": {
		3: 52,
		2: 3,
	},
	"选二": {
		2: 19,
	},
	"选一": {
		1: 4.5,
	},
}

func prizeKey(play string, hit int) string {
	playCode := map[string]string{
		"选十": "x10",
		"选九": "x9",
		"选八": "x8",
		"选七": "x7",
		"选六": "x6",
		"选五": "x5",
		"选四": "x4",
		"选三": "x3",
		"选二": "x2",
		"选一": "x1",
	}[play]
	if playCode == "" {
		return ""
	}
	return fmt.Sprintf("%sz%d", playCode, hit)
}

func unitPrize(play string, hit int, apiPrizes map[string]float64) (amount float64, level string, ok bool) {
	table, exists := fixedPrizeTable[play]
	if !exists {
		return 0, "", false
	}
	base, exists := table[hit]
	if !exists {
		return 0, "", false
	}

	level = fmt.Sprintf("%s中%d", play, hit)
	key := prizeKey(play, hit)

	if base < 0 {
		// Floating award: prefer official issue amount from API.
		if apiPrizes != nil {
			if v, found := apiPrizes[key]; found && v > 0 {
				return v, level + "(浮动)", true
			}
		}
		return 0, level + "(浮动-待官方公布)", true
	}

	if apiPrizes != nil {
		if v, found := apiPrizes[key]; found && v > 0 {
			return v, level, true
		}
	}
	return base, level, true
}

func roundMoney(v float64) float64 {
	return math.Round(v*100) / 100
}
