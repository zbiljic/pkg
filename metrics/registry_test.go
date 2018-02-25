package metrics

import "testing"

// Check the interfaces are satisfied
func TestRegistry_impl(t *testing.T) {
	var _ Registry = new(StandardRegistry)
}

func TestRegistry(t *testing.T) {
	r := NewRegistry()
	r.Register("foo", NewCounter())
	i := 0
	r.Each(func(name string, metric Metric) {
		i++
		if "foo" != name {
			t.Fatal(name)
		}
		if _, ok := metric.(Counter); !ok {
			t.Fatal(metric)
		}
	})
	if 1 != i {
		t.Fatal(i)
	}
	r.Unregister("foo")
	i = 0
	r.Each(func(string, Metric) { i++ })
	if 0 != i {
		t.Fatal(i)
	}
}

func TestRegistryDuplicate(t *testing.T) {
	r := NewRegistry()
	if err := r.Register("foo", NewCounter()); nil != err {
		t.Fatal(err)
	}
	if err := r.Register("foo", NewGauge()); nil == err {
		t.Fatal(err)
	}
	i := 0
	r.Each(func(name string, metric Metric) {
		i++
		if _, ok := metric.(Counter); !ok {
			t.Fatal(metric)
		}
	})
	if 1 != i {
		t.Fatal(i)
	}
}

func TestRegistryGet(t *testing.T) {
	r := NewRegistry()
	r.Register("foo", NewCounter())
	if count := r.Get("foo").(Counter).Count(); 0 != count {
		t.Fatal(count)
	}
	r.Get("foo").(Counter).Inc(1)
	if count := r.Get("foo").(Counter).Count(); 1 != count {
		t.Fatal(count)
	}
}

func TestRegistryGetOrRegister(t *testing.T) {
	r := NewRegistry()

	// First metric wins with GetOrRegister
	_ = r.GetOrRegister("foo", NewCounter())
	m := r.GetOrRegister("foo", NewGauge())
	if _, ok := m.(Counter); !ok {
		t.Fatal(m)
	}

	i := 0
	r.Each(func(name string, metric Metric) {
		i++
		if name != "foo" {
			t.Fatal(name)
		}
		if _, ok := metric.(Counter); !ok {
			t.Fatal(metric)
		}
	})
	if i != 1 {
		t.Fatal(i)
	}
}

func TestRegistryGetOrRegisterWithLazyInstantiation(t *testing.T) {
	r := NewRegistry()

	// First metric wins with GetOrRegister
	_ = r.GetOrRegister("foo", NewCounter)
	m := r.GetOrRegister("foo", NewGauge)
	if _, ok := m.(Counter); !ok {
		t.Fatal(m)
	}

	i := 0
	r.Each(func(name string, metric Metric) {
		i++
		if name != "foo" {
			t.Fatal(name)
		}
		if _, ok := metric.(Counter); !ok {
			t.Fatal(metric)
		}
	})
	if i != 1 {
		t.Fatal(i)
	}
}

func TestRegistryUnregister(t *testing.T) {
	r := NewRegistry()
	r.Register("foo", NewCounter())
	r.Register("bar", NewGauge())
	r.Register("baz", NewGaugeFloat64())
	var l int
	l = 0
	r.Each(func(name string, metric Metric) {
		l++
	})
	if l != 3 {
		t.Errorf("metrics: %d != %d\n", 3, l)
	}
	r.Unregister("foo")
	r.Unregister("bar")
	r.Unregister("baz")
	l = 0
	r.Each(func(name string, metric Metric) {
		l++
	})
	if l != 0 {
		t.Errorf("metrics: %d != %d\n", 0, l)
	}
}
