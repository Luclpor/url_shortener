package main

import (
	"fmt"
	"io"
	"log"
	"os"

	"github.com/Luclpor/url_shortener.git/internal/router/server"
)

var (
	buildVersion = "N/A"
	buildDate    = "N/A"
	buildCommit  = "N/A"
)

func main() {
	printBuildInfo(os.Stdout)

	s, err := server.NewServer()
	if err != nil {
		log.Fatal(err)
	}
	if err := s.Start(); err != nil {
		log.Fatal(err)
	}
}

func printBuildInfo(w io.Writer) {
	fmt.Fprint(w, formatBuildInfo(buildVersion, buildDate, buildCommit))
}

func formatBuildInfo(version, date, commit string) string {
	return fmt.Sprintf(
		"Build version: %s\nBuild date: %s\nBuild commit: %s\n",
		version,
		date,
		commit,
	)
}
