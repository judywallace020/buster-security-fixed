package utils

import (
	"bytes"
	"errors"
	"testing"
)

type failingReader struct{}

func (failingReader) Read([]byte) (int, error) {
	return 0, errors.New("entropy unavailable")
}

func TestRandomIntUsesHalfOpenRange(t *testing.T) {
	got, err := randomInt(bytes.NewReader([]byte{4}), 5, 10)
	if err != nil {
		t.Fatalf("randomInt returned error: %v", err)
	}
	if got != 9 {
		t.Fatalf("randomInt = %d, want 9", got)
	}
}

func TestRandomIntRejectsInvalidRanges(t *testing.T) {
	for _, bounds := range [][2]int{{-1, 5}, {5, 5}, {6, 5}} {
		if _, err := randomInt(bytes.NewReader([]byte{0}), bounds[0], bounds[1]); err == nil {
			t.Errorf("randomInt(%d, %d) expected error", bounds[0], bounds[1])
		}
	}
}

func TestRandomIntPropagatesEntropyFailure(t *testing.T) {
	if _, err := randomInt(failingReader{}, 0, 10); err == nil {
		t.Fatal("expected entropy source error")
	}
}
