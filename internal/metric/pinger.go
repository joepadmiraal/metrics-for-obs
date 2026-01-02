package metric

import (
	"fmt"
	"runtime"
	"sync"
	"time"

	probing "github.com/prometheus-community/pro-bing"
)

type Pinger struct {
	domain               string
	maxRTT               time.Duration
	lastError            error
	measurementCount     int
	measurementsSinceGet int
	mu                   sync.Mutex
	interval             time.Duration
}

type PingMetrics struct {
	Timestamp time.Time
	RTT       time.Duration
	Error     error
}

func NewPinger(domain string, interval time.Duration) (*Pinger, error) {
	return &Pinger{
		domain:   domain,
		interval: interval,
	}, nil
}

func (p *Pinger) GetAndResetMaxRTT() (time.Duration, error) {
	p.mu.Lock()
	defer p.mu.Unlock()

	if p.measurementsSinceGet == 0 && p.measurementCount > 0 {
		return 0, fmt.Errorf("no new measurements collected since last read")
	}

	maxRTT := p.maxRTT
	err := p.lastError

	p.maxRTT = 0
	p.lastError = nil
	p.measurementsSinceGet = 0

	return maxRTT, err
}

func (p *Pinger) Start() error {
	fmt.Printf("Pinging %s every %v\n", p.domain, p.interval)

	ticker := time.NewTicker(p.interval)
	defer ticker.Stop()

	for range ticker.C {
		rtt, err := p.ping(p.domain)

		p.mu.Lock()
		if err != nil {
			p.lastError = err
		} else if rtt > p.maxRTT {
			p.maxRTT = rtt
		}
		p.measurementCount++
		p.measurementsSinceGet++
		p.mu.Unlock()
	}

	return nil
}

func (p *Pinger) ping(domain string) (time.Duration, error) {
	pinger, err := probing.NewPinger(domain)
	if err != nil {
		return 0, err
	}

	pinger.Count = 1
	pinger.Timeout = 1 * time.Second
	pinger.SetPrivileged(runtime.GOOS == "windows")

	err = pinger.Run()
	if err != nil {
		return 0, err
	}

	stats := pinger.Statistics()
	if stats.PacketsRecv == 0 {
		return 0, fmt.Errorf("no response received")
	}

	return stats.AvgRtt, nil
}
