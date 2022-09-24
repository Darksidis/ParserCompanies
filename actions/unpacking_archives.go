package main

import (
	"github.com/mholt/archiver/v3"
	"io/fs"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
)


func main () {

	dirWithArchives := PathForm("storage", "archives")
	dirWithFiles := PathForm("storage", "files")
	// Getting a list of archives
	archives, err := ioutil.ReadDir(dirWithArchives)
	if err != nil {
		log.Fatal (err)
	}
	if !(len(archives) == 0) {
		extractionArchives(archives, dirWithArchives, dirWithFiles )
	}
}




func extractionArchives(archives []fs.FileInfo, dirWithArchives string, dirWithFiles string) {
	for _, archive := range archives {
		unpackingFiles(archive.Name(), dirWithArchives, dirWithFiles)
	}
}

func unpackingFiles(archive string, dirWithArchives string, dirWithFiles string) {
	// Extracting files from archives

	pathArchive := filepath.Join(dirWithArchives, archive)

	//Delete the archive
	defer func(name string) {
		err := os.Remove(name)
		if err != nil {
			log.Fatal("delete archive failed: ", err)

		}
	}(pathArchive)

	typeFile := archive[len(archive)-6:]
	if !(typeFile == "tar.gz") {
		// format does not match
		return
	}
	//Unpacking the archive
	err := archiver.Unarchive(pathArchive, dirWithFiles)
	if err != nil {
		// open archive failed
		return
	}

}
