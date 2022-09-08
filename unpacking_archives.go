package ParserCompanies

import (
	"fmt"
	"github.com/mholt/archiver/v3"
	"io/fs"
	"io/ioutil"
	"log"
	"os"
)


func main () {

	dirname := "files_companies"
	// Получаем список архивов
	archives, err := ioutil.ReadDir(dirname)
	checkError(err)
	if !(len(archives) == 0) {
		extractionArchives(archives)
	}
}

func checkError(err error) {
	// Проверка на ошибки
	if err != nil {
		log.Fatal(err)
	}
}

func extractionArchives(archives []fs.FileInfo) {
	// Итерируем список файлов и раскладываем по го рутинам

	for _, archive := range archives {
		// IsDir() это проверка на каталог
		unpackingFiles(archive.Name())

	}


}

func unpackingFiles(archive string) {
	// Извлекаем файлы из архивов

	// Чистим директорию во избежание ошибок
	//cleaningFiles()

	// Основная проблема тут заключается в том, что, судя по всему, в том разделе должен вызываться log.Fatal или другое прерывание программы, а вместо этого возвращается простой return,
	// Она работает из-за этого некорректно

	//Удаляем архив
	defer func(name string) {
		fmt.Println ("он файл удалил")
		if r := recover(); r != nil {
			fmt.Println("Recovered. Error:\n", r)
		}
		err := os.Remove(name)
		if err != nil {
			log.Fatal("delete archive failed: ", err)

		}
	}("files_companies/" + archive)

	typeFile := archive[len(archive)-6:]
	fmt.Println(typeFile)
	if !(typeFile == "tar.gz") {
		fmt.Println("оно обнаружило, что формат не совпадает")
		return
	}
	//Распаковываем архив
	err := archiver.Unarchive("files_companies/" + archive, "files/") //+ year + "/"
	if err != nil {
		fmt.Println("open archive failed: ", err)
		return
	}

}
