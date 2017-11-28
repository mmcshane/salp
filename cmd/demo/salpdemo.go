package main

import (
	"fmt"
	"os"
	"os/signal"
	"time"

	"github.com/mmcshane/salp"
)

var (
	probes = salp.MakeProvider("salp-demo")

	p1 = salp.MustAddProbe(probes, "p1", salp.Int8, salp.String)
	p2 = salp.MustAddProbe(probes, "p2", salp.Uint8, salp.String)
)

func main() {
	fmt.Println("List the go probes in this demo with")
	fmt.Println("\tsudo tplist -vp \"$(pgrep salpdemo)\" \"salp-demo*\"")
	fmt.Println("Trace this process with")
	fmt.Println("\tsudo trace -p \"$(pgrep salpdemo | head -n1)\" 'u::p1 \"arg1=%d arg2=%s\", arg1, arg2' 'u::p2 \"arg1=%d\", arg1'")

	probes.Load()

	defer func() {
		probes.Unload()
		probes.Dispose()
	}()

	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)

	var i, j int8

	for {
		select {
		case <-c:
			return
		case now := <-time.After(1 * time.Second):
			s := now.Format(time.RFC1123Z)
			p1.Fire(i, s)
			p2.Fire(j, s)
			i += 1
			j += 2
		}
	}
}
