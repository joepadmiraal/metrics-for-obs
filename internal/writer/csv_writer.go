package writer

import (
	"encoding/csv"
	"fmt"
	"os"
	"runtime"
	"sync"
	"time"
)

// CSVWriter handles writing metrics to a CSV file
type CSVWriter struct {
	file   *os.File
	writer *csv.Writer
	mu     sync.Mutex
}

// NewCSVWriter creates a new CSV writer and writes the header
func NewCSVWriter(filename, obsVersion, streamDomain string) (*CSVWriter, error) {
	file, err := os.Create(filename)
	if err != nil {
		return nil, fmt.Errorf("failed to create CSV file: %w", err)
	}

	writer := csv.NewWriter(file)

	// Write header information
	headerInfo := []string{
		fmt.Sprintf("OBS Studio version: %s", obsVersion),
		fmt.Sprintf("Stream domain: %s", streamDomain),
		fmt.Sprintf("OS: %s", runtime.GOOS),
	}
	if err := writer.Write(headerInfo); err != nil {
		file.Close()
		return nil, fmt.Errorf("failed to write CSV header info: %w", err)
	}

	// Write column header
	header := []string{
		"timestamp",
		"obs_rtt_ms",
		"google_rtt_ms",
		"stream_active",
		"output_bytes",
		"output_skipped_frames",
		"output_frames",
		"obs_cpu_percent",
		"obs_memory_mb",
		"system_cpu_percent",
		"system_memory_percent",
		"errors",
	}
	if err := writer.Write(header); err != nil {
		file.Close()
		return nil, fmt.Errorf("failed to write CSV header: %w", err)
	}
	writer.Flush()

	return &CSVWriter{
		file:   file,
		writer: writer,
	}, nil
}

// WriteMetrics writes a single metrics data row to the CSV file
func (cw *CSVWriter) WriteMetrics(data MetricsData) error {
	cw.mu.Lock()
	defer cw.mu.Unlock()

	obsRttMs := ""
	if data.ObsPingError == nil && data.ObsRTT > 0 {
		obsRttMs = fmt.Sprintf("%.2f", float64(data.ObsRTT.Microseconds())/1000.0)
	}

	googleRttMs := ""
	if data.GooglePingError == nil && data.GoogleRTT > 0 {
		googleRttMs = fmt.Sprintf("%.2f", float64(data.GoogleRTT.Microseconds())/1000.0)
	}

	errors := ""
	if data.ObsPingError != nil {
		errors = fmt.Sprintf("obs_ping: %v", data.ObsPingError)
	}
	if data.GooglePingError != nil {
		if errors != "" {
			errors += "; "
		}
		errors += fmt.Sprintf("google_ping: %v", data.GooglePingError)
	}
	if data.StreamError != nil {
		if errors != "" {
			errors += "; "
		}
		errors += fmt.Sprintf("stream: %v", data.StreamError)
	}
	if data.ObsStatsError != nil {
		if errors != "" {
			errors += "; "
		}
		errors += fmt.Sprintf("obs_stats: %v", data.ObsStatsError)
	}
	if data.SystemMetricsError != nil {
		if errors != "" {
			errors += "; "
		}
		errors += fmt.Sprintf("system: %v", data.SystemMetricsError)
	}

	row := []string{
		data.Timestamp.Format(time.RFC3339),
		obsRttMs,
		googleRttMs,
		fmt.Sprintf("%t", data.StreamActive),
		fmt.Sprintf("%.0f", data.OutputBytes),
		fmt.Sprintf("%.0f", data.OutputSkippedFrames),
		fmt.Sprintf("%.0f", data.OutputFrames),
		fmt.Sprintf("%.2f", data.ObsCpuUsage),
		fmt.Sprintf("%.2f", data.ObsMemoryUsage),
		fmt.Sprintf("%.2f", data.SystemCpuUsage),
		fmt.Sprintf("%.2f", data.SystemMemoryUsage),
		errors,
	}

	if err := cw.writer.Write(row); err != nil {
		return fmt.Errorf("failed to write CSV row: %w", err)
	}
	cw.writer.Flush()

	return cw.writer.Error()
}

// Close closes the CSV file
func (cw *CSVWriter) Close() error {
	cw.mu.Lock()
	defer cw.mu.Unlock()
	cw.writer.Flush()
	return cw.file.Close()
}
