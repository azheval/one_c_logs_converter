package promtail_writer

import (
	"encoding/json"
	"log/slog"
	"os"
	"path/filepath"

	"one_c_logs_converter/models"
)

// WriteEvents appends a slice of log events to a project-specific log file in JSON format.
func WriteEvents(outputDir, projectName string, entries []models.Event) error {
	outputFileName := filepath.Join(outputDir, projectName+".json")

	// Open the file in append mode, or create it if it doesn't exist.
	outputFile, err := os.OpenFile(outputFileName, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0666)
	if err != nil {
		return err
	}
	defer outputFile.Close()

	for _, entry := range entries {
		jsonEntry, err := json.Marshal(entry)
		if err != nil {
			slog.Error("failed to marshal log entry to json", "error", err)
			continue // Skip this entry and proceed with the next one
		}

		if _, err := outputFile.Write(append(jsonEntry, '\n')); err != nil {
			// It's possible the write fails, so we log it and continue.
			slog.Error("failed to write to output file", "name", outputFileName, "error", err)
		}
	}

	return nil
}
