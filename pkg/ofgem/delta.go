package ofgem

import (
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"regexp"
	"strconv"
	"strings"
)

type deltaInfo struct {
	size    int
	ct      string
	tgt     string
	content string
}

var exportBaseRe = regexp.MustCompile(`\"ExportUrlBase\":\"(.*?)\"`)

func processDelta(resp *http.Response, fd *FormData) error {
	raw, err := io.ReadAll(resp.Body)
	if err != nil {
		return err
	}
	content := string(raw)
	var elements []deltaInfo
	for pos := 0; pos < len(content); {
		di, used := getNextInfo(string(content[pos:]))
		pos += used
		elements = append(elements, di)
	}

	if len(elements) == 0 {
		ioutil.WriteFile("delta.data", raw, 0644)
		return fmt.Errorf("Unable to process the delta response elements. Stored in delta.data for review")
	}

	if elements[0].ct != "#" {
		log.Fatal("Incorrect initial delta segment receieved?")
	}

	for _, element := range elements[1:] {
		switch element.ct {
		case "hiddenField":
			fd.setValueById(element.tgt, element.content)
		case "formAction":
			fd.actionURL = element.content
		case "pageRedirect":
			url, err := url.PathUnescape(element.content)
			if err != nil {
				return err
			}
			return fmt.Errorf("Received pageRedirect to %s", url)
		case "scriptStartupBlock":
			if strings.Contains(element.content, "ExportUrlBase") {
				match := exportBaseRe.FindStringSubmatch(element.content)
				fd.exportUrlBase = strings.ReplaceAll(match[1], "\\u0026", "&")
			}
			//		default:
			//			log.Printf("Unhandled content: %s\n", element.ct)
		}
	}
	return nil
}

func getNextInfo(content string) (deltaInfo, int) {
	di := deltaInfo{}
	pos := 0
	loops := []int{0, 1, 2}
	for i := range loops {
		idx := strings.Index(content[pos:], "|")
		switch i {
		case 0:
			sz, _ := strconv.Atoi(content[:idx])
			di.size = sz
		case 1:
			di.ct = content[pos : pos+idx]
		case 2:
			di.tgt = content[pos : pos+idx]
		}
		pos += idx + 1
	}
	if !strings.HasPrefix(content[pos+di.size:], "|") {
		xtra := strings.Index(content[pos+di.size:], "|")
		di.size += xtra
	}
	di.content = content[pos : pos+di.size]
	return di, pos + di.size + 1
}
