package ado

import (
	"github.com/microsoft/ApplicationInsights-Go/appinsights"
	"os"
	"time"
)

// Logging interface
type Logging interface {
	LogWarning(msg string)
	LogError(err error)
	LogFatal(err error)
	LogInfo(msg string)
}

// AppInsightsLogger struct will handle logging to Application Insights
type AppInsightsLogger struct {
	client appinsights.TelemetryClient
}

// LogWarning comment
func (logger *AppInsightsLogger) LogWarning(msg string) {
	logger.initializeLogger()
	logger.client.TrackTrace(msg, appinsights.Warning)
}

// LogError comment
func (logger *AppInsightsLogger) LogError(err error) {
	logger.initializeLogger()
	logger.client.TrackException(err)
}

// LogFatal comment
func (logger *AppInsightsLogger) LogFatal(err error) {
	logger.initializeLogger()
	logger.client.TrackException(err)
}

// LogInfo comment
func (logger *AppInsightsLogger) LogInfo(msg string) {
	logger.initializeLogger()
	logger.client.TrackTrace(msg, appinsights.Information)
}

func (logger *AppInsightsLogger) initializeLogger() {
	if logger.client == nil {
		telemetryConfig := appinsights.NewTelemetryConfiguration(os.Getenv("APPINSIGHTS_INSTRUMENTATIONKEY"))

		// Configure how many items can be sent in one call to the data collector:
		telemetryConfig.MaxBatchSize = 8192

		// Configure the maximum delay before sending queued telemetry:
		telemetryConfig.MaxBatchInterval = 2 * time.Second

		logger.client = appinsights.NewTelemetryClientFromConfig(telemetryConfig)
	}
}