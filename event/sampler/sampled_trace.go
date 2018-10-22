package eventsampler

import "github.com/DataDog/datadog-trace-agent/model"

// SampledTraceSampler is an event sampler that ensures that events for a sampled trace are sampled as well.
type SampledTraceSampler struct {
}

// NewSampledTraceSampler creates a new instance of a SampledTraceSampler.
func NewSampledTraceSampler() *SampledTraceSampler {
	return &SampledTraceSampler{}
}

// Sample samples the provided event (returns true) if the corresponding trace was sampled.
func (sts *SampledTraceSampler) Sample(event *model.APMEvent) SamplingDecision {
	if event.TraceSampled {
		return SAMPLE
	}

	return NO_DECISION
}
