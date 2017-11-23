package salp_test

import (
	"testing"

	"github.com/mmcshane/salp"
)

func TestDryFireAProbe(t *testing.T) {
	pv := salp.MakeProvider("foo")
	pr, err := pv.AddProbe("bar")

	require(t, err == nil, err)
	require(t, !pr.Enabled(), "expected untraced probe to be disabled")

	err = pv.Load()
	require(t, err == nil, err)

	defer func() {
		pv.Unload()
		pv.Dispose()
	}()

	pr.Fire()
}

func require(t *testing.T, b bool, msgs ...interface{}) {
	t.Helper()
	if !b {
		t.Fatal(msgs)
	}
}
