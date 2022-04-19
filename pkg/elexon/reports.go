package elexon

import (
	"log"
	"net/url"
	"strconv"
	"strings"
	"time"
)

type ElexonReport struct {
	Name         string
	Description  string
	Version      string
	Fields       map[string]string
	RqdParams    []string
	Multi        map[string]string
	updateParams func(url.Values)
}

var ElexonReports = map[string]ElexonReport{
	"b1320": {
		"B1320",
		"Congestion Management Measures: Countertrading",
		"v1",
		map[string]string{
			"timeSeriesID":     "string",
			"settlementDate":   "date",
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
		},
		[]string{"SettlementDate", "Period"},
		map[string]string{},
		nil,
	},
	"b1330": {
		"B1330",
		"Congestion Management Measures: Costs of Congestion Management",
		"v1",
		map[string]string{
			"timeSeriesID":     "string",
			"year":             "int",
			"month":            "string", // JAN, FEB etc??
			"congestionAmount": "float",
			"documentType":     "string",
			"processType":      "string",
			"businessType":     "string",
			"resolution":       "string",
			"activeFlag":       "bool",
			"documentID":       "string",
			"documentRevNum":   "string",
		},
		[]string{"Year", "Month"},
		map[string]string{},
		textMonth,
	},
	"b1420": {
		"B1420",
		"Installed Generation Capacity per Unit",
		"v1",
		map[string]string{
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
		},
		[]string{"Year"},
		map[string]string{},
		nil,
	},

	"b1610": {
		"B1610",
		"Actual Generation Output per Generation Unit",
		"v2",
		map[string]string{
			"documentType":                "string",
			"businessType":                "string",
			"processType":                 "string",
			"timeSeriesID":                "string",
			"curveType":                   "string",
			"settlementDate":              "date",
			"powerSystemResourceType":     "string",
			"registeredResourceEICCode":   "string",
			"marketGenerationUnitEICCode": "string",
			"marketGenerationBMUId":       "string",
			"marketGenerationNGCBMUId":    "string",
			"bMUnitID":                    "string",
			"nGCBMUnitID":                 "string",
			"activeFlag":                  "bool",
			"documentID":                  "string",
			"documentRevNum":              "int",

			//		<Period><timeInterval><start>2022-02-01</start><end>2022-02-01</end></timeInterval><resolution>PT30M</resolution><Point><settlementPeriod>30</settlementPeriod><quantity>63.47</quantity></Point></Period>

		},
		[]string{"SettlementDate", "Period"},
		map[string]string{},
		nil,
	},
	"b1630": {
		"B1630",
		"Actual Or Estimated Wind and Solar Power Generation",
		"v1",
		map[string]string{
			"documentType":                "string",
			"businessType":                "string",
			"processType":                 "string",
			"timeSeriesID":                "string",
			"quantity":                    "float",
			"curveType":                   "string",
			"resolution":                  "string",
			"settlementDate":              "date",
			"settlementPeriod":            "string",
			"PSRType":                     "string",
			"powerSystemResourceType":     "string",
			"registeredResourceEICCode":   "string",
			"marketGenerationUnitEICCode": "string",
			"activeFlag":                  "bool",
			"documentID":                  "string",
			"documentRevNum":              "string",
		},
		[]string{"SettlementDate", "Period"},
		map[string]string{},
		nil,
	},
	"derbmdata": {
		"DERBMDATA",
		"Derived BM Unit Data",
		"v1",
		map[string]string{},
		[]string{},
		map[string]string{},
		nil,
	},
	"fuelinst": {
		"FUELINST",
		"Generation by Fuel Type (24H Instant Data)",
		"v1",
		map[string]string{
			"recordType":                     "string",
			"startTimeOfHalfHrPeriod":        "date",
			"settlementPeriod":               "int",
			"publishingPeriodCommencingTime": "dateTime",
			"ccgt":                           "int",
			"oil":                            "int",
			"coal":                           "int",
			"nuclear":                        "int",
			"wind":                           "int",
			"ps":                             "int",
			"npshyd":                         "int",
			"ocgt":                           "int",
			"other":                          "int",
			"intfr":                          "int",
			"intirl":                         "int",
			"intned":                         "int",
			"intew":                          "int",
			"biomass":                        "int",
			"intnem":                         "int",
			"intelec":                        "int",
			"intifa2":                        "int",
			"intnsl":                         "int",
			"activeFlag":                     "bool",
		},
		[]string{},
		map[string]string{},
		nil,
	},
}

func textMonth(current url.Values) {
	for k, v := range current {
		if k == "Month" {
			num, err := strconv.Atoi(v[0])
			if err != nil {
				log.Printf("Unable to convert %s into a numeric month: %s", v, err)
			} else {
				current.Set(k, strings.ToUpper(time.Month(num).String()[:3]))
			}
		}
	}
}
