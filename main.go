package main

import (
	"fmt"
	"gore/pkg/ofgem"
	"log"
)

func main() {
	fmt.Println("Go Renewables...")

	ss := ofgem.NewStationSearch()
	ss.Scheme("REGO")
	ss.CommissionMonth(2)
	ss.CommissionYear(2020)
	ss.GetStations()
	ss.SaveToFile("station_results.xml", "xml")
	ss.SaveToFile("station_results.json", "json")

	cs := ofgem.NewCertificateSearch()
	if err := cs.SetPeriod(2, 2022); err != nil {
		log.Fatal(err)
	}
	if err := cs.GetResults(); err != nil {
		log.Fatal(err)
	}
	cs.SaveToFile("certificate_results.xml", "xml")
	cs.SaveToFile("certificate_results.json", "json")
}
