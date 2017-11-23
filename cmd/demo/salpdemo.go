package main

import (
	"time"

	"github.com/mmcshane/salp"
)

var (
	probes = salp.MakeProvider("salp-demo")

	p1 = salp.MustAddProbe(probes, "p1", salp.Int8, salp.String)
	p2 = salp.MustAddProbe(probes, "p2", salp.Int8, salp.String)
)

func main() {
	probes.Load()

	var i, j int8
	for {
		s := time.Now().Format(time.RFC850)
		p1.Fire(i, s)
		p2.Fire(j, s)
		i += 1
		j += 2
		time.Sleep(1 * time.Second)
	}
	probes.Unload()
	probes.Dispose()
}
