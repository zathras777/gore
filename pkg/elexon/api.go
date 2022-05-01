package elexon

import (
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"strings"

	"github.com/zathras777/gore/pkg/gore"
)

type ElexonAPI struct {
	Report       ElexonReport
	Result       gore.ResultSet
	MultiResults map[string]gore.ResultSet

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
	ap := ElexonAPI{Report: cfg}
	ap.Result.QueryName = ap.Report.Name
	return &ap, nil
}

func (ap *ElexonAPI) ReadKeyFile(keyFn string) error {
	content, err := ioutil.ReadFile(keyFn)
	if err != nil {
		return fmt.Errorf("Unable to read API key from %s: %s", keyFn, err)
	}
	ap.key = strings.Trim(string(content), "\n")
	ap.key = strings.Trim(ap.key, " ")
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

	xmlN, err := gore.ParseXML(content)
	if err != nil {
		return err
	}
	// Regardless of whether there is a multi dataset response or not, the results of the query
	// are only sent once. Read them here and create a gore.QueryResult that we can use for subsequent
	// gore.ReultSet creation.
	qr := queryResultFromResponse(xmlN)
	if qr.Error != nil {
		ap.Result.Query = qr
		return err
	}

	if len(ap.Report.Multi) == 0 {
		ap.Result.Query = qr
		if qr.Empty {
			return nil
		}

		items, err := xmlN.GetAll("responseBody.responseList.item")
		if err != nil {
			return err
		}

		for _, item := range items {
			itemMap := item.AsMap(ap.Report.Fields)
			if err != nil {
				return err
			}
			result := gore.ResultItem{Data: itemMap}
			ap.Result.Results = append(ap.Result.Results, result)
		}
		log.Printf("%s API call returned %d items", ap.Report.Name, len(ap.Result.Results))
	}

	return nil
}

func queryResultFromResponse(xmlN gore.XmlNode) gore.QueryResult {
	var qr gore.QueryResult
	mMap, err := xmlN.GetAsMap("responseMetadata", metadataMap)
	if err != nil {
		qr.Completed = false
		qr.Error = err
		return qr
	}
	qr.Completed = true
	if mMap["httpCode"].(int) != 200 && mMap["httpCode"].(int) != 204 {
		qr.Error = fmt.Errorf("%s: %s", mMap["errorType"].(string), mMap["description"].(string))
		return qr
	}
	if mMap["httpCode"].(int) == 204 {
		qr.Empty = true
		return qr
	}
	_, ck := mMap["cappingApplied"]
	if !ck {
		return qr
	}
	if mMap["cappingApplied"].(bool) {
		qr.Capped = true
		qr.CapLimit = mMap["cappingLimit"].(int)
	}
	return qr
}
