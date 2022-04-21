package main

type command struct {
	name        string
	description string
	formatter   formatterRow
}

var availableCommands = map[string]command{
	"b1320": {
		"Elexon: B1320",
		"Congestion Management Measures: Countertrading",
		formatterRow{
			[]formatterColumn{},
		},
	},
	"b1330": {
		"Elexon: B1330",
		"Congestion Management Measures: Costs of Congestion Management",
		formatterRow{
			[]formatterColumn{},
		},
	},
	"b1420": {
		"Elexon: B1420",
		"Installed Generation Capacity per Unit",
		formatterRow{
			[]formatterColumn{
				{"Unit ID", "registeredResourceName", "string", 20, 0},
				{"Active?", "activeFlag", "bool", 10, 0},
				{"Resource Type", "powerSystemResourceType", "string", 35, 0},
				{"Voltage Limit", "voltageLimit", "int", 13, 0},
				{"Capacity", "capacity", "int", 12, 2},
			},
		},
	},
	"b1610": {
		"Elexon: B1610",
		"Actual Generation Output per Generation Unit",
		formatterRow{
			[]formatterColumn{
				{"Unit", "nGCBMUnitID", "string", 15, 0},
				{"Date", "settlementDate", "date", 12, 0},
				{"Period", "settlementPeriod", "int", 6, 0},
				{"Output", "output", "float", 12, 2},
			},
		},
	},
	"b1630": {
		"Elexon: B1630",
		"Actual Or Estimated Wind and Solar Power Generation",
		formatterRow{
			[]formatterColumn{},
		},
	},
	"derbmdata": {
		"Elexon: DERBMDATA",
		"Derived BM Unit Data - Multiple Result Sets",
		formatterRow{
			[]formatterColumn{},
		},
	},
	"dersysdata": {
		"Elexon: DERSYSDATA",
		"Derived System Wide Data",
		formatterRow{
			[]formatterColumn{
				{"Date", "settlementDate", "date", 10, 0},
				{"Period", "settlementPeriod", "int", 6, 0},
				{"Sell Price", "systemSellPrice", "float", 10, 4},
				{"Buy Price", "systemBuyPrice", "float", 10, 4},
				{"Offer Volume", "totalSystemAcceptedOfferVolume", "float", 12, 4},
				{"Bid Volume", "totalSystemAcceptedBidVolume", "float", 12, 4},
				{"Net Imbalance Volume", "indicativeNetImbalanceVolume", "float", 20, 4},
			},
		},
	},
	"fuelinst": {
		"Elexon: FUELINST",
		"Generation by Fuel Type (24H Instant Data)",
		formatterRow{
			[]formatterColumn{
				{"Date", "publishingPeriodCommencingTime", "date", 10, 0},
				{"Time", "publishingPeriodCommencingTime", "time", 5, 0},
				{"Period", "settlementPeriod", "int", 6, 0},
				{"Biomass", "biomass", "int", 8, 0},
				{"CCGT", "ccgt", "int", 8, 0},
				{"Oil", "oil", "int", 8, 0},
				{"Coal", "coal", "int", 8, 0},
				{"Nuclear", "nuclear", "int", 8, 0},
				{"Wind", "wind", "int", 8, 0},
				{"PS", "ps", "int", 8, 0},
				{"NPSHYD", "npshyd", "int", 8, 0},
				{"OCGT", "ocgt", "int", 8, 0},
				{"Other", "other", "int", 8, 0},
				{"IntFR", "intfr", "int", 8, 0},
				{"IntIRL", "intirl", "int", 8, 0},
				{"IntNED", "intned", "int", 8, 0},
				{"IntEW", "intew", "int", 8, 0},
				{"IntNEM", "intnem", "int", 8, 0},
				{"IntElec", "intelec", "int", 8, 0},
				{"IntIFA", "intifa2", "int", 8, 0},
				{"IntNSL", "intnsl", "int", 8, 0},
			},
		},
	},
	"certificatesearch": {
		"Ofgem Certificate Search",
		"Ofgem: Search certificate database",
		formatterRow{
			[]formatterColumn{
				{"Accreditation", "AccreditationNumber", "string", 13, 0},
				{"Station Name", "Station", "string", 30, 0},
				{"Capacity", "StationTIC", "float", 12, 2},
				{"Scheme", "Scheme", "string", 6, 0},
				{"Technology", "TechnologyGroup", "string", 20, 0},
				{"Period", "OutputPeriod", "string", 20, 0},
				{"No. Certs", "NoOfCertificates", "int", 8, 0},
				{"MWh Output", "MWh", "float", 10, 1},
			},
		},
	},
	"stationsearch": {
		"Ofgem Station Search",
		"Ofgem: Search the station database",
		formatterRow{
			[]formatterColumn{
				{"Accreditation", "AccreditationNumber", "string", 13, 0},
				{"Station Name", "Station", "string", 30, 0},
				{"Capacity", "Capacity", "float", 10, 2},
				{"Scheme", "Scheme", "string", 6, 0},
				{"Technology", "TechnologyGroup", "string", 20, 0},
			},
		},
	},
}
