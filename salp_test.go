package salp_test

import (
	"strconv"
	"testing"
	"time"

	"github.com/mmcshane/salp"
)

func TestProbeWithTooManyArgs(t *testing.T) {
	pv := salp.NewProvider("foo")
	defer salp.UnloadAndDispose(pv)

	_, err := pv.AddProbe("bar", salp.Int8, salp.Int8,
		salp.Int8, salp.Int8, salp.Int8, salp.Int8)
	require(t, err == nil, "unexpected error defining probe with 6 args")

	_, err = pv.AddProbe("baz", salp.Int8, salp.Int8,
		salp.Int8, salp.Int8, salp.Int8, salp.Int8, salp.Int8)
	require(t, err != nil, "expected error defining probe with 7 args")
}

func TestDryFireAProbe(t *testing.T) {
	pv := salp.NewProvider("foo")
	defer salp.UnloadAndDispose(pv)
	pr, err := pv.AddProbe("bar", salp.String, salp.Int32)

	require(t, err == nil, err)
	require(t, !pr.Enabled(), "expected untraced probe to be disabled")

	err = pv.Load()
	require(t, err == nil, err)

	// wrong arity is not an error
	pr.Fire("bar")

	// unsupported types do not cause a crash
	pr.Fire(struct{}{}, struct{}{})

	pr.Fire("bar", 3)
}

func TestProviderName(t *testing.T) {
	pv := salp.NewProvider("foo")
	require(t, pv.Name() == "foo")
}

func TestProbeName(t *testing.T) {
	pv := salp.NewProvider("foo")
	pr, err := pv.AddProbe("bar")

	require(t, err == nil, err)
	require(t, pr.Name() == "bar")
}

var result bool

func BenchmarkEnabled(b *testing.B) {
	pv := salp.NewProvider("foo")
	pr := salp.MustAddProbe(pv, "bar")
	salp.MustLoadProvider(pv)

	var tmp bool
	for i := 0; i < b.N; i++ {
		tmp = pr.Enabled()
	}
	result = tmp
}

func BenchmarkFireDisabled(b *testing.B) {
	pv := salp.NewProvider("foo")
	pr := salp.MustAddProbe(pv, "bar", salp.Int32)
	salp.MustLoadProvider(pv)

	for i := 0; i < b.N; i++ {
		pr.Fire(3)
	}
}

func require(t *testing.T, b bool, msgs ...interface{}) {
	t.Helper()
	if !b {
		t.Fatal(msgs)
	}
}

func Example() {
	// Provider and probe creation should occur early on, probably during
	// initialization

	// Create a provider to which we will attach probes. The provider
	// acts as a namespace & container for probe instances.
	provider := salp.NewProvider("my-example-provider")
	defer salp.UnloadAndDispose(provider)

	// Create a probe that can be fired with 4 args: a string, a uint8,
	// an int16, and another string
	probe1 := salp.MustAddProbe(provider, "my-example-probe",
		salp.String, salp.Uint8, salp.Int16, salp.String)

	// Create a second probe that takes only a single string argument
	probe2 := salp.MustAddProbe(provider, "my-other-examaple-probe", salp.String)

	// Now that the probes have been created, enable the provider by calling
	// Load().
	salp.MustLoadProvider(provider)

	// Initialization of our provider and 2 probes is now complete, the probes
	// are ready to be fired. Firing probes happens after initialization, inline
	// with execution of your program.

	// Fire both probes 10 times
	for i := 0; i < 10; i++ {
		probe1.Fire(strconv.Itoa(i), 5, 10, "foo")
		probe2.Fire(time.Now().Format(time.RFC1123))
		time.Sleep(1 * time.Second)
	}
}
