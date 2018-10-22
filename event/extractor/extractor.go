package eventextractor

import (
	"github.com/DataDog/datadog-trace-agent/config"
	"github.com/DataDog/datadog-trace-agent/model"
)

// Extractor extracts APM event spans from a trace.
type Extractor interface {
	// Extract extracts APM event spans from the given weighted trace information and returns them.
	Extract(t model.ProcessedTrace, sampledTrace bool) []*model.APMEvent
}

// FromConf creates a new APM event extractor based on the provided agent configuration.
func FromConf(conf *config.AgentConfig) Extractor {
	if len(conf.AnalyzedSpansByService) > 0 {
		return NewAnalyzed(conf.AnalyzedSpansByService)
	}
	if len(conf.AnalyzedRateByServiceLegacy) > 0 {
		return NewLegacyAnalyzed(conf.AnalyzedRateByServiceLegacy)
	}

	// TODO: Replace disabled extractor with TaggedExtractor
	return &DisabledExtractor{}
}
