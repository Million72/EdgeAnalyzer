package predictor

// TPSLResult holds calculated trade levels
type TPSLResult struct {
	TP1  float64
	TP2  float64
	SL   float64
	Pips float64
}

// CalculateTPSL computes TP1/TP2/SL from ATR, tuned per instrument type.
// scalp=true uses a much tighter multiplier set intended for the 1m
// timeframe: the normal 1.5-3.5x ATR levels are swing-sized and could take
// many candles to reach, which doesn't match a scalper's much shorter
// intended holding time. R:R is kept similar (~1.3-1.5) to the normal
// profile — only the absolute distance shrinks, not the risk shape.
func CalculateTPSL(side string, price, atr float64, isSynthetic, isJPY, isGold bool, scalp bool) TPSLResult {
	slMult, tp1Mult, tp2Mult := 1.5, 2.0, 3.5
	switch {
	case scalp && isSynthetic:
		slMult, tp1Mult, tp2Mult = 0.4, 0.6, 1.0
	case scalp:
		slMult, tp1Mult, tp2Mult = 0.5, 0.7, 1.2
	case isSynthetic:
		slMult, tp1Mult, tp2Mult = 1.2, 1.8, 3.0
	}

	pipMult := 10000.0
	if isJPY || isGold {
		pipMult = 100
	} else if isSynthetic {
		pipMult = 1
	}

	if side == "bull" {
		return TPSLResult{
			SL:   price - atr*slMult,
			TP1:  price + atr*tp1Mult,
			TP2:  price + atr*tp2Mult,
			Pips: atr * pipMult,
		}
	}
	return TPSLResult{
		SL:   price + atr*slMult,
		TP1:  price - atr*tp1Mult,
		TP2:  price - atr*tp2Mult,
		Pips: atr * pipMult,
	}
}

// RiskReward returns the reward:risk ratio for a proposed trade.
func RiskReward(entry, tp1, sl float64) float64 {
	reward := tp1 - entry
	if reward < 0 {
		reward = -reward
	}
	risk := sl - entry
	if risk < 0 {
		risk = -risk
	}
	if risk == 0 {
		return 0
	}
	return reward / risk
}
