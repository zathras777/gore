package ofgem

import (
	"bytes"
	"io"
	"net/http"

	"github.com/PuerkitoBio/goquery"
	"golang.org/x/net/html"
)

type HTMLElement struct {
	Name       string
	Text       string
	attributes []html.Attribute
	DOM        *goquery.Selection
}

func (fd *FormData) Parse(resp *http.Response) error {
	content, err := io.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	rdr := bytes.NewReader(content)
	doc, err := goquery.NewDocumentFromReader(rdr)
	if err != nil {
		return err
	}

	doc.Find("form").Each(func(_ int, s *goquery.Selection) {
		action, ck := s.Attr("action")
		if ck {
			fd.actionURL = action
		}
		for _, n := range s.Nodes {
			e := NewHTMLElement(n, s)

			e.ForEach("input", func(elem *HTMLElement) {
				fd.addOrUpdateInput(elem)
			})
			e.ForEach("select", func(elem *HTMLElement) {
				fd.addOrUpdateSelect(elem)
			})
			e.ForEach("label", func(elem *HTMLElement) {
				fd.recordLabel(elem)
			})
			e.ForEach("script", func(elem *HTMLElement) {
				fd.recordScript(elem)
			})
		}
	})
	return nil
}

func NewHTMLElement(n *html.Node, s *goquery.Selection) *HTMLElement {
	return &HTMLElement{
		n.Data,
		goquery.NewDocumentFromNode(n).Text(),
		n.Attr,
		s,
	}
}

func (h *HTMLElement) Attr(k string) string {
	for _, a := range h.attributes {
		if a.Key == k {
			return a.Val
		}
	}
	return ""
}

func (h *HTMLElement) ForEach(goquerySelector string, callback func(*HTMLElement)) {
	h.DOM.Find(goquerySelector).Each(func(_ int, s *goquery.Selection) {
		for _, n := range s.Nodes {
			callback(NewHTMLElement(n, s))
		}
	})
}
