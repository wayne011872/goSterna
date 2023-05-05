package mid

import (
	"net/http"
	"runtime"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	golanggeneral "github.com/wayne011872/golangGeneral"
	"github.com/wayne011872/golangGeneral/api"
	apiErr "github.com/wayne011872/golangGeneral/api/err"
	"github.com/wayne011872/golangGeneral/db"
	"github.com/wayne011872/golangGeneral/log"
)

type DBMidDI interface {
	log.LoggerDI
	db.MongoDI
}

type DBMiddle string

func NewGinDBMid(service string) GinMiddle {
	return &dbMiddle{
		service: service,
	}
}

type dbMiddle struct {
	service string
}

func (lm *dbMiddle) GetName() string {
	return "db"
}

func(lm *dbMiddle) outputErr(c *gin.Context, err error) {
	api.GinOutputErr(c,lm.service,err)
}

func (am *dbMiddle)Handler() gin.HandlerFunc {
	return func(c *gin.Context) {
		servDi, ok := c.Get(string(golanggeneral.CtxServDiKey))
		if !ok || servDi == nil {
			am.outputErr(c,apiErr.New(http.StatusInternalServerError, "can not get di"))
			c.Abort()
			return
		}

		if dbdi, ok := servDi.(DBMidDI); ok {
			uuid := uuid.New().String()
			l := dbdi.NewLogger(uuid)

			dbclt, err := dbdi.NewMongoDBClient(c.Request.Context(), "")
			if err != nil {
				am.outputErr(c, apiErr.New(http.StatusInternalServerError, err.Error()))
				c.Abort()
				return
			}
			defer dbclt.Close()

			c.Set(string(db.CtxMongoKey), dbclt)
			c.Set(string(log.CtxLogKey), l)

			c.Next()
			runtime.GC()
		}else {
			am.outputErr(c, apiErr.New(http.StatusInternalServerError, "invalid di"))
			c.Abort()
			return
		}
	}
}