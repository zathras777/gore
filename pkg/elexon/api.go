package elexon

import (
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"strings"
)

type ElexonAPI struct {
	Result ElexonAPIResult
	Items  []ElexonAPIResult
	Report string

	key        string
	report     string
	version    string
	itemName   string
	itemFields map[string]string
}

var metadataMap = map[string]string{
	"httpCode":       "int",
	"errorType":      "string",
	"description":    "string",
	"cappingApplied": "bool",
	"cappingLimit":   "int",
	"queryString":    "string",
}

func newElexonAPI(keyFn, report, version, itemName string, fields map[string]string) (*ElexonAPI, error) {
	content, err := ioutil.ReadFile(keyFn)
	if err != nil {
		return nil, fmt.Errorf("Unable to read API key from %s: %s", keyFn, err)
	}
	return &ElexonAPI{
		key:        strings.Trim(string(content), " "),
		Report:     report,
		version:    version,
		itemName:   itemName,
		itemFields: fields}, nil
}

func (ap *ElexonAPI) GetData(args map[string]string) error {
	params := url.Values{}
	params.Add("APIKey", ap.key)
	params.Add("ServiceType", "xml")
	for k, v := range args {
		params.Add(k, v)
	}
	url := fmt.Sprintf("https://api.bmreports.com/BMRS/%s/%s?%s", ap.Report, ap.version, params.Encode())
	resp, err := http.Get(url)
	if err != nil {
		return err
	}
	content, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Printf("%s API call FAILED: %s", ap.Report, err)
		return err
	}

	ioutil.WriteFile(fmt.Sprintf("%s.xml", ap.Report), content, 0644)

	xmlN := parseXML(content)
	mMap, err := xmlN.getAsMap("responseMetadata", metadataMap)
	if err != nil {
		return err
	}
	ap.Result.data = mMap
	if ap.Result.Int("httpCode") == 200 {
		items, err := xmlN.getAll("responseBody.responseList.item")
		if err != nil {
			return err
		}
		for _, item := range items {
			itemMap := item.asMap(metadataMap)
			if err != nil {
				return err
			}
			result := ElexonAPIResult{itemMap}
			ap.Items = append(ap.Items, result)
		}
		log.Printf("%s API call returned %d items", ap.report, len(ap.Items))
	} else if ap.Result.Int("httpCode") == 204 {
		return fmt.Errorf("%s API call succeeded, but no data returned", ap.report)
	} else {
		log.Printf("%s API call failed: %s [Query: %s]", ap.report, ap.Result.String("description"), ap.Result.String("queryString"))
		return fmt.Errorf("%s API call failed: %s", ap.report, ap.Result.String("description"))
	}
	return nil
}

func (ap ElexonAPI) IsCapped() bool {
	return ap.Result.Bool("cappingApplied")
}

func (ap ElexonAPI) CapLimit() int {
	return ap.Result.Int("cappingLimit")
}

func (ap *ElexonAPI) addFromXML(nodes xmlNode) error {
	meta, err := nodes.get("responseMetadata")
	if err != nil {
		log.Print(err)
		return err
	}
	fmt.Println(meta)
	return nil
}

func BM1320(keyFn string) (*ElexonAPI, error) {
	var fields = map[string]string{
		"timeSeriesID":     "string",
		"settlementDate":   "string",
		"settlementPeriod": "string",
		"quantity":         "string",
		"flowDirection":    "string",
		"reasonCode":       "string",
		"documentType":     "string",
		"processType":      "string",
		"resolution":       "string",
		"curveType":        "string",
		"activeFlag":       "string",
		"documentID":       "string",
		"documentRevNum":   "string",
	}
	return newElexonAPI(keyFn, "B1320", "V1", "CongestionCounterTrade", fields)
}

func BM1420(keyFn string) (*ElexonAPI, error) {
	var fields = map[string]string{
		"documentType":              "string",
		"businessType":              "string",
		"processType":               "string",
		"timeSeriesID":              "string",
		"powerSystemResourceType":   "string",
		"year":                      "int",
		"bMUnitID":                  "string",
		"registeredResourceEICCode": "string",
		"nominal":                   "int",
		"nGCBMUnitID":               "string",
		"registeredResourceName":    "string",
		"activeFlag":                "bool",
		"documentID":                "string",
		"implementationDate":        "date",
	}
	return newElexonAPI(keyFn, "B1420", "V1", "ConfigurationData", fields)
}
