package generators

import (
	"bytes"
	"fmt"
	"github.com/gobuffalo/packr"
	"gitlab.com/ajithnn/baana/app"
	"os"
	"strings"
	"text/template"
	"time"
)

type routeMap struct {
	ControllerList []string
	ImportPath     string
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
	// create controllers/controllers.go
	t = template.Must(template.New("base_controller").Parse(box.String("main_controller.tmpl")))
	f, err = os.OpenFile(fmt.Sprintf("%s/controllers/controllers.go", curApp.Path), os.O_RDWR|os.O_CREATE, 0755)
	if err != nil {
		fmt.Println("Cannot open file , error: " + err.Error())
		return false
	}
	err = t.Execute(f, &curApp)
	if err != nil {
		fmt.Println("Error unable to create render template " + err.Error())
		return false
	}
	// create models/models.go
	t = template.Must(template.New("base_model").Parse(box.String("main_model.tmpl")))
	f, err = os.OpenFile(fmt.Sprintf("%s/models/models.go", curApp.Path), os.O_RDWR|os.O_CREATE, 0755)
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

func GenerateMigrations(name string) {
	var tpl bytes.Buffer
	ts := time.Now().UTC().Format("20060102150405")
	funcName := struct {
		Name string
	}{strings.Title(name) + "_" + ts}

	tmpl := template.Must(template.New("migration").Parse(box.String("migration.tmpl")))
	err := tmpl.Execute(&tpl, funcName)
	f, err := os.OpenFile("migrations/migration.go", os.O_APPEND|os.O_WRONLY, 0600)
	if err != nil {
		return
	}
	defer f.Close()
	if _, err = f.WriteString(tpl.String()); err != nil {
		return
	}
	fmt.Println("Generated func in migrations/migration.go")
}
