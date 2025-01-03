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
		// 1MB
		Throughput: 1024,
	}
}

func NewFactory() receiver.Factory {
	return receiver.NewFactory(
		component.MustNewType("telemetrygen"),
		func() component.Config {
			return createDefaultReceiverConfig()
		},
		receiver.WithMetrics(createMetricsReceiver, component.StabilityLevelDevelopment),
		receiver.WithLogs(createLogsReceiver, component.StabilityLevelDevelopment),
	)
}
