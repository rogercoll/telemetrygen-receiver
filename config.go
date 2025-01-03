// Copyright The OpenTelemetry Authors
// SPDX-License-Identifier: Apache-2.0
package telemetrygen

import (
	"go.opentelemetry.io/collector/component"
)

const (
	providersKey = "providers"
)

// Config defines configuration for HostMetrics receiver.
type Config struct {
	Throughput float64 `mapstructure:"throughput"`
}

var _ component.Config = (*Config)(nil)

// Validate checks the receiver configuration is valid
func (cfg *Config) Validate() error {
	return nil
}
