//go:build ci
// +build ci

// This file is used only by CI's compile/test jobs. Production builds must
// generate clientbin.go from the real native client executable via Makefile.
package main

import (
	"errors"
	"os"
)

func Asset(string) ([]byte, error) {
	return nil, errors.New("embedded client asset is unavailable in CI stub mode")
}

func AssetInfo(string) (os.FileInfo, error) {
	return nil, errors.New("embedded client asset metadata is unavailable in CI stub mode")
}
