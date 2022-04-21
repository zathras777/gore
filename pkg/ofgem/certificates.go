package ofgem

import (
	"fmt"
	"io/ioutil"
	"strings"
	"time"

	"github.com/zathras777/gore/pkg/gore"
)

type CertificateSearch struct {
	form *form
}

var certAttrMap = map[string]string{
	"textbox4:AccreditationNumber":            "string",
	"textbox13:Station":                       "string",
	"textbox19:StationTIC":                    "float",
	"textbox5:Scheme":                         "string",
	"textbox12:Country":                       "string",
	"textbox15:TechnologyGroup":               "string",
	"textbox31:GenerationType":                "string",
	"textbox18:OutputPeriod":                  "string",
	"textbox21:NoOfCertificates":              "int",
	"textbox24:StartCertificateNo":            "string",
	"textbox27:EndCertificateNo":              "string",
	"textbox37:MWhPerCertificate":             "float",
	"textbox30:IssueDate":                     "dateTime",
	"textbox33:CertificateStatus":             "string",
	"textbox36:StatusDate":                    "dateTime",
	"textbox39:CurrentHolderOrganisationName": "string",
	"textbox45:CompanyRegistrationNumber":     "string",
}

func NewCertificateSearch() *CertificateSearch {
	return &CertificateSearch{
		form: newForm("ReportViewer.aspx?ReportPath=/DatawarehouseReports/CertificatesExternalPublicDataWarehouse&ReportVisibility=1&ReportCategory=2"),
	}
}

func (cs *CertificateSearch) Debug(onoff bool) {
	cs.form.debugDelta = onoff
}

func (cs *CertificateSearch) SetYear(year int) error {
	if err := cs.form.setValueByLabel("Output Period \"Year From\":", fmt.Sprintf("%d", year)); err != nil {
		return err
	}
	if err := cs.form.setValueByLabel("Output Period \"Year To\":", fmt.Sprintf("%d", year)); err != nil {
		return err
	}
	return nil
}

func (cs *CertificateSearch) SetMonth(month int) error {
	if err := cs.form.setValueByLabel("Output Period \"Month From\":", time.Month(month).String()[:3]); err != nil {
		return err
	}
	if err := cs.form.setValueByLabel("Output Period \"Month To\":", time.Month(month).String()[:3]); err != nil {
		return err
	}
	return nil
}

func (cs *CertificateSearch) SetPeriod(month, year int) error {
	if err := cs.SetMonth(month); err != nil {
		return err
	}
	return cs.SetYear(year)
}

func (cs *CertificateSearch) Scheme(schemes string) error {
	return cs.form.setValueByLabel("Scheme:", strings.ToUpper(schemes))
}

func (cs *CertificateSearch) Countries(countries string) error {
	return cs.form.setValueByLabel("Country:", countries)
}

func (cs *CertificateSearch) GetResults() (result gore.ResultSet) {
	result.QueryName = "certificatesearch"
	if err := cs.form.Submit("ReportViewer$ctl09$Reserved_AsyncLoadTarget"); err != nil {
		result.Query.Error = err
		return
	}
	if !cs.form.ExportAvailable() {
		result.Query.Error = fmt.Errorf("Unable to retrieve data as no export URL available")
		return
	}
	data, err := cs.form.getData("XML")
	if err != nil {
		result.Query.Error = err
		return
	}
	result.Query.Completed = true
	ioutil.WriteFile("cert_data.xml", data, 0644)
	xmlData, err := gore.ParseXML(data)
	if err != nil {
		result.Query.Error = err
		return
	}
	details, err := xmlData.GetAll("table1.Detail_Collection.Detail")
	for _, detail := range details {
		info := detail.AttrAsMap(certAttrMap)
		info["MWh"] = float64(info["NoOfCertificates"].(int)) * info["MWhPerCertificate"].(float64)
		result.Results = append(result.Results, gore.ResultItem{Data: info})
	}
	return
}
