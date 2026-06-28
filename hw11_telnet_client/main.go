package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"io"
	"net"
	"os"
	"os/signal"
	"syscall"
	"time"
)

func main() {
	timeout := flag.Duration("timeout", 10*time.Second, "connection timeout, default 10s")
	flag.Parse()

	args := flag.Args()
	if len(args) != 2 {
		fmt.Fprintln(os.Stderr, "Usage: go-telnet [--timeout=<duration>] <host> <port>")
		os.Exit(1)
	}

	address := net.JoinHostPort(args[0], args[1])

	client := NewTelnetClient(address, *timeout, os.Stdin, os.Stdout)

	if err := client.Connect(); err != nil {
		fmt.Fprintf(os.Stderr, "...Connection error: %v\n", err)
		os.Exit(1)
	}
	defer client.Close()

	fmt.Fprintf(os.Stderr, "...Connected to %s\n", address)

	ctx, cancel := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer cancel()

	// from stdin
	go func() {
		err := client.Send()
		if err != nil {
			if errors.Is(err, io.EOF) {
				fmt.Fprintln(os.Stderr, "...EOF")
			} else {
				fmt.Fprintf(os.Stderr, "...Send error: %v\n", err)
			}
		}
		client.Close()
		cancel()
	}()

	// to stdout
	go func() {
		err := client.Receive()
		if err != nil {
			if errors.Is(err, io.EOF) {
				fmt.Fprintln(os.Stderr, "...Connection was closed by peer")
			} else {
				fmt.Fprintf(os.Stderr, "...Receive error: %v\n", err)
			}
		}
		client.Close()
		cancel()
	}()

	<-ctx.Done()
}
