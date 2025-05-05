package main

import (
	"bufio"
	"log"
	"os"
)

const cfgPath = "sunny_5_skiers/config.json"

func main() {
	var err error

	cfg, err := ParseConfig(cfgPath)
	if err != nil {
		log.Fatalln("Config error:", err)
	}

	events, err := ParseEvents(os.Stdin)
	if err != nil {
		log.Fatalln("Events error:", err)
	}

	f, err := os.Create("output.log")
	if err != nil {
		log.Fatalf("cannot create output.log file: %v", err)
	}
	defer f.Close()
	writeLog := bufio.NewWriter(f)

	ProcessEvents(writeLog, cfg, events)

	if err := writeLog.Flush(); err != nil {
		log.Fatalf("cannot flush output.log: %v", err)
	}
}
