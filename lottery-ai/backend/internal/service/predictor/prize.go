package predictor

import (
	"fmt"
	"math"
	"strings"

	"youthlab/lottery-ai/internal/consts"
)

// StakeYuan 评估按单注 2 元计投入，不用当期浮动奖池。
const StakeYuan = 2.0

// ScoreVersion 奖级/奖金加权计分版本；低于此值的历史评估会重算。
const ScoreVersion = 2

const (
	weightDLTFirst  = 5_000_000.0
	weightDLTSecond = 200_000.0
)

type prizeOutcome struct {
	Level    string
	Prize    float64
	Win      bool
	Floating bool
}

func dltPrize(frontHit, backHit int) prizeOutcome {
	switch {
	case frontHit == 5 && backHit == 2:
		return prizeOutcome{"一等奖 5+2", 0, true, true}
	case frontHit == 5 && backHit == 1:
		return prizeOutcome{"二等奖 5+1", 0, true, true}
	case frontHit == 5 && backHit == 0:
		return prizeOutcome{"三等奖 5+0", 10000, true, false}
	case frontHit == 4 && backHit == 2:
		return prizeOutcome{"四等奖 4+2", 3000, true, false}
	case frontHit == 4 && backHit == 1:
		return prizeOutcome{"五等奖 4+1", 300, true, false}
	case frontHit == 3 && backHit == 2:
		return prizeOutcome{"六等奖 3+2", 200, true, false}
	case frontHit == 4 && backHit == 0:
		return prizeOutcome{"七等奖 4+0", 100, true, false}
	case (frontHit == 3 && backHit == 1) || (frontHit == 2 && backHit == 2):
		return prizeOutcome{fmt.Sprintf("八等奖 %d+%d", frontHit, backHit), 15, true, false}
	case (frontHit == 3 && backHit == 0) || (frontHit == 2 && backHit == 1) || (frontHit == 1 && backHit == 2) || (frontHit == 0 && backHit == 2):
		return prizeOutcome{fmt.Sprintf("九等奖 %d+%d", frontHit, backHit), 5, true, false}
	default:
		return prizeOutcome{fmt.Sprintf("%d+%d", frontHit, backHit), 0, false, false}
	}
}

func p3Prize(exactHits int) prizeOutcome {
	if exactHits == 3 {
		return prizeOutcome{"直选", 1040, true, false}
	}
	return prizeOutcome{fmt.Sprintf("直选命中%d位", exactHits), 0, false, false}
}

func kl8Pick10Prize(hits int) prizeOutcome {
	switch hits {
	case 10:
		return prizeOutcome{"选十中10", 0, true, true}
	case 9:
		return prizeOutcome{"选十中9", 8000, true, false}
	case 8:
		return prizeOutcome{"选十中8", 720, true, false}
	case 7:
		return prizeOutcome{"选十中7", 80, true, false}
	case 6:
		return prizeOutcome{"选十中6", 5, true, false}
	case 5:
		return prizeOutcome{"选十中5", 3, true, false}
	case 0:
		return prizeOutcome{"选十中0", 2, true, false}
	default:
		return prizeOutcome{fmt.Sprintf("选十中%d", hits), 0, false, false}
	}
}

func maxWeightYuan(lotteryCode string) float64 {
	if lotteryCode == consts.LotteryP3 {
		return 1040
	}
	return weightDLTFirst
}

// weightYuan 计分用单注奖金。浮动奖用封顶口径，不取当期奖池。
func (o prizeOutcome) weightYuan() float64 {
	if !o.Win {
		return 0
	}
	if !o.Floating {
		return o.Prize
	}
	if strings.Contains(o.Level, "二等奖") {
		return weightDLTSecond
	}
	return weightDLTFirst
}

func weightScore(weightYuan, maxWeight float64) float64 {
	if weightYuan <= 0 || maxWeight <= 0 {
		return 0
	}
	s := math.Log10(1+weightYuan) / math.Log10(1+maxWeight)
	if s > 1 {
		s = 1
	}
	return math.Round(s*10000) / 10000
}
