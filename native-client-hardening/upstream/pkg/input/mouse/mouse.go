package mouse

import (
	"errors"
	"fmt"

	"github.com/go-vgo/robotgo"
)

// MoveMouse returns an error when RobotGo reports that smooth movement failed.
func MoveMouse(x, y int) error {
	if !robotgo.MoveMouseSmooth(x, y, 0.5, 1.0, 0) {
		return errors.New("RobotGo could not move the mouse")
	}
	return nil
}

// ClickMouse converts panics into errors. The pinned RobotGo MouseClick API
// returns no status, so a normal return cannot prove that the OS delivered
// the click.
func ClickMouse() (err error) {
	defer func() {
		if recovered := recover(); recovered != nil {
			err = fmt.Errorf("mouse click failed: %v", recovered)
		}
	}()
	robotgo.MouseClick()
	return nil
}
