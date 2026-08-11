// Ticket config loading and bet expansion (单式 / 复式).
package main

import (
	"fmt"
	"os"
	"sort"

	"gopkg.in/yaml.v3"
)

type TicketsFile struct {
	Tickets []Ticket `yaml:"tickets"`
}

type Ticket struct {
	ID         string  `yaml:"id"`
	Play       string  `yaml:"play"`       // 选九 / 选十 / ...
	Mode       string  `yaml:"mode"`       // 单式 / 复式
	Multiplier int     `yaml:"multiplier"` // 倍数
	Numbers    []int   `yaml:"numbers"`    // used by 复式
	Bets       [][]int `yaml:"bets"`       // used by 单式 (each line)
}

func loadTickets(path string) (*TicketsFile, error) {
	raw, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	var cfg TicketsFile
	if err := yaml.Unmarshal(raw, &cfg); err != nil {
		return nil, err
	}
	if len(cfg.Tickets) == 0 {
		return nil, fmt.Errorf("no tickets found in %s", path)
	}
	for i := range cfg.Tickets {
		if err := normalizeTicket(&cfg.Tickets[i]); err != nil {
			return nil, fmt.Errorf("ticket[%d] %s: %w", i, cfg.Tickets[i].ID, err)
		}
	}
	return &cfg, nil
}

func normalizeTicket(t *Ticket) error {
	if t.ID == "" {
		return fmt.Errorf("id is required")
	}
	if t.Play == "" {
		return fmt.Errorf("play is required")
	}
	if _, ok := fixedPrizeTable[t.Play]; !ok {
		return fmt.Errorf("unsupported play: %s", t.Play)
	}
	if t.Multiplier <= 0 {
		t.Multiplier = 1
	}
	if t.Mode == "" {
		if len(t.Bets) > 0 {
			t.Mode = "单式"
		} else {
			t.Mode = "复式"
		}
	}

	need := playPickCount(t.Play)
	switch t.Mode {
	case "单式":
		if len(t.Bets) == 0 {
			return fmt.Errorf("单式 requires bets")
		}
		for i, bet := range t.Bets {
			nums := uniqueSorted(bet)
			if len(nums) != need {
				return fmt.Errorf("bet[%d] expects %d numbers, got %d", i, need, len(nums))
			}
			t.Bets[i] = nums
		}
	case "复式":
		nums := uniqueSorted(t.Numbers)
		if len(nums) < need {
			return fmt.Errorf("复式 needs at least %d numbers, got %d", need, len(nums))
		}
		t.Numbers = nums
		t.Bets = combinations(nums, need)
	default:
		return fmt.Errorf("unsupported mode: %s", t.Mode)
	}
	return nil
}

func playPickCount(play string) int {
	switch play {
	case "选十":
		return 10
	case "选九":
		return 9
	case "选八":
		return 8
	case "选七":
		return 7
	case "选六":
		return 6
	case "选五":
		return 5
	case "选四":
		return 4
	case "选三":
		return 3
	case "选二":
		return 2
	case "选一":
		return 1
	default:
		return 0
	}
}

func uniqueSorted(nums []int) []int {
	set := map[int]struct{}{}
	for _, n := range nums {
		if n < 1 || n > 80 {
			continue
		}
		set[n] = struct{}{}
	}
	out := make([]int, 0, len(set))
	for n := range set {
		out = append(out, n)
	}
	sort.Ints(out)
	return out
}

func combinations(nums []int, k int) [][]int {
	var out [][]int
	var dfs func(start int, path []int)
	dfs = func(start int, path []int) {
		if len(path) == k {
			cp := append([]int{}, path...)
			out = append(out, cp)
			return
		}
		for i := start; i < len(nums); i++ {
			dfs(i+1, append(path, nums[i]))
		}
	}
	dfs(0, nil)
	return out
}
