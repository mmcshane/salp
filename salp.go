// Package salp enables the definition and firing of USDT probes at runtime by
// Go programs running on Linux. These probes impose little or no overhead when
// not in use and are available for use by any tool that is able to monitor USDT
// probe points (e.g. the trace tool from the bcc project).
package salp

/*
#cgo LDFLAGS: -lstapsdt

#include <stdint.h>
#include <stdlib.h>
#include <libstapsdt.h>

#if 6 != MAX_ARGUMENTS
#	error "libstapsdt max arguments has changed (expected 6)"
#endif

// This wrapper is necessary only because CGO cannot invoke variadic functions
SDTProbe_t* salp_providerAddProbe(
		SDTProvider_t* p, const char* name, uint32_t c, ArgType_t* at){
	switch(c) {
		case 0:
			return providerAddProbe(p, name, 0);
		case 1:
			return providerAddProbe(p, name, 1, at[0]);
		case 2:
			return providerAddProbe(p, name, 2, at[0], at[1]);
		case 3:
			return providerAddProbe(p, name, 3, at[0], at[1], at[2]);
		case 4:
			return providerAddProbe(p, name, 4, at[0], at[1], at[2], at[3]);
		case 5:
			return providerAddProbe(
				p, name, 5, at[0], at[1], at[2], at[3], at[4]);
		case 6:
			return providerAddProbe(
				p, name, 5, at[0], at[1], at[2], at[3], at[4], at[5]);
		default:
			return NULL;
	}
}

// This wrapper is necessary only because CGO cannot invoke variadic functions
void salp_probeFire(SDTProbe_t* p, void** args) {
	switch(p->argCount) {
		case 0:
			probeFire(p);
			return;
		case 1:
			probeFire(p, args[0]);
			return;
		case 2:
			probeFire(p, args[0], args[1]);
			return;
		case 3:
			probeFire(p, args[0], args[1], args[2]);
			return;
		case 4:
			probeFire(p, args[0], args[1], args[2], args[3]);
			return;
		case 5:
			probeFire(p, args[0], args[1], args[2], args[3], args[4]);
			return;
		case 6:
			probeFire(p, args[0], args[1], args[2], args[3], args[4], args[5]);
			return;
	}
}

*/
import "C"
import (
	"fmt"
	"unsafe"
)

type stapsdtError struct {
	code int
	msg  string
}

// Provider represents a named collection of probes
type Provider = C.struct_SDTProvider

// Probe is a location in Go code that can be "fired" with a set of arguments
// such that extrenal tools (e.g. the `trace` tool from bcc) can attach to a
// running process and inspect the values at runtime.
type Probe = C.struct_SDTProbe

// ProbeArgType specifies the type of each individual parameter than can be
// specified when firing a Probe.
type ProbeArgType C.ArgType_t

// ProbeArgTypes are used to specify the type of parameters accepted when firing
// a Probe
const (
	// Probe argument should be treated as a uint8
	Uint8 = ProbeArgType(C.uint8)

	// Probe argument should be treated as a bool
	Bool = Uint8

	//Probe argument should be treated as a byte
	Byte = Uint8

	// Probe argument should be treated as an int8
	Int8 = ProbeArgType(C.int8)

	// Probe argument should be treated as a uint16
	Uint16 = ProbeArgType(C.uint16)

	// Probe argument should be treated as an int16
	Int16 = ProbeArgType(C.int16)

	// Probe argument should be treated as a uint32
	Uint32 = ProbeArgType(C.uint32)

	// Probe argument should be treated as an int32
	Int32 = ProbeArgType(C.int32)

	// Probe argument should be treated as a uint64
	Uint64 = ProbeArgType(C.uint64)

	// Probe argument should be treated as an int64
	Int64 = ProbeArgType(C.int64)

	// Probe argument should be treated as a uint64
	String = ProbeArgType(C.uint64)

	// Probe argument should be treated as a Go error
	Error = String
)

// Error returns a string describing the error condition. The string will
// include an error code and a message.
func (e stapsdtError) Error() string {
	return fmt.Sprintf("libstapsdt error [%v]: %v", e.code, e.msg)
}

// NewProvider creates a libstapsdt error probe provider with the supplied name.
// Provider instances are in either a loaded or an unloaded state. When Provders
// are unloaded (their initial state), probes can be created via AddProbe. Once
// the Provider is loaded via the Load() function, the probe set should not be
// changed. Probes can be cleared from the Provider instance by unloading it via
// the Unload() function.  Probe addition is not threadsafe steps must be taken
// by clients of this library to ensure that at most one thread is adding a
// Probe at a time.
func NewProvider(name string) *Provider {
	clbl := C.CString(name)
	defer C.free(unsafe.Pointer(clbl))
	return C.providerInit(clbl)
}

// Name returns the name of the provider as a string
func (p *Provider) Name() string {
	return C.GoString(p.name)
}

func (p *Provider) err() error {
	return stapsdtError{
		code: int(p.errno),
		msg:  C.GoString(p.error),
	}
}

// Load transitions the provider from the unloaded state into the loaded state
// which causes associated Probes to become active (i.e. calling Fire() on the
// probe will actually work).
func (p *Provider) Load() error {
	rc := C.providerLoad(p)
	if int(rc) != 0 {
		return p.err()
	}
	return nil
}

// AddProbe creates a new Probe instance with the supplied name and assiciates
// it with this Provider. The argTypes describe the arguments that are expected
// to be supplied when Fire is called on this Probe.
func (p *Provider) AddProbe(name string, argTypes ...ProbeArgType) (*Probe, error) {
	cname := C.CString(name)
	defer C.free(unsafe.Pointer(cname))

	var pp *Probe
	if l := len(argTypes); l == 0 {
		pp = C.salp_providerAddProbe(p, cname, 0, nil)
	} else {
		pp = C.salp_providerAddProbe(
			p, cname, C.uint32_t(l), (*C.ArgType_t)(&argTypes[0]))
	}
	if pp == nil {
		return nil, p.err()
	}
	return pp, nil
}

// MustAddProbe is a helper function that either adds a probe with the supplied
// name and argument types to the specified provider or, in the case of an
// error, panics.
func MustAddProbe(p *Provider, name string, argTypes ...ProbeArgType) *Probe {
	prb, err := p.AddProbe(name, argTypes...)
	if err != nil {
		panic(err)
	}
	return prb
}

// MustLoadProvider is a helper function the either calls Load() on the supplied
// Provider instance or in the case of an error, panics
func MustLoadProvider(p *Provider) {
	err := p.Load()
	if err != nil {
		panic(err)
	}
}

// Unload transitions this Provider from the loaded to the unloaded state.
// Associated probes are detached and must be re-attached in order to function.
func (p *Provider) Unload() {
	C.providerUnload(p)
}

// Dispose cleans up the Provider datastructures and frees associated memory
// from the underlying C library (libstapsdt). The Provider instance is useless
// after this method is invoked.
func (p *Provider) Dispose() {
	C.providerDestroy(p)
}

// Enabled returns true iff the provider associated with this Probe is in a
// loaded state and the Probe is being monitored by an agent such as the bcc
// trace tool.
func (p *Probe) Enabled() bool {
	// don't do this ...
	//return int(C.probeIsEnabled(p)) == 1

	// ~100x lower overhead for this implementation, probably due to avoiding
	// the CGO context switch and making inlining possible.
	ptr := p._fire
	return uintptr(ptr) != 0 && *(*uint8)(ptr)&0x90 != 0x90
}

// Fire invokes the Probe with the provided arguments. The type and arity of
// this invocation should match what was described by the ProbeArgType arguments
// originally given to the Provider.AddProbe invocation that created this Probe.
// Cheap to invoke if the probe is not enabled (see: Enabled())
func (p *Probe) Fire(args ...interface{}) {
	if !p.Enabled() || len(args) != int(p.argCount) {
		return
	}

	// Fire is ~28% faster (linux, x86_64, go 1,10.3) to call with no probes
	// attached if fireImpl is a separate function. I assume this has to do with
	// compiler inlining but who cares eabout the reason?  Most of the time Fire
	// will be calld _without_ a probe attached so we're going to optimize for
	// that case.

	p.fireImpl(args...)
}

func (p *Probe) fireImpl(args ...interface{}) {
	ba := [6]unsafe.Pointer{}
	for i := 0; i < len(args); i++ {
		switch ta := args[i].(type) {
		case bool:
			var arg uint8
			if ta {
				arg = 1
			}
			ba[i] = unsafe.Pointer(uintptr(arg))
		case int8:
			ba[i] = unsafe.Pointer(uintptr(ta))
		case uint8: // catches byte
			ba[i] = unsafe.Pointer(uintptr(ta))
		case int16:
			ba[i] = unsafe.Pointer(uintptr(ta))
		case uint16:
			ba[i] = unsafe.Pointer(uintptr(ta))
		case int:
			ba[i] = unsafe.Pointer(uintptr(ta))
		case uint:
			ba[i] = unsafe.Pointer(uintptr(ta))
		case int32:
			ba[i] = unsafe.Pointer(uintptr(ta))
		case uint32:
			ba[i] = unsafe.Pointer(uintptr(ta))
		case int64:
			ba[i] = unsafe.Pointer(uintptr(ta))
		case uint64:
			ba[i] = unsafe.Pointer(uintptr(ta))
		case uintptr:
			ba[i] = unsafe.Pointer(ta)
		case string:
			strptr := unsafe.Pointer(C.CString(ta))
			defer C.free(strptr)
			ba[i] = strptr
		case error:
			cstr := unsafe.Pointer(C.CString(ta.Error()))
			defer C.free(cstr)
			ba[i] = cstr
		default:
			return
		}
	}
	C.salp_probeFire(p, &ba[0])
}

// Name gets the name of this Probe as provided when it was originally created.
func (p *Probe) Name() string {
	return C.GoString(p.name)
}

// UnloadAndDispose is a convenience function suitable for deferred invocation
// that calls p.Unload() and then p.Dispose().
func UnloadAndDispose(p *Provider) {
	p.Unload()
	p.Dispose()
}
