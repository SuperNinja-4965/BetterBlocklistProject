package helpers

import "os"

// Helper function to get current directory
func GetCurrentDir() string {
	if cwd, err := os.Getwd(); err == nil {
		return cwd + "/"
	}
	return "./"
}

func FileExists(path string) (bool, error) {
	_, err := os.Stat(path)
	if err == nil {
		return true, nil
	}
	if os.IsNotExist(err) {
		return false, nil
	}
	return false, err
}

func JoinLines(lines []string) string {
	if len(lines) == 0 {
		return ""
	}

	result := lines[0]
	for i := 1; i < len(lines); i++ {
		result += "\n" + lines[i]
	}

	return result
}
