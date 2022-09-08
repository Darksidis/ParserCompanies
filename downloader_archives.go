package ParserCompanies

import (
	"fmt"
	"github.com/anaskhan96/soup"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"os"
)
func sendingRequest (url string) (*http.Response, error) {

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
	// Получаем содержимое страницы, и читаем

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {

		return "", err
	}

	defer resp.Body.Close()

	return string(body), nil


}

func downloadFile(filepath string, resp *http.Response) (err error) {
	// В этой функции мы скачиваем файл из ссылки

	// Создаем файл
	out, err := os.Create(filepath)
	if err != nil  {
		return err
	}
	defer out.Close()

	// Получаем ответ сервера
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("bad status: %s", resp.Status)
	}

	// Записываем файл
	_, err = io.Copy(out, resp.Body)
	if err != nil  {
		return err
	}

	return nil
}

func findElements (doc soup.Root) []soup.Root {
	// Здесь мы парсим разметку, и находим необходимые элементы

	elements := doc.Find("table", "summary", "heding").FindAll("a")

	return elements
}

func main() {
	url := "https://www.sec.gov/Archives/edgar/Feed/"

	response, err := sendingRequest(url)
	respBody, err := getHtmlPage(response)
	if err != nil {
		log.Fatal("parsing years failed: ", err)
	}
	doc := soup.HTMLParse(respBody)
	fmt.Println("doc: ", doc)

	years := findElements(doc)

	for _, year := range years {

		year := year.Attrs()["href"]
		url = "https://www.sec.gov/Archives/edgar/Feed/" + year
		response, err = sendingRequest(url)
		respBody, err = getHtmlPage(response)
		if err != nil {
			log.Fatal("parsing semesters failed: ", err)
		}
		doc := soup.HTMLParse(respBody)
		fmt.Println("doc: ", doc)
		semesters := findElements(doc)

		for _, sem := range semesters {

			semester := sem.Attrs()["href"]
			url = "https://www.sec.gov/Archives/edgar/Feed/" + year + semester
			response, err = sendingRequest(url)
			respBody, err = getHtmlPage(response)
			if err != nil {
				log.Fatal("parsing archives failed: ", err)
			}
			doc = soup.HTMLParse(respBody)
			archives := findElements(doc)

			for _, arch := range archives {

				archive := arch.Attrs()["href"]
				url = "https://www.sec.gov/Archives/edgar/Feed/" + year + semester + archive
				response, err = sendingRequest(url)
				err = downloadFile("files_companies/" + archive, response)

				if err != nil {
					log.Fatal("download archive failed: ", err)
				}

			}

		}
	}
}
