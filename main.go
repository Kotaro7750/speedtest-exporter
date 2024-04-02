package main

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"os"

	"github.com/caarlos0/env/v10"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/robfig/cron"
	"github.com/showwin/speedtest-go/speedtest"
)

type Config struct {
	MetricPort            int    `env:"METRICS_PORT" envDefault:"8080"`
	SpeedTestCronSchedule string `env:"SPEEDTEST_CRON_SCHEDULE" envDefault:"@every 30m"`
	SpeedtestThreadCount  int    `env:"SPEEDTEST_THREAD_COUNT" envDefault:"64"`
	LogLevel              string `env:"LOG_LEVEL" envDefault:"INFO"`
}

type Metrics struct {
	DLSpeed prometheus.Gauge
	ULSpeed prometheus.Gauge
}

var registry prometheus.Registry

func main() {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))
	slog.SetDefault(logger)

	slog.Info("Start speedtest-exporter")

	config := Config{}
	if err := env.Parse(&config); err != nil {
		slog.Error(err.Error())
		os.Exit(1)
	}

	var logLevel slog.Level
	err := logLevel.UnmarshalText([]byte(config.LogLevel))
	if err != nil {
		slog.Error(err.Error())
		os.Exit(1)
	}

	logger = slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: logLevel}))
	slog.SetDefault(logger)

	slog.Info("speedtest-exporter config", "config", config)

	registry = *prometheus.NewRegistry()

	metrics := Metrics{
		prometheus.NewGauge(prometheus.GaugeOpts{
			Namespace: "speedtest",
			Name:      "download_speed_mbps",
		}),
		prometheus.NewGauge(prometheus.GaugeOpts{
			Namespace: "speedtest",
			Name:      "upload_speed_mbps",
		}),
	}

	err = registry.Register(metrics.DLSpeed)
	if err != nil {
		slog.Error(err.Error())
		os.Exit(1)
	}

	err = registry.Register(metrics.ULSpeed)
	if err != nil {
		slog.Error(err.Error())
		os.Exit(1)
	}

	slog.Info(fmt.Sprintf("Speedtest run schedule is %s", config.SpeedTestCronSchedule))

	c := cron.New()
	err = c.AddFunc(config.SpeedTestCronSchedule, func() {
		speedtestClient := speedtest.New()
		speedtestClient.SetNThread(config.SpeedtestThreadCount)

		user, err := speedtestClient.FetchUserInfo()
		if err != nil {
			slog.Error(err.Error())
			return
		}

		slog.Debug("Fetch userInfo done", "user", user)

		doSpeedTestMulti(*speedtestClient, &metrics)
		if err != nil {
			slog.Error(err.Error())
			return
		}
	})

	if err != nil {
		slog.Error(err.Error())
		os.Exit(1)
	}

	c.Start()

	http.Handle("/metrics", promhttp.HandlerFor(&registry, promhttp.HandlerOpts{Registry: &registry}))
	if err := http.ListenAndServe(fmt.Sprintf(":%d", config.MetricPort), nil); err != nil {
		slog.Error(err.Error())
		os.Exit(1)
	}

}

func doSpeedTestMulti(speedtestClient speedtest.Speedtest, metrics *Metrics) error {
	servers, err := speedtestClient.FetchServers()
	if err != nil {
		return err
	}

	slog.Debug("Fetch servers done", "servers", servers)

	target := (*servers.Available())[0]
	if err != nil {
		return err
	}

	slog.Debug("Target server is decided", "server", target)

	slog.Info("Speedtest started")

	slog.Debug("Start download test")
	err = target.MultiDownloadTestContext(context.Background(), servers)
	if err != nil {
		return err
	}
	slog.Debug("Download test is done")

	slog.Debug("Start upload test")
	err = target.MultiUploadTestContext(context.Background(), servers)
	if err != nil {
		return err
	}
	slog.Debug("Upload test is done")

	metrics.DLSpeed.Set(target.DLSpeed)
	metrics.ULSpeed.Set(target.ULSpeed)

	slog.Info("Speedtest is done", "latencyMs", target.Latency.Milliseconds(), "download", target.DLSpeed, "upload", target.ULSpeed)
	target.Context.Reset()

	return nil
}
