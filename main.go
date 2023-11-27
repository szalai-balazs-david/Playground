// OPC UA client working against Unified Automation ANSI C Demo OPC UA server: https://www.unified-automation.com/downloads/opc-ua-servers.html
// The app subscribes to a set of Node IDs and logs their value to a SQLite database

package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"
)

func main() {
	log.Print("Start")

	log.Print("Initializing DataStore")
	ds := InitializeDataStore()

	exit := make(chan os.Signal, 1)
	signal.Notify(exit, os.Interrupt, syscall.SIGTERM)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	log.Print("Starting OPC monitoring")
	go RunOpcMonitoring(ctx, &ds)

	log.Print("Starting HTTP server")
	go RunRestServer(ctx, &ds)

	<-exit
	//Need some better cleanup?
}
