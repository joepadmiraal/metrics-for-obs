package integration

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/joepadmiraal/metrics-for-obs/internal/monitor"
)

func TestMonitor_Integration_BasicFlow(t *testing.T) {
	mockServer := NewMockOBSServer()
	defer mockServer.Close()

	tmpDir := t.TempDir()
	csvFile := filepath.Join(tmpDir, "test-metrics.csv")

	host := strings.Replace(mockServer.URL(), "ws://", "", 1)

	connInfo := monitor.ObsConnectionInfo{
		Password:       "",
		Host:           host,
		CSVFile:        csvFile,
		MetricInterval: 100,
		WriterInterval: 200,
	}

	mon, err := monitor.NewMonitor(connInfo)
	if err != nil {
		t.Fatalf("Failed to create monitor: %v", err)
	}

	if err := mon.Start(); err != nil {
		t.Fatalf("Failed to start monitor: %v", err)
	}
	defer mon.Close()

	time.Sleep(500 * time.Millisecond)

	mon.Shutdown()
	select {
	case <-mon.Done():
	case <-time.After(2 * time.Second):
		t.Fatal("Monitor did not shut down in time")
	}

	if _, err := os.Stat(csvFile); os.IsNotExist(err) {
		t.Errorf("CSV file was not created")
	}
}

func TestMonitor_Integration_MetricCollection(t *testing.T) {
	tests := []struct {
		name          string
		streamActive  bool
		outputBytes   float64
		skippedFrames float64
		totalFrames   float64
		cpuUsage      float64
		memoryUsage   float64
		expectData    bool
	}{
		{
			name:          "stream active with data",
			streamActive:  true,
			outputBytes:   1000000,
			skippedFrames: 10,
			totalFrames:   1000,
			cpuUsage:      25.5,
			memoryUsage:   512.0,
			expectData:    true,
		},
		{
			name:          "stream inactive",
			streamActive:  false,
			outputBytes:   0,
			skippedFrames: 0,
			totalFrames:   0,
			cpuUsage:      5.0,
			memoryUsage:   128.0,
			expectData:    true,
		},
		{
			name:          "high cpu usage",
			streamActive:  true,
			outputBytes:   5000000,
			skippedFrames: 100,
			totalFrames:   5000,
			cpuUsage:      85.5,
			memoryUsage:   1024.0,
			expectData:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockServer := NewMockOBSServer()
			defer mockServer.Close()

			mockServer.SetStreamActive(tt.streamActive)
			mockServer.SetStats(tt.cpuUsage, tt.memoryUsage)
			mockServer.IncrementStreamMetrics(tt.outputBytes, tt.skippedFrames, tt.totalFrames)

			tmpDir := t.TempDir()
			csvFile := filepath.Join(tmpDir, "test-metrics.csv")

			host := strings.Replace(mockServer.URL(), "ws://", "", 1)

			connInfo := monitor.ObsConnectionInfo{
				Password:       "",
				Host:           host,
				CSVFile:        csvFile,
				MetricInterval: 100,
				WriterInterval: 200,
			}

			mon, err := monitor.NewMonitor(connInfo)
			if err != nil {
				t.Fatalf("Failed to create monitor: %v", err)
			}

			if err := mon.Start(); err != nil {
				t.Fatalf("Failed to start monitor: %v", err)
			}

			time.Sleep(500 * time.Millisecond)

			mon.Shutdown()
			select {
			case <-mon.Done():
			case <-time.After(2 * time.Second):
				t.Fatal("Monitor did not shut down in time")
			}
			mon.Close()

			if tt.expectData {
				data, err := os.ReadFile(csvFile)
				if err != nil {
					t.Fatalf("Failed to read CSV file: %v", err)
				}
				if len(data) == 0 {
					t.Error("CSV file is empty")
				}
			}
		})
	}
}

func TestMonitor_Integration_GracefulShutdown(t *testing.T) {
	mockServer := NewMockOBSServer()
	defer mockServer.Close()

	tmpDir := t.TempDir()
	csvFile := filepath.Join(tmpDir, "test-metrics.csv")

	host := strings.Replace(mockServer.URL(), "ws://", "", 1)

	connInfo := monitor.ObsConnectionInfo{
		Password:       "",
		Host:           host,
		CSVFile:        csvFile,
		MetricInterval: 50,
		WriterInterval: 100,
	}

	mon, err := monitor.NewMonitor(connInfo)
	if err != nil {
		t.Fatalf("Failed to create monitor: %v", err)
	}

	if err := mon.Start(); err != nil {
		t.Fatalf("Failed to start monitor: %v", err)
	}

	time.Sleep(300 * time.Millisecond)

	mon.Shutdown()

	done := make(chan struct{})
	go func() {
		<-mon.Done()
		close(done)
	}()

	select {
	case <-done:
	case <-time.After(3 * time.Second):
		t.Fatal("Monitor did not shut down gracefully within timeout")
	}

	mon.Close()
}

func TestMonitor_Integration_ContextCancellation(t *testing.T) {
	mockServer := NewMockOBSServer()
	defer mockServer.Close()

	tmpDir := t.TempDir()
	csvFile := filepath.Join(tmpDir, "test-metrics.csv")

	host := strings.Replace(mockServer.URL(), "ws://", "", 1)

	connInfo := monitor.ObsConnectionInfo{
		Password:       "",
		Host:           host,
		CSVFile:        csvFile,
		MetricInterval: 50,
		WriterInterval: 100,
	}

	mon, err := monitor.NewMonitor(connInfo)
	if err != nil {
		t.Fatalf("Failed to create monitor: %v", err)
	}

	if err := mon.Start(); err != nil {
		t.Fatalf("Failed to start monitor: %v", err)
	}

	time.Sleep(200 * time.Millisecond)

	mon.Shutdown()

	select {
	case <-mon.Done():
	case <-time.After(2 * time.Second):
		t.Fatal("Monitor did not respond to context cancellation")
	}

	mon.Close()
}

func TestMonitor_Integration_OBSDisconnection(t *testing.T) {
	t.Skip("Skipping OBS disconnection test - requires fix in goobs client disconnect detection")
	mockServer := NewMockOBSServer()
	defer mockServer.Close()

	tmpDir := t.TempDir()
	csvFile := filepath.Join(tmpDir, "test-metrics.csv")

	host := strings.Replace(mockServer.URL(), "ws://", "", 1)

	connInfo := monitor.ObsConnectionInfo{
		Password:       "",
		Host:           host,
		CSVFile:        csvFile,
		MetricInterval: 50,
		WriterInterval: 100,
	}

	mon, err := monitor.NewMonitor(connInfo)
	if err != nil {
		t.Fatalf("Failed to create monitor: %v", err)
	}

	if err := mon.Start(); err != nil {
		t.Fatalf("Failed to start monitor: %v", err)
	}

	time.Sleep(200 * time.Millisecond)

	mockServer.DisconnectClients()

	select {
	case <-mon.Done():
	case <-time.After(5 * time.Second):
		t.Fatal("Monitor did not detect disconnection")
	}

	mon.Close()
}

func TestMonitor_Integration_ConcurrentMetricCollection(t *testing.T) {
	mockServer := NewMockOBSServer()
	defer mockServer.Close()

	tmpDir := t.TempDir()
	csvFile := filepath.Join(tmpDir, "test-metrics.csv")

	host := strings.Replace(mockServer.URL(), "ws://", "", 1)

	connInfo := monitor.ObsConnectionInfo{
		Password:       "",
		Host:           host,
		CSVFile:        csvFile,
		MetricInterval: 50,
		WriterInterval: 100,
	}

	mon, err := monitor.NewMonitor(connInfo)
	if err != nil {
		t.Fatalf("Failed to create monitor: %v", err)
	}

	if err := mon.Start(); err != nil {
		t.Fatalf("Failed to start monitor: %v", err)
	}

	done := make(chan struct{})
	go func() {
		defer close(done)
		for i := 0; i < 10; i++ {
			mockServer.IncrementStreamMetrics(100000, 1, 100)
			mockServer.SetStats(float64(10+i), float64(100+i*10))
			time.Sleep(50 * time.Millisecond)
		}
	}()

	<-done

	time.Sleep(200 * time.Millisecond)

	mon.Shutdown()
	select {
	case <-mon.Done():
	case <-time.After(2 * time.Second):
		t.Fatal("Monitor did not shut down in time")
	}
	mon.Close()

	data, err := os.ReadFile(csvFile)
	if err != nil {
		t.Fatalf("Failed to read CSV file: %v", err)
	}

	lines := strings.Split(string(data), "\n")
	if len(lines) < 3 {
		t.Errorf("Expected at least 3 lines (header + data), got %d", len(lines))
	}
}

func TestMonitor_Integration_StreamStateChanges(t *testing.T) {
	mockServer := NewMockOBSServer()
	defer mockServer.Close()

	tmpDir := t.TempDir()
	csvFile := filepath.Join(tmpDir, "test-metrics.csv")

	host := strings.Replace(mockServer.URL(), "ws://", "", 1)

	connInfo := monitor.ObsConnectionInfo{
		Password:       "",
		Host:           host,
		CSVFile:        csvFile,
		MetricInterval: 50,
		WriterInterval: 100,
	}

	mon, err := monitor.NewMonitor(connInfo)
	if err != nil {
		t.Fatalf("Failed to create monitor: %v", err)
	}

	if err := mon.Start(); err != nil {
		t.Fatalf("Failed to start monitor: %v", err)
	}

	time.Sleep(150 * time.Millisecond)
	mockServer.SetStreamActive(true)

	time.Sleep(150 * time.Millisecond)
	mockServer.SetStreamActive(false)

	time.Sleep(150 * time.Millisecond)

	mon.Shutdown()
	select {
	case <-mon.Done():
	case <-time.After(2 * time.Second):
		t.Fatal("Monitor did not shut down in time")
	}
	mon.Close()

	data, err := os.ReadFile(csvFile)
	if err != nil {
		t.Fatalf("Failed to read CSV file: %v", err)
	}

	if len(data) == 0 {
		t.Error("CSV file is empty")
	}
}

func TestMonitor_Integration_NoCSVWriter(t *testing.T) {
	mockServer := NewMockOBSServer()
	defer mockServer.Close()

	host := strings.Replace(mockServer.URL(), "ws://", "", 1)

	connInfo := monitor.ObsConnectionInfo{
		Password:       "",
		Host:           host,
		CSVFile:        "",
		MetricInterval: 50,
		WriterInterval: 100,
	}

	mon, err := monitor.NewMonitor(connInfo)
	if err != nil {
		t.Fatalf("Failed to create monitor: %v", err)
	}

	if err := mon.Start(); err != nil {
		t.Fatalf("Failed to start monitor: %v", err)
	}

	time.Sleep(300 * time.Millisecond)

	mon.Shutdown()
	select {
	case <-mon.Done():
	case <-time.After(2 * time.Second):
		t.Fatal("Monitor did not shut down in time")
	}
	mon.Close()
}

func TestMonitor_Integration_ConnectionFailure(t *testing.T) {
	connInfo := monitor.ObsConnectionInfo{
		Password:       "",
		Host:           "localhost:9999",
		CSVFile:        "",
		MetricInterval: 50,
		WriterInterval: 100,
	}

	mon, err := monitor.NewMonitor(connInfo)
	if err != nil {
		t.Fatalf("Failed to create monitor: %v", err)
	}

	err = mon.Start()
	if err == nil {
		t.Fatal("Expected error when connecting to invalid host, got nil")
		mon.Close()
	}
}

func TestMonitor_Integration_FullMonitoringCycle(t *testing.T) {
	mockServer := NewMockOBSServer()
	defer mockServer.Close()

	mockServer.SetStreamActive(true)
	mockServer.SetStats(15.5, 256.0)
	mockServer.IncrementStreamMetrics(1000000, 5, 500)

	tmpDir := t.TempDir()
	csvFile := filepath.Join(tmpDir, "full-cycle-metrics.csv")

	host := strings.Replace(mockServer.URL(), "ws://", "", 1)

	connInfo := monitor.ObsConnectionInfo{
		Password:       "",
		Host:           host,
		CSVFile:        csvFile,
		MetricInterval: 100,
		WriterInterval: 300,
	}

	mon, err := monitor.NewMonitor(connInfo)
	if err != nil {
		t.Fatalf("Failed to create monitor: %v", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := mon.Start(); err != nil {
		t.Fatalf("Failed to start monitor: %v", err)
	}

	go func() {
		ticker := time.NewTicker(100 * time.Millisecond)
		defer ticker.Stop()
		for i := 0; i < 10; i++ {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
				mockServer.IncrementStreamMetrics(50000, 0, 50)
				mockServer.SetStats(float64(15+i), float64(256+i*5))
			}
		}
	}()

	time.Sleep(1 * time.Second)

	mon.Shutdown()
	select {
	case <-mon.Done():
	case <-time.After(2 * time.Second):
		t.Fatal("Monitor did not shut down in time")
	}
	mon.Close()

	data, err := os.ReadFile(csvFile)
	if err != nil {
		t.Fatalf("Failed to read CSV file: %v", err)
	}

	lines := strings.Split(strings.TrimSpace(string(data)), "\n")
	if len(lines) < 2 {
		t.Errorf("Expected at least 2 lines in CSV (header + data), got %d", len(lines))
	}

	header := lines[1]
	expectedColumns := []string{
		"timestamp",
		"obs_rtt_ms",
		"google_rtt_ms",
		"stream_active",
		"output_bytes",
		"output_skipped_frames",
		"output_frames",
		"obs_cpu",
		"obs_memory",
		"system_cpu",
		"system_memory",
	}

	for _, col := range expectedColumns {
		if !strings.Contains(header, col) {
			t.Errorf("Expected header to contain '%s', got: %s", col, header)
		}
	}
}
