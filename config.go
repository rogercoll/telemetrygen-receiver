// Copyright The OpenTelemetry Authors
// SPDX-License-Identifier: Apache-2.0
package telemetrygen

import (
	"time"

	"go.opentelemetry.io/collector/component"
)

type (
	Throughput float64
	JsonFile   string
)

// Config defines configuration for HostMetrics receiver.
type Config struct {
	Metrics MetricsConfig `mapstructure:"metrics"`
	Logs    LogsConfig    `mapstructure:"logs"`
	Traces  TracesConfig  `mapstructure:"traces"`
}

type MetricsConfig struct {
	Throughput `mapstructure:"throughput"`
	// JsonFile is an optional configuration option to specify the path to
	// get the base generated signals from.
	JsonFile `mapstructure:"json_file"`
}

type LogsConfig struct {
	Throughput `mapstructure:"throughput"`
	// JsonFile is an optional configuration option to specify the path to
	// get the base generated signals from.
	JsonFile `mapstructure:"json_file"`
}

type TracesConfig struct {
	Throughput `mapstructure:"throughput"`
	// JsonFile is an optional configuration option to specify the path to
	// get the base generated signals from.
	JsonFile `mapstructure:"json_file"`
	// Spans duration timestamps are randomly set between 0ms and the
	// max_spans_interval duration value. Defaults to 3000ms.
	MaxSpansInterval time.Duration  `mapstructure:"max_spans_interval"`
	Services         ServicesConfig `mapstructure:"services"`
}

type ServicesConfig struct {
	// RandomizedNameCount specifies the number of randomized service names to generate.
	// This field is only relevant when is greater than 0. If not set or less than 1,
	// pre-loaded services will be used.
	RandomizedNameCount int `mapstructure:"randomized_name_count"`
}

var _ component.Config = (*Config)(nil)

// Validate checks the receiver configuration is valid
func (cfg *Config) Validate() error {
	return nil
}
