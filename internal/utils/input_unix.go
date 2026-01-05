//go:build !windows

package utils

import (
	"fmt"
	"os"
	"syscall"
	"time"

	"golang.org/x/term"
)

func RequestInput(prompt string) string {
	fmt.Print(prompt + ": ")
	fd := int(os.Stdin.Fd())
	oldState, err := term.MakeRaw(fd)
	if err != nil {
		fmt.Println("Error setting terminal to raw mode:", err)
		return ""
	}
	defer term.Restore(fd, oldState)

	// Set stdin to non-blocking
	syscall.SetNonblock(fd, true)
	defer syscall.SetNonblock(fd, false)

	var input []byte
	char := make([]byte, 1)

	startTime := time.Now()
	for {
		n, err := syscall.Read(fd, char)
		if err == nil && n > 0 {
			if char[0] == '\r' || char[0] == '\n' {
				break
			}

			if char[0] == 127 || char[0] == 8 { // Backspace
				if len(input) > 0 {
					input = input[:len(input)-1]
					fmt.Print("\b \b")
				}
				continue
			}

			// Ctrl+C handling (standard in raw mode)
			if char[0] == 3 {
				os.Exit(0)
			}

			// Printable characters
			if char[0] >= 32 && char[0] <= 126 { 
				input = append(input, char[0])
				fmt.Print(string(char[0]))
			}
			startTime = time.Now() // Reset timeout on input
		} else {
			if time.Since(startTime) > 2 * time.Second {
				if len(input) == 0 {
					return "POLL_TIMEOUT"
				}
				// If user has typed something, we wait longer or don't timeout
				// For now, let's say we don't timeout if there's partial input
				// to avoid messy screen refreshes.
				time.Sleep(50 * time.Millisecond)
				continue
			}
			time.Sleep(50 * time.Millisecond)
		}
	}
	fmt.Print("\r\n")
	return string(input)
}