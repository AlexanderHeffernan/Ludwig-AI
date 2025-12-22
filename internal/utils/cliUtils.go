package utils

import (
	"fmt"
	"os"
	"golang.org/x/term"
)

type KeyAction struct {
	Key byte
	Action func()
	Description string
}

func OnKeyPress(actions []KeyAction) {
	for _, ka := range actions {
		fmt.Printf("[%c] %s  ", ka.Key, ka.Description)
	}
	fmt.Print("\n")
	fd := int(os.Stdin.Fd())
	char := make([]byte, 1)

	oldState, err := term.MakeRaw(fd)
	if err != nil {
		fmt.Println("Error setting terminal to raw mode:", err)
		return
	}
	defer term.Restore(fd, oldState)

	os.Stdin.Read(char)
	for _, ka := range actions {
		if (char[0] != ka.Key) { continue }
		ka.Action()
		break
	}
}
