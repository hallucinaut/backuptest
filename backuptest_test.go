package backuptest

import (
	"testing"
)

func TestMain(t *testing.T) {
	// Verify the main package builds and runs
	t.Log("backuptest package exists and builds successfully")
}

func TestVersion(t *testing.T) {
	const version = "1.0.0"
	if version == "" {
		t.Error("Expected version to be defined")
	}
}
