package common

import (
	"os"

	"github.com/zepollabot/media-rating-overlay/internal/httpclient/plugins"
)

// SetupLogging configures request logging for the client
func SetupLogging(client HTTPClient, logFilePath string) error {
	logFile, err := os.OpenFile(logFilePath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return err
	}

	requestLogger := plugins.NewRequestLogger(logFile, logFile)
	client.AddPlugin(requestLogger)
	return nil
}
