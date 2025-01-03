// Copyright The OpenTelemetry Authors
// SPDX-License-Identifier: Apache-2.0

package telemetrygen

import (
	"bufio"
	"bytes"
	"context"
	_ "embed"
	"math/rand"
	"time"

	"go.opentelemetry.io/collector/component"
	"go.opentelemetry.io/collector/consumer"
	"go.opentelemetry.io/collector/pdata/pcommon"
	"go.opentelemetry.io/collector/pdata/ptrace"
	"go.opentelemetry.io/collector/receiver"
	"go.uber.org/zap"
)

//go:embed demo-data/traces.json
var demoTraces []byte

type receiverTraces struct {
	traces   ptrace.Traces
	jsonSize int
}

type tracesGenerator struct {
	cfg    *Config
	logger *zap.Logger

	sampleTraces    []receiverTraces
	lastSampleIndex int
	consumer        consumer.Traces

	cancelFn context.CancelFunc
}

func createTracesReceiver(
	ctx context.Context,
	set receiver.Settings,
	config component.Config,
	consumer consumer.Traces,
) (receiver.Traces, error) {
	genConfig := config.(*Config)
	recv := tracesGenerator{
		cfg:             genConfig,
		logger:          set.Logger,
		consumer:        consumer,
		sampleTraces:    make([]receiverTraces, 0),
		lastSampleIndex: 0,
	}

	parser := ptrace.JSONUnmarshaler{}
	scanner := bufio.NewScanner(bytes.NewReader(demoTraces))
	for scanner.Scan() {
		traceBytes := scanner.Bytes()
		lineTraces, err := parser.UnmarshalTraces(traceBytes)
		if err != nil {
			return nil, err
		}
		recv.sampleTraces = append(recv.sampleTraces, receiverTraces{
			traces:   lineTraces,
			jsonSize: len(traceBytes),
		})
	}

	return &recv, nil
}

func (ar *tracesGenerator) Start(ctx context.Context, _ component.Host) error {
	startCtx, cancelFn := context.WithCancel(ctx)
	ar.cancelFn = cancelFn

	go func() {
		ticker := time.NewTicker(1 * time.Second)
		var throughput, totalSeconds, totalSendBytes float64
		for {
			select {

			case <-startCtx.Done():
				break
			case <-ticker.C:
				totalSeconds += 1
				throughput = totalSendBytes / totalSeconds
				for throughput < ar.cfg.Throughput {
					nTraces, nSize, err := ar.nextTraces()
					if err != nil {
						ar.logger.Error(err.Error())
						continue
					}
					err = ar.consumer.ConsumeTraces(startCtx, nTraces)
					if err != nil {
						ar.logger.Error(err.Error())
						continue
					}

					totalSendBytes += float64(nSize)
					throughput = totalSendBytes / totalSeconds
				}
				ar.logger.Info("Consumed traces", zap.Float64("bytes", totalSendBytes))
			}
		}
	}()
	return nil
}

func (ar *tracesGenerator) Shutdown(context.Context) error {
	if ar.cancelFn != nil {
		ar.cancelFn()
	}
	return nil
}

func (ar *tracesGenerator) nextTraces() (ptrace.Traces, int, error) {
	nextLogs := ptrace.NewTraces()

	ar.sampleTraces[ar.lastSampleIndex].traces.CopyTo(nextLogs)
	sampledSize := ar.sampleTraces[ar.lastSampleIndex].jsonSize

	rm := nextLogs.ResourceSpans()
	for i := 0; i < rm.Len(); i++ {
		for j := 0; j < rm.At(i).ScopeSpans().Len(); j++ {
			for k := 0; k < rm.At(i).ScopeSpans().At(j).Spans().Len(); k++ {
				sspan := rm.At(i).ScopeSpans().At(j).Spans().At(k)

				now := time.Now()
				// Generate a random duration between 0 and 3 seconds
				sspan.SetStartTimestamp(pcommon.NewTimestampFromTime(now.Add(-time.Duration(rand.Intn(3000)) * time.Millisecond)))
				sspan.SetEndTimestamp(pcommon.NewTimestampFromTime(now))
			}
		}
	}

	ar.lastSampleIndex = (ar.lastSampleIndex + 1) % len(ar.sampleTraces)

	return nextLogs, sampledSize, nil
}

func (hmr *tracesGenerator) shutdown(_ context.Context) error {
	return nil
}
