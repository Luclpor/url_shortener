package forbidden

import (
	"log"
	stdlog "log"
	"os"
	stdos "os"
)

func reportsForbiddenCalls() {
	panic("boom")      // want "use of built-in panic is prohibited"
	log.Fatal("fatal") // want "log.Fatal is allowed only in function main of package main"
	os.Exit(1)         // want "os.Exit is allowed only in function main of package main"
}

func reportsAliasedCalls() {
	stdlog.Fatal("fatal") // want "log.Fatal is allowed only in function main of package main"
	stdos.Exit(1)         // want "os.Exit is allowed only in function main of package main"
}

func ignoresShadowedNames() {
	panic := func(any) {}
	panic("ok")

	log := struct {
		Fatal func(...any)
	}{Fatal: func(...any) {}}
	log.Fatal("ok")

	os := struct {
		Exit func(int)
	}{Exit: func(int) {}}
	os.Exit(0)
}
