package monitor

import (
	"fmt"
	"net/url"
	"strings"
	"time"

	"github.com/andreykaipov/goobs"
	"github.com/joepadmiraal/obs-monitor/internal/metric"
	"github.com/joepadmiraal/obs-monitor/internal/writer"
)

type ObsConnectionInfo struct {
	Password string
	Host     string
	CSVFile  string
}

type Monitor struct {
	client         *goobs.Client
	connectionInfo ObsConnectionInfo
	obsPinger      *metric.Pinger
	googlePinger   *metric.Pinger
	streamMetrics  *metric.StreamMetrics
	csvWriter      *writer.CSVWriter
	consoleWriter  *writer.ConsoleWriter
}

// NewMonitor Connects to OBS and
func NewMonitor(connectionInfo ObsConnectionInfo) (*Monitor, error) {

	return &Monitor{
		connectionInfo: connectionInfo,
	}, nil
}

// connect establishes a connection to OBS (internal use only)
func (m *Monitor) connect() error {
	var err error
	m.client, err = goobs.New(m.connectionInfo.Host, goobs.WithPassword(m.connectionInfo.Password))
	if err != nil {
		return err
	}
	return nil
}

// Start connects to OBS and starts all monitoring components
func (m *Monitor) Start() error {
	// Connect to OBS
	if err := m.connect(); err != nil {
		return fmt.Errorf("failed to connect to OBS: %w", err)
	}

	// Get OBS stream server domain
	streamSettings, err := m.client.Config.GetStreamServiceSettings()
	if err != nil {
		return fmt.Errorf("failed to get stream settings: %w", err)
	}

	serverURL := streamSettings.StreamServiceSettings.Server
	if serverURL == "" {
		return fmt.Errorf("stream server URL not found in settings")
	}

	obsDomain, err := extractDomain(serverURL)
	if err != nil {
		return fmt.Errorf("failed to extract domain from URL: %w", err)
	}

	if err := m.initializePingers(obsDomain); err != nil {
		return err
	}

	// Initialize stream metrics
	m.streamMetrics, err = metric.NewStreamMetrics(m.client)
	if err != nil {
		return fmt.Errorf("failed to initialize stream metrics: %w", err)
	}

	// Initialize CSV writer if filename is provided
	if m.connectionInfo.CSVFile != "" {
		m.csvWriter, err = writer.NewCSVWriter(m.connectionInfo.CSVFile)
		if err != nil {
			return fmt.Errorf("failed to initialize CSV writer: %w", err)
		}
		fmt.Printf("Writing metrics to CSV file: %s\n", m.connectionInfo.CSVFile)
	}

	// Initialize console writer
	m.consoleWriter = writer.NewConsoleWriter()

	m.PrintInfo()

	// Start stream metrics monitoring in a goroutine
	go func() {
		if err := m.streamMetrics.Start(); err != nil {
			fmt.Printf("Stream metrics error: %v\n", err)
		}
	}()

	// Start metrics collector
	go m.collectAndWriteMetrics()

	return nil
}

func (m *Monitor) initializePingers(obsDomain string) error {
	var err error

	m.obsPinger, err = metric.NewPinger(obsDomain)
	if err != nil {
		return fmt.Errorf("failed to initialize OBS pinger: %w", err)
	}

	m.googlePinger, err = metric.NewPinger("google.com")
	if err != nil {
		return fmt.Errorf("failed to initialize Google pinger: %w", err)
	}

	go func() {
		if err := m.obsPinger.Start(); err != nil {
			fmt.Printf("OBS pinger error: %v\n", err)
		}
	}()

	go func() {
		if err := m.googlePinger.Start(); err != nil {
			fmt.Printf("Google pinger error: %v\n", err)
		}
	}()

	return nil
}

func (m *Monitor) PrintInfo() {
	version, err := m.client.General.GetVersion()
	if err != nil {
		panic(err)
	}

	fmt.Printf("OBS Studio version: %s\n", version.ObsVersion)
	fmt.Printf("Server protocol version: %s\n", version.ObsWebSocketVersion)
	fmt.Printf("Client protocol version: %s\n", goobs.ProtocolVersion)
	fmt.Printf("Client library version: %s\n", goobs.LibraryVersion)
}

func (m *Monitor) Close() {
	if m.csvWriter != nil {
		if err := m.csvWriter.Close(); err != nil {
			fmt.Printf("Error closing CSV writer: %v\n", err)
		}
	}
	m.client.Disconnect()
}

// collectAndWriteMetrics collects metrics from both pingers and stream metrics and writes to CSV
func (m *Monitor) collectAndWriteMetrics() {
	var lastObsPingMetrics metric.PingMetrics
	var lastGooglePingMetrics metric.PingMetrics
	var lastStreamMetrics metric.StreamMetricsData
	var haveObsPing, haveGooglePing, haveStream bool

	obsPingChan := m.obsPinger.GetMetricsChan()
	googlePingChan := m.googlePinger.GetMetricsChan()
	streamChan := m.streamMetrics.GetMetricsChan()
	ticker := time.NewTicker(1 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case obsPingMetrics := <-obsPingChan:
			lastObsPingMetrics = obsPingMetrics
			haveObsPing = true

		case googlePingMetrics := <-googlePingChan:
			lastGooglePingMetrics = googlePingMetrics
			haveGooglePing = true

		case streamMetrics := <-streamChan:
			lastStreamMetrics = streamMetrics
			haveStream = true

		case <-ticker.C:
			// Write metrics once per second with the latest data from all sources
			if haveObsPing || haveGooglePing || haveStream {
				m.writeMetrics(lastObsPingMetrics, lastGooglePingMetrics, lastStreamMetrics)
			}
		}
	}
}

// writeMetrics writes a combined metrics row to CSV and console
func (m *Monitor) writeMetrics(obsPingMetrics metric.PingMetrics, googlePingMetrics metric.PingMetrics, streamMetrics metric.StreamMetricsData) {
	// Use the more recent timestamp
	timestamp := obsPingMetrics.Timestamp
	if googlePingMetrics.Timestamp.After(timestamp) {
		timestamp = googlePingMetrics.Timestamp
	}
	if streamMetrics.Timestamp.After(timestamp) {
		timestamp = streamMetrics.Timestamp
	}

	data := writer.MetricsData{
		Timestamp:           timestamp,
		ObsRTT:              obsPingMetrics.RTT,
		ObsPingError:        obsPingMetrics.Error,
		GoogleRTT:           googlePingMetrics.RTT,
		GooglePingError:     googlePingMetrics.Error,
		StreamActive:        streamMetrics.Active,
		OutputBytes:         streamMetrics.OutputBytes,
		OutputSkippedFrames: streamMetrics.OutputSkippedFrames,
		StreamError:         streamMetrics.Error,
	}

	// Write to CSV if enabled
	if m.csvWriter != nil {
		if err := m.csvWriter.WriteMetrics(data); err != nil {
			fmt.Printf("Error writing to CSV: %v\n", err)
		}
	}

	// Write to console
	if err := m.consoleWriter.WriteMetrics(data); err != nil {
		fmt.Printf("Error writing to console: %v\n", err)
	}
}

func extractDomain(rawURL string) (string, error) {
	if !strings.Contains(rawURL, "://") {
		rawURL = "rtmp://" + rawURL
	}

	parsedURL, err := url.Parse(rawURL)
	if err != nil {
		return "", err
	}

	host := parsedURL.Hostname()
	if host == "" {
		return "", fmt.Errorf("no hostname found in URL")
	}

	return host, nil
}
