package config

import (
	"flag"
	"io"
	"os"
	"testing"
)

func TestInitConfigEnablesHTTPSWithFlag(t *testing.T) {
	resetConfigTestState(t, []string{"shortener", "-s"})
	unsetEnv(t, "ENABLE_HTTPS")

	cfg, err := InitConfig()
	if err != nil {
		t.Fatalf("InitConfig() error = %v", err)
	}
	if !cfg.EnableHTTPS {
		t.Fatal("InitConfig() should enable HTTPS with -s flag")
	}
}

func TestInitConfigEnablesHTTPSWithEnv(t *testing.T) {
	resetConfigTestState(t, []string{"shortener"})
	t.Setenv("ENABLE_HTTPS", "true")

	cfg, err := InitConfig()
	if err != nil {
		t.Fatalf("InitConfig() error = %v", err)
	}
	if !cfg.EnableHTTPS {
		t.Fatal("InitConfig() should enable HTTPS with ENABLE_HTTPS=true")
	}
}

func TestInitConfigFlagEnablesHTTPSWhenEnvIsFalse(t *testing.T) {
	resetConfigTestState(t, []string{"shortener", "-s"})
	t.Setenv("ENABLE_HTTPS", "false")

	cfg, err := InitConfig()
	if err != nil {
		t.Fatalf("InitConfig() error = %v", err)
	}
	if !cfg.EnableHTTPS {
		t.Fatal("InitConfig() should enable HTTPS when -s is set")
	}
}

func resetConfigTestState(t *testing.T, args []string) {
	t.Helper()

	oldArgs := os.Args
	oldCommandLine := flag.CommandLine
	flag.CommandLine = flag.NewFlagSet(args[0], flag.ContinueOnError)
	flag.CommandLine.SetOutput(io.Discard)
	os.Args = args

	t.Cleanup(func() {
		os.Args = oldArgs
		flag.CommandLine = oldCommandLine
	})
}

func unsetEnv(t *testing.T, key string) {
	t.Helper()

	oldValue, wasSet := os.LookupEnv(key)
	if err := os.Unsetenv(key); err != nil {
		t.Fatalf("Unsetenv(%q) error = %v", key, err)
	}
	t.Cleanup(func() {
		if wasSet {
			if err := os.Setenv(key, oldValue); err != nil {
				t.Fatalf("Setenv(%q) error = %v", key, err)
			}
			return
		}
		if err := os.Unsetenv(key); err != nil {
			t.Fatalf("Unsetenv(%q) error = %v", key, err)
		}
	})
}
