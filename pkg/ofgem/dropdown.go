package ofgem

import (
	"fmt"
	"log"
	"net/url"
	"strconv"
	"strings"
)

type dropDown struct {
	ID       string
	selected []int
	options  map[int]string
}

func dropdownId(id string) string {
	return id[:24]
}

func dropdownOptionId(id string) int {
	did, err := strconv.Atoi(id[len(id)-2:])
	if err != nil {
		log.Fatalf("Unable to get numeric ID from %s: %s", id, err)
	}
	return did
}

func (dd *dropDown) addOptionLabel(id, label string) {
	dd.options[dropdownOptionId(id)] = label
}

func (dd *dropDown) updateSelected(idlist string) {
	for _, v := range strings.Split(idlist, ",") {
		n, err := strconv.Atoi(v)
		if err != nil {
			log.Fatalf("Cannot convert %s into a number: %s", v, err)
		}
		dd.selected = append(dd.selected, n)
	}
}

func (dd dropDown) getTextValue() string {
	vals := make([]string, len(dd.selected))
	for i, n := range dd.selected {
		vals[i] = dd.options[n+2]
	}
	return strings.Trim(strings.Join(vals, ", "), " ")
}

func (dd dropDown) selectedAsString() string {
	nums := make([]string, len(dd.selected))
	for i, num := range dd.selected {
		nums[i] = fmt.Sprintf("%d", num)
	}
	return strings.Join(nums, ",")
}

func (dd *dropDown) setValues(vals string) error {
	var newopts []int
	for _, v := range strings.Split(vals, ",") {
		val := strings.Trim(v, " ")
		for opt, name := range dd.options {
			if name == val {
				newopts = append(newopts, opt-2)
			}
		}
	}
	if len(newopts) == 0 {
		return fmt.Errorf("Unable to find any options that match '%s' in DropDown", vals)
	}
	log.Printf("DropDown: Changing selected from %v to %v", dd.selected, newopts)
	dd.selected = newopts
	return nil
}

func (dd dropDown) addPostValues(values url.Values) {
	//	log.Printf("POST: %s => %s", dd.ID+"$divDropDown$ctl01$HiddenIndices", dd.selectedAsString())
	//	log.Printf("POST: %s => '%s'", dd.ID+"$divDropDown$txtValue", dd.getTextValue())
	values.Set(dd.ID+"$divDropDown$ctl01$HiddenIndices", dd.selectedAsString())
	values.Set(dd.ID+"$divDropDown$txtValue", dd.getTextValue())
}
