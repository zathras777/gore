package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/zathras777/gore/pkg/elexon"
	"github.com/zathras777/gore/pkg/gore"
	"github.com/zathras777/gore/pkg/ofgem"
)

var elexonKeyFn string
var ofgemFlags *flag.FlagSet = flag.NewFlagSet("ofgem", flag.ExitOnError)
var elexonFlags *flag.FlagSet = flag.NewFlagSet("elexon", flag.ExitOnError)
var stdFlags *flag.FlagSet = flag.NewFlagSet("common", flag.ExitOnError)

func createTitle(title string) string {
	return "\n" + title + "\n" + strings.Repeat("=", len(title)) + "\n\n"
}

func showUsage() {
	fmt.Fprintf(os.Stderr, "Usage of %s:\n", os.Args[0])
	fmt.Printf("\n%s command [parameters] [flags]\n", os.Args[0])
	printAvailableCommands()
	fmt.Println(createTitle("Options"))
	fmt.Println("Options available for all commands:")
	stdFlags.PrintDefaults()
	fmt.Println("\nOptions available for Ofgem commands:")
	ofgemFlags.PrintDefaults()
	fmt.Println("\nOptions available for Elexon commands:")
	elexonFlags.PrintDefaults()
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
		cmd           command
	)

	flag.Usage = showUsage

	ofgemFlags.IntVar(&year, "year", -1, "Specify a year")
	ofgemFlags.IntVar(&month, "month", -1, "Specify a month")
	ofgemFlags.StringVar(&scheme, "scheme", "", "Ofgem Scheme (RO, REGO)")

	elexonFlags.StringVar(&elexonKeyFn, "elexonkey", "elexon.key", "Elexon API Key (required for all Elexon commands)")
	elexonFlags.StringVar(&bmunit, "bmunit", "", "BMUnit to search for (Elexon or Ofgem)")
	elexonFlags.StringVar(&date, "date",
		time.Now().Add(time.Hour*-24).Format("2006-01-02"),
		"Date to process for (format is YYYY-MM-DD) (defaults to yesterday)")
	elexonFlags.IntVar(&period, "period", -1, "Settlement Period for Elexon (1-50)")

	stdFlags.StringVar(&logFn, "log", "gore.log", "Log filename to write to")
	stdFlags.BoolVar(&verbose, "v", false, "Verbose output (disables logging to a file)")
	stdFlags.StringVar(&name, "name", "", "Name to search for")
	stdFlags.StringVar(&xportFormat, "exportformat", "", "Export format [json, xml, csv]")
	stdFlags.StringVar(&xportFilename, "exportfilename", "", "Filename for exported data")

	if len(os.Args) < 2 {
		fmt.Println("At least a command MUST be supplied.")
		showUsage()
		return
	}

	for arg, possCmd := range availableCommands {
		if os.Args[1] == arg {
			cmd = possCmd
			break
		}
	}

	if cmd.flags == nil {
		if strings.Contains(os.Args[1], "help") {
			showUsage()
			return
		}
		fmt.Printf("\nUnknown command: %s\n", os.Args[1])
		printAvailableCommands()
		return
	}

	// As we can't combine flag sets, we just copy in the flags we need....
	for _, arg := range os.Args[2:] {
		if !strings.HasPrefix(arg, "-") {
			continue
		}
		flag := stdFlags.Lookup(arg[1:])
		if flag != nil {
			cmd.flags.Var(flag.Value, arg[1:], "")
		}
	}

	fmt.Printf("Running %s: %s\n", cmd.name, cmd.description)
	cmd.flags.Parse(os.Args[2:])

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

	var result gore.ResultSet
	switch cmd.reportTag {
	case "stationsearch":
		result, err = doStationSearch(params)
	case "certificatesearch":
		result, err = doCertificateSearch(params)
	default:
		var ap *elexon.ElexonAPI
		ap, err = elexon.NewElexonReport(cmd.reportTag)
		if err != nil {
			result = gore.ResultSet{QueryName: cmd.reportTag}
			break
		}
		fmt.Printf("Getting data for Elexon Report %s [ %s ]...\n", ap.Report.Name, ap.Report.Description)
		err = ap.ReadKeyFile(elexonKeyFn)
		if err == nil {
			if err = ap.GetData(params); err == nil {
				result = ap.Result
			}
		}
	}

	if err != nil {
		fmt.Println(err)
		result = gore.ResultSet{}
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
