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
}

var box = packr.NewBox("../templates/")

func Generate(curApp app.App) bool {
	// Create <app_name>.go using the tempalte
	t := template.Must(template.New("app").Parse(box.String("app.tmpl")))
	f, err := os.OpenFile(fmt.Sprintf("%s/%s.go", curApp.Path, curApp.Name), os.O_RDWR|os.O_CREATE, 0755)
	if err != nil {
		fmt.Println("Cannot open file , error: " + err.Error())
		return false
	}
	err = t.Execute(f, curApp)
	if err != nil {
		fmt.Println("Error unable to create render template " + err.Error())
		return false
	}

	errs := make([]error, 0)
	// Create Models,Controllers, migrations folders
	errs = append(errs, os.Mkdir(fmt.Sprintf("%s/%s", curApp.Path, "models"), 0755))
	errs = append(errs, os.Mkdir(fmt.Sprintf("%s/%s", curApp.Path, "controllers"), 0755))
	errs = append(errs, os.Mkdir(fmt.Sprintf("%s/%s", curApp.Path, "migrations"), 0755))
	errs = append(errs, os.Mkdir(fmt.Sprintf("%s/%s", curApp.Path, "server"), 0755))
	errs = append(errs, os.Mkdir(fmt.Sprintf("%s/%s", curApp.Path, "route"), 0755))
	errs = append(errs, os.Mkdir(fmt.Sprintf("%s/%s", curApp.Path, "config"), 0755))
	if len(errs) != 0 {
		for _, e := range errs {
			if e != nil {
				fmt.Println("Folder create errors " + e.Error())
				return false
			}
		}
	}
	// create server/server.go
	st := template.Must(template.New("server").Parse(box.String("server.tmpl")))
	sf, er := os.OpenFile(fmt.Sprintf("%s/server/server.go", curApp.Path), os.O_RDWR|os.O_CREATE, 0755)
	if er != nil {
		fmt.Println("Cannot open file , error: " + er.Error())
		return false
	}
	er = st.Execute(sf, curApp)
	if er != nil {
		fmt.Println("Error unable to create render template " + er.Error())
		return false
	}
	// create route/route.go
	rt := template.Must(template.New("route").Parse(box.String("routemap.tmpl")))
	rf, e := os.OpenFile(fmt.Sprintf("%s/route/route.go", curApp.Path), os.O_RDWR|os.O_CREATE, 0755)
	if e != nil {
		fmt.Println("Cannot open file , error: " + e.Error())
		return false
	}
	data := routeMap{
		[]string{},
		curApp.ImportPath,
	}
	e = rt.Execute(rf, &data)
	if e != nil {
		fmt.Println("Error unable to create render template " + e.Error())
		return false
	}
	// create migrations/migration.go
	t = template.Must(template.New("base_migration").Parse(box.String("base_migration.tmpl")))
	f, err = os.OpenFile(fmt.Sprintf("%s/migrations/migration.go", curApp.Path), os.O_RDWR|os.O_CREATE, 0755)
	if err != nil {
		fmt.Println("Cannot open file , error: " + err.Error())
		return false
	}
	err = t.Execute(f, &curApp)
	if err != nil {
		fmt.Println("Error unable to create render template " + err.Error())
		return false
	}

	// create models/migration.go
	t = template.Must(template.New("model_migration").Parse(box.String("migrate_model.tmpl")))
	f, err = os.OpenFile(fmt.Sprintf("%s/models/migration.go", curApp.Path), os.O_RDWR|os.O_CREATE, 0755)
	if err != nil {
		fmt.Println("Cannot open file , error: " + err.Error())
		return false
	}
	err = t.Execute(f, &curApp)
	if err != nil {
		fmt.Println("Error unable to create render template " + err.Error())
		return false
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
		fmt.Sprintf("/%s/{id} PUT", inflection.Plural(strings.ToLower(name))),
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

	data := routeMap{
		controllerNames,
		curApp.ImportPath,
	}

	rt := template.Must(template.New("route").Parse(box.String("routemap.tmpl")))
	rf, e := os.OpenFile(fmt.Sprintf("%s/route/route.go", curApp.Path), os.O_RDWR|os.O_CREATE, 0755)
	if e != nil {
		fmt.Println("Cannot open file , error: " + e.Error())
		return false
	}

	e = rt.Execute(rf, &data)
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
