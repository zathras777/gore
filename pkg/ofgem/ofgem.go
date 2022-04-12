package ofgem

import (
	"log"
	"net/url"
	"strings"
)

var (
	baseOfgemURL   *url.URL
	publicOfgemURL *url.URL
)

func MakeUrl(uri string, public bool) string {
	if baseOfgemURL == nil {
		u, err := url.Parse("https://renewablesandchp.ofgem.gov.uk")
		if err != nil {
			log.Fatalf("Unable to parse base Ofgem URL: %s", err)
		}
		baseOfgemURL = u
		u, err = baseOfgemURL.Parse("Public/")
		if err != nil {
			log.Fatalf("Unable to create Public URL: %s", err)
		}
		publicOfgemURL = u
	}
	if strings.HasPrefix(uri, "http") {
		return uri
	}
	if strings.HasPrefix(uri, "/") {
		var u *url.URL
		var err error
		if public {
			u, err = publicOfgemURL.Parse(uri)
		} else {
			u, err = baseOfgemURL.Parse(uri[1:])
		}
		if err != nil {
			log.Fatalf("Unable to makeURL from %s: %s", uri, err)
		}
		return u.String()
	}
	if uri[:2] == "./" {
		uri = uri[2:]
	}
	if public == true {
		u, err := publicOfgemURL.Parse(uri)
		if err != nil {
			log.Fatalf("Unable to makeURL from %s: %s", uri, err)
		}
		return u.String()
	}
	u, err := baseOfgemURL.Parse(uri)
	if err != nil {
		log.Fatalf("Unable to makeURL from %s: %s", uri, err)
	}
	return u.String()
}
