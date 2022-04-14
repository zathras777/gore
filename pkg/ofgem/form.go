package ofgem

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"net/http/cookiejar"
	"net/url"
	"os"
	"sort"
	"strconv"
	"strings"
)

type form struct {
	startURL string
	action   string
	client   http.Client
	cookies  *cookiejar.Jar

	inputs        map[string]*input
	dropdowns     map[string]*dropDown
	selects       map[string]*selector
	types         map[string]string
	labels        map[string]string
	postbacks     map[string]bool
	actionURL     string
	updateRqd     bool
	exportUrlBase string

	debugDelta bool
}

func newForm(start string) *form {
	jar, err := cookiejar.New(nil)
	if err != nil {
		log.Fatalf("Unable to create a cookie jar??: %s", err)
	}

	// Try and improve performance
	t := http.DefaultTransport.(*http.Transport).Clone()
	t.MaxIdleConns = 100
	t.MaxConnsPerHost = 100
	t.MaxIdleConnsPerHost = 100

	form := &form{
		startURL: MakeUrl(start, true),
		client:   http.Client{Jar: jar, Transport: t},
		cookies:  jar,

		inputs:    make(map[string]*input),
		dropdowns: make(map[string]*dropDown),
		selects:   make(map[string]*selector),
		types:     make(map[string]string),
		postbacks: make(map[string]bool),
		labels:    make(map[string]string),
	}
	if err := form.get(); err != nil {
		log.Fatalf("Unable to create a new form instance for URL %s", start)
	}
	return form
}

func (f *form) get() error {
	if len(f.cookies.Cookies(baseOfgemURL)) == 0 {
		if err := f.doGet(MakeUrl("ReportManager.aspx?ReportVisibility=1&ReportCategory=0", true)); err != nil {
			return err
		}
	}

	if err := f.doGet(f.startURL); err != nil {
		return err
	}

	// Set some standard values
	f.setValueById("ReportViewer$ctl10", "ltr")
	f.setValueById("ReportViewer$ctl11", "standards")
	// Add the control fields we need that are normally added by scripts.
	f.addInput("__ASYNCPOST", "hidden", "true")
	f.addInput("__LASTFOCUS", "hidden", "")
	f.addInput("__EVENTTARGET", "hidden", "")
	f.addInput("__EVENTARGUMENT", "hidden", "")
	f.addInput("ScriptManager1", "hidden", "")

	return nil
}

func (f *form) Submit(tgt string) error {
	f.setValueByLabel("Page Size", "25")
	return f.doPost(tgt)
}

func (f form) ExportAvailable() bool {
	return len(f.exportUrlBase) > 0
}

func (f form) getData(xfmt string) ([]byte, error) {
	url := MakeUrl(f.exportUrlBase+xfmt, false)

	log.Printf("GET Data: %s\n", url)
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}
	resp, err := f.client.Do(req)
	if err != nil {
		return nil, err
	}
	content, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	return content, err
}

func (f *form) setValueForLabel(lbl string, val string) error {
	if err := f.setValueByLabel(lbl, val); err != nil {
		return err
	}
	return nil
}

func saveResponseBody(resp *http.Response, fn string) error {
	content, err := io.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	if err := os.WriteFile(fn, content, 0644); err != nil {
		return err
	}
	return nil
}

func logPostData(postdata url.Values) {
	keys := make([]string, 0, len(postdata))
	for k := range postdata {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	log.Print("POST DATA:")
	for _, k := range keys {
		val := postdata[k][0]
		if len(val) > 100 {
			val = val[:100]
		}
		log.Printf("  %s : %s\n", k, val)
	}
	log.Print("END OF POST DATA")
}

func (f *form) doGet(url string) error {
	log.Printf("GET: %s\n", url)
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return err
	}
	resp, err := f.client.Do(req)
	if err != nil {
		return err
	}
	if err := f.parseResponse(resp); err != nil {
		return err
	}
	return nil
}

func (f *form) doPost(tgt string) error {
	log.Printf("POST: %s\n", f.actionURL)

	f.setValueById("__EVENTTARGET", tgt)
	f.setValueById("ScriptManager1", "ScriptManager1|"+tgt)

	postdata := f.getPostValues()
	//logPostData(postdata)
	actionUrl := MakeUrl(f.actionURL, true)

	req, err := http.NewRequest("POST", actionUrl, strings.NewReader(postdata.Encode()))
	if err != nil {
		log.Fatal(err)
	}
	req.Header.Add("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Add("Content-Length", strconv.Itoa(len(postdata.Encode())))
	req.Header.Add("User-Agent", "Mozilla")
	req.Header.Add("X-Requested-With", "XMLHttpRequest")
	req.Header.Add("X-MicrosoftAjax", "Delta=true")
	ref, err := url.QueryUnescape(f.actionURL)
	if err == nil {
		req.Header.Add("Referer", ref)
	}
	req.Header.Add("Origin", "https://renewablesandchp.ofgem.gov.uk")

	resp, err := f.client.Do(req)
	if err != nil {
		return err
	}

	if err := processDelta(resp, f); err != nil {
		return err
	}
	return nil
}

// formatRequest generates ascii representation of a request
func formatRequest(r *http.Request) string {
	var request []string
	url := fmt.Sprintf("%v %v %v", r.Method, r.URL, r.Proto)
	request = append(request, url)
	request = append(request, fmt.Sprintf("Host: %v", r.Host))

	for name, headers := range r.Header {
		for _, h := range headers {
			request = append(request, fmt.Sprintf("%v: %v", name, h))
		}
	}
	// Return the request as a string
	return strings.Join(request, "\n")
}
