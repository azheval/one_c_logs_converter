package log_event

import (
	"encoding/xml"
	"os"

	"one_c_logs_converter/models"
)

// Parse reads an XML log file and extracts the log events.
func Parse(filePath string) ([]models.Event, error) {
	xmlFile, err := os.ReadFile(filePath)
	if err != nil {
		return nil, err
	}

	var eventLog models.EventLog
	if err := xml.Unmarshal(xmlFile, &eventLog); err != nil {
		return nil, err
	}

	return eventLog.Events, nil
}
