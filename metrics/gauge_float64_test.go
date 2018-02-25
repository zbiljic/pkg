package metrics

import "testing"

// Check the interfaces are satisfied
func TestGaugeFloat64_impl(t *testing.T) {
	var _ GaugeFloat64 = new(NilGaugeFloat64)
	var _ GaugeFloat64 = new(GaugeFloat64Snapshot)
	var _ GaugeFloat64 = new(StandardGaugeFloat64)
}

func TestGaugeFloat64(t *testing.T) {
	g := NewGaugeFloat64()
	g.Update(float64(47.0))
	if v := g.Value(); float64(47.0) != v {
		t.Errorf("g.Value(): 47.0 != %v\n", v)
	}
}

func TestGaugeFloat64Snapshot(t *testing.T) {
	g := NewGaugeFloat64()
	g.Update(float64(47.0))
	snapshot := g.Snapshot()
	g.Update(float64(0))
	if v := snapshot.Value(); float64(47.0) != v {
		t.Errorf("g.Value(): 47.0 != %v\n", v)
	}
}

func TestGetOrRegisterGaugeFloat64(t *testing.T) {
	r := NewRegistry()
	NewRegisteredGaugeFloat64("foo", r).Update(float64(47.0))
	t.Logf("registry: %v", r)
	if g := GetOrRegisterGaugeFloat64("foo", r); float64(47.0) != g.Value() {
		t.Fatal(g)
	}
}
