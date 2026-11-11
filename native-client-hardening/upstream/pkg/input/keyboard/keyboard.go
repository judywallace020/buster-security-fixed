package keyboard

import (
	"errors"
	"fmt"

	"github.com/go-vgo/robotgo"

	"buster-client/pkg/input/utils"
)

// TypeText reports panics, but the pinned RobotGo TypeStr API exposes no
// operation status. A nil result therefore means the call returned normally,
// not that the operating system confirmed every character was delivered.
func TypeText(text string) (err error) {
	defer func() {
		if recovered := recover(); recovered != nil {
			err = fmt.Errorf("text input failed: %v", recovered)
		}
	}()
	for _, char := range text {
		robotgo.TypeStr(string(char))
		if err := utils.RandSleep(60, 80); err != nil {
			return fmt.Errorf("text input delay failed: %v", err)
		}
	}
	return nil
}

func PressKey(key string) error {
	return toggleKey(key, true)
}

func ReleaseKey(key string) error {
	return toggleKey(key, false)
}

func TapKey(key string) error {
	if err := PressKey(key); err != nil {
		return err
	}
	if err := utils.RandSleep(35, 40); err != nil {
		releaseErr := ReleaseKey(key)
		if releaseErr != nil {
			return fmt.Errorf("tap delay failed (%v); key release also failed: %w", err, releaseErr)
		}
		return fmt.Errorf("tap delay failed: %v", err)
	}
	if err := ReleaseKey(key); err != nil {
		return fmt.Errorf("key was pressed but release failed: %w", err)
	}
	return nil
}

func toggleKey(key string, down bool) error {
	state := "up"
	if down {
		state = "down"
	}
	if result := robotgo.KeyToggle(key, state); result != "" {
		return errors.New(result)
	}
	return nil
}
