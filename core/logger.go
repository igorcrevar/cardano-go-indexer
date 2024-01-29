package core

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/hashicorp/go-hclog"
)

type LoggerConfig struct {
	LogLevel            hclog.Level
	JSONLogFormat       bool
	OpenOrCreateNewFile bool
	LogsDirectory       string
	LogFile             string
	Name                string
}

func NewLogger(config LoggerConfig) (l hclog.Logger, err error) {
	var logFileWriter *os.File

	if config.LogFile != "" {
		fullFilePath := config.LogFile

		if config.LogsDirectory != "" {
			if dirErr := os.MkdirAll(config.LogsDirectory, os.ModePerm); dirErr == nil {
				fullFilePath = filepath.Join(config.LogsDirectory, fullFilePath)
			}
		}

		if !config.OpenOrCreateNewFile {
			timestamp := strings.Replace(strings.Replace(time.Now().UTC().Format(time.RFC3339), ":", "_", -1), "-", "_", -1)
			fullFilePath = fullFilePath + "_" + timestamp
		}

		logFileWriter, err = os.OpenFile(fullFilePath+".log", os.O_APPEND|os.O_WRONLY|os.O_CREATE, 0600)
		if err != nil {
			return nil, fmt.Errorf("could not create or open log file, %w", err)
		}
	}

	return hclog.New(&hclog.LoggerOptions{
		Name:       config.Name,
		Level:      config.LogLevel,
		Output:     logFileWriter,
		JSONFormat: config.JSONLogFormat,
	}), nil
}
