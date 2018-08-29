# Salp

Salp is a library that enables Go programs to create and fire USDT probes at
runtime. Such probes allow API-stable (i.e. not dependent on function names) 
tracing of executables written in Go - especially important since function
tracing of Go code requires some unappealing (though not unimpressive)
[hacks](http://www.brendangregg.com/blog/2017-01-31/golang-bcc-bpf-function-tracing.html).
Salp uses the [libstapsdt](https://github.com/sthima/libstapsdt) library to create
and fire these probes.

[![GoDoc](https://godoc.org/github.com/mmcshane/salp?status.svg)](https://godoc.org/github.com/mmcshane/salp)

## Build

### Dependencies

Salp depends on `libstapsdt` which in turn depends on `libelf` and `libdl`.
`libstapsdt` can be built from source or (for debian based distros) installed
via an Ubuntu PPA. Full instructions are available in the [docs for
libstapsdt](http://libstapsdt.readthedocs.io/en/latest/getting-started/getting-started.html)

### Build and Test

If `libstdapsdt` is installed globally (e.g. from the PPA above or via `make
install`), you should be able to simply `go build` or `go test`. However if you
have built `libstapsdt` from source then you will need to tell the `cgo` tool
how to find the headers and .so files for `libstapsdt` using the `CGO_CFLAGS`,
`CGO_LDFLAGS`, and `LD_LIBRARY_PATH` environment variables.

```bash
export CGO_CFLAGS="-I/path/to/libstapsdt/src"
export CGO_LDFLAGS="-L/path/to/libstapsdt/out"
export LD_LIBRARY_PATH="/path/to/libstapsdt/out"
```

## Demo

This repository contains a demo executable that will fire two different probes
every second. The demo can be observed using the `trace` and `tplist` tools from
the [bcc](https://github.com/iovisor/bcc) project. Use two terminals to run the
demo - one to execute the tracable salpdemo go program and one to run the bcc
tools and see their output. In the first window run

```bash
go run internal/salpdemo.go
```

This program will print out how to monitor itself but then won't print out
anything after that. In the second window run the bcc trace program on the
`salpdemo` process, monitoring probes `p1` and `p2`.

```bash
sudo trace -p "$(pgrep salpdemo | head -n1)" \
    'u::p1 "arg1=%d arg2=%s", arg1, arg2' \
    'u::p2 "arg1=%d", arg1'
```

`trace` will output the values of `arg1` and `arg2` from probe `p1` and the
value of `arg1` from probe `p2`.
