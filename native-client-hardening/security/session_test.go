package security

import "testing"

func TestSessionTokenRoundTrip(t *testing.T) {
	token, err := NewSessionToken()
	if err != nil {
		t.Fatal(err)
	}
	if err := ValidateSessionToken(token); err != nil {
		t.Fatal(err)
	}
	if !SessionTokenEqual(token, token) {
		t.Fatal("matching token was rejected")
	}
	if SessionTokenEqual(token, token+"x") {
		t.Fatal("different token was accepted")
	}
	if SessionTokenEqual(token, "") {
		t.Fatal("empty token was accepted")
	}
}
