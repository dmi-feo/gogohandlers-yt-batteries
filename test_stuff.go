package gogohandlers_yt_batteries

import "log/slog"

type TestAppServiceProvider struct {
	logger          *slog.Logger
	YtClientFactory YtClientFactory
}

func (tsp TestAppServiceProvider) GetYtClientFactory() YtClientFactory {
	return tsp.YtClientFactory
}

func NewTestAppServiceProvider(logger *slog.Logger, ytSettings YtClientSettings) *TestAppServiceProvider {
	return &TestAppServiceProvider{
		logger: logger,
		YtClientFactory: YtClientFactory{
			logger:   logger,
			settings: ytSettings,
		},
	}
}
