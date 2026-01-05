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
	Action func(Text string) string
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

func RequestAction(actions []Command) string {
	Println("")
	commandText := RequestInput("Command Pallete")

	if commandText == "" { return "" }
	if commandText == "POLL_TIMEOUT" { return "POLL_TIMEOUT" }
	if commandText == "help" {
		PrintHelp(actions)
		return ""
	}

	for _, cmd := range actions {
		if strings.Fields(commandText)[0] != cmd.Text { continue }
		cmd.Action(commandText)
		return cmd.Text
	}
	return ""
}

func PrintHelp(actions []Command) string {
	builder := strings.Builder{}
	//fmt.Println("Available Commands:")
	builder.WriteString("Available Commands:\n")
	maxLength := 0
	for _, cmd := range actions {
		if len(cmd.Text) > maxLength {
			maxLength = len(cmd.Text)
		}
	}
	for _, cmd := range actions {
		//fmt.Printf(" %-*s: %s\r\n", maxLength, cmd.Text, cmd.Description)
		builder.WriteString(cmd.Text + ": " + cmd.Description)
	}
	//fmt.Println("\r\nPress any key to continue...")
	
	return builder.String()
}

func ClearScreen() {
	/*
	fmt.Print("\033[3J") // Clear scrollback (if supported)
    fmt.Print("\033[H\033[2J") // Home + clear visible screen
	*/
}

func Println(text string) {
	fmt.Print(text + "\r\n")
}

func GenerateTopBubbleBorder(width int) string {
	borderWidth := width - 4
	if borderWidth < 10 {
		borderWidth = 10 // Minimum border width
	}
	return " " + strings.Repeat("╭", 1) + strings.Repeat("─", borderWidth) + strings.Repeat("╮", 1) + " \n"
}

func GenerateBottomBubbleBorder(width int) string {
	borderWidth := width - 4
	if borderWidth < 10 {
		borderWidth = 10 // Minimum border width
	}
	return " " + strings.Repeat("╰", 1) + strings.Repeat("─", borderWidth) + strings.Repeat("╯", 1) + " \n"
}

func TermWidth() int {
	width, _, err := term.GetSize(int(os.Stdout.Fd()))
	if err != nil {
		return 80 // Default width
	}
	return width
}

func TermHeight() int {
	_, height, err := term.GetSize(int(os.Stdout.Fd()))
	if err != nil {
		return 24 // Default height
	}
	return height
}

func LeftRightBorderedString(name string, length int, visLength int, truncate bool, borderColor string) string {
	if (truncate && len(name) + 5 > length) {
		truncatedName := name[:length - 8] + "... "
		numSpaces := max(length - visLength - 4, 0)
		return ColoredString(" │ ", borderColor) + truncatedName + strings.Repeat(" ", numSpaces) + ColoredString("│", borderColor)
	}

	numSpaces := max(length - visLength - 4, 0)

	return ColoredString(" │ ", borderColor) + name + strings.Repeat(" ", numSpaces) + ColoredString("│", borderColor)
}

func InsertLineBreaks(s string, n int) string {
	if n <= 0 || len(s) == 0 {
        return s
    }
    var b strings.Builder
    for i := 0; i < len(s); i += n {
        end := min(i + n, len(s))
        if i > 0 {
            b.WriteByte('\n')
        }
        b.WriteString(s[i:end])
    }
    return b.String()
}

func BoldColoredString(s string, colorCode string) string {
	return fmt.Sprintf("\033[1;%sm%s\033[0m", colorCode, s)
}

func ColoredString(s string, colorCode string) string {
	return fmt.Sprintf("\033[%sm%s\033[0m", colorCode, s)
}

func BoldString(s string) string {
	return fmt.Sprintf("\033[1m%s\033[0m", s)
}
