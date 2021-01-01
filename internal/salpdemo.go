package main

import (
	"fmt"
	"os"
	"os/signal"
	"time"

	"github.com/mmcshane/salp"
)

var (
	probes = salp.NewProvider("salp-demo")

	p1 = salp.MustAddProbe(probes, "p1", salp.Int32, salp.Error, salp.String)
	p2 = salp.MustAddProbe(probes, "p2", salp.Uint8, salp.Bool)
)

func main() {
	defer salp.UnloadAndDispose(probes)
	fmt.Println(`List the go probes in this demo
	# With bcc tools
	sudo tplist -vp "$(pgrep -n salpdemo)" "*salp-demo*"

	# With bpftrace
	sudo bpftrace -p "$(pgrep -n salpdemo)" -l "usdt:*:salp-demo:*"

Trace this process
	# With bcc tools
	sudo trace -p "$(pgrep -n salpdemo)" 'u::p1 "i=%d err='%s' date='%s'", arg1, arg2, arg3' 'u::p2 "j=%d flag=%d", arg1, arg2'

	# With bpftrace
	sudo bpftrace -p "$(pgrep -n salpdemo)" -e 'usdt:p1 { printf("%d (%s)\n", arg0, str(arg2)); }'
	sudo bpftrace -p "$(pgrep -n salpdemo)" -e 'usdt:p2 { if(arg1) { printf("Truthy! %d\n", arg0); } }'`)

	salp.MustLoadProvider(probes)

	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)

	var i, j int8

	for {
		select {
		case <-c:
			return
		case now := <-time.After(1 * time.Second):
			s := now.Format(time.RFC1123Z)
			p1.Fire(i, fmt.Errorf("An error: %d", i), s)
			p2.Fire(j, j%4 == 0)
			i++
			j += 2
		}
	}
}
