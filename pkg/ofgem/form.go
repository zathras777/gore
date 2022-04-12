package ofgem

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"net/http/cookiejar"
	"net/url"
	"os"
	"strconv"
	"strings"
)

type Form struct {
	StartURL string
	action   string
	client   http.Client
	cookies  *cookiejar.Jar
	data     *FormData
}

func NewForm(start string) *Form {
	jar, err := cookiejar.New(nil)
	if err != nil {
		log.Fatalf("Unable to create a cookie jar??: %s", err)
	}

	// Try and improve performance
	t := http.DefaultTransport.(*http.Transport).Clone()
	t.MaxIdleConns = 100
	t.MaxConnsPerHost = 100
	t.MaxIdleConnsPerHost = 100

	form := &Form{
		StartURL: MakeUrl(start, true),
		client:   http.Client{Jar: jar, Transport: t},
		cookies:  jar,
		data:     NewFormData(),
	}
	if err := form.Get(); err != nil {
		log.Fatalf("Unable to create a new form instance for URL %s", start)
	}
	return form
}

func (f *Form) Get() error {
	if len(f.cookies.Cookies(baseOfgemURL)) == 0 {
		err := f.doGet(MakeUrl("ReportManager.aspx?ReportVisibility=1&ReportCategory=0", true))
		if err != nil {
			return err
		}
	}
	err := f.doGet(f.StartURL)
	if err != nil {
		return err
	}

	f.data.setValueById("ReportViewer$ctl10", "ltr")
	f.data.setValueById("ReportViewer$ctl11", "standards")
	f.data.addInput("__ASYNCPOST", "hidden", "true")

	return nil
}

func (f *Form) Post() error {
	return f.doPost()
}

func (f *Form) Submit(script string) error {
	f.data.setValueById("__EVENTTARGET", "ReportViewer$ctl09$Reserved_AsyncLoadTarget")
	f.data.addInput("ScriptManager1", "hidden", script)
	f.data.setValueByLabel("Page Size", "25")
	return f.doPost()
}

func (f Form) ExportAvailable() bool {
	return len(f.data.exportUrlBase) > 0
}

func (f Form) GetData(xfmt string) ([]byte, error) {
	url := MakeUrl(f.data.exportUrlBase+xfmt, false)

	fmt.Printf("Data - GET: %s\n", url)
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

func (f *Form) SetValueByLabel(lbl string, val string) error {
	if err := f.data.setValueByLabel(lbl, val); err != nil {
		return err
	}
	if f.data.updateRqd {
		fmt.Println("Need to update the form...")
		if err := f.doPost(); err != nil {
			return err
		}
	}
	return nil
}

func (f *Form) doGet(url string) error {
	log.Printf("GET: %s\n", url)

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return err
	}
	resp, err := f.client.Do(req)
	if err != nil {
		return err
	}
	err = f.data.Parse(resp)
	//f.data.Dump()
	return err
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

func (f *Form) doPost() error {
	log.Printf("POST: %s\n", f.data.actionURL)
	postdata := f.data.GetPostValues()

	// Debugging...
	/*
		keys := make([]string, 0, len(postdata))
		for k := range postdata {
			keys = append(keys, k)
		}
		sort.Strings(keys)

		for _, k := range keys {
			val := postdata[k][0]
			if len(val) > 100 {
				val = val[:100]
			}
			fmt.Printf("-> %s : %s\n", k, val)
		}
	*/
	actionUrl := MakeUrl(f.data.actionURL, true)

	req, err := http.NewRequest("POST", actionUrl, strings.NewReader(postdata.Encode()))
	if err != nil {
		log.Fatal(err)
	}
	req.Header.Add("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Add("Content-Length", strconv.Itoa(len(postdata.Encode())))
	req.Header.Add("User-Agent", "Mozilla")
	req.Header.Add("X-Requested-With", "XMLHttpRequest")
	req.Header.Add("X-MicrosoftAjax", "Delta=true")
	ref, err := url.QueryUnescape(f.data.actionURL)
	if err == nil {
		req.Header.Add("Referer", ref)
	}
	req.Header.Add("Origin", "https://renewablesandchp.ofgem.gov.uk")

	resp, err := f.client.Do(req)
	if err != nil {
		return err
	}

	if err := processDelta(resp, f.data); err != nil {
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
