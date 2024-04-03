# speedtest-exporter 
speedtest-exporter is prometheus exporter for network speedtest using [speedtest-go](https://github.com/showwin/speedtest-go).

## Exported Metrics

|Metrics|Type|Description|
|:--|:--|:--|
|`speedtest_download_speed_mbps`|Gauge|Download rate of last test in Mbps format|
|`speedtest_upload_speed_mbps`|Gauge|Upload rate of last test in Mbps format|

## Configuration
You can configure exporter via environment variables below.

|Environment Variable Name|Default|Description|
|:--|:--|:--|
|`METRICS_PORT`|`8080`|Port number of metrics endpoint|
|`SPEEDTEST_CRON_SCHEDULE`|`@every 30m`|Cron Schedule for executing speedtest. Unless speedtest is executed, metrics values won't be changed. Acceptable format is described in [here](https://pkg.go.dev/github.com/robfig/cron#hdr-CRON_Expression_Format).|
|`SPEEDTEST_DURATION`|`15s`|Duration time of each test (upload and download). So whole duration of speedtest is `2 * SPEEDTEST_DURATION`. Acceptable format is described [here](https://pkg.go.dev/time#ParseDuration).|
|`SPEEDTEST_THREAD_COUNT`|`64`|Goroutine thread count for executing speedtest.|
|`LOG_LEVEL`|`INFO`|Log level of this exporter. Acceptable format is describe [here](https://pkg.go.dev/log/slog#Level.String).|
