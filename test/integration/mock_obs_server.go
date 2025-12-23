package integration

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"

	"github.com/gorilla/websocket"
)

type MockOBSServer struct {
	server           *httptest.Server
	upgrader         websocket.Upgrader
	clients          []*websocket.Conn
	clientsMu        sync.Mutex
	streamActive     bool
	streamActiveMu   sync.RWMutex
	outputBytes      float64
	skippedFrames    float64
	totalFrames      float64
	statsMu          sync.RWMutex
	cpuUsage         float64
	memoryUsage      float64
	disconnectClient bool
	disconnectMu     sync.RWMutex
}

func NewMockOBSServer() *MockOBSServer {
	mock := &MockOBSServer{
		upgrader: websocket.Upgrader{
			CheckOrigin: func(r *http.Request) bool { return true },
		},
		streamActive: false,
		outputBytes:  0,
		cpuUsage:     10.5,
		memoryUsage:  256.0,
	}

	mock.server = httptest.NewServer(http.HandlerFunc(mock.handleWebSocket))
	return mock
}

func (m *MockOBSServer) URL() string {
	return strings.Replace(m.server.URL, "http://", "ws://", 1)
}

func (m *MockOBSServer) Close() {
	m.clientsMu.Lock()
	for _, client := range m.clients {
		client.Close()
	}
	m.clients = nil
	m.clientsMu.Unlock()
	m.server.Close()
}

func (m *MockOBSServer) SetStreamActive(active bool) {
	m.streamActiveMu.Lock()
	defer m.streamActiveMu.Unlock()
	m.streamActive = active
}

func (m *MockOBSServer) SetStats(cpu, memory float64) {
	m.statsMu.Lock()
	defer m.statsMu.Unlock()
	m.cpuUsage = cpu
	m.memoryUsage = memory
}

func (m *MockOBSServer) IncrementStreamMetrics(bytes, skipped, frames float64) {
	m.statsMu.Lock()
	defer m.statsMu.Unlock()
	m.outputBytes += bytes
	m.skippedFrames += skipped
	m.totalFrames += frames
}

func (m *MockOBSServer) DisconnectClients() {
	m.disconnectMu.Lock()
	m.disconnectClient = true
	m.disconnectMu.Unlock()

	m.clientsMu.Lock()
	defer m.clientsMu.Unlock()
	for _, client := range m.clients {
		client.Close()
	}
	m.clients = nil
}

func (m *MockOBSServer) handleWebSocket(w http.ResponseWriter, r *http.Request) {
	conn, err := m.upgrader.Upgrade(w, r, nil)
	if err != nil {
		return
	}

	m.clientsMu.Lock()
	m.clients = append(m.clients, conn)
	m.clientsMu.Unlock()

	m.sendHello(conn)

	for {
		m.disconnectMu.RLock()
		shouldDisconnect := m.disconnectClient
		m.disconnectMu.RUnlock()

		if shouldDisconnect {
			conn.Close()
			return
		}

		_, message, err := conn.ReadMessage()
		if err != nil {
			return
		}

		var request map[string]interface{}
		if err := json.Unmarshal(message, &request); err != nil {
			continue
		}

		m.handleRequest(conn, request)
	}
}

func (m *MockOBSServer) sendHello(conn *websocket.Conn) {
	hello := map[string]interface{}{
		"op": 0,
		"d": map[string]interface{}{
			"obsWebSocketVersion": "5.0.0",
			"rpcVersion":          1,
			"authentication":      map[string]interface{}{},
		},
	}
	conn.WriteJSON(hello)
}

func (m *MockOBSServer) handleRequest(conn *websocket.Conn, request map[string]interface{}) {
	op, ok := request["op"].(float64)
	if !ok {
		return
	}

	if op == 1 {
		m.sendIdentified(conn)
		return
	}

	if op != 6 {
		return
	}

	d, ok := request["d"].(map[string]interface{})
	if !ok {
		return
	}

	requestType, ok := d["requestType"].(string)
	if !ok {
		return
	}

	requestId, _ := d["requestId"].(string)

	var responseData map[string]interface{}

	switch requestType {
	case "GetVersion":
		responseData = m.getVersionResponse()
	case "GetStats":
		responseData = m.getStatsResponse()
	case "GetStreamStatus":
		responseData = m.getStreamStatusResponse()
	case "GetStreamServiceSettings":
		responseData = m.getStreamServiceSettingsResponse()
	}

	response := map[string]interface{}{
		"op": 7,
		"d": map[string]interface{}{
			"requestType":   requestType,
			"requestId":     requestId,
			"requestStatus": map[string]interface{}{"result": true, "code": 100},
			"responseData":  responseData,
		},
	}

	conn.WriteJSON(response)
}

func (m *MockOBSServer) sendIdentified(conn *websocket.Conn) {
	response := map[string]interface{}{
		"op": 2,
		"d": map[string]interface{}{
			"negotiatedRpcVersion": 1,
		},
	}
	conn.WriteJSON(response)
}

func (m *MockOBSServer) getVersionResponse() map[string]interface{} {
	return map[string]interface{}{
		"obsVersion":          "30.0.0",
		"obsWebSocketVersion": "5.0.0",
		"platform":            "linux",
		"platformDescription": "Linux",
	}
}

func (m *MockOBSServer) getStatsResponse() map[string]interface{} {
	m.statsMu.RLock()
	defer m.statsMu.RUnlock()

	return map[string]interface{}{
		"cpuUsage":                         m.cpuUsage,
		"memoryUsage":                      m.memoryUsage,
		"availableDiskSpace":               100000.0,
		"activeFps":                        60.0,
		"averageFrameRenderTime":           5.0,
		"renderSkippedFrames":              0.0,
		"renderTotalFrames":                3600.0,
		"outputSkippedFrames":              0.0,
		"outputTotalFrames":                3600.0,
		"webSocketSessionIncomingMessages": 100.0,
		"webSocketSessionOutgoingMessages": 100.0,
	}
}

func (m *MockOBSServer) getStreamStatusResponse() map[string]interface{} {
	m.streamActiveMu.RLock()
	active := m.streamActive
	m.streamActiveMu.RUnlock()

	m.statsMu.RLock()
	defer m.statsMu.RUnlock()

	return map[string]interface{}{
		"outputActive":        active,
		"outputReconnecting":  false,
		"outputTimecode":      "00:10:30",
		"outputDuration":      630000.0,
		"outputCongestion":    0.0,
		"outputBytes":         m.outputBytes,
		"outputSkippedFrames": m.skippedFrames,
		"outputTotalFrames":   m.totalFrames,
	}
}

func (m *MockOBSServer) getStreamServiceSettingsResponse() map[string]interface{} {
	return map[string]interface{}{
		"streamServiceType": "rtmp_common",
		"streamServiceSettings": map[string]interface{}{
			"server": "rtmp://test-ingest.example.com/app",
			"key":    "test-stream-key",
		},
	}
}

func (m *MockOBSServer) SendExitStartedEvent(conn *websocket.Conn) error {
	event := map[string]interface{}{
		"op": 5,
		"d": map[string]interface{}{
			"eventType":   "ExitStarted",
			"eventIntent": 1,
			"eventData":   map[string]interface{}{},
		},
	}
	return conn.WriteJSON(event)
}

func (m *MockOBSServer) ActiveClientCount() int {
	m.clientsMu.Lock()
	defer m.clientsMu.Unlock()
	count := 0
	for _, client := range m.clients {
		if client != nil {
			count++
		}
	}
	return count
}
