package main

import (
	"./app"
	"./generators"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"strings"
)

const (
	METAFOLDER = ".baana"
)

func main() {
	if len(os.Args) < 3 {
		fmt.Println("Incorrect command, usage: baana <command> <options>")
		fmt.Println("          baana init  <app_name>")
		fmt.Println("          baana reset <app_name>")
		os.Exit(1)
	}

	switch os.Args[1] {
	case "init":
		// Get PWD
		pwd, err := os.Getwd()
		if err != nil {
			fmt.Println("Unable to extract current working directory.")
			os.Exit(1)
		}

		// Create .baana folder - Exit if already exists
		err = os.Mkdir(fmt.Sprintf("%s/%s", pwd, METAFOLDER), 0755)
		if err != nil {
			if os.IsExist(err) {
				fmt.Println("Cannot re-init an existing project. Please run, baana reset (this will delete all metadata , may cause errors in generation.)")
				os.Exit(1)
			}
			fmt.Println("Unable to init folder.")
			os.RemoveAll(fmt.Sprintf("%s/%s", pwd, METAFOLDER))
			os.Exit(1)
		}
		// Inside folder create app.json with format as
		err = ioutil.WriteFile(fmt.Sprintf("%s/%s/%s", pwd, METAFOLDER, "app.json"), []byte("{}"), 0644)
		if err != nil {
			fmt.Println("Unable to init folder.")
			os.Remove(fmt.Sprintf("%s/%s/%s", pwd, METAFOLDER, "app.json"))
			os.Exit(1)
		}
		importPath := strings.Replace(pwd, fmt.Sprintf("%s/%s/", os.Getenv("GOPATH"), "src"), "", 1)
		/*
			{
				"app_name": os.Args[1],
				"app_path": PWD,
				"import_path":PWD - GOPATH
			}
			TODO: Add fields as required
		*/
		appData := app.App{
			Name:       os.Args[2],
			Path:       pwd,
			ImportPath: importPath,
		}
		appString, _ := json.MarshalIndent(&appData, "", "  ")
		err = ioutil.WriteFile(fmt.Sprintf("%s/%s/%s", pwd, METAFOLDER, "app.json"), appString, 0644)
		if err != nil {
			fmt.Println("Unable to write app data, write error.")
			os.Remove(fmt.Sprintf("%s/%s/%s", pwd, METAFOLDER, "app.json"))
			os.Exit(1)
		}
		// Create "AppName.go" file with appropriate template
		if !generators.Generate(appData) {
			os.Exit(1)
		}
		// Create Models,Controllers, migrations folders
		// create server/server.go
		// create route/route.go
	case "reset":
		pwd, err := os.Getwd()
		if err != nil {
			fmt.Println("Unable to extract current working directory.")
			os.Exit(1)
		}
		os.RemoveAll(fmt.Sprintf("%s/%s", pwd, METAFOLDER))
	default:
		fmt.Println("Unknown Command")
		os.Exit(1)
	}

	/*
		baana generate model <name>
			- Creates Model, Controllers and CRUD routes
		baana generate migration <name>
			- Create Migration file with timestamp_Name as the function name.
		baana generate route <Type> <Controller#Action>
			- Create route, add a function to controller for action.
		Cannot create just controller without model as of now.

		baana run
			- Compile app based on its path.
			- Run Migration
			- Run Server
	*/
}
