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
		default:
			// any args after the first six are ignored
			return providerAddProbe(
				p, name, 6, at[0], at[1], at[2], at[3], at[4], at[5]);
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
type Provider struct {
	p *C.struct_SDTProvider
}

// Probe is a location in Go code that can be "fired" with a set of arguments
// such that extrenal tools (e.g. the `trace` tool from bcc) can attach to a
// running process and inspect the values at runtime.
type Probe struct {
	p *C.struct_SDTProbe
}

// ProbeArgType specifies the type of each individual parameter than can be
// specified when firing a Probe.
type ProbeArgType C.ArgType_t

// ProbeArgTypes are used to specify the type of parameters accepted when firing
// a Probe
const (
	Uint8  = ProbeArgType(C.uint8)
	Int8   = ProbeArgType(C.int8)
	Uint16 = ProbeArgType(C.uint16)
	Int16  = ProbeArgType(C.int16)
	Uint32 = ProbeArgType(C.uint32)
	Int32  = ProbeArgType(C.int32)
	Uint64 = ProbeArgType(C.uint64)
	Int64  = ProbeArgType(C.int64)
	String = ProbeArgType(C.uint64)
)

// Error returns a string describing the error condition. The string will
// include an error code and a message.
func (e stapsdtError) Error() string {
	return fmt.Sprintf("libstapsdt error [%v]: %v", e.code, e.msg)
}

// MakeProvider creates a libstapsdt error probe provider with the supplied
// name. The returned type is a reference type (like the golang map or slice
// type) in that it is small enough to pass by value and can be copied without
// making a copy of the underlying provider. Provider instances are in either a
// loaded or an unloaded state. When Provders are unloaded (their initial
// state), probes can be created via AddProbe. Once the Provider is loaded via
// the Load() function, the probe set should not be changed. Probes can be
// cleared from the Provider instance by unloading it via the Unload() function.
// Probe addition is not threadsafe steps must be taken by clients of this
// library to ensure that at most one thread is adding a Probe at a time.
func MakeProvider(name string) Provider {
	clbl := C.CString(name)
	defer C.free(unsafe.Pointer(clbl))
	return Provider{C.providerInit(clbl)}
}

// Name returns the name of the provider as a string
func (p *Provider) Name() string {
	return C.GoString(p.p.name)
}

func (p *Provider) err() error {
	return stapsdtError{
		code: int(p.p.errno),
		msg:  C.GoString(p.p.error),
	}
}

// Load transitions the provider from the unloaded state into the loaded state
// which causes associated Probes to become active (i.e. calling Fire() on the
// probe will actually work).
func (p *Provider) Load() error {
	rc := C.providerLoad(p.p)
	if int(rc) != 0 {
		return p.err()
	}
	return nil
}

// AddProbe creates a new Probe instance with the supplied name and assiciates
// it with this Provider. The argTypes describe the arguments that are expected
// to be supplied when Fire is called on this Probe.
func (p *Provider) AddProbe(name string, argTypes ...ProbeArgType) (Probe, error) {
	cname := C.CString(name)
	defer C.free(unsafe.Pointer(cname))

	var pp *C.struct_SDTProbe
	if l := len(argTypes); l == 0 {
		pp = C.salp_providerAddProbe(p.p, cname, 0, nil)
	} else {
		pp = C.salp_providerAddProbe(
			p.p, cname, C.uint32_t(l), (*C.ArgType_t)(&argTypes[0]))
	}
	if pp == nil {
		return Probe{}, p.err()
	}
	return Probe{pp}, nil
}

// MustAddProbe is a helper function that either adds a probe with the supplied
// name and argument types to the specified provider or, in the case of an
// error, panics.
func MustAddProbe(p Provider, name string, argTypes ...ProbeArgType) Probe {
	prb, err := p.AddProbe(name, argTypes...)
	if err != nil {
		panic(err)
	}
	return prb
}

// Unload transitions this Provider from the loaded to the unloaded state.
// Associated probes are detached and must be re-attached in order to function.
func (p *Provider) Unload() {
	C.providerUnload(p.p)
}

// Dispose cleans up the Provider datastructures and frees associated memory
// from the underlying C library (libstapsdt). The Provider instance is useless
// after this method is invoked.
func (p *Provider) Dispose() {
	C.providerDestroy(p.p)
}

// Enabled returns true iff the provider assiciated with this Probe is in a
// loaded state and the Probe is being monitored by an agent such as the bcc
// trace tool.
func (p *Probe) Enabled() bool {
	return int(C.probeIsEnabled(p.p)) == 1
}

// Fire invokes the Probe with the provided arguments. The type and arity of
// this invocation should match what was described by the ProbeArgType arguments
// orginally given to the Provider.AddProbe invocation that created this Probe.
func (p *Probe) Fire(args ...interface{}) {
	if len(args) != int(p.p.argCount) {
		return
	}
	ba := [6]unsafe.Pointer{}
	for i := 0; i < len(args) && i < 5; i++ {
		switch ta := args[i].(type) {
		case int8:
			ba[i] = unsafe.Pointer(uintptr(ta))
		case uint8:
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
		default:
			return
		}
	}
	C.salp_probeFire(p.p, &ba[0])
}

// Name gets the name of this Probe as provided when it was originally created.
func (p *Probe) Name() string {
	return C.GoString(p.p.name)
}
