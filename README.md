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

## Issues

- ~~need to add ability to set dropdown values~~
- ~~add elexon api options from [pywind](https://github.com/zathras777/pywind)~~
- improve the export formatting
- ~~add a full command line app~~
- reduce the verbosity and improve the logging

## Updates

### 16th April 2022
I've added the start of a command line app and also a couple of the Elexon API commands. It's a work in progress but things should work and this seems to give me a base to build from in a similar manner to the python code I wrote. The app is basic and needs a lot of code sorted out and moved around :-)

### 14th April 2022
I decided to make the entire form a private structure, so took the opportunity to flatten out the Form & FormData resulting in one single form. This works better and allows better integration between them. The updating of elements now works better and the form generation has been improved. DropDown elements can now have values set, but setting the Scheme to anything but the default results in a page redirect? Country seems to work and I will look at the others, but see no reason why setting the scheme should fail - but it does :-(

## Pull Requests

I'm still new to publishing go packages, so feel free to contact me with suggestions and improvements. :-)

All help welcome.
