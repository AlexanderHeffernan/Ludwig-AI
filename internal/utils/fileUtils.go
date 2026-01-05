package utils
import (
	"os"
)
func ReadFileAsString(path string) string {
	data, err := os.ReadFile(path)
    if err != nil {
		return "Error reading file: " + err.Error() + "\npath: " + path
    }

    content := string(data) // Convert []byte to string
	return content
}
