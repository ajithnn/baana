package handlers

import (
	"fmt"
	perr "github.com/amagimedia/paalaka/err"
	"github.com/amagimedia/paalaka/logger"
	"github.com/gin-gonic/gin"
	"github.com/jinzhu/gorm"
	"reflect"
	"strings"
)

var err error

var mapping = map[string]handlerInit{
	"Template":       Template,
	"Event":          Event,
	"Intent":         Intent,
	"Ingest":         Ingest,
	"IngestInstance": IngestInstance,
	"IngestEgress":   IngestEgress,
	"Track":          Track,
	"EgressTrack":    EgressTrack,
	"User":           User,
	"Subscriber":     Subscriber,
}

type handlerInit func(*gorm.DB) Handler

type Handler interface {
	Create(*gin.Context)
	Update(*gin.Context)
	Read(*gin.Context)
	Delete(*gin.Context)
}

type query struct {
	Limit        int  `json:"limit" form:"limit"`
	Offset       int  `json:"offset" form:"offset"`
	Force        bool `json:"force" form:"force"`
	SubscriberID int  `json:"subscriber_id" form:"subscriber_id"`
}

type Router struct {
	DB *gorm.DB
}

func (r *Router) RouteToHandler(handlerAction string) gin.HandlerFunc {
	return func(c *gin.Context) {
		tx := r.DB.Begin()
		callDetails := strings.Split(handlerAction, "#")
		if _, ok := mapping[callDetails[0]]; ok {
			funcToCall := reflect.ValueOf(mapping[callDetails[0]])
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

func (r *Router) callAction(action string, result reflect.Value, c *gin.Context, tx *gorm.DB) {
	e := result.Elem()
	if e.IsValid() {
		caller := e.MethodByName(action)
		if caller.IsValid() {
			in := make([]reflect.Value, 1)
			in[0] = reflect.ValueOf(c)
			caller.Call(in)
			if c.Errors != nil && len(c.Errors) > 0 {
				r.setErrorResponse(c)
				tx.Rollback()
				return
			}
			tx.Commit()
			return
		}
	}
	r.NotFound(c, tx)
}

func (r *Router) NotFound(c *gin.Context, tx *gorm.DB) {
	tx.Rollback()
	c.JSON(404, gin.H{"error": "Not found"})
	return
}

func (r *Router) setErrorResponse(c *gin.Context) {
	var status int
	var errString string
	for _, err := range c.Errors {
		status = perr.ResponseStatus(err)
		errString = err.Error()
	}
	logger.Logger().Infof(fmt.Sprintf("Handler#setErrorResponse: Setting status:%d for %s", status, c.Request.URL))
	c.JSON(status, gin.H{"error": errString})
	return
}
