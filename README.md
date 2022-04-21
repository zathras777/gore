# gore
Access UK renewable data using Go

## Background
This is based on my [pywind](https://github.com/zathras777/pywind) library but doesn't yet offer the same functionality.
Presently it is very verbose and some of the exporting isn't ideal.

## Usage

Presently there are a few things working :-)

```
gore $ cd cmd/gore
gore/cmd/gore $ go build
gore/cmd/gore $ ./gore.exe -help
gore.exe [parameters] [flags] [command]


Available Commands
==================

               b1320 - Elexon: B1320
                       Congestion Management Measures: Countertrading
               b1330 - Elexon: B1330
                       Congestion Management Measures: Costs of Congestion Management
               b1420 - Elexon: B1420
                       Installed Generation Capacity per Unit
               b1610 - Elexon: B1610
                       Actual Generation Output per Generation Unit
               b1630 - Elexon: B1630
                       Actual Or Estimated Wind and Solar Power Generation
   certificatesearch - Ofgem Certificate Search
                       Ofgem: Search certificate database
           derbmdata - Elexon: DERBMDATA
                       Derived BM Unit Data - Multiple Result Sets
          dersysdata - Elexon: DERSYSDATA
                       Derived System Wide Data
            fuelinst - Elexon: FUELINST
                       Generation by Fuel Type (24H Instant Data)
       stationsearch - Ofgem Station Search
                       Ofgem: Search the station database


Parameters and Flag Options
===========================

-bmunit string
        BMUnit to search for (Elexon or Ofgem)
  -date string
        Date to process for (format is YYYY-MM-DD)
  -elexonkey string
        Elexon API Key (required for all Elexon commands) (default "elexon.key")
  -exportfilename string
        Filename for exported data
  -exportformat string
        Export format [json, xml, csv]
  -log string
        Log filename to write to (default "gore.log")
  -month int
        Specify a month (default -1)
  -name string
        Name to search for (Elexon or Ofgem)
  -period int
        Settlement Period for Elexon (1-50) (default -1)
  -scheme string
        Ofgem Scheme (RO, REGO)
  -v    Verbose output (disables logging to a file)
  -year int
        Specify a year (default -1)
```

Some of it even works :-) For example,

```
$ gore\cmd\gore> .\gore.exe -elexonkey ..\..\elexon.key -date 2022-02-01 -settlementperiod 30 b1320   
UK Renewables App
=================

Getting data for Elexon Report B1320 [ Congestion Management Measures: Countertrading ]...
B1320 API call succeeded, but no data returned
```

```
$ gore\cmd\gore> ./gore -elexonkey ..\..\elexon.key -exportformat json -exportfilename "test.json" fuelinst

UK Renewables App
=================


Getting data for Elexon Report FUELINST [ Generation by Fuel Type (24H Instant Data) ]...
Query succeeded. 283 items returned
Date       Time  Period  Biomass     CCGT      Oil     Coal  Nuclear     Wind       PS   NPSHYD     OCGT    Other    IntFR   IntIRL   IntNED    IntEW   IntNEM  IntElec   IntIFA   IntNSL 
========== ===== ====== ======== ======== ======== ======== ======== ======== ======== ======== ======== ======== ======== ======== ======== ======== ======== ======== ======== ========
2022-04-18 21:45     46     2165    15280        0        0     4955     3458        0      229        4      263        0        0        0        0        0        0        0        0
2022-04-18 21:50     46     2164    15106        0        0     4921     3428        0      229        4      266        0        0        0        0        0        0        0        0
2022-04-18 21:55     46     2163    15067        0        0     4886     3405        0      229        4      249        0        0        0        0        0        0        0        0
2022-04-18 22:00     46     2163    15055        0        0     4860     3375        0      218        5      244        0        0        0        0        0        0        0        0
...

 Exporting data to test.json as json
Export completed
```

## Issues

- ~~need to add ability to set dropdown values~~
- ~~add elexon api options from [pywind](https://github.com/zathras777/pywind)~~
- ~~improve the export formatting~~
- ~~add a full command line app~~
- ~~reduce the verbosity and improve the logging~~
- add testing
- add docs on how to use command line app
- add csv export


## Updates

### 21st April 2022
Found some issues with displaying Ofgem Station searches and so have adjusted the output to correct them. This led to realisation the date wasn't correct for the Certificate Search so that's also been updated :-) Along the way I rationalised the flags to remove the period as the year and month provide a better intreface. This allows settlementeriod to become simply period, which should make things simpler to use. Added the DERSYSDATA report from Elexon. Version 0.1.3-alpha pushed to capture these changes.

### 19th April 2022
Tidied up the cmd app and actually started making it useful. Exporting to XML and JSON should now work, though not for the multiple return reports (presently only DERBMDATA) as I'm still trying to get my head around how to do that in a way that makes sense. There is a lot of potential for tidying up yet, but as Justin used to remind everyone, "premature optimisation is evil" so things are just as they are. The data that is returned seems reasonable and should be simple enough to understand. It's faster than pywind for all operations and while it doesn't quite have the same functionality it's not far from it. 
Could it be time to push an alpha release :-)

### 17th April 2022
Changed up how the Elexon reports are handled to make it simpler to add any that are needed. I need to work out how to handle the reports that return multiple blocks, e.g. DERBMDATA. Also need to look at how to handle the multi level results that get returned and how to easily access them in a sensible way. Not yet added any decene output to the command line app but hopefully will get to that soon.

### 16th April 2022
I've added the start of a command line app and also a couple of the Elexon API commands. It's a work in progress but things should work and this seems to give me a base to build from in a similar manner to the python code I wrote. The app is basic and needs a lot of code sorted out and moved around :-)

### 14th April 2022
I decided to make the entire form a private structure, so took the opportunity to flatten out the Form & FormData resulting in one single form. This works better and allows better integration between them. The updating of elements now works better and the form generation has been improved. DropDown elements can now have values set, but setting the Scheme to anything but the default results in a page redirect? Country seems to work and I will look at the others, but see no reason why setting the scheme should fail - but it does :-(

## Pull Requests

I'm still new to publishing go packages, so feel free to contact me with suggestions and improvements. :-)

All help welcome.
