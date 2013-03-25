package main

import (
	"flag"
	"fmt"
	"github.com/ActiveState/log"
	"logyard"
	"logyard/cli"
	"logyard/util/pubsub"
	"math/rand"
	"os"
	"os/signal"
	"stackato/server"
	"strings"
	"time"
)

type stream struct {
	raw bool
}

func (cmd *stream) Name() string {
	return "stream"
}

func (cmd *stream) DefineFlags(fs *flag.FlagSet) {
	fs.BoolVar(&cmd.raw, "raw", false,
		"Show unformatted logs, including logyard INFO records (skipped by default)")
}

func (cmd *stream) Run(args []string) error {
	ipaddr, err := server.LocalIP()
	if err != nil {
		return err
	}

	rand.Seed(time.Now().UnixNano())
	port := 7000 + rand.Intn(1000)
	addr := fmt.Sprintf("%s:%d", ipaddr, port)

	srv, err := cli.NewLineServer("tcp", addr)
	if err != nil {
		return err
	}

	go srv.Start()

	name := fmt.Sprintf("tmp.logyardctl.%s-%d", ipaddr, port)

	// REFACTOR: extract URI construction of add.go and then use
	// logyard.AddDrain directly.
	(&add{
		uri:     "tcp://" + addr,
		filters: Filters(args),
		params:  Options(map[string]string{"format": "raw"}),
	}).Run(
		[]string{name})

	deleteDrain := func() {
		if err := logyard.Config.DeleteDrain(name); err != nil {
			log.Fatal(err)
		}
		log.Infof("Deleted drain %s", name)
	}
	defer deleteDrain()

	handleKeyboardInterrupt(func() {
		deleteDrain()
		os.Exit(1)
	})

	// REFACTOR: do we need MessagePrinter at all? all it does to
	// provide abstraction over color formatting; most other things
	// (formatting, skipping) happens in stream_handler.go.
	printer := cli.NewMessagePrinter(cli.MessagePrinterOptions{cmd.raw})
	printer.AddFormat("systail",
		"@m{{.Name}}@|@@@c{{.NodeID}}@|: {{.Text}}")
	printer.AddFormat("event",
		"@g{{.Type}}@|[@m{{.Process}}@!]@@@c{{.NodeID}}@|: {{.Desc}}")
	printer.AddFormat("apptail",
		"@b{{.AppName}}[{{.Source}}]@|@@@c{{.NodeID}}@|: {{.Text}}")

	printer.SetPrePrintHook(streamHandler)

	// Print incoming records
	for line := range srv.Ch {
		parts := strings.SplitN(string(line), " ", 2)
		if len(parts) != 2 {
			log.Fatal("received invalid message: %s", string(line))
		}
		msg := pubsub.Message{parts[0], parts[1]}
		if err := printer.Print(msg); err != nil {
			log.Fatalf("Error -- %s -- printing message %v", err, msg)
		}
	}

	return nil
}

func handleKeyboardInterrupt(cleanupFn func()) {
	// Handle Ctrl+C
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, os.Interrupt)
	go func() {
		for _ = range sigCh {
			cleanupFn()
		}
	}()
}
