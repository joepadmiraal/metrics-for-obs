# OBS Monitor

A utility that monitors an OBS Studio instance via a WebSocket connection. Measures stream metrics, network latency and generic performance metrics.

It connects to OBS via a WebSocket connection.
All generic metrics are collected from the machine that runs OBS Monitor so it makes sense to run this on the same machine as OBS itself.

The metric collection interval is configurable.
This enables very precise metrics which is usefull for troubleshooting low latency video streams.
It's possible that within one write window, multiple measurements are collected.
In which case it will take the max value of the measurements.

## Usage

```bash
obs-monitor -password <websocket password>
```

Example output

```bash
Press Ctrl-C to exit
Pinging google.com every 1s
Pinging a.rtmp.youtube.com every every 1s
OBS Studio version: 32.0.4
Server protocol version: 5.6.3
Client protocol version: 5.5.6
Client library version: 1.5.6

timestamp                 | obs_rtt_ms | google_rtt_ms | stream_active | output_bytes | output_skipped_frames | output_frames | obs_cpu_% | obs_mem_mb | sys_cpu_% | sys_mem_% | errors
--------------------------|------------|---------------|---------------|--------------|-----------------------|---------------|-----------|------------|-----------|-----------|--------
2025-12-23T15:01:21+01:00 |       4.74 |         12.38 |         false |            0 |                     0 |             0 |       2.8 |        400 |      18.1 |      71.6 | 
2025-12-23T15:01:22+01:00 |       4.31 |          4.98 |         false |            0 |                     0 |             0 |       3.1 |        397 |      15.8 |      74.6 | 
2025-12-23T15:01:23+01:00 |       3.91 |          5.13 |         false |            0 |                     0 |             0 |       3.3 |        398 |      11.7 |      74.7 | 
2025-12-23T15:01:24+01:00 |       3.88 |          4.45 |          true |            0 |                     0 |             0 |       3.8 |        418 |      12.4 |      73.4 | 
2025-12-23T15:01:25+01:00 |       4.31 |          6.08 |          true |       327347 |                     0 |            28 |       3.9 |        419 |      13.6 |      71.5 | 
2025-12-23T15:01:26+01:00 |       4.89 |          9.36 |          true |       330688 |                     0 |            30 |       3.6 |        419 |      13.2 |      71.6 | 
2025-12-23T15:01:27+01:00 |       4.89 |          4.19 |          true |       792085 |                     0 |            30 |       3.4 |        420 |      12.3 |      72.9 | 
2025-12-23T15:01:28+01:00 |       4.13 |          5.09 |          true |       694144 |                     0 |            30 |       3.4 |        420 |      13.6 |      71.4 | 
2025-12-23T15:01:29+01:00 |       4.33 |          4.30 |          true |       549395 |                     0 |            30 |       3.4 |        420 |      12.7 |      72.8 | 
```

### Flags

- `-password` (optional): OBS WebSocket password, the program will ask for it if it's not provided
- `-host` (optional): OBS WebSocket host (default: localhost)
- `-port` (optional): OBS WebSocket port (default: 4455)
- `-csv` (optional): CSV file to write metrics to, set to empty to prevent csv file generation (default: obs-monitor.csv)
- `-metric-interval` (optional): Metric collection interval in milliseconds (default: 1000ms)
- `-writer-interval` (optional): Writer interval in milliseconds (default: 1000ms)

## CSV Export

The monitor will write one line per second to the CSV file containing:

- `timestamp`: ISO 8601 timestamp
- `obs_rtt_ms`: Round-trip time to the streaming server in milliseconds
- `google_rtt_ms`: Round-trip time to Google in milliseconds
- `stream_active`: Whether the stream is currently active
- `output_bytes`: Total bytes sent to the streaming server during the writer-interval
- `output_skipped_frames`: Number of frames skipped in the output process during the writer-interval
- `output_frames`: Total number of frames rendered in the output process during the writer-interval
- `obs_cpu_percent`: CPU usage of the OBS process in percent
- `obs_memory_mb`: Memory usage of the OBS process in MB
- `system_cpu_percent`: Overall system CPU usage in percent
- `system_memory_percent`: Overall system memory usage in percent
- `errors`: Semicolon-separated list of any errors that occurred during metric collection

Example:
```bash
obs-monitor -password mypassword -csv metrics.csv
```

## OBS

The WebSocket password can be set and read from `Tools->WebSocket Server Settings`.

![OBS WebSocket Server Settings 1](docs/obs-1.png)
![OBS WebSocket Server Settings 2](docs/obs-2.png)

## Development

- OBS WebSocket client: https://github.com/andreykaipov/goobs
- Golang project setup: https://github.com/golang-standards/project-layout

Tests can be run via `go test ./...`
