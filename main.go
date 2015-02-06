package main

import (
	"fmt"
	"github.com/oliveagle/go-collectors/collectors"
	"os"
	"os/signal"
	"syscall"
)

func main() {

	// filter by collectors name
	// c := collectors.Search("proc")
	// list(c)
	// cdp := collectors.Run(c)

	cdp := collectors.Run(nil)

	interrupt := make(chan os.Signal, 1)
	signal.Notify(interrupt, os.Interrupt, os.Kill, syscall.SIGTERM)

	for {
		select {
		case dp := <-cdp:
			fmt.Println(dp)
			// fmt.Printf(".")

		case killSignal := <-interrupt:
			if killSignal == os.Interrupt {
				fmt.Println("Daemon was interruped by system signal")
				os.Exit(1)
			}
			fmt.Println("Daemon was killed")
			os.Exit(1)
		}
	}
}

func list(cs []collectors.Collector) {
	for _, c := range cs {
		fmt.Println(c.Name())
	}
}
