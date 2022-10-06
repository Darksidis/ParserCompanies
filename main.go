package main

import (
	"encoding/json"
	"fmt"
	"github.com/joho/godotenv"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"sync"
)

func main() {
	fmt.Println("start app")
	jsonData := goDotEnvVariable("actionsList", "ACTIONS")
	var actions map[string]bool
	err := json.Unmarshal([]byte(jsonData), &actions)
	if err != nil {
		panic(err)
	}
	countActions := 0
	var listActions []string

	for action, statusAction := range actions {

		if statusAction {
			countActions += 1
			listActions = append(listActions, action)

		}
	}

	if countActions != 0 {
		var wg sync.WaitGroup
		wg.Add(countActions)
		for i := 0; i < countActions; i++ {
			go execCom(listActions[i], &wg)
		}

		wg.Wait()
	}

}

func check(err error) {
	if err != nil {
		panic(err)
	}
}

func execCom(action string, wg *sync.WaitGroup) {
	//defer wg.Done()

	files, err := ioutil.ReadDir("actions")
	check(err)
	file := files[0].Name()
	path := filepath.Join("actions", action)
	command := []string{"./" + path}
	if string(file[len(file)-3:]) == ".go" {
		action += ".go"
		pathDb := filepath.Join("actions", "database.go")
		path = filepath.Join("actions", action)
		command = []string{"go", "run", path, pathDb}
	}

	cmd := exec.Command(command[0], command[1:]...)
	out, err := cmd.Output()
	fmt.Println(string(out))
	check(err)

}

func goDotEnvVariable(nameFile string, key string) string {

	// load actionsList.env file
	err := godotenv.Load(nameFile + ".env")

	if err != nil {
		log.Fatalf("Error loading actionsList.env file")
	}

	return os.Getenv(key)
}

func PathForm(parentFolder string, path string) string {

	path = filepath.Join(parentFolder, path)

	return path

}
