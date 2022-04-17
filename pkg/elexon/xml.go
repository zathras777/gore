package elexon

import (
	"bytes"
	"encoding/xml"
	"fmt"
	"log"
	"strconv"
	"strings"
	"time"
)

type xmlNode struct {
	XMLName xml.Name
	Content []byte    `xml:",innerxml"`
	Nodes   []xmlNode `xml:",any"`
}

func parseXML(content []byte) (nodes xmlNode, err error) {
	if len(content) == 0 {
		err = fmt.Errorf("Cannot parse empty content!")
		return
	}
	dec := xml.NewDecoder(bytes.NewReader(content))
	err = dec.Decode(&nodes)
	return
}

func (n xmlNode) get(name string) (node xmlNode, err error) {
	var found int
	if strings.Contains(name, ".") {
		parts := strings.SplitN(name, ".", 2)
		found = n.searchNodes(parts[0])
		if found != -1 {
			return n.Nodes[found].get(parts[1])
		}
	} else {
		found = n.searchNodes(name)
	}
	if found == -1 {
		return node, fmt.Errorf("Unable to find a node matching '%s'", name)
	}
	return n.Nodes[found], nil
}

func (n xmlNode) getAll(name string) (nodes []xmlNode, err error) {
	if strings.Contains(name, ".") {
		parts := strings.SplitN(name, ".", 2)
		var node xmlNode
		node, err = n.get(parts[0])
		if err != nil {
			return
		}
		return node.getAll(parts[1])
	}
	for _, nn := range n.Nodes {
		if nn.XMLName.Local == name {
			nodes = append(nodes, nn)
		}
	}
	return nodes, nil
}

func (n xmlNode) getAsMap(name string, mapInfo map[string]string) (info map[string]interface{}, err error) {
	var node xmlNode
	node, err = n.get(name)
	if err != nil {
		return
	}
	info = node.asMap(mapInfo)
	return
}

func (n xmlNode) asMap(mapInfo map[string]string) (info map[string]interface{}) {
	info = make(map[string]interface{})
	for _, elem := range n.Nodes {
		t, ck := mapInfo[elem.XMLName.Local]
		if !ck {
			continue
		}
		var cStr = string(elem.Content)
		switch t {
		case "int":
			num, err := strconv.Atoi(cStr)
			if err == nil {
				info[elem.XMLName.Local] = num
			} else {
				log.Printf("Unable to convert %s to a number", cStr)
			}
		case "float":
			num, err := strconv.ParseFloat(cStr, 64)
			if err == nil {
				info[elem.XMLName.Local] = num
			} else {
				log.Printf("Unable to convert %s to a number", cStr)
			}
		case "bool":
			info[elem.XMLName.Local] = strings.Contains("YesTrue", cStr)
		case "string":
			info[elem.XMLName.Local] = cStr
		case "date":
			tm, err := time.Parse("2006-01-02", cStr)
			if err == nil {
				info[elem.XMLName.Local] = tm
			} else {
				log.Printf("Unable to convert '%s' into date: %s", cStr, err)
			}
		case "dateTime":
			tm, err := time.Parse("2006-01-02 15:04:05", cStr)
			if err == nil {
				info[elem.XMLName.Local] = tm
			} else {
				log.Printf("Unable to convert '%s' into date/time: %s", cStr, err)
			}
		default:
			log.Printf("Unhandled content type: %s", t)
		}
	}
	return
}

func (n xmlNode) searchNodes(name string) int {
	for i, nn := range n.Nodes {
		if nn.XMLName.Local == name {
			return i
		}
	}
	return -1
}
