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
		utils.RandSleep(60, 80)
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
	utils.RandSleep(35, 40)
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
