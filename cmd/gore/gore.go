package main

import (
	"flag"
	"fmt"
	"gore/pkg/elexon"
	"log"
	"os"
	"strconv"
	"strings"
)

func createTitle(title string) string {
	return title + "\n" + strings.Repeat("=", len(title)) + "\n"
}

func main() {
	fmt.Println(createTitle("UK Renewables App"))

	var (
		elexonKeyFn      string
		logFn            string
		verbose          bool
		year             int
		month            int
		period           string
		date             string
		settlementPeriod int
		ap               *elexon.ElexonAPI
		err              error
	)

	flag.StringVar(&elexonKeyFn, "elexonkey", "elexon.key", "Elexon API Key (required for all Elexon commands)")
	flag.StringVar(&logFn, "log", "gore.log", "Log filename to write to")
	flag.BoolVar(&verbose, "v", false, "Verbose output (disables logging to a file)")
	flag.IntVar(&year, "year", -1, "Specify a year")
	flag.IntVar(&month, "month", -1, "Specify a month")
	flag.StringVar(&date, "date", "", "Date to process for (format is YYYY-MM-DD)")
	flag.IntVar(&settlementPeriod, "settlementperiod", -1, "Settlement Period for Elexon (1-50)")
	flag.StringVar(&period, "period", "", "Period to use (as string, YYYYMM). Overrides year & month")

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

	if period != "" {
		if verbose {
			fmt.Println("Overriding any set year and month as period supplied")
		}
		if len(period) != 6 {
			fmt.Printf("Period MUST be 6 characters in the format YYYYMM, e.g. 202201, not %s\n", period)
			fmt.Printf("Did you mean to use -settlementperiod ?")
			return
		}
		year, err = strconv.Atoi(period[:4])
		if err != nil {
			fmt.Printf("Invalid YYYY for period, %s: %s\n", period[:4], err)
			return
		}
		month, err = strconv.Atoi((period[4:]))
		if err != nil {
			fmt.Printf("Invalid MM for period, %s: %s\n", period[4:], err)
			return
		}
		if month < 1 || month > 12 {
			fmt.Printf("Invalid month. Must be between 1 and 12, not %d\n", month)
			return
		}
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
	if settlementPeriod != -1 {
		if settlementPeriod < 1 || settlementPeriod > 50 {
			fmt.Printf("Settlement Period must be between 1 and 50 - not %d\n", settlementPeriod)
			return
		}
		params["Period"] = fmt.Sprintf("%d", settlementPeriod)
	}
	if date != "" {
		params["SettlementDate"] = date
	}

	if len(flag.Args()) == 0 {
		fmt.Println("No command entered.")
		flag.Usage()
		return
	}

	for _, cmd := range flag.Args() {
		switch strings.ToLower(cmd) {
		case "bm1320":
			ap, err = elexon.BM1320(elexonKeyFn)
		case "bm1420":
			ap, err = elexon.BM1420(elexonKeyFn)
		default:
			fmt.Printf("Unhandled command: %s\n", cmd)
		}
		if err != nil {
			fmt.Printf("Unable to process command %s\nError: %s", cmd, err)
			return
		}
	}

	if ap != nil {
		fmt.Printf("Getting data for Elexon Report %s...\n", ap.Report)
		err = ap.GetData(params)
		if err != nil {
			fmt.Println(err)
			return
		}
		fmt.Printf("API call succeeded. %d items available\n", len(ap.Items))
		if ap.IsCapped() {
			fmt.Printf("API response was capped at %d items.\n", ap.CapLimit())
		}
	}
}
