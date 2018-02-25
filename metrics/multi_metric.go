package metrics

import (
	"reflect"
	"sync"
)

// MultiMetric is a Metric that has a name, multiple metrics, and tags.
type MultiMetric interface {
	Metric

	GetOrAdd(string, Metric) Metric
	Metrics() map[string]Metric
	Snapshot() MultiMetric
	Tags() map[string]string
}

// GetOrRegisterMultiMetric returns an existing MultiMetric or constructs and
// registers a new StandardMultiMetric.
func GetOrRegisterMultiMetric(name string, tags map[string]string, r Registry) MultiMetric {
	if nil == r {
		r = DefaultRegistry
	}
	return r.GetOrRegister(name, func() MultiMetric { return NewMultiMetric(tags) }).(MultiMetric)
}

// NewMultiMetric constructs a new StandardMultiMetric.
func NewMultiMetric(tags map[string]string) MultiMetric {
	if UseNilMetrics {
		return NilMultiMetric{}
	}
	return &StandardMultiMetric{
		metrics: make(map[string]Metric),
		tags:    tags,
	}
}

// NewRegisteredMultiMetric constructs and registers a new StandardMultiMetric.
func NewRegisteredMultiMetric(name string, tags map[string]string, r Registry) MultiMetric {
	c := NewMultiMetric(tags)
	if nil == r {
		r = DefaultRegistry
	}
	r.Register(name, c)
	return c
}

// MultiMetricSnapshot is a read-only copy of another MultiMetric.
type MultiMetricSnapshot struct {
	m MultiMetric
}

// GetOrAdd panics.
func (mm *MultiMetricSnapshot) GetOrAdd(name string, m Metric) Metric {
	panic("GetOrAdd called on a MultiMetricSnapshot")
}

// Metrics returns all the metrics in the multi metric as a map.
func (mm *MultiMetricSnapshot) Metrics() map[string]Metric {
	return mm.m.Metrics()
}

// Snapshot returns the snapshot.
func (mm *MultiMetricSnapshot) Snapshot() MultiMetric {
	return mm
}

// Tags returns tag map for the multi metric.
func (mm *MultiMetricSnapshot) Tags() map[string]string {
	return mm.m.Tags()
}

// NilMultiMetric is a no-op MultiMetric.
type NilMultiMetric struct{}

// GetOrAdd is a no-op.
func (NilMultiMetric) GetOrAdd(name string, m Metric) Metric { return nil }

// Metrics is a no-op.
func (NilMultiMetric) Metrics() map[string]Metric { return map[string]Metric{} }

// Snapshot is a no-op.
func (NilMultiMetric) Snapshot() MultiMetric { return NilMultiMetric{} }

// Tags is a no-op.
func (NilMultiMetric) Tags() map[string]string { return map[string]string{} }

// StandardMultiMetric is the standard implementation of a MultiMetric.
type StandardMultiMetric struct {
	metrics map[string]Metric
	tags    map[string]string
	mu      sync.Mutex
}

// GetOrAdd gets an existing metric or adds a new one to multi metric.
func (mm *StandardMultiMetric) GetOrAdd(name string, m Metric) Metric {
	mm.mu.Lock()
	defer mm.mu.Unlock()
	if metric, ok := mm.metrics[name]; ok {
		return metric
	}
	if v := reflect.ValueOf(m); v.Kind() == reflect.Func {
		m = v.Call(nil)[0].Interface()
	}
	mm.metrics[name] = m
	return m
}

// Metrics returns all the metrics in the multi metric as a map.
//
// Returned map should not be changed by the user.
func (mm *StandardMultiMetric) Metrics() map[string]Metric {
	return mm.metrics
}

// Snapshot returns a read-only copy of the multi metric.
func (mm *StandardMultiMetric) Snapshot() MultiMetric {
	mm.mu.Lock()
	defer mm.mu.Unlock()
	metrics := make(map[string]Metric)
	for k, v := range mm.metrics {
		switch metric := v.(type) {
		// WHEN NEW METRIC IS ADDED, CASE FOR IT MUST BE ADDED HERE ALSO.
		case Counter:
			metrics[k] = metric.Snapshot()
		case Gauge:
			metrics[k] = metric.Snapshot()
		case GaugeFloat64:
			metrics[k] = metric.Snapshot()
		case Histogram:
			metrics[k] = metric.Snapshot()
		}
	}
	tags := make(map[string]string)
	for k, v := range mm.tags {
		tags[k] = v
	}
	return &MultiMetricSnapshot{&StandardMultiMetric{metrics: metrics, tags: tags}}
}

// Tags returns tag map for the multi metric.
//
// Returned map should not be changed by the user.
func (mm *StandardMultiMetric) Tags() map[string]string {
	return mm.tags
}
