// Copyright The OpenTelemetry Authors
// SPDX-License-Identifier: Apache-2.0

package telemetrygen

import (
	"bufio"
	"bytes"
	"context"
	_ "embed"
	"time"

	"go.opentelemetry.io/collector/component"
	"go.opentelemetry.io/collector/consumer"
	"go.opentelemetry.io/collector/pdata/pcommon"
	"go.opentelemetry.io/collector/pdata/pmetric"
	"go.opentelemetry.io/collector/receiver"
	"go.uber.org/zap"
)

//go:embed demo-data/metrics.json
var demoMetrics []byte

type metricsGenerator struct {
	cfg    *Config
	logger *zap.Logger

	sampleMetrics   []pmetric.Metrics
	lastSampleIndex int
	consumer        consumer.Metrics

	cancelFn context.CancelFunc
}

func createMetricsReceiver(
	ctx context.Context,
	set receiver.Settings,
	config component.Config,
	consumer consumer.Metrics,
) (receiver.Metrics, error) {
	genConfig := config.(*Config)
	recv := metricsGenerator{
		cfg:             genConfig,
		logger:          set.Logger,
		consumer:        consumer,
		sampleMetrics:   make([]pmetric.Metrics, 0),
		lastSampleIndex: 0,
	}

	parser := pmetric.JSONUnmarshaler{}
	scanner := bufio.NewScanner(bytes.NewReader(demoMetrics))
	for scanner.Scan() {
		lineMetrics, err := parser.UnmarshalMetrics(scanner.Bytes())
		if err != nil {
			return nil, err
		}
		recv.sampleMetrics = append(recv.sampleMetrics, lineMetrics)
	}

	return &recv, nil
}

func (ar *metricsGenerator) Start(ctx context.Context, _ component.Host) error {
	startCtx, cancelFn := context.WithCancel(ctx)
	ar.cancelFn = cancelFn

	go func() {
		ticker := time.NewTicker(1 * time.Second)
		parser := pmetric.JSONMarshaler{}
		var throughput, totalSeconds, totalSendBytes float64
		for {
			select {

			case <-startCtx.Done():
				break
			case <-ticker.C:
				totalSeconds += 1
				throughput = totalSendBytes / totalSeconds
				for throughput < ar.cfg.Throughput {
					nMetrics, err := ar.nextMetrics()
					if err != nil {
						ar.logger.Error(err.Error())
						continue
					}
					err = ar.consumer.ConsumeMetrics(startCtx, nMetrics)
					if err != nil {
						ar.logger.Error(err.Error())
						continue
					}

					jsonMetrics, err := parser.MarshalMetrics(nMetrics)
					if err != nil {
						ar.logger.Error(err.Error())
						continue
					}

					totalSendBytes += float64(len(jsonMetrics))
					throughput = totalSendBytes / totalSeconds
				}
				ar.logger.Info("Consumed metrics", zap.Float64("bytes", totalSendBytes))
			}
		}
	}()
	return nil
}

func (ar *metricsGenerator) Shutdown(context.Context) error {
	if ar.cancelFn != nil {
		ar.cancelFn()
	}
	return nil
}

func (ar *metricsGenerator) nextMetrics() (pmetric.Metrics, error) {
	now := pcommon.NewTimestampFromTime(time.Now())

	nextMetrics := pmetric.NewMetrics()

	ar.sampleMetrics[ar.lastSampleIndex].CopyTo(nextMetrics)
	rm := nextMetrics.ResourceMetrics()
	for i := 0; i < rm.Len(); i++ {
		for j := 0; j < rm.At(i).ScopeMetrics().Len(); j++ {
			for k := 0; k < rm.At(i).ScopeMetrics().At(j).Metrics().Len(); k++ {
				smetric := rm.At(i).ScopeMetrics().At(j).Metrics().At(k)
				switch smetric.Type() {
				case pmetric.MetricTypeGauge:
					dps := smetric.Gauge().DataPoints()
					for i := 0; i < dps.Len(); i++ {
						dps.At(i).SetTimestamp(now)
						dps.At(i).SetStartTimestamp(now)
					}
				case pmetric.MetricTypeSum:
					dps := smetric.Sum().DataPoints()
					for i := 0; i < dps.Len(); i++ {
						dps.At(i).SetTimestamp(now)
						dps.At(i).SetStartTimestamp(now)
					}
				case pmetric.MetricTypeHistogram:
					dps := smetric.Histogram().DataPoints()
					for i := 0; i < dps.Len(); i++ {
						dps.At(i).SetTimestamp(now)
						dps.At(i).SetStartTimestamp(now)
					}
				case pmetric.MetricTypeExponentialHistogram:
					dps := smetric.ExponentialHistogram().DataPoints()
					for i := 0; i < dps.Len(); i++ {
						dps.At(i).SetTimestamp(now)
						dps.At(i).SetStartTimestamp(now)
					}
				case pmetric.MetricTypeSummary:
					dps := smetric.Summary().DataPoints()
					for i := 0; i < dps.Len(); i++ {
						dps.At(i).SetTimestamp(now)
						dps.At(i).SetStartTimestamp(now)
					}
				case pmetric.MetricTypeEmpty:
				}
			}
		}
	}

	ar.lastSampleIndex = (ar.lastSampleIndex + 1) % len(ar.sampleMetrics)

	return nextMetrics, nil
}
