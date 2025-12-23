package writer

import (
	"fmt"
	"testing"
	"time"
)

func TestMetricsData_Initialization(t *testing.T) {
	data := MetricsData{
		Timestamp:           time.Date(2025, 12, 23, 10, 0, 0, 0, time.UTC),
		ObsRTT:              50 * time.Millisecond,
		GoogleRTT:           25 * time.Millisecond,
		StreamActive:        true,
		OutputBytes:         1024.0,
		OutputSkippedFrames: 5.0,
		OutputFrames:        100.0,
		ObsCpuUsage:         15.5,
		ObsMemoryUsage:      512.0,
		SystemCpuUsage:      45.2,
		SystemMemoryUsage:   60.0,
	}

	if data.Timestamp.IsZero() {
		t.Error("Expected Timestamp to be set")
	}
	if data.ObsRTT != 50*time.Millisecond {
		t.Errorf("Expected ObsRTT to be 50ms, got %v", data.ObsRTT)
	}
	if data.StreamActive != true {
		t.Error("Expected StreamActive to be true")
	}
	if data.OutputBytes != 1024.0 {
		t.Errorf("Expected OutputBytes to be 1024.0, got %f", data.OutputBytes)
	}
}

func TestMetricsData_DefaultValues(t *testing.T) {
	data := MetricsData{}

	if !data.Timestamp.IsZero() {
		t.Error("Expected default Timestamp to be zero")
	}
	if data.ObsRTT != 0 {
		t.Errorf("Expected default ObsRTT to be 0, got %v", data.ObsRTT)
	}
	if data.GoogleRTT != 0 {
		t.Errorf("Expected default GoogleRTT to be 0, got %v", data.GoogleRTT)
	}
	if data.StreamActive != false {
		t.Error("Expected default StreamActive to be false")
	}
	if data.OutputBytes != 0 {
		t.Errorf("Expected default OutputBytes to be 0, got %f", data.OutputBytes)
	}
	if data.OutputSkippedFrames != 0 {
		t.Errorf("Expected default OutputSkippedFrames to be 0, got %f", data.OutputSkippedFrames)
	}
	if data.OutputFrames != 0 {
		t.Errorf("Expected default OutputFrames to be 0, got %f", data.OutputFrames)
	}
	if data.ObsCpuUsage != 0 {
		t.Errorf("Expected default ObsCpuUsage to be 0, got %f", data.ObsCpuUsage)
	}
	if data.ObsMemoryUsage != 0 {
		t.Errorf("Expected default ObsMemoryUsage to be 0, got %f", data.ObsMemoryUsage)
	}
	if data.SystemCpuUsage != 0 {
		t.Errorf("Expected default SystemCpuUsage to be 0, got %f", data.SystemCpuUsage)
	}
	if data.SystemMemoryUsage != 0 {
		t.Errorf("Expected default SystemMemoryUsage to be 0, got %f", data.SystemMemoryUsage)
	}
}

func TestMetricsData_ZeroValuesForOptionalFields(t *testing.T) {
	data := MetricsData{
		Timestamp:    time.Now(),
		StreamActive: true,
	}

	if data.ObsPingError != nil {
		t.Error("Expected ObsPingError to be nil")
	}
	if data.GooglePingError != nil {
		t.Error("Expected GooglePingError to be nil")
	}
	if data.StreamError != nil {
		t.Error("Expected StreamError to be nil")
	}
	if data.ObsStatsError != nil {
		t.Error("Expected ObsStatsError to be nil")
	}
	if data.SystemMetricsError != nil {
		t.Error("Expected SystemMetricsError to be nil")
	}
}

func TestMetricsData_WithErrors(t *testing.T) {
	obsErr := fmt.Errorf("obs ping failed")
	googleErr := fmt.Errorf("google ping failed")
	streamErr := fmt.Errorf("stream error")
	obsStatsErr := fmt.Errorf("obs stats error")
	systemErr := fmt.Errorf("system metrics error")

	data := MetricsData{
		Timestamp:          time.Now(),
		ObsPingError:       obsErr,
		GooglePingError:    googleErr,
		StreamError:        streamErr,
		ObsStatsError:      obsStatsErr,
		SystemMetricsError: systemErr,
	}

	if data.ObsPingError == nil {
		t.Error("Expected ObsPingError to be set")
	}
	if data.GooglePingError == nil {
		t.Error("Expected GooglePingError to be set")
	}
	if data.StreamError == nil {
		t.Error("Expected StreamError to be set")
	}
	if data.ObsStatsError == nil {
		t.Error("Expected ObsStatsError to be set")
	}
	if data.SystemMetricsError == nil {
		t.Error("Expected SystemMetricsError to be set")
	}
}

func TestMetricsData_HighValues(t *testing.T) {
	data := MetricsData{
		Timestamp:           time.Now(),
		ObsRTT:              5000 * time.Millisecond,
		GoogleRTT:           3000 * time.Millisecond,
		StreamActive:        true,
		OutputBytes:         999999999.0,
		OutputSkippedFrames: 10000.0,
		OutputFrames:        100000.0,
		ObsCpuUsage:         99.9,
		ObsMemoryUsage:      8192.0,
		SystemCpuUsage:      100.0,
		SystemMemoryUsage:   99.9,
	}

	if data.ObsRTT != 5000*time.Millisecond {
		t.Errorf("Expected high ObsRTT value, got %v", data.ObsRTT)
	}
	if data.OutputBytes != 999999999.0 {
		t.Errorf("Expected high OutputBytes value, got %f", data.OutputBytes)
	}
	if data.SystemCpuUsage != 100.0 {
		t.Errorf("Expected SystemCpuUsage 100.0, got %f", data.SystemCpuUsage)
	}
}

func TestMetricsData_NegativeValues(t *testing.T) {
	data := MetricsData{
		Timestamp:           time.Now(),
		OutputBytes:         -100.0,
		OutputSkippedFrames: -5.0,
		ObsCpuUsage:         -10.0,
	}

	if data.OutputBytes >= 0 {
		t.Logf("Note: Negative OutputBytes accepted: %f", data.OutputBytes)
	}
}

func TestMetricsData_TimestampPrecision(t *testing.T) {
	now := time.Date(2025, 12, 23, 10, 30, 45, 123456789, time.UTC)
	data := MetricsData{
		Timestamp: now,
	}

	if data.Timestamp != now {
		t.Error("Timestamp precision not preserved")
	}
	if data.Timestamp.Nanosecond() != 123456789 {
		t.Errorf("Expected nanosecond precision, got %d", data.Timestamp.Nanosecond())
	}
}

func TestMetricsData_PartialData(t *testing.T) {
	data := MetricsData{
		Timestamp:    time.Now(),
		ObsRTT:       50 * time.Millisecond,
		StreamActive: true,
	}

	if data.GoogleRTT != 0 {
		t.Error("Expected unset GoogleRTT to be zero")
	}
	if data.OutputBytes != 0 {
		t.Error("Expected unset OutputBytes to be zero")
	}
}
