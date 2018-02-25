package metrics

import (
	"fmt"
	"reflect"
	"sync"
)

// DuplicateMetric is the error returned by Registry.Register when a metric
// already exists.  If you mean to Register that metric you must first
// Unregister the existing metric.
type DuplicateMetric string

func (err DuplicateMetric) Error() string {
	return fmt.Sprintf("duplicate metric: %s", string(err))
}

// A Registry holds references to a set of metrics by name and can iterate
// over them, calling callback functions provided by the user.
//
// This is an interface so as to encourage other structs to implement
// the Registry API as appropriate.
type Registry interface {

	// Call the given function for each registered metric.
	Each(func(string, Metric))

	// Get the metric by the given name or nil if none is registered.
	Get(string) Metric

	// Gets an existing metric or registers the given one.
	// The interface can be the metric to register if not found in registry,
	// or a function returning the metric for lazy instantiation.
	GetOrRegister(string, Metric) Metric

	// Register the given metric under the given name.
	Register(string, Metric) error

	// Unregister the metric with the given name.
	Unregister(string)

	// Unregister all metrics.  (Mostly for testing.)
	UnregisterAll()
}

// StandardRegistry is the standard implementation of a Registry is
// a mutex-protected map of names to metrics.
type StandardRegistry struct {
	metrics map[string]Metric

	mutex sync.Mutex
}

// NewRegistry creates a new registry.
func NewRegistry() Registry {
	return &StandardRegistry{metrics: make(map[string]Metric)}
}

// Each calls the given function for each registered metric.
func (r *StandardRegistry) Each(fn func(string, Metric)) {
	for name, m := range r.registered() {
		fn(name, m)
	}
}

func (r *StandardRegistry) registered() map[string]Metric {
	r.mutex.Lock()
	defer r.mutex.Unlock()
	metrics := make(map[string]Metric, len(r.metrics))
	for name, m := range r.metrics {
		metrics[name] = m
	}
	return metrics
}

// Get the metric by the given name or nil if none is registered.
func (r *StandardRegistry) Get(name string) Metric {
	r.mutex.Lock()
	defer r.mutex.Unlock()
	return r.metrics[name]
}

// GetOrRegister gets an existing metric or creates and registers a new one.
// Threadsafe alternative to calling Get and Register on failure.
// The interface can be the metric to register if not found in registry,
// or a function returning the metric for lazy instantiation.
func (r *StandardRegistry) GetOrRegister(name string, m Metric) Metric {
	r.mutex.Lock()
	defer r.mutex.Unlock()
	if metric, ok := r.metrics[name]; ok {
		return metric
	}
	if v := reflect.ValueOf(m); v.Kind() == reflect.Func {
		m = v.Call(nil)[0].Interface()
	}
	r.register(name, m)
	return m
}

// Register the given metric under the given name.  Returns a DuplicateMetric
// if a metric by the given name is already registered.
func (r *StandardRegistry) Register(name string, m Metric) error {
	r.mutex.Lock()
	defer r.mutex.Unlock()
	return r.register(name, m)
}

// Unregister the metric with the given name.
func (r *StandardRegistry) Unregister(name string) {
	r.mutex.Lock()
	defer r.mutex.Unlock()
	delete(r.metrics, name)
}

// UnregisterAll unregisters all metrics.  (Mostly for testing.)
func (r *StandardRegistry) UnregisterAll() {
	r.mutex.Lock()
	defer r.mutex.Unlock()
	for name := range r.metrics {
		delete(r.metrics, name)
	}
}

func (r *StandardRegistry) register(name string, m Metric) error {
	if _, ok := r.metrics[name]; ok {
		return DuplicateMetric(name)
	}
	r.metrics[name] = m
	return nil
}

var DefaultRegistry Registry = NewRegistry()

// Each calls the given function for each registered metric.
func Each(fn func(string, Metric)) {
	DefaultRegistry.Each(fn)
}

// Get the metric by the given name or nil if none is registered.
func Get(name string) Metric {
	return DefaultRegistry.Get(name)
}

// GetOrRegister gets an existing metric or creates and registers a new one.
// Threadsafe alternative to calling Get and Register on failure.
func GetOrRegister(name string, m Metric) Metric {
	return DefaultRegistry.GetOrRegister(name, m)
}

// Register the given metric under the given name.  Returns a DuplicateMetric
// if a metric by the given name is already registered.
func Register(name string, m Metric) error {
	return DefaultRegistry.Register(name, m)
}

// MustRegister register the given metric under the given name.  Panics if
// a metric by the given name is already registered.
func MustRegister(name string, m Metric) {
	if err := Register(name, m); err != nil {
		panic(err)
	}
}

// Unregister the metric with the given name.
func Unregister(name string) {
	DefaultRegistry.Unregister(name)
}
