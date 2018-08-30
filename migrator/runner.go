package db

import (
	"bytes"
	"errors"
	"fmt"
	"github.com/amagimedia/paalaka/models"
	"github.com/jinzhu/gorm"
	"os"
	"reflect"
	"runtime"
	"sort"
	"strings"
	"text/template"
	"time"
)

type By func(m1, m2 *migrator) bool

var ByTime = func(m1, m2 *migrator) bool {
	return m1.TS.Before(m2.TS)
}

var ByTimeRev = func(m1, m2 *migrator) bool {
	return m2.TS.Before(m1.TS)
}

type migrator struct {
	Name  string
	DBVal string
	TS    time.Time
}

type migratorSorter struct {
	migrations []migrator
	by         By
}

func (m *migratorSorter) Len() int {
	return len(m.migrations)
}

func (m *migratorSorter) Swap(i, j int) {
	m.migrations[i], m.migrations[j] = m.migrations[j], m.migrations[i]
}

func (m *migratorSorter) Less(i, j int) bool {
	return m.by(&m.migrations[i], &m.migrations[j])
}

func (by By) Sort(migrators []migrator) {
	mg := &migratorSorter{
		migrations: migrators,
		by:         by, // The Sort method's receiver is the function (closure) that defines the sort order.
	}
	sort.Sort(mg)
}

func Migrate(dbConn *gorm.DB, version, mode string) error {
	if dbConn != nil {
		dbConn.AutoMigrate(&models.Subscriber{})
		dbConn.AutoMigrate(&models.Intent{})
		dbConn.AutoMigrate(&models.Track{})
		dbConn.AutoMigrate(&models.Ingest{})
		dbConn.AutoMigrate(&models.Template{})
		dbConn.AutoMigrate(&models.Event{})
		dbConn.AutoMigrate(&models.User{})
		dbConn.AutoMigrate(&models.IngestEgress{})
		dbConn.AutoMigrate(&models.IngestInstance{})
		dbConn.AutoMigrate(&models.EgressTrack{})
		dbConn.AutoMigrate(&models.Migration{})
		runMigrations(dbConn, version, mode)
		return nil
	}
	return errors.New("DB Connection not initialized, call Init before Migrate")
}

func GenerateMigrations(name string) {
	var tpl bytes.Buffer
	ts := time.Now().UTC().Format("20060102150405")
	funcName := struct {
		Name string
	}{strings.Title(name) + "_" + ts}

	tmpl := template.Must(template.ParseFiles("db/template.txt"))
	err := tmpl.Execute(&tpl, funcName)
	f, err := os.OpenFile("db/migrations.go", os.O_APPEND|os.O_WRONLY, 0600)
	if err != nil {
		return
	}
	defer f.Close()
	if _, err = f.WriteString(tpl.String()); err != nil {
		return
	}
	fmt.Println("Generated func in db/migrations.go")
}

func runMigrations(dbConn *gorm.DB, version, mode string) {
	var versions []string
	list := make([]migrator, 0)
	migratorType := reflect.TypeOf(migrator{})
	for i := 0; i < migratorType.NumMethod(); i++ {
		method := migratorType.Method(i)
		name := method.Name
		dbvals := strings.Split(name, "_")
		dbval := dbvals[len(dbvals)-1]
		ts, err := time.Parse("20060102150405", dbval)
		if err != nil {
			return
		}
		list = append(list, migrator{name, dbval, ts})
	}

	if mode == "up" {
		By(ByTime).Sort(list)
	} else {
		By(ByTimeRev).Sort(list)
	}

	caller := reflect.ValueOf(&migrator{}).Elem()
	inputs := make([]reflect.Value, 2)
	inputs[0] = reflect.ValueOf(mode)
	inputs[1] = reflect.ValueOf(dbConn)

	err := dbConn.Model(&models.Migration{}).Pluck("version", &versions).Error
	if err != nil {
		fmt.Println("Unable to access Migration table")
		return
	}

	for _, l := range list {
		if version != "" && l.DBVal != version {
			continue
		}
		if stringInSlice(l.DBVal, versions) && mode == "down" {
			caller.MethodByName(l.Name).Call(inputs)
		}
		if !stringInSlice(l.DBVal, versions) && mode == "up" {
			caller.MethodByName(l.Name).Call(inputs)
		}
	}
}

func update_migrations(dbConn *gorm.DB, mode string) {
	fpcs := make([]uintptr, 1)
	_ = runtime.Callers(2, fpcs)
	fun := runtime.FuncForPC(fpcs[0] - 1)
	ver := strings.Split(fun.Name(), "_")
	switch mode {
	case "up":
		dbConn.Create(&models.Migration{ver[len(ver)-1]})
	case "down":
		dbConn.Delete(&models.Migration{ver[len(ver)-1]})
	}
}

func stringInSlice(s string, sl []string) bool {
	for _, ss := range sl {
		if ss == s {
			return true
		}
	}
	return false
}
