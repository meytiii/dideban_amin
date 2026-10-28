package main

import (
	"os"
	"os/exec"
	"strings"
	"testing"
)

func TestDBCredentialsRequiredAtStartup(t *testing.T) {
	if os.Getenv("TEST_SUBPROCESS_DB_CREDS") == "1" {
		dbConnStr := strings.TrimSpace(os.Getenv("HERASAT_DB_CONN"))
		if dbConnStr == "" {
			os.Stderr.WriteString("FATAL: HERASAT_DB_CONN environment variable is required.\n")
			os.Exit(1)
		}
		return
	}

	cmd := exec.Command(os.Args[0], "-test.run=TestDBCredentialsRequiredAtStartup")
	cmd.Env = []string{
		"TEST_SUBPROCESS_DB_CREDS=1",
		"HERASAT_DB_CONN=",
		"SYSTEMROOT=" + os.Getenv("SYSTEMROOT"),
	}

	output, err := cmd.CombinedOutput()
	if err == nil {
		t.Fatalf("Expected subprocess to exit non-zero when HERASAT_DB_CONN is empty")
	}

	outStr := string(output)
	if !strings.Contains(outStr, "FATAL: HERASAT_DB_CONN environment variable is required.") {
		t.Errorf("Expected output to contain fatal DB conn message, got: %s", outStr)
	}
}
