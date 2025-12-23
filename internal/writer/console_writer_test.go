package writer

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"strings"
	"testing"
	"time"
)

func TestConsoleWriter_WriteMetrics_HighRTT(t *testing.T) {
	// Capture stdout
	old := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	cw := NewConsoleWriter()

	// Test with 1001ms RTT (1001000 microseconds)
	data := MetricsData{
		Timestamp:           time.Date(2025, 12, 16, 10, 0, 0, 0, time.UTC),
		ObsRTT:              1001 * time.Millisecond,
		ObsPingError:        nil,
		GoogleRTT:           25 * time.Millisecond,
		GooglePingError:     nil,
		StreamActive:        true,
		OutputBytes:         123456.0,
		OutputSkippedFrames: 10.0,
		StreamError:         nil,
	}

	err := cw.WriteMetrics(data)
	if err != nil {
		t.Fatalf("WriteMetrics failed: %v", err)
	}

	// Close writer and restore stdout
	w.Close()
	os.Stdout = old

	// Read captured output
	var buf bytes.Buffer
	io.Copy(&buf, r)
	output := buf.String()

	lines := strings.Split(strings.TrimSpace(output), "\n")
	if len(lines) < 3 {
		t.Fatalf("Expected at least 3 lines (header, separator, data), got %d", len(lines))
	}

	// Check that the data line has properly formatted RTT
	dataLine := lines[2]

	// The RTT value should be formatted as a 6-character string (right-aligned)
	// For 1001.00ms, it should appear in the output
	if !strings.Contains(dataLine, "1001.00") {
		t.Errorf("Expected RTT value '1001.00' in output, got: %s", dataLine)
	}

	// Verify the columns are aligned by checking pipe positions
	headerLine := lines[0]

	// All lines should have pipes at the same positions
	headerPipes := findPipePositions(headerLine)
	dataPipes := findPipePositions(dataLine)

	if len(headerPipes) != len(dataPipes) {
		t.Errorf("Number of columns mismatch: header has %d pipes, data has %d pipes", len(headerPipes), len(dataPipes))
	}

	for i := range headerPipes {
		if i < len(dataPipes) && headerPipes[i] != dataPipes[i] {
			t.Errorf("Column %d misaligned: header pipe at %d, data pipe at %d", i, headerPipes[i], dataPipes[i])
			t.Logf("Header: %s", headerLine)
			t.Logf("Data:   %s", dataLine)
		}
	}
}

func findPipePositions(line string) []int {
	positions := []int{}
	for i, ch := range line {
		if ch == '|' {
			positions = append(positions, i)
		}
	}
	return positions
}

func TestConsoleWriter_WriteMetrics_NormalRTT(t *testing.T) {
	old := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	cw := NewConsoleWriter()

	data := MetricsData{
		Timestamp:           time.Date(2025, 12, 23, 10, 0, 0, 0, time.UTC),
		ObsRTT:              50 * time.Millisecond,
		GoogleRTT:           25 * time.Millisecond,
		StreamActive:        true,
		OutputBytes:         1024.0,
		OutputSkippedFrames: 5.0,
	}

	err := cw.WriteMetrics(data)
	if err != nil {
		t.Fatalf("WriteMetrics failed: %v", err)
	}

	w.Close()
	os.Stdout = old

	var buf bytes.Buffer
	io.Copy(&buf, r)
	output := buf.String()

	if !strings.Contains(output, "50.00") {
		t.Errorf("Expected RTT value '50.00' in output, got: %s", output)
	}
}

func TestConsoleWriter_WriteMetrics_ZeroValues(t *testing.T) {
	old := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	cw := NewConsoleWriter()

	data := MetricsData{
		Timestamp:           time.Date(2025, 12, 23, 10, 0, 0, 0, time.UTC),
		ObsRTT:              0,
		GoogleRTT:           0,
		StreamActive:        false,
		OutputBytes:         0,
		OutputSkippedFrames: 0,
	}

	err := cw.WriteMetrics(data)
	if err != nil {
		t.Fatalf("WriteMetrics failed: %v", err)
	}

	w.Close()
	os.Stdout = old

	var buf bytes.Buffer
	io.Copy(&buf, r)
	output := buf.String()

	lines := strings.Split(strings.TrimSpace(output), "\n")
	if len(lines) < 3 {
		t.Fatalf("Expected at least 3 lines, got %d", len(lines))
	}
}

func TestConsoleWriter_WriteMetrics_WithErrors(t *testing.T) {
	old := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	cw := NewConsoleWriter()

	obsErr := fmt.Errorf("obs ping error")
	googleErr := fmt.Errorf("google ping error")

	data := MetricsData{
		Timestamp:       time.Date(2025, 12, 23, 10, 0, 0, 0, time.UTC),
		ObsRTT:          0,
		ObsPingError:    obsErr,
		GoogleRTT:       0,
		GooglePingError: googleErr,
		StreamActive:    false,
	}

	err := cw.WriteMetrics(data)
	if err != nil {
		t.Fatalf("WriteMetrics failed: %v", err)
	}

	w.Close()
	os.Stdout = old

	var buf bytes.Buffer
	io.Copy(&buf, r)
	output := buf.String()

	if len(output) == 0 {
		t.Error("Expected some output even with errors")
	}
}

func TestConsoleWriter_WriteMetrics_InactiveStream(t *testing.T) {
	old := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	cw := NewConsoleWriter()

	data := MetricsData{
		Timestamp:    time.Date(2025, 12, 23, 10, 0, 0, 0, time.UTC),
		ObsRTT:       50 * time.Millisecond,
		GoogleRTT:    25 * time.Millisecond,
		StreamActive: false,
	}

	err := cw.WriteMetrics(data)
	if err != nil {
		t.Fatalf("WriteMetrics failed: %v", err)
	}

	w.Close()
	os.Stdout = old

	var buf bytes.Buffer
	io.Copy(&buf, r)
	output := buf.String()

	if len(output) == 0 {
		t.Error("Expected output for inactive stream")
	}
}

func TestConsoleWriter_WriteMetrics_FormattingConsistency(t *testing.T) {
	old := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	cw := NewConsoleWriter()

	testCases := []MetricsData{
		{
			Timestamp:    time.Date(2025, 12, 23, 10, 0, 0, 0, time.UTC),
			ObsRTT:       50 * time.Millisecond,
			GoogleRTT:    25 * time.Millisecond,
			StreamActive: true,
		},
		{
			Timestamp:    time.Date(2025, 12, 23, 10, 1, 0, 0, time.UTC),
			ObsRTT:       100 * time.Millisecond,
			GoogleRTT:    50 * time.Millisecond,
			StreamActive: true,
		},
		{
			Timestamp:    time.Date(2025, 12, 23, 10, 2, 0, 0, time.UTC),
			ObsRTT:       75 * time.Millisecond,
			GoogleRTT:    30 * time.Millisecond,
			StreamActive: false,
		},
	}

	for _, data := range testCases {
		err := cw.WriteMetrics(data)
		if err != nil {
			t.Fatalf("WriteMetrics failed: %v", err)
		}
	}

	w.Close()
	os.Stdout = old

	var buf bytes.Buffer
	io.Copy(&buf, r)
	output := buf.String()

	if len(output) == 0 {
		t.Error("Expected output from multiple WriteMetrics calls")
	}

	lines := strings.Split(output, "\n")
	if len(lines) < 3 {
		t.Error("Expected multiple output lines")
	}
}

func TestConsoleWriter_NewConsoleWriter(t *testing.T) {
	cw := NewConsoleWriter()

	if cw == nil {
		t.Fatal("NewConsoleWriter returned nil")
	}
}

func TestConsoleWriter_WriteMetrics_AllMetrics(t *testing.T) {
	old := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	cw := NewConsoleWriter()

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

	err := cw.WriteMetrics(data)
	if err != nil {
		t.Fatalf("WriteMetrics failed: %v", err)
	}

	w.Close()
	os.Stdout = old

	var buf bytes.Buffer
	io.Copy(&buf, r)
	output := buf.String()

	lines := strings.Split(strings.TrimSpace(output), "\n")
	if len(lines) < 3 {
		t.Fatalf("Expected at least 3 lines, got %d", len(lines))
	}

	dataLine := lines[2]
	if !strings.Contains(dataLine, "50.00") {
		t.Error("Expected to find OBS RTT in output")
	}
}
