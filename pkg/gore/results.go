package gore

import (
	"bytes"
	"encoding/json"
	"encoding/xml"
	"fmt"
	"io/ioutil"
	"strings"
	"time"
)

type ResultSet struct {
	QueryName string
	Query     QueryResult
	Results   []ResultItem
}

type QueryResult struct {
	Completed bool
	Error     error
	Capped    bool
	CapLimit  int
}

type ResultItem struct {
	Data map[string]interface{} `json:"ResultItem"`
}

func (ri ResultItem) Int(name string) int {
	v, ck := ri.Data[name]
	if !ck {
		return -1
	}
	return v.(int)
}

func (ri ResultItem) Float(name string) float64 {
	v, ck := ri.Data[name]
	if !ck {
		return -1
	}
	return v.(float64)
}

func (ri ResultItem) String(name string) string {
	v, ck := ri.Data[name]
	if !ck {
		return ""
	}
	return v.(string)
}

func (ri ResultItem) Bool(name string) bool {
	v, ck := ri.Data[name]
	if !ck {
		return false
	}
	return v.(bool)
}

func (ri ResultItem) Date(name string) time.Time {
	v, ck := ri.Data[name]
	if !ck {
		return time.Now()
	}
	return v.(time.Time)
}

func (rs ResultSet) Export(filename, xFmt string) (err error) {
	var content []byte
	var prefix []byte
	var suffix []byte
	switch strings.ToLower(xFmt) {
	case "xml":
		prefix = []byte("<?xml version=\"1.0\" encoding=\"UTF-8\" standalone=\"yes\"?><Results>")
		suffix = []byte("</Results>")
		content, err = xml.Marshal(rs.Results)
	case "json":
		prefix = []byte("{\"Results\":")
		suffix = []byte("}")
		content, err = json.Marshal(rs.Results)
	case "csv":
		err = fmt.Errorf("CSV exporting not yet implemented")
	}
	if err != nil {
		return
	}

	return ioutil.WriteFile(filename, bytes.Join([][]byte{prefix, content, suffix}, []byte("")), 0644)
}

type xmlMapEntry struct {
	XMLName xml.Name
	Value   interface{} `xml:",chardata"`
}

// map to xml
func (ri ResultItem) MarshalXML(e *xml.Encoder, start xml.StartElement) error {
	if len(ri.Data) == 0 {
		return nil
	}

	err := e.EncodeToken(start)
	if err != nil {
		return err
	}

	for k, v := range ri.Data {
		e.Encode(xmlMapEntry{XMLName: xml.Name{Local: k}, Value: v})
	}

	return e.EncodeToken(start.End())
}
