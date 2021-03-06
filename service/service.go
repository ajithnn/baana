package service

import (
	"encoding/json"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/mysql"
	_ "github.com/jinzhu/gorm/dialects/postgres"
	_ "github.com/jinzhu/gorm/dialects/sqlite"
	"os"
	"reflect"
	"strings"
)

var err error
var ControllerFuncs map[string]HandlerInit

type HandlerInit func(*gorm.DB) interface{}
type ErrorFunc func(*gin.Context)

type Query struct {
	Limit  int `json:"limit" form:"limit"`
	Offset int `json:"offset" form:"offset"`
}

type Service struct {
	DB       *gorm.DB
	SetError ErrorFunc
}

type dbConfig struct {
	User     string `json:"user"`
	Password string `json:"password"`
	Host     string `json:"host"`
	Port     string `json:"port"`
	DB       string `json:"db"`
	Type     string `json:"type"`
	Env      string `json:"env"`
}

type configs struct {
	Db []dbConfig `json:"db"`
}

func New(dbConf []byte) (*Service, error) {
	var conf configs
	err := json.Unmarshal(dbConf, &conf)
	if err != nil {
		return &Service{}, err
	}
	curConf, err := getConfigByEnv(conf)
	if err != nil {
		return &Service{}, err
	}
	return &Service{
		DB: getDB(curConf.User, curConf.Password, curConf.Host, curConf.Port, curConf.DB, curConf.Type),
	}, nil
}

func (r *Service) LoadRoutes(eng gin.IRoutes, routes []byte) {
	// Read the JSON of routes
	var routeMap map[string]string
	err := json.Unmarshal(routes, &routeMap)
	if err != nil {
		panic(err)
	}
	// For each route, Call Route to Handler with correct string
	for rt, runner := range routeMap {
		rsplit := strings.Split(rt, "#")
		method := rsplit[0]
		route := rsplit[1]
		controllerAction := runner
		switch method {
		case "GET":
			eng.GET(route, r.RouteToHandler(controllerAction))
		case "PUT":
			eng.PUT(route, r.RouteToHandler(controllerAction))
		case "POST":
			eng.POST(route, r.RouteToHandler(controllerAction))
		case "DELETE":
			eng.DELETE(route, r.RouteToHandler(controllerAction))
		case "PATCH":
			eng.PATCH(route, r.RouteToHandler(controllerAction))
		case "Any":
			eng.Any(route, r.RouteToHandler(controllerAction))
		default:
			err = fmt.Errorf("Invalid Method type - %s. Please use one of GET|POST|PUT|PATCH|DELETE|Any", method)
			panic(err)
		}
	}
}

func (r *Service) Terminate() {
	r.DB.Close()
}

func (r *Service) RouteToHandler(handlerAction string) gin.HandlerFunc {
	return func(c *gin.Context) {
		tx := r.DB
		c.Set("controller_action", handlerAction)
		callDetails := strings.Split(handlerAction, "#")
		if _, ok := ControllerFuncs[callDetails[0]]; ok {
			funcToCall := reflect.ValueOf(ControllerFuncs[callDetails[0]])
			in := make([]reflect.Value, 1)
			in[0] = reflect.ValueOf(tx)
			result := funcToCall.Call(in)
			if len(result) == 1 {
				r.callAction(callDetails[1], result[0], c, tx)
				return
			}
		}
		r.NotFound(c, tx)
	}
}

func (r *Service) callAction(action string, result reflect.Value, c *gin.Context, tx *gorm.DB) {
	e := result.Elem()
	if e.IsValid() {
		caller := e.MethodByName(action)
		if caller.IsValid() {
			in := make([]reflect.Value, 1)
			in[0] = reflect.ValueOf(c)
			caller.Call(in)
			if c.Errors != nil && len(c.Errors) > 0 {
				r.SetError(c)
				//tx.Rollback()
				return
			}
			//tx.Commit()
			return
		}
	}
	r.NotFound(c, tx)
}

func (r *Service) NotFound(c *gin.Context, tx *gorm.DB) {
	//tx.Rollback()
	controllerAction := "NotKnown"
	if val, ok := c.Get("controller_action"); ok {
		controllerAction = val.(string)
	}
	c.JSON(404, gin.H{"error": "Invalid Controller#Action " + controllerAction + " , Not found."})
	return
}

func getConfigByEnv(cfg configs) (dbConfig, error) {
	for _, c := range cfg.Db {
		if c.Env == os.Getenv("BAANA_ENV") {
			return c, nil
		}
	}
	return dbConfig{}, fmt.Errorf("Invalid Environment BAANA_ENV, please set one of development|test|staging|production. Unable to Set DB credentials")
}

// Panics if DB init fails
func getDB(user, pwd, host, port, dbName, dbType string) *gorm.DB {
	var db *gorm.DB
	var err error
	if user == "" || pwd == "" || port == "" || dbName == "" || dbType == "" {
		panic("Invalid DB credentials,None of the parameters can be empty.")
	}
	switch dbType {
	case "mysql":
		db, err = gorm.Open(dbType, fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?charset=utf8&parseTime=True&loc=Local", user, pwd, host, port, dbName))
	case "postgres":
		db, err = gorm.Open(dbType, fmt.Sprintf("host=%s port=%s user=%s dbname=%s password=%s", host, port, user, dbName, pwd))
	case "sqlite3":
		db, err = gorm.Open(dbType, dbName)
	}
	if err != nil {
		panic(err)
	}
	return db
}
