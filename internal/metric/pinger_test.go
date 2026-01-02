package metric

import (
	"fmt"
	"sync"
	"testing"
	"time"
)

func TestPinger_GetAndResetMaxRTT_ReturnsCorrectMaxValue(t *testing.T) {
	p := &Pinger{
		maxRTT:               150 * time.Millisecond,
		measurementCount:     3,
		measurementsSinceGet: 1,
	}

	rtt, err := p.GetAndResetMaxRTT()

	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}
	if rtt != 150*time.Millisecond {
		t.Errorf("Expected RTT to be 150ms, got %v", rtt)
	}
}

func TestPinger_GetAndResetMaxRTT_ResetsValue(t *testing.T) {
	p := &Pinger{
		maxRTT:               150 * time.Millisecond,
		measurementCount:     2,
		measurementsSinceGet: 1,
	}

	_, _ = p.GetAndResetMaxRTT()

	if p.maxRTT != 0 {
		t.Errorf("Expected maxRTT to be reset to 0, got %v", p.maxRTT)
	}
	if p.lastError != nil {
		t.Error("Expected lastError to be reset to nil")
	}
}

func TestPinger_GetAndResetMaxRTT_ErrorHandling(t *testing.T) {
	testError := fmt.Errorf("ping error")
	p := &Pinger{
		maxRTT:               100 * time.Millisecond,
		lastError:            testError,
		measurementCount:     4,
		measurementsSinceGet: 2,
	}

	rtt, err := p.GetAndResetMaxRTT()

	if err == nil {
		t.Error("Expected error to be returned")
	}
	if err != testError {
		t.Error("Expected error pointer to match")
	}
	if rtt != 100*time.Millisecond {
		t.Errorf("Expected RTT to be returned even with error, got %v", rtt)
	}

	if p.lastError != nil {
		t.Error("Expected lastError to be reset to nil after GetAndResetMaxRTT")
	}
}

func TestPinger_GetAndResetMaxRTT_ConcurrentAccess(t *testing.T) {
	p := &Pinger{
		maxRTT:               200 * time.Millisecond,
		measurementCount:     1,
		measurementsSinceGet: 1,
	}

	var wg sync.WaitGroup
	iterations := 100

	for i := 0; i < iterations; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			_, _ = p.GetAndResetMaxRTT()
		}()
	}

	wg.Wait()
}

func TestPinger_GetAndResetMaxRTT_ZeroValue(t *testing.T) {
	p := &Pinger{
		maxRTT: 0,
	}

	rtt, err := p.GetAndResetMaxRTT()

	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}
	if rtt != 0 {
		t.Errorf("Expected RTT to be 0, got %v", rtt)
	}
}

func TestPinger_GetAndResetMaxRTT_TracksMaximum(t *testing.T) {
	p := &Pinger{
		maxRTT:               50 * time.Millisecond,
		measurementCount:     1,
		measurementsSinceGet: 1,
	}

	rtt1, _ := p.GetAndResetMaxRTT()
	if rtt1 != 50*time.Millisecond {
		t.Errorf("Expected first call to return 50ms, got %v", rtt1)
	}

	p.maxRTT = 200 * time.Millisecond
	p.measurementsSinceGet = 1

	rtt2, _ := p.GetAndResetMaxRTT()
	if rtt2 != 200*time.Millisecond {
		t.Errorf("Expected second call to return 200ms (new max), got %v", rtt2)
	}
}

func TestPinger_GetAndResetMaxRTT_NoNewMeasurements(t *testing.T) {
	p := &Pinger{
		maxRTT:               100 * time.Millisecond,
		measurementCount:     5,
		measurementsSinceGet: 0,
	}

	rtt, err := p.GetAndResetMaxRTT()

	if err == nil {
		t.Error("Expected error when no new measurements collected")
	}
	if err != nil && err.Error() != "no new measurements collected since last read" {
		t.Errorf("Expected specific error message, got: %v", err)
	}
	if rtt != 0 {
		t.Errorf("Expected RTT to be 0, got %v", rtt)
	}
}

func TestPinger_NewPinger(t *testing.T) {
	domain := "example.com"
	interval := 1 * time.Second

	p, err := NewPinger(domain, interval)

	if err != nil {
		t.Fatalf("NewPinger returned error: %v", err)
	}
	if p == nil {
		t.Fatal("NewPinger returned nil")
	}
	if p.domain != domain {
		t.Errorf("Expected domain %s, got %s", domain, p.domain)
	}
	if p.interval != interval {
		t.Errorf("Expected interval %v, got %v", interval, p.interval)
	}
}

func TestPinger_GetAndResetMaxRTT_HighRTT(t *testing.T) {
	p := &Pinger{
		maxRTT:               1500 * time.Millisecond,
		measurementCount:     1,
		measurementsSinceGet: 1,
	}

	rtt, err := p.GetAndResetMaxRTT()

	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}
	if rtt != 1500*time.Millisecond {
		t.Errorf("Expected RTT to be 1500ms, got %v", rtt)
	}
}

func TestPinger_GetAndResetMaxRTT_MultipleResets(t *testing.T) {
	p := &Pinger{
		maxRTT:               100 * time.Millisecond,
		measurementCount:     1,
		measurementsSinceGet: 1,
	}

	rtt1, _ := p.GetAndResetMaxRTT()
	if rtt1 != 100*time.Millisecond {
		t.Errorf("Expected first call to return 100ms, got %v", rtt1)
	}

	rtt2, _ := p.GetAndResetMaxRTT()
	if rtt2 != 0 {
		t.Errorf("Expected second call to return 0 after reset, got %v", rtt2)
	}
}
