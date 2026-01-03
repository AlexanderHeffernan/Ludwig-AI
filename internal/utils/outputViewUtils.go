package utils

import (
	"strings"
	"encoding/json"
	"time"
	"ludwig/internal/orchestrator"
	"ludwig/internal/types"
	//tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"crypto/sha256"
	"sync"
	"regexp"
)

var OUTPUT_STYLE lipgloss.Style = lipgloss.NewStyle().Padding(0, 0)

func removeSurroundingQuotes(s string) string {
	s = strings.TrimSpace(s)
	if len(s) >= 2 {
		return s[1 : len(s)-1]
	}
	return s
}

/*
func CheckFileUpdated(filePath string, lastHash []byte) (bool, []byte) {
	newHash, fileContent := GetFileContentHash(filePath)

	return bytes.Equal(newHash, lastHash), newHash
}
*/

func GetFileContentHash(filePath string) ([]byte, string) {
	h := sha256.New()
	fileContent := ReadFileAsString(filePath)
	h.Write([]byte(fileContent))
	return h.Sum(nil), fileContent
}

func GetFileHash(filePath string) []byte {
	h := sha256.New()
	fileContent := ReadFileAsString(filePath)
	h.Write([]byte(fileContent))
	return h.Sum(nil)
}

type DelayedTask struct {
	timer *time.Timer
	mu sync.Mutex
	canceled bool
}

func NewDelayedTask(duration time.Duration, task func()) *DelayedTask {
	dt := &DelayedTask{}

	dt.timer = time.AfterFunc(duration, func() {
		dt.mu.Lock()
		alreadyCanceled := dt.canceled
		dt.mu.Unlock()
		if !alreadyCanceled {
			task()
		}
	})
	return dt
}

func OutputLine(line string) string {
	// remove surrounding {}
	if len(line) == 0 {
		return line
	}
	if line[0] != '{' || len(line) < 2 {
		return line
	}
	if line == "---" {
		return "---"
	}
	var object map[string]any
	err := json.Unmarshal([]byte(line), &object)
	builder := strings.Builder{}
	if err != nil {
		return ""
	}
	if object["type"] == "init" {
		return ""
	}
	if object["type"] == "message" {
		timestamp := object["timestamp"].(string)
		styledTimestamp := FormatTimestamp(timestamp)
		builder.WriteString(styledTimestamp)
		output := strings.Builder{}
		output.WriteString(object["content"].(string))
		builder.WriteString(OUTPUT_STYLE.Render(output.String()))
		return builder.String()
	}
	if object["type"] == "tool_use" {
		timestamp := object["timestamp"].(string)
		timestamp = FormatTimestamp(timestamp)
		builder.WriteString(timestamp)

		output := strings.Builder{}
		toolName := object["tool_name"].(string)
		params := object["parameters"].(map[string]any)
		output.WriteString("Using tool: " + toolName + "\n")
		writeParams(&output, params)
		builder.WriteString(OUTPUT_STYLE.Render(output.String()))

		return builder.String()
	}
	if object["type"] == "tool_result" {
		output := strings.Builder{}
		result := object["status"].(string)
		output.WriteString("Tool result: ")
		output.WriteString(result)
		output.WriteString("\n")
		builder.WriteString(OUTPUT_STYLE.Render(output.String()))

		return builder.String()
	}
	if object["type"] == "result" {
		return builder.String()
	}

	return line
}

func writeParams(builder *strings.Builder, params map[string]any) {
	builder.WriteString("With parameters:\n")
	for key, value := range params {
		paramLine := "  - " + key + ": " + value.(string) + "\n"
		builder.WriteString(paramLine)
	}
}

var ORDERED_LIST_REGEX = regexp.MustCompile(`\n *([0-9]+\. )`)
var LIST_STYLE = lipgloss.NewStyle().Foreground(lipgloss.Color("#cc6600"))

func colouredOrderedLists(output string) string {
	return ORDERED_LIST_REGEX.ReplaceAllStringFunc(output, func(match string) string {
		return LIST_STYLE.Render(match)
	})
}

var UNORDERED_LIST_REGEX = regexp.MustCompile(`\n +(\+|\*|-) `)
func colouredUnorderedLists(output string) string {
	return UNORDERED_LIST_REGEX.ReplaceAllStringFunc(output, func(match string) string {
		return LIST_STYLE.Render(match)
	})
}

//var STRING_REGEX = regexp.MustCompile(`"([^"\\]*(\\.[^"\\]*)*)"`)
var STRING_REGEX = regexp.MustCompile(`("([^"\\]*(\\.[^"\\]*)*)")|(` + "`" + `[^` + "`" + `]*` + "`" + `)`)
var STRING_STYLE = lipgloss.NewStyle().Foreground(lipgloss.Color("#CE7354"))
func colouredStrings(output string) string {
	return STRING_REGEX.ReplaceAllStringFunc(output, func(match string) string {
		return STRING_STYLE.Render(match)
	})
}

func OutputLines(lines []string) string {
	output := strings.Builder{}
	if len(lines) == 0 {
		return "no output"
	}
	started := false
	linesToSkip := 2
	for _, line := range lines {
		if line == "---" && started {
			break
		}
		if line == "---" {
			started = true
			continue
		}
		if !started {
			continue
		}
		if line == "" {
			continue
		}
		if linesToSkip > 0 {
			linesToSkip--
			continue
		}
		output.WriteString(OutputLine(line))
		output.WriteString("\n")
	}
	outputStr := colouredUnorderedLists(output.String())
	outputStr = colouredStrings(outputStr)
	return colouredOrderedLists(outputStr)
}

func ThinkingString(task types.Task) string {
	if orchestrator.IsRunning() && task.Status == types.InProgress {
		return " (thinking...)"
	}
	return ""
}

func GetTaskByPath(tasks []types.Task, path string) *types.Task {
	// remove the './.ludwig/' prefix from path
	path = strings.TrimPrefix(path, "./.ludwig/")
	DebugLog("GetTaskByPath: looking for path: " + path)
	for _, task := range tasks {
		if task.ResponseFile != path { continue; }
		return &task
	}
	return nil
}

