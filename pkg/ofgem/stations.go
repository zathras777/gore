package ofgem

import (
	"encoding/json"
	"encoding/xml"
	"fmt"
	"io/ioutil"
	"strings"
	"time"
)

type StationSearch struct {
	Results []OfgemStation

	form *Form
}

type OfgemStation struct {
	GeneratorID           string `xml:"GeneratorID,attr"`
	StatusName            string `xml:",attr"`
	GeneratorName         string `xml:",attr"`
	SchemeName            string `xml:",attr"`
	Capacity              string `xml:",attr"`
	Country               string `xml:",attr"`
	TechnologyName        string `xml:",attr"`
	OutputType            string `xml:",attr"`
	AccreditationDate     string `xml:",attr"`
	CommissionDate        string `xml:",attr"`
	OrganisationContact   string `xml:"textbox6,attr"`
	OrganisationAddress   string `xml:"textbox61,attr"`
	OrganisationFaxNumber string `xml:"FaxNumber,attr"`
	StationAddress        string `xml:"textbox65,attr"`
}

type StationDetailCollection struct {
	Details []OfgemStation `xml:"Detail"`
}

type StationTable struct {
	DetailCollection StationDetailCollection `xml:"Detail_Collection"`
}

type StationReport struct {
	Table StationTable `xml:"tableAccreditation"`
}

func NewStationSearch() *StationSearch {
	return &StationSearch{form: NewForm("ReportViewer.aspx?ReportPath=/Renewables/Accreditation/AccreditedStationsExternalPublic&ReportVisibility=1&ReportCategory=1")}
}

func (ss *StationSearch) Scheme(scheme string) error {
	return ss.form.SetValueByLabel("Scheme", scheme)
}

func (ss *StationSearch) CommissionYear(year int) error {
	return ss.form.SetValueByLabel("Commission Year", fmt.Sprintf("%d", year))
}

func (ss *StationSearch) CommissionMonth(month int) error {
	return ss.form.SetValueByLabel("Commission Month", time.Month(month).String()[:3])
}

func (ss *StationSearch) GetStations() error {
	if err := ss.form.Submit("ScriptManager1|ReportViewer$ctl04$ctl00"); err != nil {
		return err
	}
	if !ss.form.ExportAvailable() {
		return fmt.Errorf("Unable to retrieve data?")
	}
	data, err := ss.form.GetData("XML")
	if err != nil {
		return err
	}
	ioutil.WriteFile("station_data.xml", data, 0644)
	return ss.parseXML(data)
}

func (ss StationSearch) SaveToFile(fn, xfmt string) (err error) {
	var xdata []byte
	switch strings.ToLower(xfmt) {
	case "json":
		xdata, err = json.Marshal(ss.Results)
	case "xml":
		xdata, err = xml.MarshalIndent(ss.Results, "", "    ")
	default:
		err = fmt.Errorf("Unknown export format: %s", xfmt)
	}
	if err != nil {
		return err
	}
	return ioutil.WriteFile(fn, xdata, 0644)
}

func (ss *StationSearch) parseXML(content []byte) error {
	var report StationReport
	err := xml.Unmarshal(content, &report)
	if err != nil {
		return err
	}
	ss.Results = report.Table.DetailCollection.Details
	return nil

}
