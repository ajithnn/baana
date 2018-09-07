package generators

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/gobuffalo/packr"
	"github.com/jinzhu/inflection"
	"gitlab.com/ajithnn/baana/app"
	"io/ioutil"
	"os"
	"strings"
	"text/template"
	"time"
)

type routeMap struct {
	ControllerList []string
	ImportPath     string
}

type Model struct {
	Name      string
	ShortDesc string
	LongDesc  string
}

type Controller struct {
	Name       string
	PluralName string
	Endpoint   string
	ImportPath string
}

var box = packr.NewBox("../templates/")

func Init(curApp app.App) bool {

	var err error

	// Create folders
	folders := []string{"models", "controllers", "migrations", "server", "route", "config"}
	for _, folder := range folders {
		err = os.Mkdir(fmt.Sprintf("%s/%s", curApp.Path, folder), 0755)
		if err != nil {
			fmt.Println("Folder create error " + err.Error())
			return false
		}
	}

	// Create Config Files
	err = createConfigFiles()
	if err != nil {
		fmt.Println("Error: " + err.Error())
		return false
	}

	// Create Go Code Files
	codeFiles := [][]string{
		{"server.tmpl", "server/server.go"}, {"base_migration.tmpl", "migrations/migration.go"},
		{"migrate_model.tmpl", "models/migration.go"}, {"routemap.tmpl", "route/route.go"},
		{"app.tmpl", curApp.Name + ".go"},
	}

	for _, params := range codeFiles {
		err = createCodeFile(params[0], params[1], curApp)
		if err != nil {
			fmt.Println("Failed to create " + params[1])
			return false
		}
	}

	return true
}

func GenerateMigrations(name string) bool {
	var tpl bytes.Buffer
	ts := time.Now().UTC().Format("20060102150405")
	funcName := struct {
		Name string
	}{name + "_" + ts}

	tmpl := template.Must(template.New("migration").Parse(box.String("migration.tmpl")))
	err := tmpl.Execute(&tpl, funcName)
	f, err := os.OpenFile("migrations/migration.go", os.O_APPEND|os.O_WRONLY, 0600)
	if err != nil {
		fmt.Println("Error: " + err.Error())
		return false
	}
	defer f.Close()
	if _, err = f.WriteString(tpl.String()); err != nil {
		fmt.Println("Error: " + err.Error())
		return false
	}
	fmt.Println("Generated func in migrations/migration.go")
	return true
}

func GenerateModels(curApp app.App, name string) bool {
	var tpl bytes.Buffer
	// Create Model file by passing the name
	m := Model{
		name,
		fmt.Sprintf("%s denotes ", name),
		fmt.Sprintf("%s is used for ", name),
	}
	tmpl := template.Must(template.New("model").Parse(box.String("model.tmpl")))
	err := tmpl.Execute(&tpl, m)
	f, err := os.OpenFile("models/"+strings.ToLower(name)+".go", os.O_CREATE|os.O_WRONLY, 0600)
	if err != nil {
		fmt.Println("Error: " + err.Error())
		return false
	}
	defer f.Close()
	if _, err = f.WriteString(tpl.String()); err != nil {
		fmt.Println("Error: " + err.Error())
		return false
	}
	// Check if Controller file exists, If yes, skip.
	// IF no, create controller file with name and app data.
	val := GenerateControllers(curApp, name)
	if !val {
		return val
	}
	// Load routes.json file and add
	// Create,Read,Update, Delete to routes.
	return GenerateRoutes(curApp, name)
}

func GenerateRoutes(curApp app.App, name string) bool {
	fmt.Println("Generating Routes for " + name)
	var curRoutes = make(map[string]string)
	action := map[string]string{
		"GET":    "Read",
		"PUT":    "Update",
		"POST":   "Create",
		"DELETE": "Delete",
	}
	// Read routes.json to a map[string]string
	c, err := ioutil.ReadFile("config/routes.json")
	if err != nil {
		fmt.Println("Error in reading routes.json, error: " + err.Error())
		return false
	}
	err = json.Unmarshal(c, &curRoutes)
	if err != nil {
		fmt.Println("Error in reading routes.json, error: " + err.Error())
		return false
	}

	paths := []string{
		fmt.Sprintf("/%s/ POST", inflection.Plural(strings.ToLower(name))),
		fmt.Sprintf("/%s/:id PUT", inflection.Plural(strings.ToLower(name))),
		fmt.Sprintf("/%s/*id GET", inflection.Plural(strings.ToLower(name))),
		fmt.Sprintf("/%s/*id DELETE", inflection.Plural(strings.ToLower(name))),
	}

	// Add routes for create, read,update and delete
	for _, p := range paths {
		pm := strings.Split(p, " ")
		curRoutes[fmt.Sprintf("%s#%s", pm[1], pm[0])] = fmt.Sprintf("%s#%s", name, action[pm[1]])
	}

	rjson, err := json.MarshalIndent(curRoutes, "", "  ")
	err = ioutil.WriteFile("config/routes.json", rjson, 0755)
	// Update routeMap template by reading unique controllers from routes.json
	controllers := make(map[string]bool)
	controllerNames := make([]string, 0)
	for _, r := range curRoutes {
		ca := strings.Split(r, "#")
		controllers[ca[0]] = true
	}
	for names, _ := range controllers {
		controllerNames = append(controllerNames, names)
	}

	curApp.ControllerList = controllerNames

	rt := template.Must(template.New("route").Parse(box.String("routemap.tmpl")))
	rf, e := os.OpenFile(fmt.Sprintf("%s/route/route.go", curApp.Path), os.O_RDWR|os.O_CREATE, 0755)
	if e != nil {
		fmt.Println("Cannot open file , error: " + e.Error())
		return false
	}

	e = rt.Execute(rf, &curApp)
	if e != nil {
		fmt.Println("Error unable to create render template " + e.Error())
		return false
	}

	return true

}

func GenerateControllers(curApp app.App, name string) bool {
	fmt.Println("Generating Controller for " + name)
	controller := Controller{
		name,
		inflection.Plural(name),
		inflection.Plural(strings.ToLower(name)),
		curApp.ImportPath,
	}
	// Check if Controller file exists, If yes, skip.
	f, err := os.Create("controllers/" + strings.ToLower(name) + ".go")
	// IF no, create controller file with name and app data.
	if err == nil {
		tmpl := template.Must(template.New("controller").Parse(box.String("controller.tmpl")))
		err = tmpl.Execute(f, controller)
		if err != nil {
			fmt.Println("Error: " + err.Error())
			return false
		}
		return true
	}
	return false
}

func createConfigFiles() error {
	dbjson := map[string]interface{}{
		"db": []map[string]string{
			{
				"user":     "",
				"password": "",
				"host":     "",
				"port":     "",
				"db":       "",
				"type":     "",
				"env":      "development",
			},
			{
				"user":     "",
				"password": "",
				"host":     "",
				"port":     "",
				"db":       "",
				"type":     "",
				"env":      "test",
			},
			{
				"user":     "",
				"password": "",
				"host":     "",
				"port":     "",
				"db":       "",
				"type":     "",
				"env":      "staging",
			},
			{
				"user":     "",
				"password": "",
				"host":     "",
				"port":     "",
				"db":       "",
				"type":     "",
				"env":      "production",
			},
		},
	}

	dbjsonBytes, err := json.MarshalIndent(dbjson, "", "  ")
	err = ioutil.WriteFile("config/db.json", dbjsonBytes, 0755)
	err = ioutil.WriteFile("config/routes.json", []byte("{}"), 0755)
	return err
}

func createCodeFile(tmplName, codePath string, curApp app.App) error {
	tmpl := template.Must(template.New(tmplName).Parse(box.String(tmplName)))
	file, err := os.OpenFile(fmt.Sprintf("%s/%s", curApp.Path, codePath), os.O_RDWR|os.O_CREATE, 0755)
	if err != nil {
		return err
	}
	err = tmpl.Execute(file, curApp)
	if err != nil {
		return err
	}
	return nil
}
