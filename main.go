package main

import (
	"fmt"
	"log"
	"os"
)

func main() {
	if err := run(); err != nil {
		log.Fatalf("execution finished with error: %s", err.Error())
		os.Exit(1)
	}
}

func run() error {
	fmt.Println("Hello world")
	return nil
}
