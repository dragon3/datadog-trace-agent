package main

import (
	"fmt"
	"testing"

	"github.com/DataDog/datadog-trace-agent/model"
	"github.com/DataDog/datadog-trace-agent/sampler"
	"github.com/DataDog/datadog-trace-agent/testutil"
	"github.com/stretchr/testify/assert"
)

func createTrace(serviceName string, operationName string, topLevel bool, hasPriority bool, priority int) processedTrace {
	ws := model.WeightedSpan{TopLevel: topLevel, Span: &model.Span{Service: serviceName, Name: operationName}}
	if hasPriority {
		ws.SetMetric(sampler.SamplingPriorityKey, float64(priority))
	}
	wt := model.WeightedTrace{&ws}
	return processedTrace{WeightedTrace: wt, Root: ws.Span}
}

func TestTransactionSampler(t *testing.T) {
	assert := assert.New(t)

	config := make(map[string]map[string]float64)
	config["myService"] = make(map[string]float64)
	config["myService"]["myOperation"] = 1

	config["mySampledService"] = make(map[string]float64)
	config["mySampledService"]["myOperation"] = 0

	tests := []struct {
		name             string
		trace            processedTrace
		expectedSampling bool
	}{
		{"Top-level service and span name match", createTrace("myService", "myOperation", true, false, 0), true},
		{"Top-level service name doesn't match", createTrace("otherService", "myOperation", true, false, 0), false},
		{"Top-level span name doesn't match", createTrace("myService", "otherOperation", true, false, 0), false},
		{"Top-level service and span name don't match", createTrace("otherService", "otherOperation", true, false, 0), false},
		{"Non top-level service and span name match", createTrace("myService", "myOperation", false, false, 0), true},
		{"Non top-level service name doesn't match", createTrace("otherService", "myOperation", false, false, 0), false},
		{"Non top-level span name doesn't match", createTrace("myService", "otherOperation", false, false, 0), false},
		{"Non top-level service and span name don't match", createTrace("otherService", "otherOperation", false, false, 0), false},
		{"Match, sampling rate 0, no priority", createTrace("mySampledService", "myOperation", true, false, 0), false},
		{"Match, sampling rate 0, priority 1", createTrace("mySampledService", "myOperation", true, true, 1), false},
		{"Match, sampling rate 0, priority 2", createTrace("mySampledService", "myOperation", true, true, 2), true},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			ts := newTransactionSampler(config, 10)
			analyzedSpans := ts.Extract(test.trace)

			if test.expectedSampling {
				assert.Len(analyzedSpans, 1, fmt.Sprintf("Trace %v should have been sampled", test.trace))
			} else {
				assert.Len(analyzedSpans, 0, fmt.Sprintf("Trace %v should not have been sampled", test.trace))
			}
		})
	}
}

func TestMaxEventsPerTrace(t *testing.T) {
	testSpan := testutil.RandomSpan()

	config := map[string]map[string]float64{
		testSpan.Service: {
			testSpan.Name: 1,
		},
	}

	spans := make([]*model.Span, 100)

	for i := 0; i < len(spans); i++ {
		spans[i] = testSpan
	}

	trace := model.Trace(spans)

	processedTrace := processedTrace{
		Env:           "test",
		Root:          spans[0],
		Sublayers:     nil,
		Trace:         trace,
		WeightedTrace: model.NewWeightedTrace(trace, spans[0]),
	}

	t.Run("no limit", func(t *testing.T) {
		ts := newTransactionSampler(config, -1)
		assert.Len(t, ts.Extract(processedTrace), len(spans))
	})

	t.Run("no events", func(t *testing.T) {
		ts := newTransactionSampler(config, 0)
		assert.Len(t, ts.Extract(processedTrace), 0)
	})

	t.Run("5 events", func(t *testing.T) {
		ts := newTransactionSampler(config, 5)
		assert.Len(t, ts.Extract(processedTrace), 5)
	})

	t.Run("exact num spans", func(t *testing.T) {
		ts := newTransactionSampler(config, len(spans))
		assert.Len(t, ts.Extract(processedTrace), len(spans))
	})
}
