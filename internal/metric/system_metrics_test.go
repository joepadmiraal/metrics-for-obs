package metric

import (
	"fmt"
	"sync"
	"testing"
	"time"
)

func TestSystemMetrics_GetAndResetMaxValues_ReturnsCorrectMaxValues(t *testing.T) {
	sm := &SystemMetrics{
		maxCpuUsage:    75.5,
		maxMemoryUsage: 85.2,
	}

	data := sm.GetAndResetMaxValues()

	if data.CpuUsage != 75.5 {
		t.Errorf("Expected CpuUsage to be 75.5, got %f", data.CpuUsage)
	}
	if data.MemoryUsage != 85.2 {
		t.Errorf("Expected MemoryUsage to be 85.2, got %f", data.MemoryUsage)
	}
}

func TestSystemMetrics_GetAndResetMaxValues_ResetsValues(t *testing.T) {
	sm := &SystemMetrics{
		maxCpuUsage:    75.5,
		maxMemoryUsage: 85.2,
	}

	_ = sm.GetAndResetMaxValues()

	if sm.maxCpuUsage != 0 {
		t.Errorf("Expected maxCpuUsage to be reset to 0, got %f", sm.maxCpuUsage)
	}
	if sm.maxMemoryUsage != 0 {
		t.Errorf("Expected maxMemoryUsage to be reset to 0, got %f", sm.maxMemoryUsage)
	}
	if sm.lastError != nil {
		t.Error("Expected lastError to be reset to nil")
	}
}

func TestSystemMetrics_GetAndResetMaxValues_ErrorHandling(t *testing.T) {
	testError := fmt.Errorf("system error")
	sm := &SystemMetrics{
		maxCpuUsage: 50.0,
		lastError:   testError,
	}

	data := sm.GetAndResetMaxValues()

	if data.Error == nil {
		t.Error("Expected error to be returned")
	}
	if data.Error != testError {
		t.Error("Expected error pointer to match")
	}

	if sm.lastError != nil {
		t.Error("Expected lastError to be reset to nil after GetAndResetMaxValues")
	}
}

func TestSystemMetrics_GetAndResetMaxValues_ConcurrentAccess(t *testing.T) {
	sm := &SystemMetrics{
		maxCpuUsage:    60.0,
		maxMemoryUsage: 70.0,
	}

	var wg sync.WaitGroup
	iterations := 100

	for i := 0; i < iterations; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			_ = sm.GetAndResetMaxValues()
		}()
	}

	wg.Wait()
}

func TestSystemMetrics_GetAndResetMaxValues_ZeroValues(t *testing.T) {
	sm := &SystemMetrics{
		maxCpuUsage:    0,
		maxMemoryUsage: 0,
	}

	data := sm.GetAndResetMaxValues()

	if data.CpuUsage != 0 {
		t.Errorf("Expected CpuUsage to be 0, got %f", data.CpuUsage)
	}
	if data.MemoryUsage != 0 {
		t.Errorf("Expected MemoryUsage to be 0, got %f", data.MemoryUsage)
	}
	if data.Error != nil {
		t.Error("Expected no error with zero values")
	}
}

func TestSystemMetrics_GetAndResetMaxValues_TracksMaximum(t *testing.T) {
	sm := &SystemMetrics{
		maxCpuUsage: 40.0,
	}

	data1 := sm.GetAndResetMaxValues()
	if data1.CpuUsage != 40.0 {
		t.Errorf("Expected first call to return 40.0, got %f", data1.CpuUsage)
	}

	sm.maxCpuUsage = 80.0

	data2 := sm.GetAndResetMaxValues()
	if data2.CpuUsage != 80.0 {
		t.Errorf("Expected second call to return 80.0 (new max), got %f", data2.CpuUsage)
	}
}

func TestSystemMetrics_NewSystemMetrics(t *testing.T) {
	interval := 100 * time.Millisecond

	sm, err := NewSystemMetrics(interval)

	if err != nil {
		t.Fatalf("NewSystemMetrics returned error: %v", err)
	}
	if sm == nil {
		t.Fatal("NewSystemMetrics returned nil")
	}
	if sm.interval != interval {
		t.Errorf("Expected interval %v, got %v", interval, sm.interval)
	}
}

func TestSystemMetrics_GetAndResetMaxValues_TimestampSet(t *testing.T) {
	sm := &SystemMetrics{
		maxCpuUsage: 50.0,
	}

	before := time.Now()
	data := sm.GetAndResetMaxValues()
	after := time.Now()

	if data.Timestamp.Before(before) || data.Timestamp.After(after) {
		t.Error("Expected timestamp to be set to current time")
	}
}

func TestSystemMetrics_GetAndResetMaxValues_HighMemoryUsage(t *testing.T) {
	sm := &SystemMetrics{
		maxMemoryUsage: 99.9,
	}

	data := sm.GetAndResetMaxValues()

	if data.MemoryUsage != 99.9 {
		t.Errorf("Expected MemoryUsage to be 99.9, got %f", data.MemoryUsage)
	}
}
