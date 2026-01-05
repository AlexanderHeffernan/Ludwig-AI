//go:build windows

package utils

import (
	"bufio"
	"fmt"
	"os"
	"strings"
	"time"
)

func RequestInput(prompt string) string {
	fmt.Print(prompt + ": ")

	// Use a simpler approach on Windows
	// We'll use a channel-based timeout mechanism instead of raw terminal mode
	inputChan := make(chan string, 1)
	
	go func() {
		reader := bufio.NewReader(os.Stdin)
		text, _ := reader.ReadString('\n')
		inputChan <- strings.TrimSpace(text)
	}()
	
	select {
	case input := <-inputChan:
		return input
	case <-time.After(2 * time.Second):
		return "POLL_TIMEOUT"
	}
}