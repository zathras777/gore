package ofgem

import (
	"encoding/json"
	"encoding/xml"
	"fmt"
	"io/ioutil"
	"log"
	"strings"
	"time"
)

type CertificateSearch struct {
	Results []OfgemCertificate

	form *Form
}

type OfgemCertificate struct {
	AccreditationNumber           string  `xml:"textbox4,attr"`
	Station                       string  `xml:"textbox13,attr"`
	StationTIC                    float64 `xml:"textbox19,attr"`
	Scheme                        string  `xml:"textbox5,attr"`
	Country                       string  `xml:"textbox12,attr"`
	TechnologyGroup               string  `xml:"textbox15,attr"`
	GenerationType                string  `xml:"textbox31,attr"`
	OutputPeriod                  string  `xml:"textbox18,attr"`
	NoOfCertificates              float64 `xml:"textbox21,attr"`
	StartCertificateNo            string  `xml:"textbox24,attr"`
	EndCertificateNo              string  `xml:"textbox27,attr"`
	MWhPerCertificate             float64 `xml:"textbox37,attr"`
	IssueDateStr                  string  `xml:"textbox30,attr" json:"IssueDate"`
	CertificateStatus             string  `xml:"textbox33,attr"`
	StatusDateStr                 string  `xml:"textbox36,attr" json:"StatusDate"`
	CurrentHolderOrganisationName string  `xml:"textbox39,attr"`
	CompanyRegistrationNumber     string  `xml:"textbox45,attr"`
}

type CertificateDetailCollection struct {
	Details []OfgemCertificate `xml:"Detail"`
}

type CertificateTable struct {
	DetailCollection CertificateDetailCollection `xml:"Detail_Collection"`
}

type CertificateReport struct {
	Table CertificateTable `xml:"table1"`
}

func NewCertificateSearch() *CertificateSearch {
	return &CertificateSearch{form: NewForm("ReportViewer.aspx?ReportPath=/DatawarehouseReports/CertificatesExternalPublicDataWarehouse&ReportVisibility=1&ReportCategory=2")}
}

func (cs *CertificateSearch) SetPeriod(month, year int) error {
	if err := cs.form.SetValueByLabel("Output Period \"Month From\":", time.Month(month).String()[:3]); err != nil {
		return err
	}
	if err := cs.form.SetValueByLabel("Output Period \"Month To\":", time.Month(month).String()[:3]); err != nil {
		return err
	}
	if err := cs.form.SetValueByLabel("Output Period \"Year From\":", fmt.Sprintf("%d", year)); err != nil {
		return err
	}
	if err := cs.form.SetValueByLabel("Output Period \"Year To\":", fmt.Sprintf("%d", year)); err != nil {
		return err
	}
	return nil
}

func (cs *CertificateSearch) Scheme(schemes string) error {
	return cs.form.SetValueByLabel("Scheme:", strings.ToUpper(schemes))
}

func (cs *CertificateSearch) GetResults() error {
	if err := cs.form.Submit("ScriptManager1|ReportViewer$ctl09$Reserved_AsyncLoadTarget"); err != nil {
		return err
	}
	if !cs.form.ExportAvailable() {
		return fmt.Errorf("Unable to retrieve data as no export URL available")
	}
	data, err := cs.form.GetData("XML")
	if err != nil {
		return err
	}
	ioutil.WriteFile("cert_data.xml", data, 0644)
	return cs.parseXML(data)
}

func (cs CertificateSearch) SaveToFile(fn, xfmt string) (err error) {
	var xdata []byte
	switch strings.ToLower(xfmt) {
	case "json":
		xdata, err = json.Marshal(cs.Results)
	case "xml":
		xdata, err = xml.MarshalIndent(cs.Results, "", "    ")
	default:
		err = fmt.Errorf("Unknown export format: %s", xfmt)
	}
	if err != nil {
		return err
	}
	return ioutil.WriteFile(fn, xdata, 0644)
}

func (cs *CertificateSearch) parseXML(content []byte) error {
	var report CertificateReport
	err := xml.Unmarshal(content, &report)
	if err != nil {
		return err
	}
	cs.Results = report.Table.DetailCollection.Details
	return nil
}

func (oc OfgemCertificate) IssueDate() time.Time {
	tm, err := time.Parse("2006-02-01T15:04:05", oc.IssueDateStr)
	if err != nil {
		log.Fatalf("Unable to parse Ofgem date: %s", err)
	}
	return tm
}

func (oc OfgemCertificate) StatusDate() time.Time {
	tm, err := time.Parse("2006-02-01T15:04:05", oc.StatusDateStr)
	if err != nil {
		log.Fatalf("Unable to parse Ofgem date: %s", err)
	}
	return tm
}

func (oc OfgemCertificate) MWh() float64 {
	return oc.NoOfCertificates * oc.MWhPerCertificate
}

func (oc OfgemCertificate) JSON() []byte {
	js, err := json.Marshal(oc)
	if err != nil {
		log.Fatalf("Unable to create JSON from OfgemCertificate: %s", err)
	}
	return js
}

func (oc OfgemCertificate) String() string {
	var lines []string
	lines = append(lines, fmt.Sprintf("  Accreditation Number: %s", oc.AccreditationNumber))
	lines = append(lines, fmt.Sprintf("          Station Name: %s", oc.Station))
	lines = append(lines, fmt.Sprintf("           Station TIC: %f", oc.StationTIC))
	lines = append(lines, fmt.Sprintf("                Scheme: %s", oc.Scheme))
	lines = append(lines, fmt.Sprintf("               Country: %s", oc.Country))
	lines = append(lines, fmt.Sprintf("      Technology Group: %s", oc.TechnologyGroup))
	lines = append(lines, fmt.Sprintf("       Generation Type: %s", oc.GenerationType))
	lines = append(lines, fmt.Sprintf("         Output Period: %s", oc.OutputPeriod))
	lines = append(lines, fmt.Sprintf("   No. of Certificates: %f [ %f MWh ]", oc.NoOfCertificates, oc.MWh()))
	lines = append(lines, fmt.Sprintf("       1st Cert Number: %s", oc.StartCertificateNo))
	lines = append(lines, fmt.Sprintf("      Last Cert Number: %s", oc.EndCertificateNo))
	lines = append(lines, fmt.Sprintf("          MWh per Cert: %f", oc.MWhPerCertificate))
	lines = append(lines, fmt.Sprintf("            Issue Date: %s", oc.IssueDate().Format("2006-01-02")))
	lines = append(lines, fmt.Sprintf("           Cert Status: %s", oc.CertificateStatus))
	lines = append(lines, fmt.Sprintf("           Status Date: %s", oc.StatusDate().Format("2006-01-02")))
	lines = append(lines, fmt.Sprintf("        Current Holder: %s [%s]", oc.CurrentHolderOrganisationName, oc.CompanyRegistrationNumber))
	return strings.Join(lines, "\n")
}
