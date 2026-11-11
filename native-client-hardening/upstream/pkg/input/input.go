package input

import (
	"buster-client/pkg/input/keyboard"
	"buster-client/pkg/input/mouse"
)

func MoveMouse(x, y int) error {
	return mouse.MoveMouse(x, y)
}

func ClickMouse() error {
	return mouse.ClickMouse()
}

func TypeText(text string) error {
	return keyboard.TypeText(text)
}

func PressKey(key string) error {
	return keyboard.PressKey(key)
}

func ReleaseKey(key string) error {
	return keyboard.ReleaseKey(key)
}

func TapKey(key string) error {
	return keyboard.TapKey(key)
}
