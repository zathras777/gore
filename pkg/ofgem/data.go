package ofgem

import (
	"fmt"
	"log"
	"net/url"
	"regexp"
	"sort"
	"strings"
)

type FormData struct {
	inputs        map[string]*input
	Dropdowns     map[string]*DropDown
	Selects       map[string]*Select
	Types         map[string]string
	Labels        map[string]string
	Postbacks     map[string]bool
	actionURL     string
	updateRqd     bool
	exportUrlBase string
}

var createRe = regexp.MustCompile("\\$create\\(Microsoft.Reporting.WebFormsClient\\._.*ParameterInputControl, .*;")
var ctlIdListRe = regexp.MustCompile("\"CustomInputControlIdList\":\\[.*\\],")
var postBackRe = regexp.MustCompile("\"PostBackOnChange\":([a-z]+),")

func NewFormData() *FormData {
	fd := &FormData{
		inputs:    make(map[string]*input),
		Dropdowns: make(map[string]*DropDown),
		Selects:   make(map[string]*Select),
		Types:     make(map[string]string),
		Postbacks: make(map[string]bool),
		Labels:    make(map[string]string),
	}
	fd.addInput("__EVENTTARGET", "hidden", "")
	fd.addInput("__EVENTARGUMENT", "hidden", "")
	return fd
}

func (fd FormData) GetPostValues() url.Values {
	values := url.Values{}
	for _, inp := range fd.inputs {
		values.Set(inp.id, inp.Value())
	}
	for _, sel := range fd.Selects {
		values.Set(sel.ID, sel.Value())
	}
	for _, dd := range fd.Dropdowns {
		dd.addPostValues(values)
	}
	return values
}

func (fd *FormData) addOrUpdateInput(elem *HTMLElement) {
	name := elem.Attr("name")
	if strings.Contains(name, "$ddDropDownButton") || strings.Contains(name, "$txtValue") {
		return
	}
	inp, ck := fd.inputs[name]
	if ck {
		inp.updateFromHTML(elem)
		return
	}
	if !strings.Contains(name, "$divDropDown") {
		inp = newInputFromHTML(elem)
		if inp == nil {
			return
		}
		fd.inputs[inp.id] = inp
		fd.Types[inp.id] = "Input"
		return
	}
	// It's a dropdown input
	dd := fd.getDropDown(name, true)
	if strings.Contains(name, "$HiddenIndices") {
		dd.updateSelected(elem.Attr("value"))
	}
}

func (fd *FormData) addOrUpdateSelect(elem *HTMLElement) {
	s, ck := fd.Selects[elem.Attr("name")]
	if ck {
		fmt.Printf("Need to update select %s\n", elem.Attr("name"))
	} else {
		s = newSelectFromHTML(elem)
		fd.Selects[s.ID] = s
		fd.Types[s.ID] = "Select"
	}
}

func (fd *FormData) recordLabel(elem *HTMLElement) {
	id := strings.ReplaceAll(elem.Attr("for"), "_", "$")
	name := strings.ReplaceAll(elem.Text, "\u00a0", " ")
	if !strings.Contains(id, "$divDropDown") {
		fd.Labels[name] = id
		return
	}
	dd := fd.getDropDown(id, false)
	dd.addOptionLabel(id, name)
}

func (fd *FormData) recordScript(elem *HTMLElement) {
	matches := createRe.FindAllString(elem.Text, -1)
	if len(matches) == 0 {
		return
	}
	for _, line := range matches {
		ids := ctlIdListRe.FindString(line)
		parts := strings.Split(ids, "\"")
		pb := postBackRe.FindString(line)
		fd.setPostback(strings.ReplaceAll(parts[3], "_", "$"), strings.Contains(string(pb), "true"))
	}
}

func (fd *FormData) setPostback(name string, val bool) {
	if strings.Contains(name, "$divDropDown") {
		name = dropdownId(name)
	}
	fd.Postbacks[name] = val
}

func (fd FormData) checkPostback(name string) bool {
	val, ck := fd.Postbacks[name]
	if !ck {
		return false
	}
	return val
}

func (fd *FormData) addInput(name, typ, val string) error {
	_, ck := fd.inputs[name]
	if ck {
		return fmt.Errorf("There is already an input with the name %s", name)
	}
	inp := &input{name, typ, val}
	fd.inputs[name] = inp
	fd.Types[name] = "Input"
	return nil
}

func (fd *FormData) setValueByLabel(lbl, val string) error {
	id, ck := fd.Labels[lbl]
	if !ck {
		return fmt.Errorf("Unable to find a label '%s'", lbl)
	}
	return fd.setValueById(id, val)
}

func (fd *FormData) setValueById(id, val string) (err error) {
	tp, ck := fd.Types[id]
	if !ck {
		return fmt.Errorf("Unable to find a field with ID %s", id)
	}
	switch tp {
	case "Input":
		fd.inputs[id].value = val
	case "Select":
		err = fd.Selects[id].setValue(val)
	case "DropDown":
		log.Printf("Implement DropDown updating!!!")
	}
	if fd.checkPostback(id) {
		fd.inputs["__EVENTTARGET"].value = id
		fd.updateRqd = true
	}
	return err
}

func (fd FormData) Dump() {
	fmt.Print("\nCurrent Form Data:\n\n")
	keys := make([]string, 0, len(fd.Types))
	for k := range fd.Types {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	for _, nm := range keys {
		switch fd.Types[nm] {
		case "Input":
			val := fmt.Sprintf("%s", fd.inputs[nm].Value())
			if len(val) > 100 {
				val = val[:100] + "..."
			}
			fmt.Printf("Input: %s -> %s\n", nm, val)
		case "DropDown":
			dd := fd.getDropDown(nm, false)
			fmt.Printf("DropDown: %s -> %d options listed, %d selected [ %v ]\n", nm, len(dd.Options), len(dd.Selected), dd.Selected)
			fmt.Println(dd.getTextValue())
		case "Select":
			fmt.Printf("Select: %s -> %d options listed -> %s\n", nm, len(fd.Selects[nm].options), fd.Selects[nm].Value())
		}
	}
	lbls := make([]string, 0, len(fd.Labels))
	for k := range fd.Labels {
		lbls = append(lbls, k)
	}
	sort.Strings(lbls)
	fmt.Println("\n\nLabels:")
	for _, l := range lbls {
		fmt.Printf("%s -> %s\n", l, fd.Labels[l])
	}
	fmt.Print("\n\nPostbacks:\n")
	for k, v := range fd.Postbacks {
		fmt.Printf("%s -> %t\n", k, v)
	}
	fmt.Print("\n\n")
}

func (fd *FormData) getDropDown(id string, create bool) *DropDown {
	did := dropdownId(id)
	dd, ck := fd.Dropdowns[did]
	if !ck {
		if !create {
			log.Fatalf("Unable to find a DropDown with an ID of %s", did)
		}
		dd = &DropDown{ID: did, Options: make(map[int]string)}
		fd.Dropdowns[did] = dd
		fd.Types[did] = "DropDown"
	}
	return dd
}
