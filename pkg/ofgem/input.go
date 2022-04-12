package ofgem

import "fmt"

type input struct {
	id    string
	typ   string
	value string
}

func newInputFromHTML(elem *HTMLElement) *input {
	id := elem.Attr("name")
	val, set := valueFromElement(elem)
	if !set {
		return nil
	}
	return &input{id, elem.Attr("type"), val}
}

func (i *input) updateFromHTML(elem *HTMLElement) {
	val, set := valueFromElement(elem)
	if !set {
		return
	}
	i.value = val
}

func valueFromElement(elem *HTMLElement) (string, bool) {
	switch elem.Attr("type") {
	case "hidden":
		return elem.Attr("value"), true
	case "checkbox":
		if elem.Attr("checked") == "checked" {
			return "on", true
		}
		return "off", true
	case "radio":
		if elem.Attr("checked") == "checked" {
			return elem.Attr("value"), true
		}
		return "", false
	default:
		fmt.Printf("Unhandled input type: %s", elem.Attr("type"))
	}
	return elem.Attr("value"), true
}

func (i input) Value() string {
	return i.value
}
