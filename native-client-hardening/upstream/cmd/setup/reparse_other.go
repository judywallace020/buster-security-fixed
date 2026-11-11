//go:build !windows
// +build !windows

package main

func isReparsePoint(string) (bool, error) {
	return false, nil
}
