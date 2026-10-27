package main

import (
	"os"
	"testing"
)

func TestMain(m *testing.M) {
	InitCrypto("8f14e45fceea167a5a36dedd4bea2543")
	os.Exit(m.Run())
}
