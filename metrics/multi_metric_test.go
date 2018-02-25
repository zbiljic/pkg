package metrics

import (
	"testing"
)

// Check the interfaces are satisfied
func TestMultiMetric_impl(t *testing.T) {
	var _ MultiMetric = new(NilMultiMetric)
	var _ MultiMetric = new(MultiMetricSnapshot)
	var _ MultiMetric = new(StandardMultiMetric)
}

func TestMultiMetricSnapshot(t *testing.T) {
	mm := NewMultiMetric(map[string]string{})
	c := NewCounter()
	mm.GetOrAdd("counter", c)
	c.Inc(1)
	g := NewGauge()
	mm.GetOrAdd("gauge", g)
	g.Update(int64(47))
	gf64 := NewGaugeFloat64()
	mm.GetOrAdd("gaugeFloat64", gf64)
	gf64.Update(float64(47.0))

	snapshot := mm.Snapshot()
	metrics := snapshot.Metrics()

	if metricsCount := len(metrics); metricsCount != 3 {
		t.Errorf("len(snapshot.Metrics()): 3 != %d\n", metricsCount)
	}

	c.Inc(1)
	switch v := metrics["counter"].(type) {
	case Counter:
		if 1 != v.Count() {
			t.Errorf("counter.Count(): 1 != %d\n", v.Count())
		}
	default:
		t.Errorf("metric is not 'Counter': %v\n", v)
	}

	g.Update(int64(74))
	switch v := metrics["gauge"].(type) {
	case Gauge:
		if 47 != v.Value() {
			t.Errorf("gauge.Value(): 47 != %d\n", v.Value())
		}
	default:
		t.Errorf("metric is not 'Gauge': %v\n", v)
	}

	gf64.Update(float64(74.0))
	switch v := metrics["gaugeFloat64"].(type) {
	case GaugeFloat64:
		if 47 != v.Value() {
			t.Errorf("gaugeFloat64.Value(): 47 != %f\n", v.Value())
		}
	default:
		t.Errorf("metric is not 'GaugeFloat64': %v\n", v)
	}

}
