package main

import (
	"encoding/json"
	"fmt"
	"gitlab.com/ajithnn/baana/app"
	"gitlab.com/ajithnn/baana/generators"
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
		fmt.Println("          baana generate model|controller|migration <name>")
		os.Exit(1)
	}

	var status bool
	// Get PWD
	pwd, err := os.Getwd()
	if err != nil {
		fmt.Println("Unable to extract current working directory.")
		os.Exit(1)
	}

	switch os.Args[1] {
	case "init":

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
			Name:           os.Args[2],
			Path:           pwd,
			ImportPath:     importPath,
			ControllerList: []string{},
		}
		appString, _ := json.MarshalIndent(&appData, "", "  ")
		err = ioutil.WriteFile(fmt.Sprintf("%s/%s/%s", pwd, METAFOLDER, "app.json"), appString, 0644)
		if err != nil {
			fmt.Println("Unable to write app data, write error.")
			os.Remove(fmt.Sprintf("%s/%s/%s", pwd, METAFOLDER, "app.json"))
			os.Exit(1)
		}

		// Create Init set of folders and files
		if !generators.Init(appData) {
			os.Exit(1)
		}
	case "reset":
		if err != nil {
			fmt.Println("Unable to extract current working directory.")
			os.Exit(1)
		}
		os.RemoveAll(fmt.Sprintf("%s/%s", pwd, METAFOLDER))
	case "generate":
		/*
			baana generate model <name>
				- Creates Model, Controllers and CRUD routes
			baana generate controller <name>
				- Create Migration file with timestamp_Name as the function name.
			baana generate migration <name>
				- Create Migration file with timestamp_Name as the function name.
			To Create a independent route edit config/routes.json directly.

		*/

		if len(os.Args) < 4 {
			fmt.Println("Invalid command use. Usage: baana generate (model|migration|controller|route) (name)")
			os.Exit(1)
		}

		var appData app.App
		appBytes, err := ioutil.ReadFile(fmt.Sprintf("%s/%s/%s", pwd, METAFOLDER, "app.json"))
		err = json.Unmarshal(appBytes, &appData)
		if err != nil {
			fmt.Println("Error in reading app data " + err.Error())
			os.Exit(1)
		}

		name := strings.Title(os.Args[3])

		switch os.Args[2] {
		case "migration":
			status = generators.GenerateMigrations(name)
		case "model":
			status = generators.GenerateModels(appData, name)
		case "controller":
			status = generators.GenerateControllers(appData, name)
		default:
			fmt.Println("Unknown Sub Command to generate: use one of migration|model|controller|route")
			os.Exit(1)
		}
	default:
		fmt.Println("Unknown Command. Use one of init|reset|generate")
		os.Exit(1)
	}
	if !status {
		os.Exit(1)
	}

}
