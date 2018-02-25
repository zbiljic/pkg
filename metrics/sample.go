package metrics

import (
	"math"
	"math/rand"
	"sort"
	"sync"
)

// Sample maintains a statistically-significant selection of values from
// a stream.
type Sample interface {
	Clear()
	Count() int64
	Max() int64
	Mean() float64
	Min() int64
	Percentile(float64) float64
	Percentiles([]float64) []float64
	Size() int
	Snapshot() Sample
	StdDev() float64
	Sum() int64
	Update(int64)
	Values() []int64
	Variance() float64
}

// NilSample is a no-op Sample.
type NilSample struct{}

// Clear is a no-op.
func (NilSample) Clear() {}

// Count is a no-op.
func (NilSample) Count() int64 { return 0 }

// Max is a no-op.
func (NilSample) Max() int64 { return 0 }

// Mean is a no-op.
func (NilSample) Mean() float64 { return 0.0 }

// Min is a no-op.
func (NilSample) Min() int64 { return 0 }

// Percentile is a no-op.
func (NilSample) Percentile(p float64) float64 { return 0.0 }

// Percentiles is a no-op.
func (NilSample) Percentiles(ps []float64) []float64 {
	return make([]float64, len(ps))
}

// Size is a no-op.
func (NilSample) Size() int { return 0 }

// Snapshot is a no-op.
func (NilSample) Snapshot() Sample { return NilSample{} }

// StdDev is a no-op.
func (NilSample) StdDev() float64 { return 0.0 }

// Sum is a no-op.
func (NilSample) Sum() int64 { return 0 }

// Update is a no-op.
func (NilSample) Update(v int64) {}

// Values is a no-op.
func (NilSample) Values() []int64 { return []int64{} }

// Variance is a no-op.
func (NilSample) Variance() float64 { return 0.0 }

// SampleSnapshot is a read-only copy of another Sample.
type SampleSnapshot struct {
	count  int64
	values []int64
}

// Clear panics.
func (*SampleSnapshot) Clear() {
	panic("Clear called on a SampleSnapshot")
}

// Count returns the count of inputs at the time the snapshot was taken.
func (s *SampleSnapshot) Count() int64 { return s.count }

// Max returns the maximal value at the time the snapshot was taken.
func (s *SampleSnapshot) Max() int64 { return SampleMax(s.values) }

// Mean returns the mean value at the time the snapshot was taken.
func (s *SampleSnapshot) Mean() float64 { return SampleMean(s.values) }

// Min returns the minimal value at the time the snapshot was taken.
func (s *SampleSnapshot) Min() int64 { return SampleMin(s.values) }

// Percentile returns an arbitrary percentile of values at the time the
// snapshot was taken.
func (s *SampleSnapshot) Percentile(p float64) float64 {
	return SamplePercentile(s.values, p)
}

// Percentiles returns a slice of arbitrary percentiles of values at the time
// the snapshot was taken.
func (s *SampleSnapshot) Percentiles(ps []float64) []float64 {
	return SamplePercentiles(s.values, ps)
}

// Size returns the size of the sample at the time the snapshot was taken.
func (s *SampleSnapshot) Size() int { return len(s.values) }

// Snapshot returns the snapshot.
func (s *SampleSnapshot) Snapshot() Sample { return s }

// StdDev returns the standard deviation of values at the time the snapshot was
// taken.
func (s *SampleSnapshot) StdDev() float64 { return SampleStdDev(s.values) }

// Sum returns the sum of values at the time the snapshot was taken.
func (s *SampleSnapshot) Sum() int64 { return SampleSum(s.values) }

// Update panics.
func (*SampleSnapshot) Update(int64) {
	panic("Update called on a SampleSnapshot")
}

// Values returns a copy of the values in the sample.
func (s *SampleSnapshot) Values() []int64 {
	values := make([]int64, len(s.values))
	copy(values, s.values)
	return values
}

// A UniformSample using Vitter's Algorithm R.
//
// <http://www.cs.umd.edu/~samir/498/vitter.pdf>
type UniformSample struct {
	count         int64
	mutex         sync.Mutex
	reservoirSize int
	values        []int64
}

// NewUniformSample constructs a new uniform sample with the given reservoir
// size.
func NewUniformSample(reservoirSize int) Sample {
	if UseNilMetrics {
		return NilSample{}
	}
	return &UniformSample{
		reservoirSize: reservoirSize,
		values:        make([]int64, 0, reservoirSize),
	}
}

// Clear clears all samples.
func (s *UniformSample) Clear() {
	s.mutex.Lock()
	defer s.mutex.Unlock()
	s.count = 0
	s.values = make([]int64, 0, s.reservoirSize)
}

// Count returns the number of samples recorded, which may exceed the
// reservoir size.
func (s *UniformSample) Count() int64 {
	s.mutex.Lock()
	defer s.mutex.Unlock()
	return s.count
}

// Max returns the maximum value in the sample, which may not be the maximum
// value ever to be part of the sample.
func (s *UniformSample) Max() int64 {
	s.mutex.Lock()
	defer s.mutex.Unlock()
	return SampleMax(s.values)
}

// Mean returns the mean of the values in the sample.
func (s *UniformSample) Mean() float64 {
	s.mutex.Lock()
	defer s.mutex.Unlock()
	return SampleMean(s.values)
}

// Min returns the minimum value in the sample, which may not be the minimum
// value ever to be part of the sample.
func (s *UniformSample) Min() int64 {
	s.mutex.Lock()
	defer s.mutex.Unlock()
	return SampleMin(s.values)
}

// Percentile returns an arbitrary percentile of values in the sample.
func (s *UniformSample) Percentile(p float64) float64 {
	s.mutex.Lock()
	defer s.mutex.Unlock()
	return SamplePercentile(s.values, p)
}

// Percentiles returns a slice of arbitrary percentiles of values in the
// sample.
func (s *UniformSample) Percentiles(ps []float64) []float64 {
	s.mutex.Lock()
	defer s.mutex.Unlock()
	return SamplePercentiles(s.values, ps)
}

// Size returns the size of the sample, which is at most the reservoir size.
func (s *UniformSample) Size() int {
	s.mutex.Lock()
	defer s.mutex.Unlock()
	return len(s.values)
}

// Snapshot returns a read-only copy of the sample.
func (s *UniformSample) Snapshot() Sample {
	s.mutex.Lock()
	defer s.mutex.Unlock()
	values := make([]int64, len(s.values))
	copy(values, s.values)
	return &SampleSnapshot{
		count:  s.count,
		values: values,
	}
}

// StdDev returns the standard deviation of the values in the sample.
func (s *UniformSample) StdDev() float64 {
	s.mutex.Lock()
	defer s.mutex.Unlock()
	return SampleStdDev(s.values)
}

// Sum returns the sum of the values in the sample.
func (s *UniformSample) Sum() int64 {
	s.mutex.Lock()
	defer s.mutex.Unlock()
	return SampleSum(s.values)
}

// Update samples a new value.
func (s *UniformSample) Update(v int64) {
	s.mutex.Lock()
	defer s.mutex.Unlock()
	s.count++
	if len(s.values) < s.reservoirSize {
		s.values = append(s.values, v)
	} else {
		r := rand.Int63n(s.count)
		if r < int64(len(s.values)) {
			s.values[int(r)] = v
		}
	}
}

// Values returns a copy of the values in the sample.
func (s *UniformSample) Values() []int64 {
	s.mutex.Lock()
	defer s.mutex.Unlock()
	values := make([]int64, len(s.values))
	copy(values, s.values)
	return values
}

// Variance returns the variance of the values in the sample.
func (s *UniformSample) Variance() float64 {
	s.mutex.Lock()
	defer s.mutex.Unlock()
	return SampleVariance(s.values)
}

// SampleMax returns the maximum value of the slice of int64.
func SampleMax(values []int64) int64 {
	if 0 == len(values) {
		return 0
	}
	var max int64 = math.MinInt64
	for _, v := range values {
		if max < v {
			max = v
		}
	}
	return max
}

// SampleMean returns the mean value of the slice of int64.
func SampleMean(values []int64) float64 {
	if 0 == len(values) {
		return 0.0
	}
	return float64(SampleSum(values)) / float64(len(values))
}

// SampleMin returns the minimum value of the slice of int64.
func SampleMin(values []int64) int64 {
	if 0 == len(values) {
		return 0
	}
	var min int64 = math.MaxInt64
	for _, v := range values {
		if min > v {
			min = v
		}
	}
	return min
}

// SamplePercentile returns an arbitrary percentile of the slice of int64.
func SamplePercentile(values int64Slice, p float64) float64 {
	return SamplePercentiles(values, []float64{p})[0]
}

// SamplePercentiles returns a slice of arbitrary percentiles of the slice of
// int64.
func SamplePercentiles(values int64Slice, ps []float64) []float64 {
	scores := make([]float64, len(ps))
	size := len(values)
	if size > 0 {
		sort.Sort(values)
		for i, p := range ps {
			pos := p * float64(size+1)
			if pos < 1.0 {
				scores[i] = float64(values[0])
			} else if pos >= float64(size) {
				scores[i] = float64(values[size-1])
			} else {
				lower := float64(values[int(pos)-1])
				upper := float64(values[int(pos)])
				scores[i] = lower + (pos-math.Floor(pos))*(upper-lower)
			}
		}
	}
	return scores
}

// Variance returns the variance of values at the time the snapshot was taken.
func (s *SampleSnapshot) Variance() float64 { return SampleVariance(s.values) }

// SampleStdDev returns the standard deviation of the slice of int64.
func SampleStdDev(values []int64) float64 {
	return math.Sqrt(SampleVariance(values))
}

// SampleSum returns the sum of the slice of int64.
func SampleSum(values []int64) int64 {
	var sum int64
	for _, v := range values {
		sum += v
	}
	return sum
}

// SampleVariance returns the variance of the slice of int64.
func SampleVariance(values []int64) float64 {
	if 0 == len(values) {
		return 0.0
	}
	m := SampleMean(values)
	var sum float64
	for _, v := range values {
		d := float64(v) - m
		sum += d * d
	}
	return sum / float64(len(values))
}

type int64Slice []int64

func (p int64Slice) Len() int           { return len(p) }
func (p int64Slice) Less(i, j int) bool { return p[i] < p[j] }
func (p int64Slice) Swap(i, j int)      { p[i], p[j] = p[j], p[i] }
