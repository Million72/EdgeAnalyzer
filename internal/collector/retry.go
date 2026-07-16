package collector

import (
	"time"

	"otc-predictor/pkg/types"
)

// FetchCandlesWithRetry attempts the fetch, retrying once after a short backoff
// if the first attempt fails (handles transient Deriv WS hiccups).
func (d *DerivClient) FetchCandlesWithRetry(symbol string, granularity, count int) ([]types.Candle, error) {
	candles, err := d.FetchCandles(symbol, granularity, count)
	if err == nil {
		return candles, nil
	}
	time.Sleep(1500 * time.Millisecond)
	return d.FetchCandles(symbol, granularity, count)
}
