package eventsampler

import "github.com/DataDog/datadog-trace-agent/model"

// SamplerChain is an APMEvent sampler that submits an event to a sequence of samplers, returning as soon as one of
// the samplers makes a decision.
type SamplerChain struct {
	samplers []Sampler
	callback func(decision SamplingDecision)
}

// NewSamplerChain creates a new sampler chain from the provided samplers and with the specified decision callback.
func NewSamplerChain(samplers []Sampler, callback func(decision SamplingDecision)) *SamplerChain {
	return &SamplerChain{
		samplers: samplers,
		callback: callback,
	}
}

// Sample returns the first !NO_DECISION sampling decision from calls to the underlying samplers, in order.
func (sc *SamplerChain) Sample(event *model.APMEvent) SamplingDecision {
	for _, sampler := range sc.samplers {
		decision := sampler.Sample(event)

		if decision != NO_DECISION {
			return decision
		}
	}

	return NO_DECISION
}
