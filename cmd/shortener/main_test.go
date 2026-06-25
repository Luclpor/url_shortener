package main

import "testing"

func TestFormatBuildInfoUsesNAForEmptyValues(t *testing.T) {
	got := formatBuildInfo("", "", "")
	want := "Build version: N/A\nBuild date: N/A\nBuild commit: N/A\n"

	if got != want {
		t.Fatalf("formatBuildInfo() = %q, want %q", got, want)
	}
}

func TestFormatBuildInfoUsesProvidedValues(t *testing.T) {
	got := formatBuildInfo("v1.0.0", "2026-06-25", "abc123")
	want := "Build version: v1.0.0\nBuild date: 2026-06-25\nBuild commit: abc123\n"

	if got != want {
		t.Fatalf("formatBuildInfo() = %q, want %q", got, want)
	}
}
