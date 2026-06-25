package main

import (
	"log"
	"os"
)

func main() {
	log.Fatal("allowed")
	os.Exit(0)
	panic("still forbidden")
}

func helper() {
	log.Fatal("forbidden")
	os.Exit(1)
}
