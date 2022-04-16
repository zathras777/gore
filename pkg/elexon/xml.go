package elexon

import (
	"bytes"
	"encoding/xml"
	"fmt"
	"io/ioutil"
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

func TestXML(fn string) {
	content, err := ioutil.ReadFile(fn)
	if err != nil {
		panic(err)
	}
	n := parseXML(content)
	meta, err := n.get("responseMetadata")
	fmt.Println(meta)
	code, err := n.get("responseMetadata.httpCode")
	fmt.Println(code)
	nodes, err := n.getAll("responseBody.responseList.item")
	fmt.Printf("Total of %d nodes found.\n", len(nodes))
	fmt.Println(n.getAsMap("responseMetadata", metadataMap))
}

func parseXML(content []byte) xmlNode {
	dec := xml.NewDecoder(bytes.NewReader(content))

	var n xmlNode
	err := dec.Decode(&n)
	if err != nil {
		panic(err)
	}
	return n
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
		case "bool":
			info[elem.XMLName.Local] = strings.Contains("YesTrue", cStr)
		case "string":
			info[elem.XMLName.Local] = cStr
		case "date":
			tm, err := time.Parse("2006-02-01", cStr)
			if err == nil {
				info[elem.XMLName.Local] = tm
			} else {
				log.Printf("Unable to convert '%s' into date: %s", cStr, err)
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

//<responseBody>
//  <dataItem>B1420</dataItem>
//  <responseList>
//    <item>
//      <documentType>Configuration document</documentType>
//      <businessType>Production unit</businessType>
//      <processType>Creation</processType>
//      <timeSeriesID>NGET-EMFIP-CONF-TS-00740033</timeSeriesID>
//      <powerSystemResourceType>&quot;Wind Offshore&quot;</powerSystemResourceType>
//      <year>2021</year>
//      <bMUnitID>NA</bMUnitID>
//      <registeredResourceEICCode>48W00000HOWBO-3H</registeredResourceEICCode>
//      <nominal>440</nominal>
//      <nGCBMUnitID>HOWBO-1</nGCBMUnitID>
//      <registeredResourceName>HOWBO-1</registeredResourceName>
//      <activeFlag>Y</activeFlag>
//      <documentID>NGET-EMFIP-CONF-00688826</documentID>
//      <implementationDate>2021-10-11</implementationDate>
//    </item>
