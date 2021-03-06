package statsd

import (
	"fmt"

	"github.com/DataDog/datadog-go/statsd"
	"github.com/DataDog/datadog-trace-agent/config"
)

// StatsClient represents a client capable of sending stats to some stat endpoint.
type StatsClient interface {
	Gauge(name string, value float64, tags []string, rate float64) error
	Count(name string, value int64, tags []string, rate float64) error
	Histogram(name string, value float64, tags []string, rate float64) error
}

// Client is a global Statsd client. When a client is configured via Configure,
// that becomes the new global Statsd client in the package.
var Client StatsClient = (*statsd.Client)(nil)

// Configure creates a statsd client for the given agent's configuration, using the specified global tags.
func Configure(conf *config.AgentConfig, tags []string) error {
	client, err := statsd.New(fmt.Sprintf("%s:%d", conf.StatsdHost, conf.StatsdPort))
	if err != nil {
		return err
	}
	client.Tags = tags
	Client = client
	return nil
}
