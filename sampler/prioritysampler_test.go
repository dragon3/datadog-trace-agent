package sampler

import (
	"math"
	"math/rand"
	"testing"

	log "github.com/cihub/seelog"

	"github.com/DataDog/datadog-trace-agent/config"
	"github.com/DataDog/datadog-trace-agent/model"
	"github.com/stretchr/testify/assert"
)

const (
	testServiceA = "service-a"
	testServiceB = "service-b"
)

func getTestPriorityEngine() *PriorityEngine {
	// Disable debug logs in these tests
	log.UseLogger(log.Disabled)

	// No extra fixed sampling, no maximum TPS
	extraRate := 1.0
	maxTPS := 0.0

	rateByService := config.RateByService{}
	return NewPriorityEngine(extraRate, maxTPS, &rateByService)
}

func getTestTraceWithService(t *testing.T, service string, s *PriorityEngine) (model.Trace, *model.Span) {
	tID := randomTraceID()
	trace := model.Trace{
		model.Span{TraceID: tID, SpanID: 1, ParentID: 0, Start: 42, Duration: 1000000, Service: service, Type: "web", Meta: map[string]string{"env": defaultEnv}},
		model.Span{TraceID: tID, SpanID: 2, ParentID: 1, Start: 100, Duration: 200000, Service: service, Type: "sql"},
	}
	r := rand.Float64()
	priority := 0.0
	rates := s.getRateByService()
	key := byServiceKey(trace[0].Service, defaultEnv)
	var rate float64
	if r, ok := rates[key]; ok {
		rate = r
	} else {
		rate = 1
	}
	if r <= rate {
		priority = 1
	}
	trace[0].Metrics = map[string]float64{samplingPriorityKey: priority}
	return trace, &trace[0]
}

func TestMaxTPSByService(t *testing.T) {
	// Test the "effectiveness" of the maxTPS option.
	assert := assert.New(t)
	s := getTestPriorityEngine()

	type testCase struct {
		maxTPS        float64
		tps           float64
		relativeError float64
	}
	testCases := []testCase{
		{maxTPS: 10.0, tps: 20.0, relativeError: 0.2},
	}
	if !testing.Short() {
		testCases = append(testCases,
			testCase{maxTPS: 5.0, tps: 50.0, relativeError: 0.2},
			testCase{maxTPS: 3.0, tps: 200.0, relativeError: 0.2},
			testCase{maxTPS: 1.0, tps: 1000.0, relativeError: 0.2},
			testCase{maxTPS: 10.0, tps: 10.0, relativeError: 0.001},
			testCase{maxTPS: 10.0, tps: 3.0, relativeError: 0.001})
	}

	// To avoid the edge effects from an non-initialized sampler, wait a bit before counting samples.
	const (
		initPeriods = 50
		periods     = 500
	)

	for _, tc := range testCases {
		t.Logf("testing maxTPS=%0.1f tps=%0.1f", tc.maxTPS, tc.tps)
		s.Sampler.maxTPS = tc.maxTPS
		periodSeconds := s.Sampler.Backend.decayPeriod.Seconds()
		tracesPerPeriod := tc.tps * periodSeconds
		// Set signature score offset high enough not to kick in during the test.
		s.Sampler.signatureScoreOffset = 2 * tc.tps
		s.Sampler.signatureScoreFactor = math.Pow(s.Sampler.signatureScoreSlope, math.Log10(s.Sampler.signatureScoreOffset))

		sampledCount := 0
		handledCount := 0

		for period := 0; period < initPeriods+periods; period++ {
			s.Sampler.Backend.DecayScore()
			s.Sampler.AdjustScoring()
			for i := 0; i < int(tracesPerPeriod); i++ {
				trace, root := getTestTraceWithService(t, "service-a", s)
				sampled := s.Sample(trace, root, defaultEnv)
				// Once we got into the "supposed-to-be" stable "regime", count the samples
				if period > initPeriods {
					handledCount++
					if sampled {
						sampledCount++
					}
				}
			}
		}

		// When tps is lower than maxTPS it means that we are actually not sampling
		// anything, so the target is the original tps, and not maxTPS.
		// Also, in that case, results should be more precise.
		targetTPS := tc.maxTPS
		relativeError := 0.01
		if tc.maxTPS > tc.tps {
			targetTPS = tc.tps
		} else {
			relativeError = 0.1 + s.Sampler.Backend.decayFactor - 1
		}

		// Check that the sampled score is roughly equal to maxTPS. This is different from
		// the score sampler test as here we run adjustscoring on a regular basis so the converges to maxTPS.
		assert.InEpsilon(targetTPS, s.Sampler.Backend.GetSampledScore(), relativeError)

		// We should have keep the right percentage of traces
		assert.InEpsilon(targetTPS/tc.tps, float64(sampledCount)/float64(handledCount), relativeError)

		// We should have a throughput of sampled traces around maxTPS
		// Check for 1% epsilon, but the precision also depends on the backend imprecision (error factor = decayFactor).
		// Combine error rates with L1-norm instead of L2-norm by laziness, still good enough for tests.
		assert.InEpsilon(targetTPS, float64(sampledCount)/(float64(periods)*periodSeconds), relativeError)
	}
}

// Ensure PriorityEngine implements engine.
var testPriorityEngine Engine = &PriorityEngine{}
