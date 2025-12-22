package utils

import (
	"fmt"
	"os"
	"golang.org/x/term"
	"strings"
)

type KeyAction struct {
	Key byte
	Action func()
	Description string
}

type Command struct {
	Text string
	Action func(Text string)
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
		return
	}
}

func RequestAction(actions []Command) {
	Println("")
	commandText := RequestInput("Command Pallete")

	if commandText == "" { return }
	if commandText == "help" {
		PrintHelp(actions)
		return
	}

	for _, cmd := range actions {
		if strings.Fields(commandText)[0] != cmd.Text { continue }
		cmd.Action(commandText)
		return
	}
}

func PrintHelp(actions []Command) {
	fmt.Println("Available Commands:")
	maxLength := 0
	for _, cmd := range actions {
		if len(cmd.Text) > maxLength {
			maxLength = len(cmd.Text)
		}
	}
	for _, cmd := range actions {
		fmt.Printf(" %-*s: %s\r\n", maxLength, cmd.Text, cmd.Description)
	}
	RequestAction(actions)
}

func RequestInput(prompt string) string {
	fmt.Print(prompt + ": ")
	fd := int(os.Stdin.Fd())
	oldState, err := term.MakeRaw(fd)
	if err != nil {
		fmt.Println("Error setting terminal to raw mode:", err)
		return ""
	}
	defer term.Restore(fd, oldState)

	var input []byte
	char := make([]byte, 1)

	for {
		_, err := os.Stdin.Read(char)
		if err != nil {
			fmt.Println("Error reading input:", err)
			return ""
		}

		if char[0] == '\r' || char[0] == '\n' {
			break
		}

		if char[0] == 127 { // Backspace
			if len(input) > 0 {
				input = input[:len(input)-1]
				fmt.Print("\b \b")
			}
			continue
		}
		// Printable characters
		if char[0] >= 32 && char[0] <= 126 { 
			input = append(input, char[0])
			fmt.Print(string(char[0]))
		}
	}
	fmt.Print("\r\n")
	return string(input)
}

func ClearScreen() {
	fmt.Print("\033[3J") // Clear scrollback (if supported)
    fmt.Print("\033[H\033[2J") // Home + clear visible screen
}

func Println(text string) {
	fmt.Print(text + "\r\n")
}
