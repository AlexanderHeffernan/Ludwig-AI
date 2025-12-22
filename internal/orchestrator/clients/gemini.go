package clients

import (
	"bytes"
	"os/exec"
)

type GeminiClient struct{}

func (g *GeminiClient) SendPrompt(prompt string) (string, error) {
	cmd := exec.Command("gemini", "--yolo", prompt)
	var out bytes.Buffer
	var stderror bytes.Buffer
	cmd.Stdout = &out
	cmd.Stderr = &stderror

	err := cmd.Run()
	if err != nil {
		return "", err
	}
	return out.String(), nil
}
