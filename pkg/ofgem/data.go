package ofgem

import (
	"fmt"
	"log"
	"net/url"
	"regexp"
	"sort"
	"strings"
)

var createRe = regexp.MustCompile("\\$create\\(Microsoft.Reporting.WebFormsClient\\._.*ParameterInputControl, .*;")
var ctlIdListRe = regexp.MustCompile("\"CustomInputControlIdList\":\\[.*\\],")
var postBackRe = regexp.MustCompile("\"PostBackOnChange\":([a-z]+),")

func (f form) getPostValues() url.Values {
	values := url.Values{}
	for _, inp := range f.inputs {
		values.Set(inp.id, inp.Value())
	}
	for _, sel := range f.selects {
		values.Set(sel.ID, sel.Value())
	}
	for _, dd := range f.dropdowns {
		dd.addPostValues(values)
	}
	return values
}

func (f *form) addOrUpdateInput(elem *HTMLElement) {
	name := elem.Attr("name")
	if strings.Contains(name, "$ddDropDownButton") || strings.Contains(name, "$txtValue") {
		return
	}
	inp, ck := f.inputs[name]
	if ck {
		inp.updateFromHTML(elem)
		return
	}
	if !strings.Contains(name, "$divDropDown") {
		inp = newInputFromHTML(elem)
		if inp == nil {
			return
		}
		f.inputs[inp.id] = inp
		f.types[inp.id] = "Input"
		return
	}
	// It's a dropdown input
	dd := f.getDropDown(name, true)
	if strings.Contains(name, "$HiddenIndices") {
		dd.updateSelected(elem.Attr("value"))
	}
}

func (f *form) addOrUpdateSelect(elem *HTMLElement) {
	s, ck := f.selects[elem.Attr("name")]
	if ck {
		log.Printf("Need to implement updating select %s\n", elem.Attr("name"))
	} else {
		s = newSelectFromHTML(elem)
		f.selects[s.ID] = s
		f.types[s.ID] = "Select"
	}
}

func (f *form) recordLabel(elem *HTMLElement) {
	id := strings.ReplaceAll(elem.Attr("for"), "_", "$")
	name := strings.ReplaceAll(elem.Text, "\u00a0", " ")
	if !strings.Contains(id, "$divDropDown") {
		if strings.Contains("TrueFalseNULL", name) {
			return
		}
		id = strings.ReplaceAll(id, "$txtValue", "")
		f.labels[name] = id
		return
	}
	dd := f.getDropDown(id, false)
	dd.addOptionLabel(id, name)
}

func (f *form) recordScript(elem *HTMLElement) {
	matches := createRe.FindAllString(elem.Text, -1)
	if len(matches) == 0 {
		return
	}
	for _, line := range matches {
		ids := ctlIdListRe.FindString(line)
		parts := strings.Split(ids, "\"")
		pb := postBackRe.FindString(line)
		f.setPostback(strings.ReplaceAll(parts[3], "_", "$"), strings.Contains(string(pb), "true"))
	}
}

func (f *form) setPostback(name string, val bool) {
	if strings.Contains(name, "$ddDropDownButton") {
		name = dropdownId(name)
	}
	f.postbacks[name] = val
}

func (f form) checkPostback(name string) bool {
	val, ck := f.postbacks[name]
	if !ck {
		return false
	}
	return val
}

func (f *form) addInput(name, typ, val string) error {
	_, ck := f.inputs[name]
	if ck {
		return fmt.Errorf("There is already an input with the name %s", name)
	}
	inp := &input{name, typ, val}
	f.inputs[name] = inp
	f.types[name] = "Input"
	return nil
}

func (f *form) setValueByLabel(lbl, val string) error {
	id, ck := f.labels[lbl]
	if !ck {
		return fmt.Errorf("Unable to find a label '%s'", lbl)
	}
	return f.setValueById(id, val)
}

func (f *form) setValueById(id, val string) (err error) {
	tp, ck := f.types[id]
	if !ck {
		return fmt.Errorf("Unable to find a field with ID %s", id)
	}
	switch tp {
	case "Input":
		f.inputs[id].value = val
	case "Select":
		err = f.selects[id].setValue(val)
	case "DropDown":
		err = f.dropdowns[id].setValues(val)
	}
	if f.checkPostback(id) == true {
		return f.doPost(id)
	}
	return err
}

func (f form) Dump() {
	fmt.Print("\nCurrent Form Data:\n\n")
	keys := make([]string, 0, len(f.types))
	for k := range f.types {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	for _, nm := range keys {
		switch f.types[nm] {
		case "Input":
			val := fmt.Sprintf("%s", f.inputs[nm].Value())
			if len(val) > 100 {
				val = val[:100] + "..."
			}
			fmt.Printf("Input: %s -> %s\n", nm, val)
		case "DropDown":
			dd := f.getDropDown(nm, false)
			fmt.Printf("DropDown: %s -> %d options listed, %d selected [ %v ]\n", nm, len(dd.options), len(dd.selected), dd.selected)
			fmt.Println(dd.getTextValue())
		case "Select":
			fmt.Printf("Select: %s -> %d options listed -> %s\n", nm, len(f.selects[nm].options), f.selects[nm].Value())
		}
	}
	lbls := make([]string, 0, len(f.labels))
	for k := range f.labels {
		lbls = append(lbls, k)
	}
	sort.Strings(lbls)
	fmt.Println("\n\nLabels:")
	for _, l := range lbls {
		fmt.Printf("%s -> %s\n", l, f.labels[l])
	}
	fmt.Print("\n\nPostbacks:\n")
	for k, v := range f.postbacks {
		fmt.Printf("%s -> %t\n", k, v)
	}
	fmt.Print("\n\n")
}

func (fd *form) getDropDown(id string, create bool) *dropDown {
	did := dropdownId(id)
	dd, ck := fd.dropdowns[did]
	if !ck {
		if !create {
			log.Fatalf("Unable to find a DropDown with an ID of %s", did)
		}
		dd = &dropDown{ID: did, options: make(map[int]string)}
		fd.dropdowns[did] = dd
		fd.types[did] = "DropDown"
	}
	return dd
}
