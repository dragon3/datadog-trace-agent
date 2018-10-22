package eventextractor

import (
	"github.com/DataDog/datadog-trace-agent/model"
	"github.com/DataDog/datadog-trace-agent/sampler"
)

type AnalyzedExtractor struct {
	analyzedSpansByService map[string]map[string]float64
}

// NewAnalyzed returns an APM event extractor that extracts APM events from a trace following the provided
// extraction rates for any spans matching a (service name, operation name) pair.
func NewAnalyzed(analyzedSpansByService map[string]map[string]float64) *AnalyzedExtractor {
	return &AnalyzedExtractor{
		analyzedSpansByService: analyzedSpansByService,
	}
}

// Extract extracts analyzed spans from the trace and returns them as a slice
func (s *AnalyzedExtractor) Extract(t model.ProcessedTrace, sampledTrace bool) []*model.APMEvent {
	var events []*model.APMEvent

	// Get the trace priority
	priority, hasPriority := t.GetSamplingPriority()

	for _, span := range t.WeightedTrace {
		if s.shouldAnalyze(span, hasPriority, priority) {
			events = append(events, &model.APMEvent{
				Span:         span.Span,
				TraceSampled: sampledTrace,
			})
		}
	}

	return events
}

func (s *AnalyzedExtractor) shouldAnalyze(span *model.WeightedSpan, hasPriority bool, priority int) bool {
	if operations, ok := s.analyzedSpansByService[span.Service]; ok {
		if analyzeRate, ok := operations[span.Name]; ok {
			// If the trace has been manually sampled, we keep all matching spans
			highPriority := hasPriority && priority >= 2
			if highPriority || sampler.SampleByRate(span.TraceID, analyzeRate) {
				return true
			}
		}
	}
	return false
}
