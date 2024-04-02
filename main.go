package main

import (
	"context"
	"log/slog"
	"net/http"
	"os"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/robfig/cron"
	"github.com/showwin/speedtest-go/speedtest"
)

type Metrics struct {
	DLSpeed prometheus.Gauge
	ULSpeed prometheus.Gauge
}

var registry prometheus.Registry

func main() {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))
	slog.SetDefault(logger)

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

	err := registry.Register(metrics.DLSpeed)
	if err != nil {
		slog.Error(err.Error())
		os.Exit(1)
	}

	err = registry.Register(metrics.ULSpeed)
	if err != nil {
		slog.Error(err.Error())
		os.Exit(1)
	}

	c := cron.New()
	c.AddFunc("@every 1m", func() {
		speedtestClient := speedtest.New()
		speedtestClient.SetNThread(64)

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

	c.Start()

	http.Handle("/metrics", promhttp.HandlerFor(&registry, promhttp.HandlerOpts{Registry: &registry}))
	http.ListenAndServe(":8080", nil)

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
