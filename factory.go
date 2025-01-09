// Copyright The OpenTelemetry Authors
// SPDX-License-Identifier: Apache-2.0

package telemetrygen

import (
	"go.opentelemetry.io/collector/component"
	"go.opentelemetry.io/collector/receiver"
)

// CreateDefaultConfig creates the default configuration for the Scraper.
func createDefaultReceiverConfig() *Config {
	return &Config{
		Metrics: MetricsConfig{
			// 1 kB
			Throughput: 1024,
		},
		Logs: LogsConfig{
			// 1 kB
			Throughput: 1024,
		},
		Traces: TracesConfig{
			// 1 kB
			Throughput:       1024,
			MaxSpansInterval: 3000,
		},
	}
}

func NewFactory() receiver.Factory {
	return receiver.NewFactory(
		component.MustNewType("telemetrygen"),
		func() component.Config {
			return createDefaultReceiverConfig()
		},
		receiver.WithMetrics(createMetricsReceiver, component.StabilityLevelDevelopment),
		receiver.WithTraces(createTracesReceiver, component.StabilityLevelDevelopment),
		receiver.WithLogs(createLogsReceiver, component.StabilityLevelDevelopment),
	)
}
