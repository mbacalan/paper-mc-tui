package utils

import (
	"fmt"
	"os"
	"path/filepath"
	"time"
)

type Logger struct {
	logFile     string
	versionFile string
}

func NewLogger() (*Logger, error) {
	// Create logs directory if it doesn't exist
	if err := os.MkdirAll("logs", 0755); err != nil {
		return nil, fmt.Errorf("failed to create logs directory: %w", err)
	}

	return &Logger{
		logFile:     filepath.Join("logs", "paper-mc.log"),
		versionFile: filepath.Join("logs", "paper-ver.txt"),
	}, nil
}

func (l *Logger) Log(message string) error {
	timestamp := time.Now().Format("2006-01-02 15:04:05")
	logEntry := fmt.Sprintf("[%s] %s\n", timestamp, message)

	f, err := os.OpenFile(l.logFile, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return err
	}
	defer f.Close()

	_, err = f.WriteString(logEntry)
	return err
}

func (l *Logger) GetLastDownloadedVersion() (string, error) {
	data, err := os.ReadFile(l.versionFile)
	if err != nil {
		if os.IsNotExist(err) {
			return "", nil
		}
		return "", err
	}
	return string(data), nil
}

func (l *Logger) SaveDownloadedVersion(version string) error {
	return os.WriteFile(l.versionFile, []byte(version), 0644)
}
