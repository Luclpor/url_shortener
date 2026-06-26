package main

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

func TestRunGeneratesResetMethods(t *testing.T) {
	root := t.TempDir()
	pkgDir := filepath.Join(root, "sample")
	if err := os.MkdirAll(pkgDir, 0755); err != nil {
		t.Fatal(err)
	}

	writeFile(t, filepath.Join(root, "go.mod"), "module resetfixture\n\ngo 1.26.0\n")
	writeFile(t, filepath.Join(pkgDir, "sample.go"), `package sample

type Custom struct {
	Value int
}

func (c *Custom) Reset() {
	c.Value = -1
}

// generate:reset
type Child struct {
	Count  int
	Names  []string
	Labels map[string]string
}

// generate:reset
type Parent[T any] struct {
	I         int
	Str       string
	Ok        bool
	StrP      *string
	Values    []int
	Labels    map[string]string
	Child     Child
	ChildPtr  *Child
	Custom    Custom
	CustomPtr *Custom
	Inline    struct {
		N int
		S []string
		M map[string]int
	}
	Numbers [2]int
	Generic T
}
`)
	writeFile(t, filepath.Join(pkgDir, "sample_test.go"), `package sample

import "testing"

func TestGeneratedResetBehavior(t *testing.T) {
	str := "value"
	parent := Parent[int]{
		I:         12,
		Str:       "filled",
		Ok:        true,
		StrP:      &str,
		Values:    []int{1, 2, 3},
		Labels:    map[string]string{"a": "b"},
		Child:     Child{Count: 2, Names: []string{"a"}, Labels: map[string]string{"x": "y"}},
		ChildPtr:  &Child{Count: 3, Names: []string{"b"}, Labels: map[string]string{"k": "v"}},
		Custom:    Custom{Value: 10},
		CustomPtr: &Custom{Value: 11},
		Inline: struct {
			N int
			S []string
			M map[string]int
		}{N: 4, S: []string{"inline"}, M: map[string]int{"n": 1}},
		Numbers: [2]int{7, 8},
		Generic: 42,
	}
	valuesCap := cap(parent.Values)
	childNamesCap := cap(parent.Child.Names)
	inlineCap := cap(parent.Inline.S)

	parent.Reset()

	if parent.I != 0 || parent.Str != "" || parent.Ok || parent.Generic != 0 {
		t.Fatalf("primitive fields were not reset: %+v", parent)
	}
	if parent.StrP == nil || *parent.StrP != "" {
		t.Fatalf("string pointer was not reset")
	}
	if len(parent.Values) != 0 || cap(parent.Values) != valuesCap {
		t.Fatalf("slice was not truncated in place")
	}
	if parent.Labels == nil || len(parent.Labels) != 0 {
		t.Fatalf("map was not cleared")
	}
	if parent.Child.Count != 0 || len(parent.Child.Names) != 0 || cap(parent.Child.Names) != childNamesCap || len(parent.Child.Labels) != 0 {
		t.Fatalf("child struct was not reset: %+v", parent.Child)
	}
	if parent.ChildPtr == nil || parent.ChildPtr.Count != 0 || len(parent.ChildPtr.Names) != 0 || len(parent.ChildPtr.Labels) != 0 {
		t.Fatalf("child pointer was not reset: %+v", parent.ChildPtr)
	}
	if parent.Custom.Value != -1 || parent.CustomPtr == nil || parent.CustomPtr.Value != -1 {
		t.Fatalf("custom Reset methods were not called")
	}
	if parent.Inline.N != 0 || len(parent.Inline.S) != 0 || cap(parent.Inline.S) != inlineCap || len(parent.Inline.M) != 0 {
		t.Fatalf("inline struct was not reset: %+v", parent.Inline)
	}
	if parent.Numbers != [2]int{} {
		t.Fatalf("array was not reset: %+v", parent.Numbers)
	}
}
`)
	testdataDir := filepath.Join(pkgDir, "testdata")
	if err := os.MkdirAll(testdataDir, 0755); err != nil {
		t.Fatal(err)
	}
	writeFile(t, filepath.Join(testdataDir, "fixture.go"), `package testdata

// generate:reset
type Fixture struct {
	Value int
}
`)

	if err := run(root); err != nil {
		t.Fatal(err)
	}

	generated, err := os.ReadFile(filepath.Join(pkgDir, generatedFileName))
	if err != nil {
		t.Fatal(err)
	}
	text := string(generated)
	for _, want := range []string{"func (r *Child) Reset()", "func (r *Parent[T]) Reset()"} {
		if !strings.Contains(text, want) {
			t.Fatalf("generated file does not contain %q:\n%s", want, text)
		}
	}
	if _, err := os.Stat(filepath.Join(testdataDir, generatedFileName)); !os.IsNotExist(err) {
		t.Fatalf("testdata generated file exists, err=%v", err)
	}

	cmd := exec.Command("go", "test", "./...")
	cmd.Dir = root
	cmd.Env = append(os.Environ(),
		"GOCACHE="+filepath.Join(root, ".gocache"),
		"GOMODCACHE="+filepath.Join(root, ".gomodcache"),
	)
	output, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("generated package does not pass go test: %v\n%s", err, output)
	}
}

func writeFile(t *testing.T, path string, content string) {
	t.Helper()
	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		t.Fatal(err)
	}
}
