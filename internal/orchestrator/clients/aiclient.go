package clients

import "io"

type AIClient interface {
	SendPrompt(prompt string, writer io.Writer) (string, error)
	SendPromptWithDir(prompt string, writer io.Writer, workDir string) (string, error)
}
