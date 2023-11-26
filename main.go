// OPC UA client working against Unified Automation ANSI C Demo OPC UA server: https://www.unified-automation.com/downloads/opc-ua-servers.html
// The app subscribes to a set of Node IDs and logs their value to a SQLite database

package main

import (
	"context"
	"log"
	"time"
)

func main() {
	log.Print("Start")

	nodes := []string{
		"ns=4;s=Demo.Dynamic.Scalar.Double",
		"ns=4;s=Demo.Dynamic.Scalar.Float",
		"ns=4;s=Demo.Dynamic.Scalar.Int32",
		"ns=4;s=Demo.SimulationActive",
		"ns=4;s=Demo.SimulationSpeed",
	}

	log.Print("Initializing DataStore")
	ds := InitializeDataStore()

	ctx := context.Background()

	log.Print("Starting OPC monitoring")
	go RunOpcMonitoring(ctx, &ds, nodes)

	log.Print("Starting HTTP server")
	go RunRestServer(ctx, &ds)

	//This is sloppy, I'm sure there's a better way.
	for {
		if ctx.Err() != nil {
			break
		}
		time.Sleep(time.Second)
	}
}
