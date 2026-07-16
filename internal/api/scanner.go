package api

import (
	"log"
	"sync"
	"time"

	"otc-predictor/internal/collector"
	"otc-predictor/internal/predictor"
	"otc-predictor/internal/spike"
	"otc-predictor/internal/storage"
	"otc-predictor/pkg/types"
)

type Scanner struct {
	cfg    *Config
	deriv  *collector.DerivClient
	store  *storage.Store
	cache  *collector.CandleCache
}

func NewScanner(cfg *Config, deriv *collector.DerivClient, store *storage.Store) *Scanner {
	return &Scanner{cfg: cfg, deriv: deriv, store: store, cache: collector.NewCandleCache()}
}

// Run performs one full scan across all markets and timeframes on a schedule.
func (s *Scanner) Run() {
	s.scanOnce()
	ticker := time.NewTicker(time.Duration(s.cfg.ScanIntervalSeconds) * time.Second)
	defer ticker.Stop()
	for range ticker.C {
		s.scanOnce()
	}
}

func (s *Scanner) scanOnce() {
	markets := s.cfg.AllMarkets()
	log.Printf("scanner: starting scan of %d markets across %d timeframes", len(markets), len(s.cfg.Timeframes))

	for tfName, tfCfg := range s.cfg.Timeframes {
		s.scanTimeframe(markets, tfName, tfCfg)
	}
	log.Println("scanner: scan complete")
}

func (s *Scanner) scanTimeframe(markets []types.Market, tfName string, tfCfg types.TimeframeConfig) {
	const batchSize = 3
	var wg sync.WaitGroup
	sem := make(chan struct{}, batchSize)

	for _, market := range markets {
		wg.Add(1)
		sem <- struct{}{}
		go func(m types.Market) {
			defer wg.Done()
			defer func() { <-sem }()
			s.scanMarket(m, tfName, tfCfg)
		}(market)
	}
	wg.Wait()
}

// fetchCandles checks the cache first, falls back to a retried live fetch,
// and populates the cache on success. This avoids hammering Deriv when the
// current candle for a granularity hasn't closed yet, while still recovering
// gracefully from transient WebSocket failures.
func (s *Scanner) fetchCandles(symbol string, granularity, count int) ([]types.Candle, error) {
	if cached, ok := s.cache.Get(symbol, granularity); ok && len(cached) >= count {
		return cached, nil
	}
	fresh, err := s.deriv.FetchCandlesWithRetry(symbol, granularity, count)
	if err != nil {
		return nil, err
	}
	s.cache.Set(symbol, granularity, fresh)
	return fresh, nil
}

// isBoomCrash routes Boom/Crash symbols to the dedicated spike engine
// (survival analysis + post-spike reaction) instead of the general
// indicator-scoring engine used for forex and other synthetics — those two
// index types have a fundamentally different underlying process and need
// fundamentally different math, not just different parameters on the same model.
func isBoomCrash(symbol string) bool {
	return spike.ContainsFold(symbol, "boom") || spike.ContainsFold(symbol, "crash")
}

func (s *Scanner) scanMarket(market types.Market, tfName string, tfCfg types.TimeframeConfig) {
	candles, err := s.fetchCandles(market.Deriv, tfCfg.Granularity, s.cfg.CandleCount)
	if err != nil {
		log.Printf("scanner: %s/%s candles error: %v", market.Symbol, tfName, err)
		s.store.SetSignal(types.Signal{
			Symbol: market.Symbol, Type: market.Type, Timeframe: tfName,
			Signal: "WAIT", Error: err.Error(), Timestamp: time.Now(),
		})
		return
	}

	// Boom/Crash indices: route to the dedicated survival-analysis + post-spike
	// reaction engine. No HTF fetch needed here — the spike model works off the
	// raw candle history of the instrument itself, not multi-timeframe trend bias.
	if isBoomCrash(market.Symbol) {
		sig := spike.RunSpikeEngine(market, candles)
		sig.Timeframe = tfName
		s.store.SetSignal(sig)
		return
	}

	htf1, err := s.fetchCandles(market.Deriv, tfCfg.HTFGran, s.cfg.HTFCandleCount)
	if err != nil {
		log.Printf("scanner: %s/%s htf1 error: %v", market.Symbol, tfName, err)
		htf1 = nil
	}

	htf2, err := s.fetchCandles(market.Deriv, tfCfg.HTF2Gran, s.cfg.HTFCandleCount)
	if err != nil {
		log.Printf("scanner: %s/%s htf2 error: %v", market.Symbol, tfName, err)
		htf2 = nil
	}

	sig := predictor.BuildSignal(market, tfName, candles, htf1, htf2)
	s.store.SetSignal(sig)
}
