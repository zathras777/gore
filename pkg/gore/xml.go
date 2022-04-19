package gore

import (
	"bytes"
	"encoding/xml"
	"fmt"
	"log"
	"strconv"
	"strings"
	"time"
)

type XmlNode struct {
	XMLName xml.Name
	Content []byte     `xml:",innerxml"`
	Attr    []xml.Attr `xml:",any,attr"`
	Nodes   []XmlNode  `xml:",any"`
}

func ParseXML(content []byte) (nodes XmlNode, err error) {
	if len(content) == 0 {
		err = fmt.Errorf("Cannot parse empty content!")
		return
	}
	dec := xml.NewDecoder(bytes.NewReader(content))
	err = dec.Decode(&nodes)
	return
}

func (n XmlNode) Get(name string) (node XmlNode, err error) {
	var found int
	if strings.Contains(name, ".") {
		parts := strings.SplitN(name, ".", 2)
		found = n.searchNodes(parts[0])
		if found != -1 {
			return n.Nodes[found].Get(parts[1])
		}
	} else {
		found = n.searchNodes(name)
	}
	if found == -1 {
		return node, fmt.Errorf("Unable to find a node matching '%s'", name)
	}
	return n.Nodes[found], nil
}

func (n XmlNode) GetAll(name string) (nodes []XmlNode, err error) {
	if strings.Contains(name, ".") {
		parts := strings.SplitN(name, ".", 2)
		var node XmlNode
		node, err = n.Get(parts[0])
		if err != nil {
			return
		}
		return node.GetAll(parts[1])
	}
	for _, nn := range n.Nodes {
		if nn.XMLName.Local == name {
			nodes = append(nodes, nn)
		}
	}
	return nodes, nil
}

func (n XmlNode) GetAsMap(name string, mapInfo map[string]string) (info map[string]interface{}, err error) {
	var node XmlNode
	node, err = n.Get(name)
	if err != nil {
		return
	}
	info = node.AsMap(mapInfo)
	return
}

func (n XmlNode) AsMap(mapInfo map[string]string) (info map[string]interface{}) {
	info = make(map[string]interface{})
	for _, elem := range n.Nodes {
		t, ck := mapInfo[elem.XMLName.Local]
		if !ck {
			continue
		}
		info[elem.XMLName.Local] = convert(string(elem.Content), t)
	}
	return
}

func attrMatch(name string, mapName string) bool {
	if !strings.Contains(mapName, ":") {
		return name == mapName
	}
	return strings.SplitN(mapName, ":", 2)[0] == name
}

func (n XmlNode) AttrAsMap(mapInfo map[string]string) (info map[string]interface{}) {
	info = make(map[string]interface{})
	for _, elem := range n.Attr {
		var final string
		var elemType string
		for final, elemType = range mapInfo {
			if attrMatch(elem.Name.Local, final) {
				if strings.Contains(final, ":") {
					final = strings.SplitN(final, ":", 2)[1]
				}
				break
			}
		}
		if len(final) == 0 {
			log.Printf("Skipping attribute %s as no match found in supplied map\n", elem.Name.Local)
			continue
		}
		info[final] = convert(string(elem.Value), elemType)
	}
	return
}

func convert(cStr string, t string) (rv interface{}) {
	switch t {
	case "int":
		num, err := strconv.Atoi(cStr)
		if err == nil {
			rv = num
		} else {
			log.Printf("Unable to convert %s to a number", cStr)
		}
	case "float":
		num, err := strconv.ParseFloat(cStr, 64)
		if err == nil {
			rv = num
		} else {
			log.Printf("Unable to convert %s to a number", cStr)
		}
	case "bool":
		rv = strings.Contains("YesTrue", cStr)
	case "string":
		rv = cStr
	case "date":
		tm, err := time.Parse("2006-01-02", cStr)
		if err == nil {
			rv = tm
		} else {
			log.Printf("Unable to convert '%s' into date: %s", cStr, err)
		}
	case "dateTime":
		tm, err := time.Parse("2006-01-02 15:04:05", cStr)
		if err == nil {
			rv = tm
		} else {
			log.Printf("Unable to convert '%s' into date/time: %s", cStr, err)
		}
	default:
		rv = cStr
		log.Printf("Unhandled content type: %s", t)
	}
	return
}

func (n XmlNode) searchNodes(name string) int {
	for i, nn := range n.Nodes {
		if nn.XMLName.Local == name {
			return i
		}
	}
	return -1
}
