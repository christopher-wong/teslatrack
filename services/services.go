package services

import (
	"database/sql"
	log "github.com/sirupsen/logrus"
	"math"
	"sync"
)

type cache struct {
	sync.RWMutex
	batteryDegradationData []CompletedChargeDBResponse
	minFullChargeRange     float64
	maxFullChargeRange     float64
}

type ServicesClient struct {
	db     *sql.DB
	logger *log.Logger
	cache  *cache
}

func New(db *sql.DB, logger *log.Logger) *ServicesClient {
	return &ServicesClient{
		db, logger, &cache{},
	}
}

func (s *ServicesClient) GetBatteryDegradationData() []CompletedChargeDBResponse {
	s.cache.RLock()
	defer s.cache.RUnlock()

	return s.cache.batteryDegradationData
}

func (s *ServicesClient) CalculateBatteryDegradationStats() {
	results, err := s.GetCompletedChargeRows(1)
	if err != nil {
		log.Fatal("failed to fetch completed charge data")
	}

	var newResults []CompletedChargeDBResponse

	// store the minimum and maximum range estimates seen
	min := math.Inf(1)
	max := math.Inf(-1)
	for _, row := range *results {
		// linearly extrapolate the estimate range at 100%
		// given current Battery Level (%) and BatteryRange(mi), calculate estimate range at a 100% BatteryLevel
		est100pct := float64(row.BatteryRange) * float64(100) / float64(row.BatteryLevel)
		row.EstBatteryRange100Pct = est100pct

		min = math.Min(float64(est100pct), min)
		max = math.Max(float64(est100pct), max)

		newResults = append(newResults, row)
	}

	s.cache.Lock()
	defer s.cache.Unlock()

	s.cache.maxFullChargeRange = max
	s.cache.minFullChargeRange = min
	s.cache.batteryDegradationData = newResults
}
