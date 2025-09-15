package models

import (
	"encoding/json"
	"encoding/xml"
	"io"
	"strings"
)

type Event struct {
	Level                   string `xml:"http://v8.1c.ru/eventLog Level"`
	Date                    string `xml:"http://v8.1c.ru/eventLog Date"`
	ApplicationName         string `xml:"http://v8.1c.ru/eventLog ApplicationName"`
	ApplicationPresentation string `xml:"http://v8.1c.ru/eventLog ApplicationPresentation"`
	Event                   string `xml:"http://v8.1c.ru/eventLog Event"`
	EventPresentation       string `xml:"http://v8.1c.ru/eventLog EventPresentation"`
	User                    string `xml:"http://v8.1c.ru/eventLog User"`
	UserName                string `xml:"http://v8.1c.ru/eventLog UserName"`
	Computer                string `xml:"http://v8.1c.ru/eventLog Computer"`
	Metadata                string `xml:"http://v8.1c.ru/eventLog Metadata"`
	MetadataPresentation    string `xml:"http://v8.1c.ru/eventLog MetadataPresentation"`
	Comment                 string `xml:"http://v8.1c.ru/eventLog Comment"`
	Data                    Data   `xml:"http://v8.1c.ru/eventLog Data"`
	DataPresentation        string `xml:"http://v8.1c.ru/eventLog DataPresentation"`
	TransactionStatus       string `xml:"http://v8.1c.ru/eventLog TransactionStatus"`
	TransactionID           string `xml:"http://v8.1c.ru/eventLog TransactionID"`
	Connection              string `xml:"http://v8.1c.ru/eventLog Connection"`
	Session                 string `xml:"http://v8.1c.ru/eventLog Session"`
	ServerName              string `xml:"http://v8.1c.ru/eventLog ServerName"`
	Port                    string `xml:"http://v8.1c.ru/eventLog Port"`
	SyncPort                string `xml:"http://v8.1c.ru/eventLog SyncPort"`
}

type EventLog struct {
	XMLName xml.Name `xml:"http://v8.1c.ru/eventLog EventLog"`
	Events  []Event  `xml:"http://v8.1c.ru/eventLog Event"`
}

type Data struct {
	Content interface{}
}

func (d Data) MarshalJSON() ([]byte, error) {
	return json.Marshal(d.Content)
}

func (d *Data) UnmarshalXML(decoder *xml.Decoder, start xml.StartElement) error {
	for _, attr := range start.Attr {
		if attr.Name.Local == "type" && strings.HasSuffix(attr.Value, "string") {
			var s string
			if err := decoder.DecodeElement(&s, &start); err != nil {
				return err
			}
			d.Content = s
			return nil
		}
	}

	var n node
	if err := decoder.DecodeElement(&n, &start); err != nil {
		if err == io.EOF {
			d.Content = nil
			return nil
		}
		return err
	}

	if n.XMLName.Local == "" && len(n.Children) == 0 && len(n.Attrs) == 0 && strings.TrimSpace(n.Content) == "" {
		d.Content = nil
	} else {
		d.Content = n.toMap()
	}

	return nil
}

type node struct {
	XMLName  xml.Name
	Attrs    []xml.Attr `xml:",attr"`
	Children []node     `xml:",any"`
	Content  string     `xml:",chardata"`
}

func (n *node) toMap() interface{} {
	m := make(map[string]interface{})
	for _, attr := range n.Attrs {
		m[attr.Name.Local] = attr.Value
	}

	if len(n.Children) > 0 {
		groups := make(map[string][]interface{})
		for _, child := range n.Children {
			groups[child.XMLName.Local] = append(groups[child.XMLName.Local], child.toMap())
		}
		for name, group := range groups {
			if len(group) == 1 {
				m[name] = group[0]
			} else {
				m[name] = group
			}
		}
		return m
	}

	if len(m) > 0 {
		if strings.TrimSpace(n.Content) != "" {
			m["#text"] = n.Content
		}
		return m
	}

	return n.Content
}
