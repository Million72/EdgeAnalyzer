package predictor

import (
	"fmt"

	"otc-predictor/internal/strategy"
)

type ValidationResult struct {
	Valid       bool
	Side        string // "bull" | "bear"
	Reason      string
	CounterTrend bool
}

const (
	MinScore          = 10.0
	MinMargin         = 4.0
	RSIOverbought     = 70.0
	RSIOversold       = 30.0
	// 62%, not 70%. MaxScore grew 34->39 as Retest/Pullback/more chart
	// patterns were added — all bonus-only factors that aren't present on
	// most scans. Keeping a 70% floor against the bigger denominator would
	// have silently raised the real bar (70% of 39 ≈ 27.3 raw points needed,
	// vs 70% of 34 ≈ 23.8 before). 62% of 39 ≈ 24.2 — the same real
	// requirement as before, not a new stricter one introduced by accident.
	MinConfidencePct = 62.0
)

// ValidateSignal applies score thresholds, RSI extremes, 3-timeframe agreement,
// and finally a hard confidence floor. The confidence floor exists because
// score/margin dominance alone does NOT guarantee a genuinely strong signal —
// a result can clear MinScore/MinMargin while still translating to a weak
// percentage (e.g. 10/34 ≈ 29%). Without this floor, low-conviction signals
// were passing validation and firing as real BUY/SELL calls.
func ValidateSignal(result strategy.EngineResult) ValidationResult {
	bullDominant := result.BullScore >= MinScore && result.BullScore > result.BearScore+MinMargin
	bearDominant := result.BearScore >= MinScore && result.BearScore > result.BullScore+MinMargin

	if !bullDominant && !bearDominant {
		return ValidationResult{Valid: false, Reason: "Insufficient confluence"}
	}

	side := "bear"
	if bullDominant {
		side = "bull"
	}

	// RSI extreme block
	if side == "bull" && result.RSI > RSIOverbought {
		return ValidationResult{Valid: false, Side: side, Reason: "RSI overbought — BUY blocked"}
	}
	if side == "bear" && result.RSI < RSIOversold {
		return ValidationResult{Valid: false, Side: side, Reason: "RSI oversold — SELL blocked"}
	}

	// Extension / chasing block — don't BUY after price has already run far
	// above its last swing low, or SELL after it's run far below its last
	// swing high. This does NOT predict a reversal (nothing can) — it blocks
	// entries with structurally worse risk/reward because the move already
	// happened before the signal fired.
	if side == "bull" && result.BullExtended {
		return ValidationResult{Valid: false, Side: side, Reason: "Price already extended far above recent swing low — this would be chasing, not an early entry"}
	}
	if side == "bear" && result.BearExtended {
		return ValidationResult{Valid: false, Side: side, Reason: "Price already extended far below recent swing high — this would be chasing, not an early entry"}
	}

	// 3-Timeframe agreement — HTF1 must confirm, HTF2 must not oppose
	want := "BEAR"
	if side == "bull" {
		want = "BULL"
	}
	htf1OK := result.HTF1Bias == want
	htf2OK := result.HTF2Bias == "NEUTRAL" || result.HTF2Bias == want

	if !htf1OK || !htf2OK {
		return ValidationResult{Valid: false, Side: side, Reason: "MTF disagreement"}
	}

	// ── Entry Model gate — REQUIRED, not a bonus ────────────────────
	// A signal may not fire unless at least one of the 11 named entry-model
	// combos is actually present on the SAME side as the dominant direction.
	// General indicator agreement (EMA/RSI/MACD/structure) can independently
	// clear the score/margin/confidence checks — that used to be enough to
	// fire a signal with zero entry models matched, which defeats the
	// purpose of having named entry models at all. This mirrors the exact
	// fix applied to MT5 Signal Pro 2's JS validator.
	matchingModel := false
	for _, m := range result.EntryModels {
		if m.Side == side {
			matchingModel = true
			break
		}
	}
	if !matchingModel {
		return ValidationResult{
			Valid:  false,
			Side:   side,
			Reason: "No entry model present — general confluence alone is not sufficient to fire a signal",
		}
	}

	// ── Confidence floor ─────────────────────────────────────────────
	score := result.BearScore
	if side == "bull" {
		score = result.BullScore
	}
	confidencePct := (score / result.MaxScore) * 100

	if confidencePct < MinConfidencePct {
		return ValidationResult{
			Valid:  false,
			Side:   side,
			Reason: fmt.Sprintf("Confidence %.0f%% below %.0f%% floor — signal too weak to act on", confidencePct, MinConfidencePct),
		}
	}

	counterTrend := (result.Trend == "BULLISH" && side == "bear") || (result.Trend == "BEARISH" && side == "bull")

	return ValidationResult{Valid: true, Side: side, CounterTrend: counterTrend}
}

// ConfirmationCandle checks the most recent candle actually supports the signal direction.
func ConfirmationCandle(candles []struct{ Open, Close, High, Low float64 }, side string) bool {
	if len(candles) == 0 {
		return false
	}
	last := candles[len(candles)-1]
	bodySize := last.Close - last.Open
	if bodySize < 0 {
		bodySize = -bodySize
	}
	rng := last.High - last.Low
	if rng == 0 {
		rng = 0.000001
	}
	bodyRatio := bodySize / rng

	isBullCandle := last.Close > last.Open
	isBearCandle := last.Close < last.Open

	if side == "bull" && isBullCandle && bodyRatio > 0.4 {
		return true
	}
	if side == "bear" && isBearCandle && bodyRatio > 0.4 {
		return true
	}
	return false
}
