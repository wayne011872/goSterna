package mid

import (
	"fmt"
	"net/http"
	"runtime"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	goSterna "github.com/wayne011872/goSterna"
	apiErr "github.com/wayne011872/goSterna/api/err"
	"github.com/wayne011872/goSterna/db"
	"github.com/wayne011872/goSterna/log"
)

type PgxMidDI interface {
	log.LoggerDI
	db.PgxDI
}

type PgxMiddle string

func NewGinPgxMid(service string) GinMiddle {
	return &pgxMiddle{
		service: service,
	}
}

type pgxMiddle struct {
	service string
}

func (pm *pgxMiddle) GetName() string {
	return "db"
}

func(pm *pgxMiddle) outputErr(c *gin.Context, err error) {
	apiErr.GinOutputErr(c,pm.service,err)
}

func (pm *pgxMiddle)Handler() gin.HandlerFunc {
	return func(c *gin.Context) {
		servDi, ok := c.Get(string(goSterna.CtxServDiKey))
		if !ok || servDi == nil {
			pm.outputErr(c,apiErr.NewApiError(http.StatusInternalServerError, "can not get di"))
			c.Abort()
			return
		}
		if dbdi, ok := servDi.(PgxMidDI); ok {
			uuid := uuid.New().String()
			l := dbdi.NewLogger(uuid)

			dbclt, err := dbdi.NewPgxClient(c.Request.Context())
			if err != nil {
				pm.outputErr(c, apiErr.NewApiError(http.StatusInternalServerError, err.Error()))
				c.Abort()
				return
			}
			defer dbclt.Close()
			defer fmt.Println("close pgx middle")
			c.Set(string(db.CtxPgxKey), dbclt)
			c.Set(string(log.CtxLogKey), l)

			c.Next()
			runtime.GC()
		}else {
			pm.outputErr(c, apiErr.NewApiError(http.StatusInternalServerError, "invalid di"))
			c.Abort()
			return
		}
	}
}