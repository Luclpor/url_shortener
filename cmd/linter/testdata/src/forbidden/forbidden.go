package forbidden

import (
	"log"
	stdlog "log"
	"os"
	stdos "os"
)

func reportsForbiddenCalls() {
	panic("boom")
	log.Fatal("fatal")
	os.Exit(1)
}

func reportsAliasedCalls() {
	stdlog.Fatal("fatal")
	stdos.Exit(1)
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
