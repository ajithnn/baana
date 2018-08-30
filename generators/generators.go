package generators

import (
	"../app"
	"fmt"
	"github.com/gobuffalo/packr"
	"os"
	"text/template"
)

type routeMap struct {
	ControllerList []string
	ImportPath     string
}

func Generate(curApp app.App) bool {
	// Create <app_name>.go using the tempalte
	box := packr.NewBox("../templates/")
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
		[]string{"Event", "Template"},
		curApp.ImportPath,
	}
	e = rt.Execute(rf, &data)
	if e != nil {
		fmt.Println("Error unable to create render template " + e.Error())
		return false
	}
	return true
}
