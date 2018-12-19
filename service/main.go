package main

import (
	"dec/service/elevator"
	"flag"
	console "github.com/AsynkronIT/goconsole"
)

var flagBind = flag.String("bind", "127.0.0.1:9000", "Bind to address")
var flagID = flag.Uint("id", 0, "ID")

func main() {
	flag.Parse()

	elevator.NewElevatorService(*flagBind, *flagID)

	for {
		console.ReadLine()
	}
}
