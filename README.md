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

You can run through the file manager
```
go run main.go
```

In the list of actions, you can also specify which files will be launched (true - the file will be launched, false - the file will be skipped, observe case)
