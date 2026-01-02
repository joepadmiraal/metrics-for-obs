package metric

import (
	"fmt"
	"sync"
	"time"

	"github.com/andreykaipov/goobs"
)

type ObsStats struct {
	client               *goobs.Client
	maxObsCpuUsage       float64
	maxObsMemoryUsage    float64
	lastError            error
	measurementCount     int
	measurementsSinceGet int
	mu                   sync.Mutex
	interval             time.Duration
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

	if s.measurementsSinceGet == 0 && s.measurementCount > 0 {
		return ObsStatsData{
			Timestamp:      time.Now(),
			ObsCpuUsage:    0,
			ObsMemoryUsage: 0,
			Error:          fmt.Errorf("no new measurements collected since last read"),
		}
	}

	s.maxObsCpuUsage = 0
	s.maxObsMemoryUsage = 0
	s.lastError = nil
	s.measurementsSinceGet = 0

	return ObsStatsData{
		Timestamp:      time.Now(),
		ObsCpuUsage:    maxCpu,
		ObsMemoryUsage: maxMemory,
		Error:          err,
	}
}

func (s *ObsStats) updateStats(cpuUsage, memoryUsage float64) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if cpuUsage > s.maxObsCpuUsage {
		s.maxObsCpuUsage = cpuUsage
	}
	if memoryUsage > s.maxObsMemoryUsage {
		s.maxObsMemoryUsage = memoryUsage
	}
	s.measurementCount++
	s.measurementsSinceGet++
}

func (s *ObsStats) recordError(err error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.lastError = err
}

func (s *ObsStats) Start() error {
	ticker := time.NewTicker(s.interval)
	defer ticker.Stop()

	for range ticker.C {
		stats, err := s.client.General.GetStats()

		if err != nil {
			s.recordError(err)
			continue
		}

		s.updateStats(stats.CpuUsage, stats.MemoryUsage)
	}

	return nil
}
