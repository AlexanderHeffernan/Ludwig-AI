package utils
import (
	"fmt"
	"time"
	"os"
	tea "github.com/charmbracelet/bubbletea"
)


/**
 * DebugLog writes a debug message to debug.log if the DEBUG environment variable is set.
 * Use `DEBUG=1 go run cmd/main.go` when debugging.
 */
func DebugLog(msg string) {
	if os.Getenv("DEBUG") == "" {
		return
	}
	f, err := tea.LogToFile("debug.log", "")
	if err != nil {
		return
	}
	defer f.Close()
	
	timestamp := time.Now().Format("2006-01-02 15:04:05")
	fmt.Fprintf(f, "[%s] %s\n", timestamp, msg)
}
