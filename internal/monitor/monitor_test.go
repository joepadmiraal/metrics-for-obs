package monitor

import (
	"testing"
	"time"
)

func TestExtractDomain_FullRTMPURL(t *testing.T) {
	tests := []struct {
		name     string
		url      string
		expected string
	}{
		{
			name:     "RTMP URL with path",
			url:      "rtmp://live.twitch.tv/app",
			expected: "live.twitch.tv",
		},
		{
			name:     "RTMPS URL",
			url:      "rtmps://live.twitch.tv/app",
			expected: "live.twitch.tv",
		},
		{
			name:     "RTMP URL with port",
			url:      "rtmp://live.twitch.tv:1935/app",
			expected: "live.twitch.tv",
		},
		{
			name:     "RTMP URL simple",
			url:      "rtmp://example.com",
			expected: "example.com",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			domain, err := extractDomain(tt.url)
			if err != nil {
				t.Fatalf("extractDomain failed: %v", err)
			}
			if domain != tt.expected {
				t.Errorf("Expected domain %s, got %s", tt.expected, domain)
			}
		})
	}
}

func TestExtractDomain_URLWithoutProtocol(t *testing.T) {
	tests := []struct {
		name     string
		url      string
		expected string
	}{
		{
			name:     "Domain with path",
			url:      "live.twitch.tv/app",
			expected: "live.twitch.tv",
		},
		{
			name:     "Domain with port and path",
			url:      "live.twitch.tv:1935/app",
			expected: "live.twitch.tv",
		},
		{
			name:     "Simple domain",
			url:      "example.com",
			expected: "example.com",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			domain, err := extractDomain(tt.url)
			if err != nil {
				t.Fatalf("extractDomain failed: %v", err)
			}
			if domain != tt.expected {
				t.Errorf("Expected domain %s, got %s", tt.expected, domain)
			}
		})
	}
}

func TestExtractDomain_InvalidURLs(t *testing.T) {
	tests := []struct {
		name string
		url  string
	}{
		{
			name: "Empty string",
			url:  "",
		},
		{
			name: "Just protocol",
			url:  "rtmp://",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := extractDomain(tt.url)
			if err == nil {
				t.Error("Expected error for invalid URL")
			}
		})
	}
}

func TestNewMonitor_Initialization(t *testing.T) {
	connInfo := ObsConnectionInfo{
		Password:       "test-password",
		Host:           "localhost:4455",
		CSVFile:        "test.csv",
		MetricInterval: 1000,
		WriterInterval: 5000,
	}

	monitor, err := NewMonitor(connInfo)

	if err != nil {
		t.Fatalf("NewMonitor failed: %v", err)
	}
	if monitor == nil {
		t.Fatal("NewMonitor returned nil")
	}
	if monitor.connectionInfo.Password != "test-password" {
		t.Error("Password not set correctly")
	}
	if monitor.metricInterval != 1000*time.Millisecond {
		t.Errorf("Expected metricInterval 1000ms, got %v", monitor.metricInterval)
	}
	if monitor.writerInterval != 5000*time.Millisecond {
		t.Errorf("Expected writerInterval 5000ms, got %v", monitor.writerInterval)
	}
}

func TestMonitor_Shutdown(t *testing.T) {
	connInfo := ObsConnectionInfo{
		Password:       "test",
		Host:           "localhost:4455",
		MetricInterval: 1000,
		WriterInterval: 5000,
	}

	monitor, err := NewMonitor(connInfo)
	if err != nil {
		t.Fatalf("NewMonitor failed: %v", err)
	}

	monitor.Shutdown()

	select {
	case <-monitor.ctx.Done():
	case <-time.After(100 * time.Millisecond):
		t.Error("Context not cancelled after Shutdown")
	}
}
