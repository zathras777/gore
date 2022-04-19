package ofgem

import (
	"fmt"
	"gore/pkg/gore"
	"io/ioutil"
	"time"
)

type StationSearch struct {
	form *form
}

var stationAttrMap = map[string]string{
	"GeneratorID":                     "string",
	"StatusName":                      "string",
	"GeneratorName":                   "string",
	"SchemeName":                      "string",
	"Capacity":                        "float",
	"Country":                         "string",
	"TechnologyName":                  "string",
	"OutputType":                      "string",
	"AccreditationDate":               "date",
	"CommissionDate":                  "date",
	"textbox6:OrganisationContact":    "string",
	"textbox61:OrganisationAddress":   "string",
	"FaxNumber:OrganisationFaxNumber": "string",
	"textbox65:StationAddress":        "string",
}

func NewStationSearch() *StationSearch {
	return &StationSearch{
		form: newForm("ReportViewer.aspx?ReportPath=/Renewables/Accreditation/AccreditedStationsExternalPublic&ReportVisibility=1&ReportCategory=1"),
	}
}

func (ss *StationSearch) Scheme(scheme string) error {
	return ss.form.setValueByLabel("Scheme", scheme)
}

func (ss *StationSearch) CommissionYear(year int) error {
	return ss.form.setValueByLabel("Commission Year", fmt.Sprintf("%d", year))
}

func (ss *StationSearch) CommissionMonth(month int) error {
	return ss.form.setValueByLabel("Commission Month", time.Month(month).String()[:3])
}

func (ss *StationSearch) AccreditationYear(year int) error {
	return ss.form.setValueByLabel("Accreditation Year", fmt.Sprintf("%d", year))
}

func (ss *StationSearch) AccreditationMonth(month int) error {
	return ss.form.setValueByLabel("Accreditation Month", time.Month(month).String()[:3])
}

func (ss *StationSearch) GetResults() (result gore.ResultSet) {
	result.QueryName = "stationsearch"
	if err := ss.form.Submit("ReportViewer$ctl04$ctl00"); err != nil {
		result.Query.Error = err
		return
	}
	if !ss.form.ExportAvailable() {
		result.Query.Error = fmt.Errorf("Unable to retrieve data?")
		return
	}
	data, err := ss.form.getData("XML")
	if err != nil {
		result.Query.Error = err
		return
	}
	result.Query.Completed = true
	ioutil.WriteFile("station_data.xml", data, 0644)
	xmlData, err := gore.ParseXML(data)
	if err != nil {
		result.Query.Error = err
		return
	}
	details, err := xmlData.GetAll("tableAccreditation.Detail_Collection.Detail")
	for _, detail := range details {
		result.Results = append(result.Results, gore.ResultItem{Data: detail.AttrAsMap(stationAttrMap)})
	}
	return
}
