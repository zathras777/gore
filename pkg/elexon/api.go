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
	Report ElexonReport
	Result ElexonAPIResult
	Items  []ElexonAPIResult

	key string
}

var metadataMap = map[string]string{
	"httpCode":       "int",
	"errorType":      "string",
	"description":    "string",
	"cappingApplied": "bool",
	"cappingLimit":   "int",
	"queryString":    "string",
}

func NewElexonReport(report string) (*ElexonAPI, error) {
	cfg, ck := ElexonReports[strings.ToLower(report)]
	if !ck {
		return nil, fmt.Errorf("Unable to find a configured report %s", report)
	}
	return &ElexonAPI{Report: cfg}, nil
}

func (ap *ElexonAPI) ReadKeyFile(keyFn string) error {
	content, err := ioutil.ReadFile(keyFn)
	if err != nil {
		return fmt.Errorf("Unable to read API key from %s: %s", keyFn, err)
	}
	ap.key = strings.Trim(string(content), " ")
	return nil
}

func (ap *ElexonAPI) GetData(args map[string]string) error {
	params := url.Values{}
	_, ck := args["APIKey"]
	if !ck && len(ap.key) == 0 {
		return fmt.Errorf("You either need to supply the APIKey parameter or call ReadKeyFile() before getting data")
	}
	if !ck {
		params.Add("APIKey", ap.key)
	}
	params.Add("ServiceType", "xml")

	for _, rqd := range ap.Report.RqdParams {
		_, ck := args[rqd]
		if !ck {
			return fmt.Errorf("Calls to Report %s require the %s parameter to be set", ap.Report.Name, rqd)
		}
	}
	for k, v := range args {
		params.Add(k, v)
	}
	if ap.Report.updateParams != nil {
		ap.Report.updateParams(params)
	}

	url := fmt.Sprintf("https://api.bmreports.com/BMRS/%s/%s?%s", ap.Report.Name, ap.Report.Version, params.Encode())
	resp, err := http.Get(url)
	if err != nil {
		return err
	}
	if resp.StatusCode != 200 {
		fmt.Println(resp)
		return fmt.Errorf("Elexon server responded with a status code %d. Url was %s", resp.StatusCode, url)
	}
	content, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Printf("%s API call FAILED: %s", ap.Report.Name, err)
		return err
	}
	if len(content) == 0 {
		log.Printf("Empty response received. Status Code %d\n\n", resp.StatusCode)
		fmt.Println(resp)
		return fmt.Errorf("Empty response receieved from Elexon")
	}

	ioutil.WriteFile(fmt.Sprintf("%s.xml", ap.Report.Name), content, 0644)

	xmlN, err := parseXML(content)
	if err != nil {
		return err
	}
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
			itemMap := item.asMap(ap.Report.Fields)
			if err != nil {
				return err
			}
			result := ElexonAPIResult{itemMap}
			ap.Items = append(ap.Items, result)
		}
		log.Printf("%s API call returned %d items", ap.Report.Name, len(ap.Items))
	} else if ap.Result.Int("httpCode") == 204 {
		return fmt.Errorf("%s API call succeeded, but no data returned", ap.Report.Name)
	} else {
		log.Printf("%s API call failed: %s [Query: %s]", ap.Report.Name, ap.Result.String("description"), ap.Result.String("queryString"))
		return fmt.Errorf("%s API call failed: %s", ap.Report.Name, ap.Result.String("description"))
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
