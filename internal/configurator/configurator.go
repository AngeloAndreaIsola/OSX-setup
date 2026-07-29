package configurator

import (
	"fmt"
	"os"
	"strings"
)

// EnsureBlock idempotently adds a block of configuration to the given file.
func EnsureBlock(filePath, blockName, content string) error {
	data, err := os.ReadFile(filePath)
	if err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("failed to read config file: %w", err)
	}
	
	contentStr := string(data)
	startMarker := fmt.Sprintf("# >>> setupper %s start >>>", blockName)
	endMarker := fmt.Sprintf("# <<< setupper %s end <<<", blockName)
	
	if strings.Contains(contentStr, startMarker) && strings.Contains(contentStr, endMarker) {
		startIdx := strings.Index(contentStr, startMarker)
		endIdx := strings.Index(contentStr, endMarker) + len(endMarker)
		
		newStr := contentStr[:startIdx] + startMarker + "\n" + content + "\n" + endMarker + contentStr[endIdx:]
		return os.WriteFile(filePath, []byte(newStr), 0644)
	}
	
	newBlock := fmt.Sprintf("\n%s\n%s\n%s\n", startMarker, content, endMarker)
	
	if len(contentStr) > 0 && !strings.HasSuffix(contentStr, "\n") {
		newBlock = "\n" + newBlock
	}
	
	f, err := os.OpenFile(filePath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return err
	}
	defer f.Close()
	
	_, err = f.WriteString(newBlock)
	return err
}

func ConfigureShell(rcPath string) error {
	content := `export PATH="$HOME/.setupper/bin:$PATH"`
	return EnsureBlock(rcPath, "core", content)
}
