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
	dirWithArchives := pathForm("archives")
	dirWithFiles := pathForm("files")
	// Getting a list of archives
	archives, err := ioutil.ReadDir(dirWithArchives)
	if err != nil {
		log.Fatal (err)
	}
	if !(len(archives) == 0) {
		extractionArchives(archives, dirWithArchives, dirWithFiles )
	}
}

func pathForm (path string) string {
	_, err := os.Stat(path)
	if os.IsNotExist(err) {
		// If the folder does not exist, then the program is launched through the binary
		wd, err := os.Getwd()
		if err != nil {
			panic(err)
		}
		parent := filepath.Dir(wd)

		path = filepath.Join(parent, path)
	}

	return path

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
