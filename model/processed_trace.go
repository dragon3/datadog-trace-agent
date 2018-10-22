package model

import "github.com/DataDog/datadog-trace-agent/constants"

type ProcessedTrace struct {
	Trace         Trace
	WeightedTrace WeightedTrace
	Root          *Span
	Env           string
	Sublayers     map[*Span][]SublayerValue
}

func (pt *ProcessedTrace) Weight() float64 {
	if pt.Root == nil {
		return 1.0
	}
	return pt.Root.Weight()
}

func (pt *ProcessedTrace) GetSamplingPriority() (int, bool) {
	if pt.Root == nil {
		return 0, false
	}
	p, ok := pt.Root.Metrics[constants.SamplingPriorityKey]
	return int(p), ok
}
