package main

import "testing"

func TestValidateInputMessage(t *testing.T) {
	valid := []message{
		{Command: "pressKey", Data: "CTRL"},
		{Command: "releaseKey", Data: "left"},
		{Command: "tapKey", Data: "F12"},
		{Command: "typeText", Data: "hello"},
		{Command: "typeText", Data: "مرحبا"},
		{Command: "moveMouse", X: -32768, Y: 32767},
		{Command: "clickMouse"},
		{Command: "ping"},
	}
	for _, msg := range valid {
		if err := validateInputMessage(&msg); err != nil {
			t.Errorf("expected valid message %+v, got %v", msg, err)
		}
	}

	invalid := []struct {
		name string
		msg  message
	}{
		{"empty key", message{Command: "tapKey", Data: ""}},
		{"key with spaces", message{Command: "pressKey", Data: "ctrl alt"}},
		{"key with control byte", message{Command: "tapKey", Data: "enter\n"}},
		{"oversized text", message{Command: "typeText", Data: string(make([]byte, maxNativeTextBytes+1))}},
		{"NUL text", message{Command: "typeText", Data: "hello\x00world"}},
		{"invalid UTF-8", message{Command: "typeText", Data: string([]byte{0xff})}},
		{"x below bound", message{Command: "moveMouse", X: -32769}},
		{"y above bound", message{Command: "moveMouse", Y: 32768}},
		{"unknown command", message{Command: "runShell", Data: "whoami"}},
	}
	for _, tc := range invalid {
		t.Run(tc.name, func(t *testing.T) {
			if err := validateInputMessage(&tc.msg); err == nil {
				t.Errorf("expected validation error for %+v", tc.msg)
			}
		})
	}
}

func TestValidateInputMessageRejectsNil(t *testing.T) {
	if err := validateInputMessage(nil); err == nil {
		t.Fatal("expected nil message to be rejected")
	}
}

func TestStagedSetupPathUsesExecutableExtensionOnWindows(t *testing.T) {
	got := stagedSetupPath(`C:\\Users\\tester\\buster-client-setup.exe`, "windows", "token")
	want := `C:\\Users\\tester\\buster-client-setup.exe.token.verified.exe`
	if got != want {
		t.Fatalf("stagedSetupPath() = %q, want %q", got, want)
	}
}

func TestStagedSetupPathForUnix(t *testing.T) {
	got := stagedSetupPath("/opt/buster/buster-client-setup", "linux", "token")
	want := "/opt/buster/buster-client-setup.token.verified"
	if got != want {
		t.Fatalf("stagedSetupPath() = %q, want %q", got, want)
	}
}
