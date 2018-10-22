package eventsampler

import (
	"github.com/DataDog/datadog-trace-agent/model"
	"github.com/DataDog/datadog-trace-agent/sampler"
)

// MaxEPSSampler (Max Events Per Second Sampler) is an event sampler that samples provided events so as to ensure no
// more than a certain amount of events is sampled per second.
type MaxEPSSampler struct {
	maxEPS      float64
	rateCounter RateCounter
}

// NewMaxEPSSampler creates a new instance of a MaxEPSSampler with the provided maximum amount of events per second.
func NewMaxEPSSampler(maxEPS float64, rateCounter RateCounter) *MaxEPSSampler {
	return &MaxEPSSampler{
		maxEPS:      maxEPS,
		rateCounter: rateCounter,
	}
}

// Sample determines whether or not we should sample the provided event in order to ensure no more than maxEPS events
// are sampled every second.
func (s *MaxEPSSampler) Sample(event *model.APMEvent) SamplingDecision {
	maxEPSRate := 1.0
	currentEPS := s.rateCounter.GetSampledRate()

	if currentEPS > s.maxEPS {
		maxEPSRate = s.maxEPS / currentEPS
	}

	sampled := sampler.SampleByRate(event.Span.TraceID, maxEPSRate)

	if sampled {
		s.rateCounter.CountSampled()
		return SAMPLE
	}

	return DONT_SAMPLE
}

// RateCounter keeps track of different event rates.
type RateCounter interface {
	Start()
	Stop()
	CountSampled()
	GetSampledRate() float64
}

// SamplerBackendRateCounter is a RateCounter backed by a sampler.Backend.
type SamplerBackendRateCounter struct {
	backend sampler.Backend
}

func NewSamplerBackendRateCounter() *SamplerBackendRateCounter {
	return &SamplerBackendRateCounter{
		backend: sampler.NewMemoryBackend(sampler.DefaultDecayPeriod, sampler.DefaultDecayFactor),
	}
}

func (sb *SamplerBackendRateCounter) Start() {
	go sb.backend.Run()
}

func (sb *SamplerBackendRateCounter) Stop() {
	sb.backend.Stop()
}

func (sb *SamplerBackendRateCounter) CountSampled() {
	sb.backend.CountSample()
}

func (sb *SamplerBackendRateCounter) GetSampledRate() float64 {
	// TODO: or should it be sb.backend.GetSampledScore()
	return sb.backend.GetUpperSampledScore()
}

// ReadOnlyRateCounter is a read-only view of a backing RateCounter.
type ReadOnlyRateCounter struct {
	rateCounter RateCounter
}

func NewReadOnlyRateCounter(rateCounter RateCounter) *ReadOnlyRateCounter {
	return &ReadOnlyRateCounter{
		rateCounter: rateCounter,
	}
}

func (ro *ReadOnlyRateCounter) Start() {
	// no-op
}

func (ro *ReadOnlyRateCounter) Stop() {
	// no-op
}

func (ro *ReadOnlyRateCounter) CountSampled() {
	// no-op
}

func (ro *ReadOnlyRateCounter) GetSampledRate() float64 {
	return ro.rateCounter.GetSampledRate()
}
