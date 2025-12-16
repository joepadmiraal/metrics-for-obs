package metric

import (
	"fmt"
	"time"

	probing "github.com/prometheus-community/pro-bing"
)

type Pinger struct {
	domain      string
	metricsChan chan PingMetrics
}

type PingMetrics struct {
	Timestamp time.Time
	RTT       time.Duration
	Error     error
}

func NewPinger(domain string) (*Pinger, error) {
	return &Pinger{
		domain:      domain,
		metricsChan: make(chan PingMetrics, 10),
	}, nil
}

func (p *Pinger) GetMetricsChan() <-chan PingMetrics {
	return p.metricsChan
}

func (p *Pinger) Start() error {
	fmt.Printf("Pinging %s\n", p.domain)

	ticker := time.NewTicker(1 * time.Second)
	defer ticker.Stop()

	for range ticker.C {
		go func() {
			timestamp := time.Now()
			rtt, err := p.ping(p.domain)

			select {
			case p.metricsChan <- PingMetrics{
				Timestamp: timestamp,
				RTT:       rtt,
				Error:     err,
			}:
			default:
			}
		}()
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
	pinger.SetPrivileged(false)

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
