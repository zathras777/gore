package ofgem

import (
	"fmt"
	"log"
	"net/url"
	"strconv"
	"strings"
)

type DropDown struct {
	ID       string
	Selected []int
	Options  map[int]string
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

func (dd *DropDown) addOptionLabel(id, label string) {
	dd.Options[dropdownOptionId(id)] = label
}

func (dd *DropDown) updateSelected(idlist string) {
	for _, v := range strings.Split(idlist, ",") {
		n, err := strconv.Atoi(v)
		if err != nil {
			log.Fatalf("Cannot convert %s into a number: %s", v, err)
		}
		dd.Selected = append(dd.Selected, n)
	}
}

func (dd DropDown) getTextValue() string {
	vals := make([]string, len(dd.Selected))
	for i, n := range dd.Selected {
		vals[i] = dd.Options[n+2]
	}
	return strings.Join(vals, ", ")
}

func (dd DropDown) selectedAsString() string {
	nums := make([]string, len(dd.Selected))
	for i, num := range dd.Selected {
		nums[i] = fmt.Sprintf("%d", num)
	}
	return strings.Join(nums, ",")
}

func (dd DropDown) addPostValues(values url.Values) {
	values.Set(dd.ID+"$divDropDown$ctl01$HiddenIndices", dd.selectedAsString())
	values.Set(dd.ID+"$divDropDown$txtValue", dd.getTextValue())
}
