package clients

import "io"

type AIClient interface {
	SendPrompt(prompt string) (string, error)
	SendPromptWithStream(prompt string, writer io.Writer) (string, error)
}
