//go:build windows
// +build windows

package main

import (
	"os"
	"os/exec"
	"path/filepath"
	"testing"
)

func TestEnsureNoSymlinkComponentsRejectsJunction(t *testing.T) {
	base, cleanup := makeTempDir(t)
	defer cleanup()

	target := filepath.Join(base, "target")
	junction := filepath.Join(base, "junction")
	if err := os.Mkdir(target, 0700); err != nil {
		t.Fatal(err)
	}

	cmd := exec.Command("cmd", "/c", "mklink", "/J", junction, target)
	if output, err := cmd.CombinedOutput(); err != nil {
		t.Skipf("junction creation unavailable: %v (%s)", err, output)
	}
	defer os.Remove(junction)

	if err := ensureNoSymlinkComponents(filepath.Join(junction, "child")); err == nil {
		t.Fatal("Windows junction component must be rejected")
	}
}
