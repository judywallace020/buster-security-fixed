package utils

import (
	"crypto/rand"
	"errors"
	"io"
	"math/big"
	"time"
)

func randomInt(reader io.Reader, min, max int) (int, error) {
	if min < 0 || max <= min {
		return 0, errors.New("invalid random range")
	}
	span := big.NewInt(int64(max - min))
	n, err := rand.Int(reader, span)
	if err != nil {
		return 0, err
	}
	return min + int(n.Int64()), nil
}

// RandSleep sleeps for a cryptographically selected duration in [min, max)
// milliseconds. It returns invalid-range and entropy-source errors.
func RandSleep(min, max int) error {
	delay, err := randomInt(rand.Reader, min, max)
	if err != nil {
		return err
	}
	time.Sleep(time.Duration(delay) * time.Millisecond)
	return nil
}
