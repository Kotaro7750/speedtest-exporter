package main

import (
	"context"
	"log/slog"
	"os"

	"github.com/showwin/speedtest-go/speedtest"
)

func main() {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))
	slog.SetDefault(logger)

	speedtestClient := speedtest.New()
	speedtestClient.SetNThread(64)

	user, err := speedtestClient.FetchUserInfo()
	if err != nil {
		slog.Error(err.Error())
		os.Exit(1)
	}

	slog.Info("Fetch userInfo done", "user", user)

	doSpeedTestMulti(*speedtestClient)
}

func doSpeedTestSingle(speedtestClient speedtest.Speedtest) {
	servers, err := speedtestClient.FetchServers()
	if err != nil {
		slog.Error(err.Error())
		os.Exit(1)
	}

	slog.Debug("Fetch servers done", "servers", servers)

	targets, err := servers.Available().FindServer([]int{})
	if err != nil {
		slog.Error(err.Error())
		os.Exit(1)
	}

	for _, s := range targets {
		slog.Info("Start speedtest", "server", s)

		slog.Info("Start download test")
		err := s.DownloadTest()
		if err != nil {
			slog.Error(err.Error())
			os.Exit(1)
		}

		slog.Info("Start upload test")
		err = s.UploadTest()
		if err != nil {
			slog.Error(err.Error())
			os.Exit(1)
		}

		slog.Info("Speedtest is done", "latencyUs", s.Latency.Microseconds(), "download", s.DLSpeed, "upload", s.ULSpeed)
		s.Context.Reset()
	}
}

func doSpeedTestMulti(speedtestClient speedtest.Speedtest) {
	servers, err := speedtestClient.FetchServers()
	if err != nil {
		slog.Error(err.Error())
		os.Exit(1)
	}

	slog.Debug("Fetch servers done", "servers", servers)

	targets := servers.Available()

	for _, target := range *targets {
		slog.Info("Start download test")
		err := target.MultiDownloadTestContext(context.Background(), servers)
		if err != nil {
			slog.Error(err.Error())
			os.Exit(1)
		}

		slog.Info("Start upload test")
		err = target.MultiUploadTestContext(context.Background(), servers)
		if err != nil {
			slog.Error(err.Error())
			os.Exit(1)
		}

		slog.Info("Speedtest is done", "latencyUs", target.Latency.Microseconds(), "download", target.DLSpeed, "upload", target.ULSpeed)
		target.Context.Reset()
	}
}
