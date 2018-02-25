package metrics

import "testing"

func BenchmarkHistogram(b *testing.B) {
	h := NewHistogram(NewUniformSample(100))
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		h.Update(int64(i))
	}
}
