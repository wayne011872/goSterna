package mid

import (
	"net/http"
	"runtime"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	goSterna "github.com/wayne011872/goSterna"
	apiErr "github.com/wayne011872/goSterna/api/err"
	"github.com/wayne011872/goSterna/db"
	"github.com/wayne011872/goSterna/log"
)

type SqlDBMidDI interface {
	log.LoggerDI
	db.SqlDI
}

type SqlDBMiddle string

func NewGinSqlDBMid(service string) GinMiddle {
	return &sqlDbMiddle{
		service: service,
	}
}

type sqlDbMiddle struct {
	service string
}

func (lm *sqlDbMiddle) GetName() string {
	return "db"
}

func(lm *sqlDbMiddle) outputErr(c *gin.Context, err error) {
	apiErr.GinOutputErr(c,lm.service,err)
}

func (am *sqlDbMiddle)Handler() gin.HandlerFunc {
	return func(c *gin.Context) {
		servDi, ok := c.Get(string(goSterna.CtxServDiKey))
		if !ok || servDi == nil {
			am.outputErr(c,apiErr.NewApiError(http.StatusInternalServerError, "can not get di"))
			c.Abort()
			return
		}

		if dbdi, ok := servDi.(SqlDBMidDI); ok {
			uuid := uuid.New().String()
			l := dbdi.NewLogger(uuid)

			dbclt, err := dbdi.NewSqlDB(c.Request.Context())
			if err != nil {
				am.outputErr(c, apiErr.NewApiError(http.StatusInternalServerError, err.Error()))
				c.Abort()
				return
			}
			defer dbclt.Close()

			c.Set(string(db.CtxSqlKey), dbclt)
			c.Set(string(log.CtxLogKey), l)

			c.Next()
			runtime.GC()
		}else {
			am.outputErr(c, apiErr.NewApiError(http.StatusInternalServerError, "invalid di"))
			c.Abort()
			return
		}
	}
}