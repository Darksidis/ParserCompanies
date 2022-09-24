# ParserCompanies

### About
-------

Parser that generates a dataset containing information about US companies.

### What did I use
-------

Golang, PostgreSQL

### How it works
--------

The parser downloads the archive from the site "sec.gov", unpacks it, and sends the data to the database using regular expressions

### How to run it
--------

You can run through the file manager
```
go run main.go
```

In the list of actions *actionsList.env*, you can also specify which files will be launched (true - the file will be launched, false - the file will be skipped, observe case)
```
{"sending_data.go" : true, "unpacking_archives.go" : true, "downloader_archives.go" : true}
```

You also need to specify in the environment variables *db.env* the data of your database
```
USERNAME_DB=
PASSWORD=
HOST=
PORT=
DB_NAME=
```
