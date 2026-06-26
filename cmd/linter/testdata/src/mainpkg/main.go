package main

import (
	"log"
	"os"
)

func main() {
	log.Fatal("allowed")
	os.Exit(0)
	func() {
		log.Fatal("forbidden in func literal") // want "log.Fatal is allowed only in function main of package main"
		os.Exit(1)                             // want "os.Exit is allowed only in function main of package main"
	}()
	panic("still forbidden") // want "use of built-in panic is prohibited"
}

func helper() {
	log.Fatal("forbidden") // want "log.Fatal is allowed only in function main of package main"
	os.Exit(1)             // want "os.Exit is allowed only in function main of package main"

}
