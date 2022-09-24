package main

import (
	"github.com/jackc/pgx/v4/pgxpool"
	"io/fs"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)


func main () {

	path := PathForm("storage", "files")
	files, err := ioutil.ReadDir(path)
	checkError(err)

	// Если архивы присутствуют, начинаем их распаковку
	if !(len(files) == 0) {
		cursor := ConnectingToTheBase()
		CreateBaseOrDoNothing(cursor)
		AddDataCSV(cursor)
		workWithGorotines(cursor, path)

	}

}

func checkError(err error) {
	// Проверка на ошибки
	if err != nil {
		log.Fatal(err)
	}
}


func workWithGorotines (cursor *pgxpool.Pool, path string) {
	// Working with goroutines, creating two channels
	//(one works with opening and reading files and the other with parsing it)

	fileChan := make(chan fs.FileInfo)
	lineChan := make(chan string)
	files, err := ioutil.ReadDir(path)
	checkError(err)

	files, err = ioutil.ReadDir(path)
	checkError(err)

	for i := 0; i < 1; i++ {
		go sendDataToChannel(fileChan, lineChan, path)
	}

	// This specifies how many parsed lines should be read at the same time.
	countLines  := 100

	for i := 0; i < countLines; i++ {
		go parsingInfoUsingRegexp(lineChan, cursor)
	}

	// We sort through the files and put them in the channel
	for _, f := range files {
		fileChan <- f
	}



}

func sendDataToChannel(fileChan chan fs.FileInfo, lineChan chan string, path string) {
	// We get a list of unpacked data, parse them using regular expressions,
	// and send them to the database

	for {
		select {
		case _, ok := <- lineChan:
			if !(ok) {
				break
			}
		case file, ok := <-fileChan:
			// If the channel is closed, then we interrupt the loop
			if !(ok) {
				break
			}

			data := OpenAndReadFile(file, path)
			if len(data) == 0 {
				continue
			}
			lineChan <- string(data)
		}
	}
}


func OpenAndReadFile(file fs.FileInfo, dirname string) []byte{
	statusFile := strings.Contains(file.Name(), ".corr")

	pathFile := filepath.Join(dirname, file.Name())

	//Delete file
	defer func(name string) {
		err := os.Remove(name)
		if err != nil {
			log.Fatal ("delete file failed: ", err)

		}
	}(pathFile)

	if statusFile {
		return []byte{}
	}

	dat, err := os.ReadFile(pathFile)

	// If some error occurs, then skip and delete the file
	if err != nil {
		// open file failed
		return []byte{}
	}

	return dat

}

func parsingInfoUsingRegexp(lineChan chan string, cursor *pgxpool.Pool) (map[string]string, map[string][]string) {
	// Here we parse the received information from the file using regular expressions
	// (Basic Information, Business Address and Mail Address)
	for {
		select {
		case data, ok := <-lineChan:
			// If the channel is closed, then we interrupt the loop
			if !(ok) {
				break
			}

			data = strings.Replace(data,"\n", "", -1)

			// We find the name of the companies, its CIK and SIC using a regular expression
			basicInfoReg, err :=  regexp.Compile("CONFORMED-NAME>(.*?)" + "<CIK>(.*?)" + "<ASSIGNED-SIC>(.*?)<")
			checkError (err)
			basicInformation := basicInfoReg.FindAllString(data, -1)

			// Finding a business address using a regular expression
			businessReg, err := regexp.Compile("<BUSINESS-ADDRESS>(.*?)" + "</BUSINESS-ADDRESS>")
			checkError (err)
			businessAddress := businessReg.FindAllString(data, -1)

			// Finding a home address with a regular expression
			mailReg, err := regexp.Compile("<MAIL-ADDRESS>(.*?)" + "</MAIL-ADDRESS>")
			checkError (err)
			mailAddress := mailReg.FindAllString(data, -1)

			routes := CreatingArraysData(basicInformation, businessAddress, mailAddress, cursor)

			tags, tagsAddress := addingDataToMap(routes)

			// If the tagged list is empty, skip this iteration
			if len(tags) == 0 {
				continue
			}
			companyScope := GetCompanyScope(cursor, tags["<ASSIGNED-SIC>"])

			AddingDataToDb(tags, tagsAddress, companyScope, cursor)
		}


	}
}


func CreatingArraysData (basicInformation []string, businessAddress []string, mailAddress []string, cursor *pgxpool.Pool) [][]string{
	// We form three lists with tags, and add them to the array of arrays for further merging

	// If the main data is empty, then skip the iteration
	if len(basicInformation) == 0 {
		var routes [][]string
		return routes
	}
	// We form a list based on the opening tag
	basicInformation = strings.Split(basicInformation[0], "<")

	status := CheckExistCik(basicInformation, cursor)
	// If the data is already in the database (there are many duplicates in the dataset), we return an empty array of arrays
	if status {
		var routes [][]string
		return routes
	}

	// Due to the specifics of parsing, we will need to clear the list from an empty element (using a slice)
	basicInformation = append(basicInformation[:len(basicInformation)-1], basicInformation[len(basicInformation):]...)

	// For ease of combining slices, we first combine them into an array of arrays.
	routes := [][]string{basicInformation}

	// If the business address exists, turn it into a string array
	if !(len(businessAddress) == 0) {
		businessAddress = strings.Split(businessAddress[0], "<")
		routes = append(routes, businessAddress)
	}
	// If the mail address exists, turn it into a string array
	if !(len(mailAddress) == 0) {
		mailAddress = strings.Split(mailAddress[0], "<")
		routes = append(routes, mailAddress)
	}

	return routes

}
func addingDataToMap(routes [][]string) (map[string]string, map[string][]string){
	// Here we add data to the map

	// If the array of arrays is empty, then either the element is already in the database, or something has happened,
	// return empty maps
	if len(routes) == 0 {
		tags :=  map[string]string{}
		tagsAddress := map[string][]string{}
		return tags, tagsAddress
	}
	var desiredList []string
	// Combining slices
	for _, route := range routes {
		desiredList = append(desiredList, route...)
	}

	// Maps. All information received will be added to them.
	tags := map[string]string{"<CONFORMED-NAME>":"",
		"<CIK>":"",
		"<ASSIGNED-SIC>":"",
		"<PHONE>": "",
	}
	tagsAddress := map[string][]string{"<STREET1>": {"", ""},
		"<CITY>":{"", ""},
		"<STATE>":{"", ""},
		"<ZIP>":{"", ""},
		"<STREET2>":{"", ""}}

	// Index number, 0 equals business address, 1 equals home
	busOrMail := 0
	for _, el := range desiredList {
		// A very strange bug, when combining several slices, empty elements appear
		if len(el) == 0 {
			continue
		}
		// Finding the index of the closing tag
		indexCloseTag := strings.Index(el, ">")
		// And we form a new slice from the slice, where the tags are affixed correctly
		tag := "<" + el[0:indexCloseTag] + ">"
		el = el[indexCloseTag+1:]

		if tag == "<MAIL-ADDRESS>" {
			busOrMail = 1
		}
		// If the tag is in the map, add it
		if _, ok :=  tags[tag]; ok {
			tags[tag] = el
		}
		if _, ok := tagsAddress[tag]; ok {
			tagsAddress[tag][busOrMail] = el
		}
	}

	return tags, tagsAddress

}


