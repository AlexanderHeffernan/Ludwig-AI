package utils
import (
	"fmt"
	"time"
	tea "github.com/charmbracelet/bubbletea"
)

func DebugLog(msg string) {
	f, err := tea.LogToFile("debug.log", "")
	if err != nil {
		return
	}
	defer f.Close()
	
	timestamp := time.Now().Format("2006-01-02 15:04:05")
	fmt.Fprintf(f, "[%s] %s\n", timestamp, msg)
}
