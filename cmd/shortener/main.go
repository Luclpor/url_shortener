package main

import (
	"fmt"
	"io"
	"log"
	"os"

	"github.com/Luclpor/url_shortener.git/internal/router/server"
)

var buildVersion string
var buildDate string
var buildCommit string

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
		buildInfoValue(version),
		buildInfoValue(date),
		buildInfoValue(commit),
	)
}

func buildInfoValue(value string) string {
	if value == "" {
		return "N/A"
	}
	return value
}
