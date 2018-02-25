package metrics

import "testing"

func BenchmarkRegistry(b *testing.B) {
	r := NewRegistry()
	r.Register("foo", NewCounter())
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		r.Each(func(string, Metric) {})
	}
}
