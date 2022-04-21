package main

import (
	"flag"
	"fmt"
	"gore/pkg/elexon"
	"gore/pkg/gore"
	"gore/pkg/ofgem"
	"log"
	"os"
	"sort"
	"strconv"
	"strings"
)

var elexonKeyFn string

func createTitle(title string) string {
	return "\n" + title + "\n" + strings.Repeat("=", len(title)) + "\n\n"
}

func main() {
	fmt.Println(createTitle("UK Renewables App"))

	var (
		logFn         string
		verbose       bool
		year          int
		month         int
		period        int
		date          string
		scheme        string
		name          string
		bmunit        string
		xportFormat   string
		xportFilename string
		err           error
	)

	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage of %s:\n", os.Args[0])
		fmt.Printf("\n%s [parameters] [flags] [command]\n", os.Args[0])
		printAvailableCommands()
		fmt.Printf(createTitle("Parameters and Flag Options"))
		flag.PrintDefaults()
	}

	flag.StringVar(&elexonKeyFn, "elexonkey", "elexon.key", "Elexon API Key (required for all Elexon commands)")
	flag.StringVar(&logFn, "log", "gore.log", "Log filename to write to")
	flag.BoolVar(&verbose, "v", false, "Verbose output (disables logging to a file)")
	flag.IntVar(&year, "year", -1, "Specify a year")
	flag.IntVar(&month, "month", -1, "Specify a month")
	flag.StringVar(&date, "date", "", "Date to process for (format is YYYY-MM-DD)")
	flag.IntVar(&period, "period", -1, "Settlement Period for Elexon (1-50)")
	flag.StringVar(&scheme, "scheme", "", "Ofgem Scheme (RO, REGO)")
	flag.StringVar(&name, "name", "", "Name to search for (Elexon or Ofgem)")
	flag.StringVar(&bmunit, "bmunit", "", "BMUnit to search for (Elexon or Ofgem)")
	flag.StringVar(&xportFormat, "exportformat", "", "Export format [json, xml, csv]")
	flag.StringVar(&xportFilename, "exportfilename", "", "Filename for exported data")

	flag.Parse()

	if verbose {
		fmt.Println("Logging to command line only. Log file disabled.")
	} else {
		f, err := os.OpenFile(logFn, os.O_WRONLY|os.O_CREATE|os.O_APPEND, 0644)
		if err != nil {
			log.Fatal(err)
		}
		defer f.Close()
		log.SetOutput(f)
	}

	params := make(map[string]string)
	if year != -1 {
		if year < 100 {
			year += 2000
		}
		params["Year"] = fmt.Sprintf("%d", year)
	}
	if month != -1 {
		if month < 1 || month > 12 {
			fmt.Printf("Month must be between 1 and 12 - not %d\n", month)
			return
		}
		params["Month"] = fmt.Sprintf("%d", month)
	}
	if period != -1 {
		if period < 1 || period > 50 {
			fmt.Printf("Settlement Period must be between 1 and 50 - not %d\n", period)
			return
		}
		params["Period"] = fmt.Sprintf("%d", period)
	}
	if date != "" {
		params["SettlementDate"] = date
	}
	if scheme != "" {
		params["Scheme"] = scheme
	}
	if name != "" {
		params["Name"] = name
	}
	if bmunit != "" {
		params["NGCBMUnitID"] = bmunit
	}

	if verbose {
		fmt.Printf("Params for Query: %v\n", params)
	}

	if len(flag.Args()) == 0 {
		fmt.Println("No command entered.")
		flag.Usage()
		return
	}

	var result gore.ResultSet

	for _, cmd := range flag.Args() {
		switch strings.ToLower(cmd) {
		case "certificatesearch":
			result, err = doCertificateSearch(params)
		case "stationsearch":
			result, err = doStationSearch(params)
		default:
			result, err = processElexonCommand(cmd, params)
		}
		if err != nil {
			fmt.Printf("Unable to process command %s\nError: %s", cmd, err)
			return
		}
	}

	if !result.Query.Completed {
		fmt.Printf("Unable to complete the requested query.\nError: %s\n", result.Query.Error)
		return
	}
	if result.Query.Error != nil {
		fmt.Printf("Query was completed but with an error. No data available.\nError: %s\n", result.Query.Error)
		return
	}
	fmt.Printf("Query succeeded. %d items returned\n", len(result.Results))
	if result.Query.Capped {
		fmt.Printf("Query response was capped at %d items.\n", result.Query.CapLimit)
	}

	cmd, ck := availableCommands[strings.ToLower(result.QueryName)]
	if ck && len(cmd.formatter.columns) > 0 {
		fmt.Println(createTitle(cmd.name + " Output"))
		fmt.Println(cmd.formatter.formatTitles())
		cmd.formatter.printRows(result.Results)
	}

	if xportFilename != "" {
		fmt.Printf("\n Exporting data to %s as %s\n", xportFilename, xportFormat)
		err = result.Export(xportFilename, xportFormat)
		if err != nil {
			fmt.Println(err)
			return
		}
		fmt.Println("Export completed")
	}
}

func printAvailableCommands() {
	fmt.Printf("\n%s", createTitle("Available Commands"))
	cmds := make([]string, 0, len(availableCommands))
	for cmd := range availableCommands {
		cmds = append(cmds, cmd)
	}
	sort.Strings(cmds)
	for _, cmd := range cmds {
		cmdData := availableCommands[cmd]
		fmt.Printf("%20s - %-40s\n", cmd, cmdData.name)
		fmt.Printf("%s%s\n", strings.Repeat(" ", 23), cmdData.description)
	}
	fmt.Println()
}

func processElexonCommand(cmd string, params map[string]string) (gore.ResultSet, error) {
	ap, err := elexon.NewElexonReport(cmd)
	if err != nil {
		return gore.ResultSet{QueryName: ap.Report.Name}, err
	}
	fmt.Printf("Getting data for Elexon Report %s [ %s ]...\n", ap.Report.Name, ap.Report.Description)
	if err = ap.ReadKeyFile(elexonKeyFn); err != nil {
		fmt.Println(err)
		return gore.ResultSet{}, err
	}

	err = ap.GetData(params)
	return ap.Result, err
}

func doCertificateSearch(params map[string]string) (gore.ResultSet, error) {
	cs := ofgem.NewCertificateSearch()
	year, ck := params["Year"]
	if ck {
		num, err := strconv.Atoi(year)
		if err != nil {
			return gore.ResultSet{}, nil
		}
		if err := cs.SetYear(num); err != nil {
			return gore.ResultSet{}, err
		}
	}
	month, ck := params["Month"]
	if ck {
		num, err := strconv.Atoi(month)
		if err != nil {
			return gore.ResultSet{}, nil
		}
		if err := cs.SetMonth(num); err != nil {
			return gore.ResultSet{}, err
		}
	}
	result := cs.GetResults()
	if result.Query.Error != nil {
		return result, result.Query.Error
	}
	return result, nil
}

func doStationSearch(params map[string]string) (gore.ResultSet, error) {
	ss := ofgem.NewStationSearch()
	year, ck := params["Year"]
	if ck {
		num, err := strconv.Atoi(year)
		if err != nil {
			return gore.ResultSet{}, nil
		}
		if err := ss.AccreditationYear(num); err != nil {
			return gore.ResultSet{}, err
		}
	}
	month, ck := params["Month"]
	if ck {
		num, err := strconv.Atoi(month)
		if err != nil {
			return gore.ResultSet{}, nil
		}
		if err := ss.AccreditationMonth(num); err != nil {
			return gore.ResultSet{}, err
		}
	}
	sch, ck := params["Scheme"]
	if ck {
		if err := ss.Scheme(sch); err != nil {
			return gore.ResultSet{}, err
		}
	}

	result := ss.GetResults()
	if result.Query.Error != nil {
		return result, result.Query.Error
	}
	return result, nil
}
