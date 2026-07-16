package types

import "time"

// Candle represents a single OHLC candle
type Candle struct {
	Open  float64 `json:"open"`
	High  float64 `json:"high"`
	Low   float64 `json:"low"`
	Close float64 `json:"close"`
	Time  int64   `json:"time"`
}

// Market defines a tradable instrument
type Market struct {
	Symbol string `yaml:"symbol" json:"symbol"`
	Deriv  string `yaml:"deriv" json:"deriv"`
	IsJPY  bool   `yaml:"is_jpy" json:"isJPY"`
	IsGold bool   `yaml:"is_gold" json:"isGold"`
	Type   string `json:"type"` // "forex" | "synthetic"
}

// TimeframeConfig defines granularities for a timeframe
type TimeframeConfig struct {
	Granularity int `yaml:"granularity" json:"granularity"`
	HTFGran     int `yaml:"htf_gran" json:"htfGran"`
	HTF2Gran    int `yaml:"htf2_gran" json:"htf2Gran"`
}

// Factor represents one scored decision-flow step
type Factor struct {
	Step   string  `json:"step"`
	Label  string  `json:"label"`
	Side   string  `json:"side"` // "bull" | "bear" | "neutral"
	Weight float64 `json:"weight"`
}

// Signal is the final output for a market+timeframe
type Signal struct {
	Symbol      string    `json:"symbol"`
	Type        string    `json:"type"` // "forex" | "synthetic"
	Timeframe   string    `json:"timeframe"`
	Price       float64   `json:"price"`
	Signal      string    `json:"signal"` // "BUY" | "SELL" | "WAIT"
	Confidence  int       `json:"confidence"`
	TP1         *float64  `json:"tp1,omitempty"`
	TP2         *float64  `json:"tp2,omitempty"`
	SL          *float64  `json:"sl,omitempty"`
	RR          *float64  `json:"rr,omitempty"`
	Pips        *float64  `json:"pips,omitempty"`
	ATR         *float64  `json:"atr,omitempty"`
	BullScore   float64   `json:"bullScore"`
	BearScore   float64   `json:"bearScore"`
	MaxScore    float64   `json:"maxScore"`
	RSI         float64   `json:"rsi"`
	MACDBull    bool      `json:"macdBull"`
	Trend       string    `json:"trend"`
	HTF1Bias    string    `json:"htf1Bias"`
	HTF2Bias    string    `json:"htf2Bias"`
	Structure   string    `json:"structure"`
	CounterTrend bool     `json:"counterTrend"`
	BlockReason string    `json:"blockReason,omitempty"`
	Factors     []Factor  `json:"factors"`
	Timestamp   time.Time `json:"timestamp"`
	Error       string    `json:"error,omitempty"`
}

// SignalOutcome tracks whether a past signal hit TP or SL
type SignalOutcome struct {
	SignalID    string    `json:"signalId"`
	Symbol      string    `json:"symbol"`
	Signal      string    `json:"signal"`
	EntryPrice  float64   `json:"entryPrice"`
	TP1         float64   `json:"tp1"`
	SL          float64   `json:"sl"`
	Outcome     string    `json:"outcome"` // "TP1_HIT" | "SL_HIT" | "PENDING" | "EXPIRED"
	ClosedAt    *time.Time `json:"closedAt,omitempty"`
	CreatedAt   time.Time `json:"createdAt"`
}

// PerformanceStats aggregates win rate per symbol
type PerformanceStats struct {
	Symbol      string  `json:"symbol"`
	TotalSignals int    `json:"totalSignals"`
	Wins        int     `json:"wins"`
	Losses      int     `json:"losses"`
	Pending     int     `json:"pending"`
	WinRate     float64 `json:"winRate"`
}
