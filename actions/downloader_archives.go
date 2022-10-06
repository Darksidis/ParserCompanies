package main

import (
	"encoding/json"
	"fmt"
	"github.com/anaskhan96/soup"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"path/filepath"
)

func main() {
	fmt.Println("download archives start")
	cursor := ConnectingToTheBase()
	CreateLogs(cursor)

	logs := make(map[string]string)

	url := "https://www.sec.gov/Archives/edgar/Feed/"
	response, err := sendingRequest(url)
	respBody, err := getHtmlPage(response)
	if err != nil {
		log.Fatal("parsing years failed: ", err)
	}
	doc := soup.HTMLParse(respBody)

	years := findElements(doc)

	dat := GetLogs(cursor)

	stopTrigger := [3]int{1, 1, 1}
	// The condition handles the case if logs are present
	if !(len(dat) == 0) {
		err = json.Unmarshal([]byte(dat), &logs)
		if err != nil {
			log.Fatal("json processing failed: ", err)
		}

		for index, _ := range stopTrigger {
			stopTrigger[index] = 0
		}

	}

	for _, year := range years {

		year := year.Attrs()["href"]

		if !(year == logs["year"]) && stopTrigger[0] == 0 {
			continue
		}
		stopTrigger[0] = 1

		url = "https://www.sec.gov/Archives/edgar/Feed/" + year
		response, err = sendingRequest(url)
		respBody, err = getHtmlPage(response)
		if err != nil {
			log.Fatal("parsing semesters failed: ", err)
		}
		doc := soup.HTMLParse(respBody)
		semesters := findElements(doc)

		logs["year"] = year

		for _, sem := range semesters {

			semester := sem.Attrs()["href"]

			if !(semester == logs["semester"]) && stopTrigger[1] == 0 {
				continue
			}

			stopTrigger[1] = 1

			url = "https://www.sec.gov/Archives/edgar/Feed/" + year + semester
			response, err = sendingRequest(url)
			respBody, err = getHtmlPage(response)
			if err != nil {
				log.Fatal("parsing archives failed: ", err)
			}
			doc = soup.HTMLParse(respBody)
			archives := findElements(doc)

			logs["semester"] = semester

			for _, arch := range archives {

				archive := arch.Attrs()["href"]

				if !(archive == logs["archive"]) && stopTrigger[2] == 0 {
					continue
				}

				stopTrigger[2] = 1

				typeFile := archive[len(archive)-6:]
				if !(typeFile == "tar.gz") {
					// format incorrect
					continue
				}

				url = "https://www.sec.gov/Archives/edgar/Feed/" + year + semester + archive
				response, err = sendingRequest(url)

				path := PathForm("storage", "archives")
				pathArchive := filepath.Join(path, archive)
				err = downloadFile(pathArchive, response)

				if err != nil {
					log.Fatal("download archive failed: ", err)
				}

				logs["archive"] = archive

				jsonString, err := json.Marshal(logs)
				if err != nil {
					log.Fatal("map to json process failed: ", err)
				}

				SendLogs(cursor, jsonString)
			}

		}
	}
}

func sendingRequest(url string) (*http.Response, error) {

	client := &http.Client{}

	req, err := http.NewRequest("GET", url, nil)
	req.Header.Add("User-Agent", "Mozilla/5.0 (Windows NT 6.3; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/104.0.0.0 Safari/537.36")

	resp, err := client.Do(req)
	if err != nil {
		return resp, err
	}
	return resp, nil
}

func getHtmlPage(resp *http.Response) (string, error) {
	// Get page content and read

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {

		return "", err
	}

	defer resp.Body.Close()

	return string(body), nil

}

func downloadFile(filepath string, resp *http.Response) (err error) {
	// In this function, we are downloading a file from a link

	// Create file
	out, err := os.Create(filepath)
	if err != nil {
		return err
	}
	defer out.Close()

	// Getting a server response
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("bad status: %s", resp.Status)
	}

	// Writing a file
	_, err = io.Copy(out, resp.Body)
	if err != nil {
		return err
	}

	return nil
}

func findElements(doc soup.Root) []soup.Root {
	// We parse the markup and find the necessary elements

	elements := doc.Find("table", "summary", "heding").FindAll("a")

	return elements
}
