package models

import "encoding/xml"

// Event represents a single event from the 1C log file.
type Event struct {
	Level                 string `xml:"http://v8.1c.ru/eventLog Level"`
	Date                  string `xml:"http://v8.1c.ru/eventLog Date"`
	ApplicationName       string `xml:"http://v8.1c.ru/eventLog ApplicationName"`
	ApplicationPresentation string `xml:"http://v8.1c.ru/eventLog ApplicationPresentation"`
	Event                 string `xml:"http://v8.1c.ru/eventLog Event"`
	EventPresentation     string `xml:"http://v8.1c.ru/eventLog EventPresentation"`
	User                  string `xml:"http://v8.1c.ru/eventLog User"`
	UserName              string `xml:"http://v8.1c.ru/eventLog UserName"`
	Computer              string `xml:"http://v8.1c.ru/eventLog Computer"`
	Metadata              string `xml:"http://v8.1c.ru/eventLog Metadata"`
	MetadataPresentation  string `xml:"http://v8.1c.ru/eventLog MetadataPresentation"`
	Comment               string `xml:"http://v8.1c.ru/eventLog Comment"`
	Data                  string `xml:"http://v8.1c.ru/eventLog Data"`
	DataPresentation      string `xml:"http://v8.1c.ru/eventLog DataPresentation"`
	TransactionStatus     string `xml:"http://v8.1c.ru/eventLog TransactionStatus"`
	TransactionID         string `xml:"http://v8.1c.ru/eventLog TransactionID"`
	Connection            string `xml:"http://v8.1c.ru/eventLog Connection"`
	Session               string `xml:"http://v8.1c.ru/eventLog Session"`
	ServerName            string `xml:"http://v8.1c.ru/eventLog ServerName"`
	Port                  string `xml:"http://v8.1c.ru/eventLog Port"`
	SyncPort              string `xml:"http://v8.1c.ru/eventLog SyncPort"`
}

// EventLog represents the root of the 1C log file.
type EventLog struct {
	XMLName xml.Name `xml:"http://v8.1c.ru/eventLog EventLog"`
	Events  []Event  `xml:"http://v8.1c.ru/eventLog Event"`
}