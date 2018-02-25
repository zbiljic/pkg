package metrics

import "testing"

func BenchmarkGaugeFloat64(b *testing.B) {
	g := NewGaugeFloat64()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		g.Update(float64(i))
	}
}
