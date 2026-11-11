//go:build windows
// +build windows

package main

import "syscall"

// FILE_ATTRIBUTE_REPARSE_POINT covers junctions, symbolic links, and other
// filesystem redirections. Rejecting all reparse points is intentionally
// conservative for installation paths.
const fileAttributeReparsePoint = 0x0400

func isReparsePoint(path string) (bool, error) {
	name, err := syscall.UTF16PtrFromString(path)
	if err != nil {
		return false, err
	}
	attributes, err := syscall.GetFileAttributes(name)
	if err != nil {
		return false, err
	}
	return attributes&fileAttributeReparsePoint != 0, nil
}
