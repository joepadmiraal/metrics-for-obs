package metric

import (
	"fmt"
	"sync"
	"testing"
	"time"
)

func TestObsStats_GetAndResetMaxValues_ReturnsCorrectMaxValues(t *testing.T) {
	obs := &ObsStats{
		maxObsCpuUsage:    25.5,
		maxObsMemoryUsage: 1024.0,
		measurementCount:  5,
	}

	data := obs.GetAndResetMaxValues()

	if data.ObsCpuUsage != 25.5 {
		t.Errorf("Expected ObsCpuUsage to be 25.5, got %f", data.ObsCpuUsage)
	}
	if data.ObsMemoryUsage != 1024.0 {
		t.Errorf("Expected ObsMemoryUsage to be 1024.0, got %f", data.ObsMemoryUsage)
	}
}

func TestObsStats_GetAndResetMaxValues_ResetsValues(t *testing.T) {
	obs := &ObsStats{
		maxObsCpuUsage:    25.5,
		maxObsMemoryUsage: 1024.0,
		measurementCount:  5,
	}

	_ = obs.GetAndResetMaxValues()

	if obs.maxObsCpuUsage != 0 {
		t.Errorf("Expected maxObsCpuUsage to be reset to 0, got %f", obs.maxObsCpuUsage)
	}
	if obs.maxObsMemoryUsage != 0 {
		t.Errorf("Expected maxObsMemoryUsage to be reset to 0, got %f", obs.maxObsMemoryUsage)
	}
	if obs.lastError != nil {
		t.Error("Expected lastError to be reset to nil")
	}
}

func TestObsStats_GetAndResetMaxValues_ErrorHandling(t *testing.T) {
	testError := fmt.Errorf("test error")
	obs := &ObsStats{
		maxObsCpuUsage:   10.0,
		lastError:        testError,
		measurementCount: 3,
	}

	data := obs.GetAndResetMaxValues()

	if data.Error == nil {
		t.Error("Expected error to be returned")
	}
	if data.Error != testError {
		t.Error("Expected error pointer to match")
	}

	if obs.lastError != nil {
		t.Error("Expected lastError to be reset to nil after GetAndResetMaxValues")
	}
}

func TestObsStats_GetAndResetMaxValues_ConcurrentAccess(t *testing.T) {
	obs := &ObsStats{
		maxObsCpuUsage:    15.0,
		maxObsMemoryUsage: 512.0,
		measurementCount:  10,
	}

	var wg sync.WaitGroup
	iterations := 100

	for i := 0; i < iterations; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			_ = obs.GetAndResetMaxValues()
		}()
	}

	wg.Wait()
}

func TestObsStats_GetAndResetMaxValues_ZeroValues(t *testing.T) {
	obs := &ObsStats{
		maxObsCpuUsage:    0,
		maxObsMemoryUsage: 0,
		measurementCount:  0,
	}

	data := obs.GetAndResetMaxValues()

	if data.ObsCpuUsage != 0 {
		t.Errorf("Expected ObsCpuUsage to be 0, got %f", data.ObsCpuUsage)
	}
	if data.ObsMemoryUsage != 0 {
		t.Errorf("Expected ObsMemoryUsage to be 0, got %f", data.ObsMemoryUsage)
	}
	if data.Error != nil {
		t.Error("Expected no error with zero values")
	}
}

func TestObsStats_GetAndResetMaxValues_TracksMaximum(t *testing.T) {
	obs := &ObsStats{
		maxObsCpuUsage:   10.0,
		measurementCount: 2,
	}

	data1 := obs.GetAndResetMaxValues()
	if data1.ObsCpuUsage != 10.0 {
		t.Errorf("Expected first call to return 10.0, got %f", data1.ObsCpuUsage)
	}

	obs.maxObsCpuUsage = 20.0
	obs.measurementCount = 3

	data2 := obs.GetAndResetMaxValues()
	if data2.ObsCpuUsage != 20.0 {
		t.Errorf("Expected second call to return 20.0 (new max), got %f", data2.ObsCpuUsage)
	}
}

func TestObsStats_NewObsStats(t *testing.T) {
	interval := 100 * time.Millisecond

	obs, err := NewObsStats(nil, interval)

	if err != nil {
		t.Fatalf("NewObsStats returned error: %v", err)
	}
	if obs == nil {
		t.Fatal("NewObsStats returned nil")
	}
	if obs.interval != interval {
		t.Errorf("Expected interval %v, got %v", interval, obs.interval)
	}
}

func TestObsStats_GetAndResetMaxValues_SetsTimestamp(t *testing.T) {
	obs := &ObsStats{
		maxObsCpuUsage: 15.0,
	}

	before := time.Now()
	data := obs.GetAndResetMaxValues()
	after := time.Now()

	if data.Timestamp.Before(before) || data.Timestamp.After(after) {
		t.Error("Expected timestamp to be set to current time")
	}
}
