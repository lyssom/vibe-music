package envfile

import (
	"bufio"
	"os"
	"path/filepath"
	"strings"

	"github.com/lyssom/vibe-music/pkg/logger"
)

// LoadEnvFile reads a .env file and sets environment variables.
func LoadEnvFile(path string) error {
	log := logger.New("env", logger.DEBUG)
	log.Debug("Loading .env from: %s", path)

	file, err := os.Open(path)
	if err != nil {
		log.Debug("Could not open .env file: %v", err)
		return err
	}
	defer file.Close()

	count := 0
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())

		// Skip empty lines and comments
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		// Parse KEY=VALUE
		if idx := strings.Index(line, "="); idx != -1 {
			key := strings.TrimSpace(line[:idx])
			value := strings.TrimSpace(line[idx+1:])
			value = strings.Trim(value, "\"'") // Remove quotes

			if key != "" {
				log.Debug("Setting: %s=%s...", key, truncate(value, 10))
				os.Setenv(key, value)
				count++
			}
		}
	}

	log.Info("Loaded %d env vars from %s", count, path)
	return nil
}

func truncate(s string, max int) string {
	if len(s) <= max {
		return s
	}
	return s[:max] + "..."
}

// FindEnvFile looks for .env in common locations.
func FindEnvFile() string {
	candidates := []string{
		".env",
		".env.local",
		filepath.Join("cmd", "vibe-echo", ".env"),
	}

	for _, path := range candidates {
		if _, err := os.Stat(path); err == nil {
			return path
		}
	}

	return ""
}