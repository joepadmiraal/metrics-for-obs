package metric

import (
	"sync"
	"time"

	"github.com/andreykaipov/goobs"
)

type ObsStats struct {
	client            *goobs.Client
	maxObsCpuUsage    float64
	maxObsMemoryUsage float64
	lastError         error
	measurementCount  int
	mu                sync.Mutex
	interval          time.Duration
}

type ObsStatsData struct {
	Timestamp      time.Time
	ObsCpuUsage    float64
	ObsMemoryUsage float64
	Error          error
}

func NewObsStats(client *goobs.Client, interval time.Duration) (*ObsStats, error) {
	return &ObsStats{
		client:   client,
		interval: interval,
	}, nil
}

func (s *ObsStats) GetAndResetMaxValues() ObsStatsData {
	s.mu.Lock()
	defer s.mu.Unlock()

	maxCpu := s.maxObsCpuUsage
	maxMemory := s.maxObsMemoryUsage
	err := s.lastError

	s.maxObsCpuUsage = 0
	s.maxObsMemoryUsage = 0
	s.lastError = nil

	return ObsStatsData{
		Timestamp:      time.Now(),
		ObsCpuUsage:    maxCpu,
		ObsMemoryUsage: maxMemory,
		Error:          err,
	}
}

func (s *ObsStats) Start() error {
	ticker := time.NewTicker(s.interval)
	defer ticker.Stop()

	for range ticker.C {
		stats, err := s.client.General.GetStats()

		s.mu.Lock()
		if err != nil {
			s.lastError = err
		} else {
			if stats.CpuUsage > s.maxObsCpuUsage {
				s.maxObsCpuUsage = stats.CpuUsage
			}
			if stats.MemoryUsage > s.maxObsMemoryUsage {
				s.maxObsMemoryUsage = stats.MemoryUsage
			}
			s.measurementCount++
		}
		s.mu.Unlock()
	}

	return nil
}
