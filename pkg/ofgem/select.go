package ofgem

import (
	"fmt"
	"log"
	"strings"
)

type Select struct {
	ID      string
	options []*selectoption
}

type selectoption struct {
	name     string
	value    string
	selected bool
}

func newSelectFromHTML(elem *HTMLElement) *Select {
	s := &Select{ID: elem.Attr("name")}
	elem.ForEach("option", func(e *HTMLElement) {
		if strings.Contains(e.Text, "&lt;Select&nbsp;a&nbsp;Value&gt;") {
			fmt.Println("Skipping not useful select option...")
			return
		}
		opt := &selectoption{value: e.Attr("value"), name: e.Text, selected: e.Attr("selected") == "selected"}
		s.options = append(s.options, opt)
	})
	return s
}

func (s Select) Value() string {
	for _, opt := range s.options {
		if opt.selected {
			return opt.value
		}
	}
	log.Fatalf("Unable to find a selected value for %s", s.ID)
	return ""
}

func (s *Select) setValue(val string) error {
	found := false
	for _, opt := range s.options {
		opt.selected = false
		if opt.name == val {
			if found {
				return fmt.Errorf("Value %s matches more than one select option??", val)
			}
			opt.selected = true
			found = true
		}
	}
	if !found {
		return fmt.Errorf("Unable to find value '%s' in the available options", val)
	}
	return nil
}
