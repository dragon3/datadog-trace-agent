package eventsampler

import (
	"github.com/DataDog/datadog-trace-agent/config"
	"github.com/DataDog/datadog-trace-agent/model"
)

type SamplingDecision int8

const (
	NO_DECISION SamplingDecision = iota
	SAMPLE                       = iota
	DONT_SAMPLE                  = iota
)

// Sampler samples APM events according to implementation-defined techniques.
type Sampler interface {
	// Sample decides whether to sample the provided event or not.
	Sample(event *model.APMEvent) SamplingDecision
}

// BatchSampler allows sampling a collection of APM events, returning only those that survived sampling.
type BatchSampler struct {
	sampler Sampler
}

// NewBatchSampler creates a new BatchSampler using the provided underlying sampler.
func NewBatchSampler(sampler Sampler) *BatchSampler {
	return &BatchSampler{
		sampler: sampler,
	}
}

// Sample takes a collection of events, makes a sampling decision for each event and returns a collection containing
// only those events that were sampled.
func (bs *BatchSampler) Sample(events []*model.APMEvent) []*model.APMEvent {
	result := make([]*model.APMEvent, 0, len(events))

	for _, event := range events {
		if bs.sampler.Sample(event) == SAMPLE {
			result = append(result, event)
		}
	}

	return result
}

// FromConf creates an event sampler based on the provided agent configuration.
func FromConf(conf *config.AgentConfig) *BatchSampler {
	rateCounter := NewSamplerBackendRateCounter()
	// Start and leave running until the end
	rateCounter.Start()

	return NewBatchSampler(
		NewSamplerChain(
			[]Sampler{
				// Sample all events for which their respective traces were sampled
				NewSampledTraceSampler(),
				// For those events that did not have its trace sampled, sample as many as possible respecting MaxEPS.
				NewMaxEPSSampler(conf.MaxEPS, NewReadOnlyRateCounter(rateCounter)),
			}, func(decision SamplingDecision) {
				if decision == SAMPLE {
					// We only increment the number of events sampled after the chain decides. This is because of the
					// chain's shortcircuiting mechanism.
					rateCounter.CountSampled()
				}
			}),
	)
}
