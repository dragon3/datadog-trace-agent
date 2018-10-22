package eventextractor

import "github.com/DataDog/datadog-trace-agent/model"

type DisabledExtractor struct{}

// NewDisabled returns a new APM event extractor that does not extract any events.
func NewDisabled() *DisabledExtractor {
	return &DisabledExtractor{}
}

func (s *DisabledExtractor) Extract(t model.ProcessedTrace, sampledTrace bool) []*model.APMEvent {
	return nil
}
