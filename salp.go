package salp

/*
#cgo LDFLAGS: -lstapsdt

#include <assert.h>
#include <stdint.h>
#include <stdlib.h>
#include <libstapsdt.h>

// This wrapper is necessary only because CGO cannot invoke varargs functions
SDTProbe_t* salp_providerAddProbe(SDTProvider_t* p, const char* name,
									uint32_t c, ArgType_t* at){
	assert(6 == MAX_ARGUMENTS); // MAX_ARGUMENTS is from libstapsdt.h
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

void salp_probeFire(SDTProbe_t* p, void** args) {
	assert(p->argCount <= MAX_ARGUMENTS); // MAX_ARGUMENTS is from libstapsdt.h
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

type Provider struct {
	p *C.struct_SDTProvider
}

type Probe struct {
	p *C.struct_SDTProbe
}

type ProbeArgType C.ArgType_t

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

func (e stapsdtError) Error() string {
	return fmt.Sprintf("libstapsdt error [%v]: %v", e.code, e.msg)
}

func MakeProvider(name string) Provider {
	clbl := C.CString(name)
	defer C.free(unsafe.Pointer(clbl))
	return Provider{C.providerInit(clbl)}
}

func (p Provider) Name() string {
	return C.GoString(p.p.name)
}

func (p Provider) err() error {
	return stapsdtError{
		code: int(p.p.errno),
		msg:  C.GoString(p.p.error),
	}
}

func (p Provider) Load() error {
	rc := C.providerLoad(p.p)
	if int(rc) != 0 {
		return p.err()
	}
	return nil
}

func (p Provider) AddProbe(name string, argTypes ...ProbeArgType) (Probe, error) {
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

func MustAddProbe(p Provider, name string, argTypes ...ProbeArgType) Probe {
	prb, err := p.AddProbe(name, argTypes...)
	if err != nil {
		panic(err)
	}
	return prb
}

func (p Provider) Unload() {
	C.providerUnload(p.p)
}

func (p Provider) Dispose() {
	C.providerDestroy(p.p)
}

func (p Probe) Enabled() bool {
	return int(C.probeIsEnabled(p.p)) == 1
}

func (p Probe) Fire(args ...interface{}) {
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
		case int32:
			ba[i] = unsafe.Pointer(uintptr(ta))
		case uint32:
			ba[i] = unsafe.Pointer(uintptr(ta))
		case int64:
			ba[i] = unsafe.Pointer(uintptr(ta))
		case uint64:
			ba[i] = unsafe.Pointer(uintptr(ta))
		case string:
			strptr := unsafe.Pointer(C.CString(ta))
			defer C.free(strptr)
			ba[i] = strptr
		}
	}
	C.salp_probeFire(p.p, &ba[0])
}

func (p Probe) Name() string {
	return C.GoString(p.p.name)
}
