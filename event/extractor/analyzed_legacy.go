package eventextractor

import (
	"github.com/DataDog/datadog-trace-agent/model"
	"github.com/DataDog/datadog-trace-agent/sampler"
)

type LegacyAnalyzedExtractor struct {
	analyzedRateByService map[string]float64
}

// NewLegacyAnalyzed returns an APM event extractor that extracts APM events from a trace following the specified
// extraction rates for any spans matching a specific service.
func NewLegacyAnalyzed(analyzedRateByService map[string]float64) *LegacyAnalyzedExtractor {
	return &LegacyAnalyzedExtractor{
		analyzedRateByService: analyzedRateByService,
	}
}

// Extract extracts analyzed spans from the trace and returns them as a slice
func (s *LegacyAnalyzedExtractor) Extract(t model.ProcessedTrace, sampledTrace bool) []*model.APMEvent {
	var events []*model.APMEvent

	// inspect the WeightedTrace so that we can identify top-level spans
	for _, span := range t.WeightedTrace {
		if s.shouldAnalyze(span) {
			events = append(events, &model.APMEvent{
				Span:         span.Span,
				TraceSampled: sampledTrace,
			})
		}
	}

	return events
}

// shouldAnalyze tells if a span should be considered as analyzed
// Only top-level spans are eligible to be analyzed
func (s *LegacyAnalyzedExtractor) shouldAnalyze(span *model.WeightedSpan) bool {
	if !span.TopLevel {
		return false
	}

	if analyzeRate, ok := s.analyzedRateByService[span.Service]; ok {
		if sampler.SampleByRate(span.TraceID, analyzeRate) {
			return true
		}
	}

	return false
}
