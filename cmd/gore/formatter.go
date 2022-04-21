package main

import (
	"fmt"
	"strings"

	"github.com/zathras777/gore/pkg/gore"
)

type formatterColumn struct {
	title    string
	field    string
	format   string
	width    int
	decimals int
}

type formatterRow struct {
	columns []formatterColumn
}

func (fr formatterRow) generateFormat() (fmtString string) {
	for _, col := range fr.columns {
		fmtString += col.formatString() + " "
	}
	fmtString += "\n"
	return
}

func (fr formatterRow) formatTitles() (titleString string) {
	underscore := ""
	for _, col := range fr.columns {
		_fmt := "%"
		if strings.Contains("stringdatetime", col.format) {
			_fmt += "-"
		}
		_fmt += fmt.Sprintf("%ds", col.width)
		titleString += fmt.Sprintf(_fmt, col.title) + " "
		underscore += strings.Repeat("=", col.width) + " "
	}
	titleString += "\n" + underscore
	return
}

func (fr formatterRow) printRows(items []gore.ResultItem) {
	rowFmt := fr.generateFormat()
	for _, item := range items {
		var data []interface{}
		for _, col := range fr.columns {
			var dd interface{}
			switch col.format {
			case "string":
				cStr := item.String(col.field)
				if len(cStr) > col.width {
					dd = cStr[:col.width-3] + "..."
				} else {
					dd = cStr
				}
			case "int":
				dd = item.Int(col.field)
			case "float":
				dd = item.Float(col.field)
			case "bool":
				if item.Bool(col.field) {
					dd = "Yes"
				} else {
					dd = "No"
				}
			case "date":
				dt := item.Date(col.field)
				dd = dt.Format("2006-01-02")
			case "time":
				dt := item.Date(col.field)
				dd = dt.Format("15:04")
			case "datetime":
				dt := item.Date(col.field)
				dd = dt.Format("2006-01-02 15:04")
			default:
				dd = "?"
			}
			data = append(data, dd)
		}
		fmt.Printf(rowFmt, data...)
	}
}

func (fc formatterColumn) formatString() (fmtString string) {
	fmtString = "%"
	switch fc.format {
	case "string", "date", "time", "datetime", "bool":
		fmtString += fmt.Sprintf("-%ds", fc.width)
	case "int":
		fmtString += fmt.Sprintf("%dd", fc.width)
	case "float":
		fmtString += fmt.Sprintf("%d.%df", fc.width, fc.decimals)
	}
	return
}
