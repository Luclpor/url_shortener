package forbidden

import (
	. "log"
	. "os"
)

func reportsDotImportedCalls() {
	Fatal("fatal")
	Exit(1)
}
