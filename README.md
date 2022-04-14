# gore
Access UK renewable data using Go

## Background
This is based on my [pywind](https://github.com/zathras777/pywind) library but doesn't yet offer the same functionality.
Presently it is very verbose and some of the exporting isn't ideal.

## Usage

Presently there isn't much to see :-)

```
$ go build
$ ./gore
```

## Issues

- ~~need to add ability to set dropdown values~~
- add elexon api options from [pywind](https://github.com/zathras777/pywind)
- improve the export formatting
- add a full command line app
- reduce the verbosity and improve the logging

## Updates

### 14th April 2022
I decided to make the entire form a private structure, so took the opportunity to flatten out the Form & FormData resulting in one single form. This works better and allows better integration between them. The updating of elements now works better and the form generation has been improved. DropDown elements can now have values set, but setting the Scheme to anything but the default results in a page redirect? Country seems to work and I will look at the others, but see no reason why setting the scheme should fail - but it does :-(

## Pull Requests

I'm still new to publishing go packages, so feel free to contact me with suggestions and improvements. :-)

All help welcome.
