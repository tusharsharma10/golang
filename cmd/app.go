package main

import (
	"flag"
	"log"
	"os"
	"restapi/internal/server"
)

func main() {
	// start app server here
	environment := flag.String("e", "development", "")
	flag.Usage = func() {
		log.Println("Usage: server -e {mode}")
		os.Exit(1)
	}

	flag.Parse()
	server.Init("." + *environment + ".env")
}
