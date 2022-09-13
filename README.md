# ParserCompanies

## About
-------

Parser that generates a dataset containing information about US companies.

## What did I use
-------

Golang, PostgreSQL

## How it works
--------

The parser downloads the archive from the site "sec.gov", unpacks it, and sends the data to the database using regular expressions

## How to run it
--------

### Through "go run"
--------
#### Download archives
```
go run downloader_archives.go base_settings.go
```
#### Unpacking archives
```
go run unpacking_archives.go
```

#### Sending Data
```
go run sending_data.go base_settings.go
```

### Through binaries
--------
```
cd bin
```
#### > Windows needs to run exe files

#### Download archives
```
./downloader_archives
```
#### Unpacking archives
```
./unpacking_archives.
```

#### Sending Data
```
./sending_data
```
