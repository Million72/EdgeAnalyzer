package strategy

import "otc-predictor/pkg/types"

type Pattern struct {
	Name     string
	Side     string // "bull" | "bear" | "neutral"
	Strength float64
}

// CandlestickPatterns detects single/multi-candle formations on the last 3 candles.
func CandlestickPatterns(candles []types.Candle) []Pattern {
	if len(candles) < 3 {
		return nil
	}
	pats := []Pattern{}
	n := len(candles)
	c0, c1, c2 := candles[n-1], candles[n-2], candles[n-3]

	if isBear(c1) && isBull(c0) && c0.Open <= c1.Close && c0.Close >= c1.Open {
		pats = append(pats, Pattern{"Bullish Engulfing", "bull", 3})
	}
	if isBull(c1) && isBear(c0) && c0.Open >= c1.Close && c0.Close <= c1.Open {
		pats = append(pats, Pattern{"Bearish Engulfing", "bear", 3})
	}
	if dnWick(c0) > body(c0)*2 && upWick(c0) < body(c0)*0.5 && body(c0) < candleRange(c0)*0.4 {
		pats = append(pats, Pattern{"Hammer", "bull", 2})
	}
	if upWick(c0) > body(c0)*2 && dnWick(c0) < body(c0)*0.5 && body(c0) < candleRange(c0)*0.4 {
		pats = append(pats, Pattern{"Shooting Star", "bear", 2})
	}
	if isBull(c2) && isBull(c1) && isBull(c0) && c1.Close > c2.Close && c0.Close > c1.Close {
		pats = append(pats, Pattern{"Three White Soldiers", "bull", 3})
	}
	if isBear(c2) && isBear(c1) && isBear(c0) && c1.Close < c2.Close && c0.Close < c1.Close {
		pats = append(pats, Pattern{"Three Black Crows", "bear", 3})
	}
	if isBear(c2) && body(c1) < candleRange(c1)*0.3 && isBull(c0) && c0.Close > (c2.Open+c2.Close)/2 {
		pats = append(pats, Pattern{"Morning Star", "bull", 3})
	}
	if isBull(c2) && body(c1) < candleRange(c1)*0.3 && isBear(c0) && c0.Close < (c2.Open+c2.Close)/2 {
		pats = append(pats, Pattern{"Evening Star", "bear", 3})
	}
	return pats
}

// ChartPatterns detects multi-swing formations: double top/bottom, H&S, BOS.
func ChartPatterns(candles []types.Candle, dec int) []Pattern {
	if len(candles) < 30 {
		return nil
	}
	pats := []Pattern{}
	start := 0
	if len(candles) > 80 {
		start = len(candles) - 80
	}
	slice := candles[start:]
	n := len(slice)
	tol := 0.012

	sH := swingHighs(slice, 3)
	sL := swingLows(slice, 3)
	curr := slice[n-1].Close

	// BOS
	if len(sH) > 0 && len(sL) > 0 {
		lastH := sH[len(sH)-1]
		lastL := sL[len(sL)-1]
		if curr > lastH.Price {
			pats = append(pats, Pattern{"BOS Bullish", "bull", 3})
		} else if curr < lastL.Price {
			pats = append(pats, Pattern{"BOS Bearish", "bear", 3})
		}
	}

	// Double Bottom
	for a := 0; a < len(sL)-1; a++ {
		found := false
		for b := a + 1; b < len(sL); b++ {
			if sL[b].Index-sL[a].Index < 5 {
				continue
			}
			if absF(sL[a].Price-sL[b].Price)/sL[a].Price < tol {
				peakBetween := false
				for _, h := range sH {
					if h.Index > sL[a].Index && h.Index < sL[b].Index {
						peakBetween = true
						break
					}
				}
				if peakBetween {
					pats = append(pats, Pattern{"Double Bottom", "bull", 3})
					found = true
					break
				}
			}
		}
		if found {
			break
		}
	}

	// Double Top
	for a := 0; a < len(sH)-1; a++ {
		found := false
		for b := a + 1; b < len(sH); b++ {
			if sH[b].Index-sH[a].Index < 5 {
				continue
			}
			if absF(sH[a].Price-sH[b].Price)/sH[a].Price < tol {
				troughBetween := false
				for _, l := range sL {
					if l.Index > sH[a].Index && l.Index < sH[b].Index {
						troughBetween = true
						break
					}
				}
				if troughBetween {
					pats = append(pats, Pattern{"Double Top", "bear", 3})
					found = true
					break
				}
			}
		}
		if found {
			break
		}
	}

	return pats
}
