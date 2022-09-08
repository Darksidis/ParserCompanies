package ParserCompanies

import (
	"fmt"
	"github.com/jackc/pgx/v4/pgxpool"
	"io/fs"
	"io/ioutil"
	"log"
	"os"
	"regexp"
	"strings"
)


func main () {

	// ЗАДАЧА - РАЗДЕЛИТЬ РАСПАКОВКУ АРХИВОВ И ЧТЕНИЕ ФАЙЛОВ НА ОТДЕЛЬНЫЕ ПОТОКИ
	files, err := ioutil.ReadDir("files")
	checkError(err)

	// Если архивы присутствуют, начинаем их распаковку
	if !(len(files) == 0) {

		cursor := ConnectingToTheBase()
		CreateBaseOrDoNothing(cursor)
		AddDataCSV(cursor)

		go workWithGorotines(cursor)

	}



	// Если что-то происходит, курсор закрывается
	//defer cursor.Close()
}

func checkError(err error) {
	// Проверка на ошибки
	if err != nil {
		log.Fatal(err)
	}
}




func workWithGorotines (cursor *pgxpool.Pool) {
	// Работаем с горутинами, создаем два канала
	// (один работает с открытием и чтением файлов, а другой с его анализом)
	// В конце очищаем папку

	fileChan := make(chan fs.FileInfo)
	lineChan := make(chan string)
	files, err := ioutil.ReadDir("files/")
	checkError(err)

	files, err = ioutil.ReadDir("files/")
	checkError(err)

	for i := 0; i < 1; i++ {
		fmt.Println (fileChan)
		go sendDataToChannel(fileChan, lineChan)
	}

	// Здесь указывается, сколько необходимо читать спарсенных строк одновременно
	countLines  := 100

	for i := 0; i < countLines; i++ {
		go parsingInfoUsingRegexp(lineChan, cursor)
	}

	//sendDataToChannel(fileChan, lineChan)
	// Итерируем файлы и ложим их в канал
	for _, f := range files {
		fileChan <- f
	}
	// Из-за того, что при выполнений горутин, как правило, остаётся один файл
	// его приходится удалять вручную, по факту это костыль

	// закрываем каналы
	//defer close(fileChan)
	//defer close(lineChan)
	//defer cleaningFiles()
	defer fmt.Println("абобус")



}

func sendDataToChannel(fileChan chan fs.FileInfo, lineChan chan string) {
	// Получаем список распакованных данных, парсим их с помощью регулярок,
	// и отправляем в базу

	//date_slice := strings.Split(date, ".")
	for {
		select {
		case _, ok := <- lineChan:
			if !(ok) {
				break
			}
		case file, ok := <-fileChan:
			// Если канал закрыт, то прерываем цикл
			if !(ok) {
				break
			}
			// Иногда возникают ситуаций, когда файла на момент вызова функций уже не существует
			files, err := ioutil.ReadDir("files/")
			if err != nil {
				log.Fatal ("read files failed: ", err)
			}
			statusExist := fileExists(files, file.Name())
			if !(statusExist) {
				fmt.Println ("файл не существует")
				continue
			}

			data := OpenAndReadFile(file, lineChan)
			if len(data) == 0 {
				continue
			}
			lineChan <- string(data)
		}
	}
}


func OpenAndReadFile(file fs.FileInfo) []byte{
	dirname := "files/" //+ year + "/"
	statusFile := strings.Contains(file.Name(), ".corr")

	//Удаляем файл
	defer func(name string) {
		err := os.Remove(name)
		if err != nil {
			log.Fatal ("delete file failed: ", err)

		}
	}(dirname + file.Name())

	if statusFile {
		return []byte{}
	}

	dat, err := os.ReadFile(dirname + file.Name())

	// Мы парсим просто огромное количество файлов,
	// все случаи ошибок предусмотреть невозможно
	// гораздо проще просто пропустить, удалить файл и идти дальше
	if err != nil {
		fmt.Println ("read file failed: ", err)
		return []byte{}
	}

	return dat

}

func fileExists(files []fs.FileInfo, file string) bool {

	for _, fil := range files {
		if fil.Name() == file {
			return true
		}
	}

	return false
}

func parsingInfoUsingRegexp(lineChan chan string, cursor *pgxpool.Pool) (map[string]string, map[string][]string) {
	// Здесь мы анализируем полученную информацию с файла с помощью регулярных выражений
	// (Базовая информация, Бизнес-Адрес и Майл-Адрес)
	for {
		select {
		case data, ok := <-lineChan:
			// Если канал закрыт, то прерываем цикл
			if !(ok) {
				break
			}
			year := "1999"

			data = strings.Replace(data,"\n", "", -1)

			// Находим имя компаний, её идентификатор и sic с помощью регулярного выражения
			basicInfoReg, err :=  regexp.Compile("CONFORMED-NAME>(.*?)" + "<CIK>(.*?)" + "<ASSIGNED-SIC>(.*?)<")
			checkError (err)
			basicInformation := basicInfoReg.FindAllString(data, -1)

			// Находим бизнес-адрес с помощью регулярного выражения
			businessReg, err := regexp.Compile("<BUSINESS-ADDRESS>(.*?)" + "</BUSINESS-ADDRESS>")
			checkError (err)
			businessAddress := businessReg.FindAllString(data, -1)

			// Находим домашний адрес с помощью регулярного выражения
			mailReg, err := regexp.Compile("<MAIL-ADDRESS>(.*?)" + "</MAIL-ADDRESS>")
			checkError (err)
			mailAddress := mailReg.FindAllString(data, -1)

			routes := CreatingArraysData(basicInformation, businessAddress, mailAddress, cursor)

			tags, tagsAddress := addingDataToMap(routes)

			//Если список с тегами пустой, пропускаем эту итерацию
			if len(tags) == 0 {
				continue
			}
			companyScope := GetCompanyScope(cursor, tags["<ASSIGNED-SIC>"])

			AddingDataToDb(year, tags, tagsAddress, companyScope, cursor)
		}


	}
}


func CreatingArraysData (basicInformation []string, businessAddress []string, mailAddress []string, cursor *pgxpool.Pool) [][]string{
	//формируем три списка с тегами, и добавляем их в массив массивов для дальнейшего объединения

	// Если основные данные пусты, то пропускаем итерацию
	if len(basicInformation) == 0 {
		var routes [][]string
		return routes
	}
	// Формируем список на основе открывающего тега
	fmt.Println(basicInformation)
	basicInformation = strings.Split(basicInformation[0], "<")

	status := CheckExistCik (basicInformation, cursor)
	// Если данные уже есть в базе (в датасете много дубликатов), возвращаем пустой массив массивов
	if status {
		var routes [][]string
		return routes
	}

	// Из-за специфики парсинга нам необходимо будет очистить список от пустого элемента (с помощью среза)
	basicInformation = append(basicInformation[:len(basicInformation)-1], basicInformation[len(basicInformation):]...)

	// Для простоты объединения слайсов предварительно объединяем их в массив массивов.
	routes := [][]string{basicInformation}

	// Если бизнес-адрес существует, превращаем его в строковый массив
	if !(len(businessAddress) == 0) {
		businessAddress = strings.Split(businessAddress[0], "<")
		routes = append(routes, businessAddress)
	}
	// Если бизнес-адрес существует, превращаем его в строковый массив
	if !(len(mailAddress) == 0) {
		mailAddress = strings.Split(mailAddress[0], "<")
		routes = append(routes, mailAddress)
	}

	return routes

}
func addingDataToMap(routes [][]string) (map[string]string, map[string][]string){
	// Здесь мы добавляем данные в map-словари

	// Если массив массивов пустой, то или элемент уже есть в базе, или что-то произошло,
	// возвращаем пустые map-словари
	if len(routes) == 0 {
		tags :=  map[string]string{}
		tagsAddress := map[string][]string{}
		return tags, tagsAddress
	}
	var desiredList []string
	// Объединяем слайсы
	for _, route := range routes {
		desiredList = append(desiredList, route...)
	}

	// Map-словари. В них будет добавляться вся полученная информация
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

	// Число-индекс, 0 равен бизнес-адресу, 1 домашнему
	busOrMail := 0
	for _, el := range desiredList {
		// Очень странный баг, при объединений нескольких слайсов появляются пустые элементы
		if len(el) == 0 {
			continue
		}
		// Находим, какой индекс у закрывающего тега
		indexCloseTag := strings.Index(el, ">")
		// И формируем новый слайс из среза, где теги проставлены корректно
		tag := "<" + el[0:indexCloseTag] + ">"
		el = el[indexCloseTag+1:]

		//fmt.Println("el: ", el)
		if tag == "<MAIL-ADDRESS>" {
			busOrMail = 1
		}
		// Если тег есть в словаре, добавляем его
		if _, ok :=  tags[tag]; ok {
			tags[tag] = el
		}
		if _, ok := tagsAddress[tag]; ok {
			tagsAddress[tag][busOrMail] = el
		}
	}

	return tags, tagsAddress

}


