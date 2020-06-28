package main

import (
	"fmt"
	"os"
	"os/signal"
	"runtime"
	"syscall"

	nats "github.com/nats-io/nats.go"
)

func main() {
	fmt.Println("Starting")

	nc, err := nats.Connect(nats.DefaultURL)
	if err != nil {
		panic(err)
	}

	nc.Subscribe("insights.get.velocity", func(m *nats.Msg) {
		fmt.Print("receiving in callback: ")
		fmt.Println(string(m.Data))
	})

	c := make(chan os.Signal, 1)
	signal.Notify(c, syscall.SIGINT)
	go func() {
		// Wait for signal
		<-c
		nc.Drain()
		os.Exit(0)
	}()
	runtime.Goexit()
}
