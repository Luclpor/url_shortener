package forbidden

import (
	. "log"
	. "os"
)

func reportsDotImportedCalls() {
	Fatal("fatal") // want "log.Fatal is allowed only in function main of package main"
	Exit(1)        // want "os.Exit is allowed only in function main of package main"
}
