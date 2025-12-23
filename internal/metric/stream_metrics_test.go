package metric

import (
	"sync"
	"testing"
	"time"
)

func TestStreamMetrics_GetAndResetMaxValues_NoMeasurements(t *testing.T) {
	sm := &StreamMetrics{
		measurementCount: 0,
	}

	data := sm.GetAndResetMaxValues()

	if data.OutputBytes != 0 {
		t.Errorf("Expected OutputBytes to be 0, got %f", data.OutputBytes)
	}
	if data.OutputSkippedFrames != 0 {
		t.Errorf("Expected OutputSkippedFrames to be 0, got %f", data.OutputSkippedFrames)
	}
	if data.OutputFrames != 0 {
		t.Errorf("Expected OutputFrames to be 0, got %f", data.OutputFrames)
	}
}

func TestStreamMetrics_GetAndResetMaxValues_OneMeasurement(t *testing.T) {
	sm := &StreamMetrics{
		maxOutputBytes:    1000.0,
		maxSkippedFrames:  10.0,
		maxTotalFrames:    100.0,
		measurementCount:  1,
	}

	data := sm.GetAndResetMaxValues()

	if data.OutputBytes != 0 {
		t.Errorf("Expected OutputBytes to be 0 (not enough measurements for delta), got %f", data.OutputBytes)
	}
	if data.OutputSkippedFrames != 0 {
		t.Errorf("Expected OutputSkippedFrames to be 0, got %f", data.OutputSkippedFrames)
	}
	if data.OutputFrames != 0 {
		t.Errorf("Expected OutputFrames to be 0, got %f", data.OutputFrames)
	}

	if sm.prevOutputBytes != 1000.0 {
		t.Errorf("Expected prevOutputBytes to be set to 1000.0, got %f", sm.prevOutputBytes)
	}
}

func TestStreamMetrics_GetAndResetMaxValues_MultipleMeasurements(t *testing.T) {
	sm := &StreamMetrics{
		prevOutputBytes:   1000.0,
		prevSkippedFrames: 10.0,
		prevTotalFrames:   100.0,
		maxOutputBytes:    2500.0,
		maxSkippedFrames:  25.0,
		maxTotalFrames:    250.0,
		measurementCount:  5,
	}

	data := sm.GetAndResetMaxValues()

	expectedBytes := 2500.0 - 1000.0
	if data.OutputBytes != expectedBytes {
		t.Errorf("Expected OutputBytes to be %f, got %f", expectedBytes, data.OutputBytes)
	}

	expectedSkipped := 25.0 - 10.0
	if data.OutputSkippedFrames != expectedSkipped {
		t.Errorf("Expected OutputSkippedFrames to be %f, got %f", expectedSkipped, data.OutputSkippedFrames)
	}

	expectedFrames := 250.0 - 100.0
	if data.OutputFrames != expectedFrames {
		t.Errorf("Expected OutputFrames to be %f, got %f", expectedFrames, data.OutputFrames)
	}

	if sm.prevOutputBytes != 2500.0 {
		t.Errorf("Expected prevOutputBytes to be updated to 2500.0, got %f", sm.prevOutputBytes)
	}
	if sm.prevSkippedFrames != 25.0 {
		t.Errorf("Expected prevSkippedFrames to be updated to 25.0, got %f", sm.prevSkippedFrames)
	}
	if sm.prevTotalFrames != 250.0 {
		t.Errorf("Expected prevTotalFrames to be updated to 250.0, got %f", sm.prevTotalFrames)
	}
}

func TestStreamMetrics_GetAndResetMaxValues_MaxValueTracking(t *testing.T) {
	sm := &StreamMetrics{
		prevOutputBytes:   1000.0,
		maxOutputBytes:    3000.0,
		measurementCount:  3,
	}

	data1 := sm.GetAndResetMaxValues()
	expectedDelta1 := 3000.0 - 1000.0
	if data1.OutputBytes != expectedDelta1 {
		t.Errorf("Expected first call to return delta %f, got %f", expectedDelta1, data1.OutputBytes)
	}

	sm.maxOutputBytes = 3500.0
	sm.measurementCount = 3

	data2 := sm.GetAndResetMaxValues()
	expectedDelta2 := 3500.0 - 3000.0
	if data2.OutputBytes != expectedDelta2 {
		t.Errorf("Expected second call to return delta %f, got %f", expectedDelta2, data2.OutputBytes)
	}
}

func TestStreamMetrics_GetAndResetMaxValues_ConcurrentAccess(t *testing.T) {
	sm := &StreamMetrics{
		prevOutputBytes:  1000.0,
		maxOutputBytes:   2000.0,
		measurementCount: 2,
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

func TestStreamMetrics_NewStreamMetrics(t *testing.T) {
	interval := 100 * time.Millisecond

	sm, err := NewStreamMetrics(nil, interval)

	if err != nil {
		t.Fatalf("NewStreamMetrics returned error: %v", err)
	}
	if sm == nil {
		t.Fatal("NewStreamMetrics returned nil")
	}
	if sm.interval != interval {
		t.Errorf("Expected interval %v, got %v", interval, sm.interval)
	}
}
