# gore
Access UK renewable data using Go

## Background
This is based on my [pywind](https://github.com/zathras777/pywind) library but doesn't yet offer the same functionality.
Presently it is very verbose and some of the exporting isn't ideal.

## Usage

Presently there isn't a lot that is usable :-)

```
gore $ cd cmd/gore
gore/cmd/gore $ go build
gore/cmd/gore $ ./gore
UK Renewables App
=================

No command entered.
Usage of gore.exe:
  -date string
        Date to process for (format is YYYY-MM-DD)
  -elexonkey string
        Elexon API Key (required for all Elexon commands) (default "elexon.key")
  -log string
        Log filename to write to (default "gore.log")
  -month int
        Specify a month (default -1)
  -period string
        Period to use (as string, YYYYMM). Overrides year & month
  -settlementperiod int
        Settlement Period for Elexon (1-50) (default -1)
  -v    Verbose output (disables logging to a file)
  -year int
        Specify a year (default -1)
```

For example,

```
$ gore\cmd\gore> .\gore.exe -elexonkey ..\..\elexon.key -date 2022-02-01 -settlementperiod 30 b1320   
UK Renewables App
=================

Getting data for Elexon Report B1320 [ Congestion Management Measures: Countertrading ]...
B1320 API call succeeded, but no data returned
```

## Issues

- ~~need to add ability to set dropdown values~~
- ~~add elexon api options from [pywind](https://github.com/zathras777/pywind)~~
- improve the export formatting
- ~~add a full command line app~~
- ~~reduce the verbosity and improve the logging~~
- add testing
- add docs on how to use command line app

## Updates

### 17th April 2022
Changed up how the Elexon reports are handled to make it simpler to add any that are needed. I need to work out how to handle the reports that return multiple blocks, e.g. DERBMDATA. Also need to look at how to handle the multi level results that get returned and how to easily access them in a sensible way. Not yet added any decene output to the command line app but hopefully will get to that soon.

### 16th April 2022
I've added the start of a command line app and also a couple of the Elexon API commands. It's a work in progress but things should work and this seems to give me a base to build from in a similar manner to the python code I wrote. The app is basic and needs a lot of code sorted out and moved around :-)

### 14th April 2022
I decided to make the entire form a private structure, so took the opportunity to flatten out the Form & FormData resulting in one single form. This works better and allows better integration between them. The updating of elements now works better and the form generation has been improved. DropDown elements can now have values set, but setting the Scheme to anything but the default results in a page redirect? Country seems to work and I will look at the others, but see no reason why setting the scheme should fail - but it does :-(

## Pull Requests

I'm still new to publishing go packages, so feel free to contact me with suggestions and improvements. :-)

All help welcome.
